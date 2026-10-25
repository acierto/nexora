// Package cli implements the standalone local Nexora workflow.
package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/acierto/nexora/internal/config"
	"github.com/acierto/nexora/internal/knowledge"
	"github.com/acierto/nexora/internal/markdown"
	"github.com/acierto/nexora/internal/registry"
	"github.com/acierto/nexora/internal/resolver"
)

const usage = `Nexora — local knowledge for AI development workflows

Usage: nexora <command> [flags]

Commands:
  init       Create nexora.yaml and knowledge/rules/project.md safely
  validate   Validate all configured knowledge
  list       List all knowledge, including inactive items
  resolve    Select active knowledge by project, agent, and files
  help       Show help (help <command> for flags)

Read commands accept --config <path> and --json.
Run nexora <command> --help for command flags.
`

type fileFlags []string

func (f *fileFlags) String() string         { return strings.Join(*f, ",") }
func (f *fileFlags) Set(value string) error { *f = append(*f, value); return nil }

func Run(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	return run(ctx, args, cwd, stdout, stderr)
}

func run(ctx context.Context, args []string, cwd string, stdout, stderr io.Writer) int {
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
	if command != "init" && command != "validate" && command != "list" && command != "resolve" {
		return fail(fmt.Errorf("unknown command %q\n%s", command, usage), 2)
	}
	var parseOutput bytes.Buffer
	flags := flag.NewFlagSet(command, flag.ContinueOnError)
	flags.SetOutput(&parseOutput)
	flags.Usage = func() { fmt.Fprintf(&parseOutput, "Usage: nexora %s [flags]\n", command); flags.PrintDefaults() }
	var configPath, project, agent string
	var asJSON bool
	var files fileFlags
	if command == "init" {
		flags.StringVar(&project, "project", "", "project ID (default: current directory name)")
	} else {
		flags.StringVar(&configPath, "config", "./nexora.yaml", "configuration file (no parent discovery)")
		flags.BoolVar(&asJSON, "json", false, "emit one JSON object")
		if command == "resolve" {
			flags.StringVar(&project, "project", "", "project ID (default: configuration project)")
			flags.StringVar(&agent, "agent", "", "agent ID for applicability matching only")
			flags.Var(&files, "file", "repository-relative file (repeatable)")
		}
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
	cfg, err := config.Load(configPath)
	if err != nil {
		return fail(err, 1)
	}
	items := []knowledge.Item{}
	for _, source := range cfg.Sources {
		loaded, err := markdown.Load(ctx, source)
		if err != nil {
			return fail(err, 1)
		}
		items = append(items, loaded...)
	}
	snapshot, err := registry.New(items)
	if err != nil {
		return fail(err, 1)
	}
	items = snapshot.List()
	var value any
	switch command {
	case "validate":
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
		}
	case "list":
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
	case "resolve":
		if !wasSet(flags, "project") {
			project = cfg.Project
		}
		result, err := resolver.Resolve(items, resolver.Request{Project: project, Agent: agent, Files: files})
		if err != nil {
			return fail(err, 1)
		}
		value = result
		if !asJSON {
			renderResult(&output, result)
		}
	}
	if asJSON {
		encoder := json.NewEncoder(&output)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(value); err != nil {
			return fail(err, 1)
		}
	}
	return write(output.Bytes())
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
