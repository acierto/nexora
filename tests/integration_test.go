package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/acierto/nexora/internal/resolver"
)

var repository, executable string

func TestMain(m *testing.M) {
	var err error
	repository, err = filepath.Abs("..")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	dir, err := os.MkdirTemp("", "nexora-integration-")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	name := "nexora"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	executable = filepath.Join(dir, name)
	build := exec.Command("go", "build", "-o", executable, "./cmd/nexora")
	build.Dir = repository
	if output, err := build.CombinedOutput(); err != nil {
		fmt.Fprintln(os.Stderr, string(output), err)
		os.RemoveAll(dir)
		os.Exit(1)
	}
	code := m.Run()
	os.RemoveAll(dir)
	os.Exit(code)
}

func command(t *testing.T, dir string, args ...string) (int, string, string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, executable, args...)
	cmd.Dir = dir
	var out, errOut bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errOut
	err := cmd.Run()
	if ctx.Err() != nil {
		t.Fatal(ctx.Err())
	}
	if err == nil {
		return 0, out.String(), errOut.String()
	}
	if exit, ok := err.(*exec.ExitError); ok {
		return exit.ExitCode(), out.String(), errOut.String()
	}
	t.Fatal(err)
	return 0, "", ""
}

func resolve(t *testing.T, args ...string) resolver.Result {
	t.Helper()
	args = append([]string{"resolve", "--json"}, args...)
	code, out, err := command(t, repository, args...)
	if code != 0 || err != "" {
		t.Fatal(code, out, err)
	}
	var result resolver.Result
	if e := json.Unmarshal([]byte(out), &result); e != nil {
		t.Fatal(e)
	}
	return result
}

func ids(selections []resolver.Selection) []string {
	values := []string{}
	for _, s := range selections {
		values = append(values, s.Item.ID)
	}
	return values
}
func expectIDs(t *testing.T, got []resolver.Selection, want ...string) {
	t.Helper()
	if want == nil {
		want = []string{}
	}
	if !reflect.DeepEqual(ids(got), want) {
		t.Fatalf("IDs %v; want %v", ids(got), want)
	}
}

func TestLocalExample(t *testing.T) {
	config := "examples/local/nexora.yaml"
	code, out, err := command(t, repository, "validate", "--config", config, "--json")
	if code != 0 || err != "" {
		t.Fatal(code, out, err)
	}
	var check struct {
		Valid bool `json:"valid"`
		Count int  `json:"item_count"`
	}
	if e := json.Unmarshal([]byte(out), &check); e != nil || !check.Valid || check.Count != 9 {
		t.Fatal(out, e)
	}
	for _, agent := range []string{"codex", "claude"} {
		result := resolve(t, "--config", config, "--agent", agent, "--file", "src/auth/service.go")
		expectIDs(t, result.RequiredRules, "security/authentication", "go/conventions", "agents/"+agent)
		expectIDs(t, result.OptionalRules, "go/naming")
		expectIDs(t, result.Skills, "skills/add-endpoint")
		expectIDs(t, result.References, "references/authentication")
		expectIDs(t, result.Decisions, "decisions/local-knowledge")
		if len(result.RequiredRules[0].Reasons[3].Matches) != 1 || result.RequiredRules[0].Item.Source.Path != "rules/authentication.md" || result.RequiredRules[0].Item.Version != "1.0.0" {
			t.Fatal("missing documented provenance")
		}
	}
	result := resolve(t, "--config", config, "--file", "src/catalog/service.go")
	expectIDs(t, result.RequiredRules, "go/conventions")
	expectIDs(t, result.References)
	expectIDs(t, result.Skills, "skills/add-endpoint")
	result = resolve(t, "--config", config, "--agent", "codex")
	expectIDs(t, result.RequiredRules, "go/conventions", "agents/codex")
	result = resolve(t, "--config", config, "--project", "unrelated", "--agent", "codex", "--file", "src/auth/service.go")
	expectIDs(t, result.RequiredRules, "go/conventions", "agents/codex")
	expectIDs(t, result.OptionalRules)
	expectIDs(t, result.Skills)
	expectIDs(t, result.References)
	expectIDs(t, result.Decisions)
	result = resolve(t, "--config", config, "--agent", "codex", "--file", "src/auth/service.go", "--file", "./src/auth/service.go", "--file", "src/catalog/service.go")
	expectIDs(t, result.RequiredRules, "security/authentication", "go/conventions", "agents/codex")
}

func TestSharedExamples(t *testing.T) {
	for _, project := range []string{"commerce-api", "billing-api"} {
		config := "examples/shared-knowledge/projects/" + project + "/nexora.yaml"
		code, out, err := command(t, repository, "validate", "--config", config)
		if code != 0 || err != "" || !strings.Contains(out, "3 knowledge items") {
			t.Fatal(code, out, err)
		}
		result := resolve(t, "--config", config, "--file", "src/auth/service.go")
		if project == "commerce-api" {
			expectIDs(t, result.RequiredRules, "security/authentication", "go/conventions", "commerce/errors")
		} else {
			expectIDs(t, result.RequiredRules, "go/conventions", "billing/errors")
		}
		if len(result.OptionalRules)+len(result.Skills)+len(result.References)+len(result.Decisions) != 0 {
			t.Fatal("unexpected optional knowledge")
		}
		other := "commerce-api"
		if project == other {
			other = "billing-api"
		}
		result = resolve(t, "--config", config, "--project", other, "--file", "src/catalog/service.go")
		expectIDs(t, result.RequiredRules, "go/conventions")
	}
	result := resolve(t, "--config", "examples/shared-knowledge/projects/commerce-api/nexora.yaml", "--project", "billing-api", "--file", "src/auth/service.go")
	expectIDs(t, result.RequiredRules, "go/conventions")
}

func TestExecutableWorkflow(t *testing.T) {
	root := t.TempDir()
	code, out, err := command(t, root, "init", "--project", "my-api")
	if code != 0 || err != "" {
		t.Fatal(code, out, err)
	}
	for _, cmd := range []string{"validate", "list", "resolve"} {
		code, out, err = command(t, root, cmd, "--json")
		if code != 0 || err != "" || !json.Valid([]byte(out)) {
			t.Fatal(cmd, code, out, err)
		}
	}
	code, out, err = command(t, root, "resolve", "--project", "other", "--json")
	if code != 0 || strings.Contains(out, "project/conventions") {
		t.Fatal(code, out, err)
	}
	if e := os.WriteFile(filepath.Join(root, "knowledge", "invalid.md"), []byte("invalid"), 0644); e != nil {
		t.Fatal(e)
	}
	for _, cmd := range []string{"validate", "list", "resolve"} {
		code, out, err = command(t, root, cmd, "--json")
		if code != 1 || out != "" || !strings.Contains(err, "invalid.md") {
			t.Fatal(cmd, code, out, err)
		}
	}
	code, out, err = command(t, root, "bad-command")
	if code != 2 || out != "" || err == "" {
		t.Fatal(code, out, err)
	}
}

func fences(text, language string) []string {
	blocks := []string{}
	parts := strings.Split(text, "```"+language+"\n")
	for _, part := range parts[1:] {
		blocks = append(blocks, strings.SplitN(part, "```", 2)[0])
	}
	return blocks
}

func shell(t *testing.T, dir, script string) string {
	t.Helper()
	sh, err := exec.LookPath("sh")
	if err != nil {
		t.Skip("POSIX documentation recipes require sh")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, sh, "-eu", "-c", script)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "PATH="+filepath.Dir(executable)+string(os.PathListSeparator)+os.Getenv("PATH"))
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("documented recipe failed:\n%s\n%s\n%v", script, output, err)
	}
	return string(output)
}

func TestDocumentedRecipes(t *testing.T) {
	for _, file := range []string{"docs/cli.md", "docs/troubleshooting.md", "examples/local/README.md", "examples/shared-knowledge/README.md", "examples/team-docs/README.md"} {
		t.Run(file, func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join(repository, filepath.FromSlash(file)))
			if err != nil {
				t.Fatal(err)
			}
			for _, block := range fences(string(data), "sh") {
				if strings.Contains(block, "python3") {
					if _, err := exec.LookPath("python3"); err != nil {
						t.Skip("optional Python JSON scripting recipe")
					}
				}
				shell(t, repository, block)
			}
		})
	}
}

func TestQuickstart(t *testing.T) {
	data, e := os.ReadFile(filepath.Join(repository, "docs", "quickstart.md"))
	if e != nil {
		t.Fatal(e)
	}
	blocks := fences(string(data), "sh")
	samples := fences(string(data), "markdown")
	if len(blocks) != 3 || len(samples) != 1 {
		t.Fatal("quickstart blocks changed; update walkthrough coverage")
	}
	root := t.TempDir()
	output := shell(t, root, blocks[0])
	if !strings.Contains(output, "Created nexora.yaml") {
		t.Fatal(output)
	}
	if e := os.WriteFile(filepath.Join(root, "knowledge", "rules", "authentication.md"), []byte(samples[0]), 0644); e != nil {
		t.Fatal(e)
	}
	shell(t, root, blocks[1])
	output = shell(t, root, blocks[2])
	for _, want := range []string{"Valid: 2 knowledge items", "handbook (docs): 1 documents, 1 passages", "Required rules (2)\n  security/authentication", "  project/conventions", "Passages (1)\n  Token rotation — tokens.md"} {
		if !strings.Contains(output, want) {
			t.Fatalf("quickstart output lacks %q:\n%s", want, output)
		}
	}
	code, out, err := command(t, root, "resolve", "--file", "src/auth/service.go", "--json")
	if code != 0 || err != "" {
		t.Fatal(code, out, err)
	}
	var result resolver.Result
	if e := json.Unmarshal([]byte(out), &result); e != nil {
		t.Fatal(e)
	}
	expectIDs(t, result.RequiredRules, "security/authentication", "project/conventions")
	code, out, err = command(t, root, "resolve", "--project", "another-project", "--file", "src/auth/service.go", "--json")
	if code != 0 || strings.Contains(out, "security/authentication") || strings.Contains(out, "project/conventions") {
		t.Fatal(code, out, err)
	}
}

func TestDocumentationLinks(t *testing.T) {
	link := regexp.MustCompile(`\[[^\]]+\]\(([^)]+)\)`)
	for _, base := range []string{"README.md", "docs", "examples"} {
		err := filepath.WalkDir(filepath.Join(repository, base), func(file string, entry os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() || filepath.Ext(file) != ".md" {
				return nil
			}
			data, err := os.ReadFile(file)
			if err != nil {
				return err
			}
			for _, match := range link.FindAllStringSubmatch(string(data), -1) {
				target := strings.SplitN(match[1], "#", 2)[0]
				if target == "" || strings.Contains(target, "://") {
					continue
				}
				if _, err := os.Stat(filepath.Join(filepath.Dir(file), filepath.FromSlash(target))); err != nil {
					t.Errorf("%s: broken link %s", file, target)
				}
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
}
