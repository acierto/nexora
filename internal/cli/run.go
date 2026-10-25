// Package cli implements the standalone Nexora workflow.
package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"text/tabwriter"
	"time"
	"unicode/utf8"

	"github.com/acierto/nexora/internal/app"
	"github.com/acierto/nexora/internal/knowledge"
	"github.com/acierto/nexora/internal/mcp"
	"github.com/acierto/nexora/internal/nativesync"
	"github.com/acierto/nexora/internal/resolver"
	"github.com/acierto/nexora/internal/retrieval"
)

// Version is reported by the MCP server.
var Version = "dev"

const usage = `Nexora — one knowledge layer for every AI agent

Usage: nexora <command> [flags]

Commands:
  init       Create nexora.yaml and knowledge/rules/project.md safely
  validate   Validate all configured knowledge
  list       List all knowledge, including inactive items
  resolve    Select knowledge by project, agent, and files (+ docs with --task)
  index      Index docs and Confluence sources for retrieval
  search     Search indexed documentation
  mcp        Serve knowledge to agents over MCP (stdio)
  sync       Write knowledge into AGENTS.md/CLAUDE.md and agent skill folders
  help       Show help (help <command> for flags)

Commands accept --config <path>; read commands also accept --json.
Run nexora <command> --help for command flags.
`

var commands = []string{"init", "validate", "list", "resolve", "index", "search", "mcp", "sync"}

type fileFlags []string

func (f *fileFlags) String() string         { return strings.Join(*f, ",") }
func (f *fileFlags) Set(value string) error { *f = append(*f, value); return nil }

// Streams lets tests supply stdin for the MCP server.
type Streams struct {
	Stdin  io.Reader
	Getenv func(string) string
	HTTP   *http.Client
}

func Run(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	return RunWith(ctx, args, cwd, Streams{Stdin: os.Stdin, Getenv: os.Getenv, HTTP: &http.Client{Timeout: 60 * time.Second}}, stdout, stderr)
}

func run(ctx context.Context, args []string, cwd string, stdout, stderr io.Writer) int {
	return RunWith(ctx, args, cwd, Streams{Stdin: strings.NewReader(""), Getenv: func(string) string { return "" }}, stdout, stderr)
}

// RunWith executes one command with explicit streams and environment.
func RunWith(ctx context.Context, args []string, cwd string, streams Streams, stdout, stderr io.Writer) int {
	fail := func(err error, code int) int { fmt.Fprintln(stderr, err); return code }
	write := func(data []byte) int {
		if _, err := io.Copy(stdout, bytes.NewReader(data)); err != nil {
			return fail(err, 1)
		}
		return 0
	}
	if len(args) == 0 || args[0] == "--help" || args[0] == "-h" || args[0] == "help" && len(args) == 1 {
		return write([]byte(usage))
	}
	if args[0] == "help" {
		args = append(append([]string{}, args[1:]...), "--help")
	}
	command := args[0]
	if !slices.Contains(commands, command) {
		return fail(fmt.Errorf("unknown command %q\n%s", command, usage), 2)
	}
	var parseOutput bytes.Buffer
	flags := flag.NewFlagSet(command, flag.ContinueOnError)
	flags.SetOutput(&parseOutput)
	flags.Usage = func() { fmt.Fprintf(&parseOutput, "Usage: nexora %s [flags]\n", command); flags.PrintDefaults() }
	var configPath, project, agent, task, query, source, dir string
	var asJSON, check bool
	var limit int
	var files fileFlags
	if command == "init" {
		flags.StringVar(&project, "project", "", "project ID (default: current directory name)")
	} else {
		flags.StringVar(&configPath, "config", "./nexora.yaml", "configuration file (no parent discovery)")
		if command != "mcp" {
			flags.BoolVar(&asJSON, "json", false, "emit one JSON object")
		}
	}
	switch command {
	case "resolve":
		flags.StringVar(&project, "project", "", "project ID (default: configuration project)")
		flags.StringVar(&agent, "agent", "", "agent ID for applicability matching only")
		flags.Var(&files, "file", "repository-relative file (repeatable)")
		flags.StringVar(&task, "task", "", "describe the task to also retrieve documentation passages")
		flags.IntVar(&limit, "limit", 5, "maximum passages with --task (1-50)")
	case "search":
		flags.StringVar(&query, "query", "", "search text (required)")
		flags.StringVar(&project, "project", "", "project ID (default: configuration project)")
		flags.IntVar(&limit, "limit", 5, "maximum passages (1-50)")
	case "index":
		flags.StringVar(&source, "source", "", "index only this docs or confluence source")
	case "sync":
		flags.StringVar(&agent, "agent", "", "codex or claude (required)")
		flags.StringVar(&project, "project", "", "project ID (default: configuration project)")
		flags.StringVar(&dir, "dir", "", "repository root to write into (default: the config directory)")
		flags.BoolVar(&check, "check", false, "write nothing; exit 1 if files are out of date")
	}
	if err := flags.Parse(args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return write(parseOutput.Bytes())
		}
		return fail(errors.New(strings.TrimSpace(parseOutput.String())), 2)
	}
	if flags.NArg() != 0 {
		return fail(fmt.Errorf("%s accepts flags only, not positional arguments; run nexora %s --help", command, command), 2)
	}
	if limit < 1 || limit > 50 {
		if wasSet(flags, "limit") {
			return fail(fmt.Errorf("--limit must be between 1 and 50"), 2)
		}
	}
	if command == "search" && strings.TrimSpace(query) == "" {
		return fail(fmt.Errorf("search requires --query <text>; run nexora search --help"), 2)
	}
	if command == "sync" && agent != "codex" && agent != "claude" {
		return fail(fmt.Errorf("sync requires --agent codex or --agent claude"), 2)
	}
	if ctx.Err() != nil {
		return fail(ctx.Err(), 1)
	}
	var output bytes.Buffer
	if command == "init" {
		if !wasSet(flags, "project") {
			project = filepath.Base(cwd)
		}
		if err := initialize(cwd, project); err != nil {
			return fail(err, 1)
		}
		fmt.Fprintln(&output, "Created nexora.yaml and knowledge/rules/project.md.\nRun nexora validate, then nexora resolve.")
		return write(output.Bytes())
	}
	if !filepath.IsAbs(configPath) {
		configPath = filepath.Join(cwd, configPath)
	}
	a := app.App{ConfigPath: configPath, Getenv: streams.Getenv, HTTP: streams.HTTP}
	var value any
	switch command {
	case "validate", "list":
		loaded, err := a.Load(ctx)
		if err != nil {
			return fail(err, 1)
		}
		items := loaded.Items
		if command == "validate" {
			value = struct {
				Valid     bool `json:"valid"`
				ItemCount int  `json:"item_count"`
			}{true, len(items)}
			if !asJSON {
				suffix := "s"
				if len(items) == 1 {
					suffix = ""
				}
				fmt.Fprintf(&output, "Valid: %d knowledge item%s\n", len(items), suffix)
				if n := len(loaded.Config.DocumentSources()); n > 0 {
					fmt.Fprintf(&output, "Documentation sources: %d (run nexora index to refresh passages)\n", n)
				}
			}
		} else {
			value = struct {
				Items []knowledge.Item `json:"items"`
			}{items}
			if !asJSON {
				w := tabwriter.NewWriter(&output, 0, 4, 2, ' ', 0)
				fmt.Fprintln(w, "ID\tKIND\tNAME\tSTATUS\tVERSION\tSOURCE")
				for _, item := range items {
					fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s:%s\n", item.ID, item.Kind, item.Name, item.Status, item.Version, item.Source.ID, item.Source.Path)
				}
				if err := w.Flush(); err != nil {
					return fail(err, 1)
				}
			}
		}
	case "resolve":
		result, err := a.Context(ctx, app.ContextRequest{Project: project, ProjectExplicit: wasSet(flags, "project"), Agent: agent, Files: files, Task: task, Limit: limit})
		if err != nil {
			return fail(err, 1)
		}
		value = result
		if !asJSON {
			renderResult(&output, result.Result)
			if strings.TrimSpace(task) != "" {
				renderHits(&output, "Passages", result.Passages)
			}
		}
	case "search":
		hits, err := a.Search(ctx, app.SearchRequest{Query: query, Project: project, ProjectExplicit: wasSet(flags, "project"), Limit: limit})
		if err != nil {
			return fail(err, 1)
		}
		value = struct {
			Hits []retrieval.Hit `json:"hits"`
		}{hits}
		if !asJSON {
			renderHits(&output, "Results", hits)
		}
	case "index":
		report, err := a.Index(ctx, source)
		if err != nil {
			return fail(err, 1)
		}
		value = report
		if !asJSON {
			model := report.Model
			if model == "" {
				model = "none (keyword search)"
			}
			fmt.Fprintf(&output, "Indexed into the %s store; embeddings: %s\n", report.Store, model)
			for _, s := range report.Sources {
				fmt.Fprintf(&output, "  %s (%s): %d documents, %d passages", s.ID, s.Type, s.Documents, s.Passages)
				if s.Embedded > 0 {
					fmt.Fprintf(&output, ", %d embedded", s.Embedded)
				}
				fmt.Fprintln(&output)
			}
		}
	case "mcp":
		loaded, err := a.Load(ctx)
		if err != nil {
			return fail(err, 1)
		}
		_ = loaded
		server := &mcp.Server{App: a, Version: Version, Log: stderr}
		if err := server.Serve(ctx, streams.Stdin, stdout); err != nil && !errors.Is(err, context.Canceled) {
			return fail(err, 1)
		}
		return 0
	case "sync":
		loaded, err := a.Load(ctx)
		if err != nil {
			return fail(err, 1)
		}
		if dir == "" {
			dir = loaded.Config.Dir
		} else if !filepath.IsAbs(dir) {
			dir = filepath.Join(cwd, dir)
		}
		plan, err := nativesync.Plan(nativesync.Options{Agent: agent, Project: app.Project(loaded.Config, project, wasSet(flags, "project")), Dir: dir, Items: loaded.Items})
		if err != nil {
			return fail(err, 1)
		}
		value = plan
		changed := plan.Changed()
		if check {
			if !asJSON {
				renderPlan(&output, plan, true)
			} else if err := encode(&output, value); err != nil {
				return fail(err, 1)
			}
			code := write(output.Bytes())
			if code == 0 && (changed || len(plan.Conflicts) > 0) {
				return 1
			}
			return code
		}
		if err := nativesync.Apply(dir, plan); err != nil {
			return fail(err, 1)
		}
		if !asJSON {
			renderPlan(&output, plan, false)
		}
	}
	if asJSON {
		if err := encode(&output, value); err != nil {
			return fail(err, 1)
		}
	}
	return write(output.Bytes())
}

func encode(w io.Writer, value any) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
}

func wasSet(flags *flag.FlagSet, name string) bool {
	found := false
	flags.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}

func renderResult(output io.Writer, result resolver.Result) {
	for _, group := range []struct {
		name  string
		items []resolver.Selection
	}{{"Required rules", result.RequiredRules}, {"Optional rules", result.OptionalRules}, {"Skills", result.Skills}, {"References", result.References}, {"Decisions", result.Decisions}} {
		fmt.Fprintf(output, "%s (%d)\n", group.name, len(group.items))
		for _, selection := range group.items {
			i := selection.Item
			fmt.Fprintf(output, "  %s — %s\n  Source: %s:%s; version: %s\n", i.ID, i.Name, i.Source.ID, i.Source.Path, i.Version)
			for _, reason := range selection.Reasons {
				if reason.Unrestricted {
					fmt.Fprintf(output, "  %s: unrestricted\n", reason.Dimension)
				} else if len(reason.Matches) > 0 {
					for _, match := range reason.Matches {
						fmt.Fprintf(output, "  paths: %s matched %s\n", match.Pattern, match.File)
					}
				} else {
					fmt.Fprintf(output, "  %s: %s\n", reason.Dimension, reason.Value)
				}
			}
			fmt.Fprintf(output, "\n%s\n\n", i.Content)
		}
	}
}

func renderHits(output io.Writer, title string, hits []retrieval.Hit) {
	fmt.Fprintf(output, "%s (%d)\n", title, len(hits))
	for _, h := range hits {
		p := h.Passage
		heading := p.Title
		if p.Heading != "" && p.Heading != p.Title {
			heading += " › " + strings.TrimPrefix(p.Heading, p.Title+" › ")
		}
		fmt.Fprintf(output, "  %s — %s\n  Source: %s (%s); id: %s\n", heading, p.Location, p.SourceID, p.SourceType, p.ID)
		for _, r := range h.Reasons {
			if len(r.Terms) > 0 {
				fmt.Fprintf(output, "  %s: rank %d, matched %s\n", r.Method, r.Rank, strings.Join(r.Terms, ", "))
			} else {
				fmt.Fprintf(output, "  %s: rank %d, similarity %.3f\n", r.Method, r.Rank, r.Score)
			}
		}
		fmt.Fprintf(output, "\n%s\n\n", excerpt(p.Text, 600))
	}
}

func excerpt(text string, n int) string {
	if utf8.RuneCountInString(text) <= n {
		return text
	}
	r := []rune(text)
	return strings.TrimSpace(string(r[:n])) + " …"
}

func renderPlan(output io.Writer, plan nativesync.Result, check bool) {
	verb := "Synced"
	if check {
		verb = "Checked"
	}
	fmt.Fprintf(output, "%s %s knowledge for project %s\n", verb, plan.Agent, plan.Project)
	for _, c := range plan.Changes {
		action := c.Action
		if check && action != "unchanged" {
			action = "would " + action
		}
		fmt.Fprintf(output, "  %-16s %s\n", action, c.Path)
	}
	for _, c := range plan.Conflicts {
		fmt.Fprintf(output, "  conflict         %s\n", c)
	}
	if check && !plan.Changed() && len(plan.Conflicts) == 0 {
		fmt.Fprintln(output, "Everything is up to date.")
	}
}
