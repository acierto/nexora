package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/acierto/nexora/internal/resolver"
)

func invoke(t *testing.T, root string, args ...string) (int, string, string) {
	t.Helper()
	var out, err bytes.Buffer
	code := run(context.Background(), args, root, &out, &err)
	return code, out.String(), err.String()
}
func put(t *testing.T, file, data string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(file), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file, []byte(data), 0644); err != nil {
		t.Fatal(err)
	}
}
func initialized(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	if code, out, err := invoke(t, root, "init", "--project", "commerce-api"); code != 0 {
		t.Fatal(code, out, err)
	}
	return root
}

func TestUsage(t *testing.T) {
	root := t.TempDir()
	for _, args := range [][]string{nil, {"help"}, {"--help"}, {"init", "--help"}, {"help", "resolve"}, {"list", "-h"}} {
		code, out, err := invoke(t, root, args...)
		if code != 0 || out == "" || err != "" {
			t.Fatalf("%v: %d %q %q", args, code, out, err)
		}
	}
	for _, args := range [][]string{{"unknown"}, {"list", "--unknown"}, {"resolve", "positional"}, {"resolve", "--file"}, {"help", "resolve", "extra"}} {
		code, out, err := invoke(t, root, args...)
		if code != 2 || out != "" || err == "" {
			t.Fatalf("%v: %d %q %q", args, code, out, err)
		}
	}
	code, out, err := invoke(t, root, "validate")
	if code != 1 || out != "" || !strings.Contains(err, "nexora init") {
		t.Fatal(code, out, err)
	}
}

func TestReadWorkflow(t *testing.T) {
	root := initialized(t)
	configFile := filepath.Join(root, "nexora.yaml")
	before, _ := os.ReadFile(configFile)
	for command, want := range map[string]string{"validate": "Valid: 1 knowledge item", "list": "project/conventions", "resolve": "Required rules (1)"} {
		code, out, err := invoke(t, root, command)
		if code != 0 || !strings.Contains(out, want) || err != "" {
			t.Fatal(code, out, err)
		}
	}
	code, out, err := invoke(t, root, "validate", "--json")
	if code != 0 || err != "" {
		t.Fatal(code, out, err)
	}
	var valid map[string]any
	if e := json.Unmarshal([]byte(out), &valid); e != nil || valid["valid"] != true || valid["item_count"] != float64(1) {
		t.Fatal(out, e)
	}
	for _, project := range []string{"commerce-api", "other"} {
		code, out, err := invoke(t, root, "resolve", "--json", "--project", project, "--agent", "codex", "--file", "src/main.go", "--file", "src/auth/a.go")
		if code != 0 || err != "" {
			t.Fatal(code, out, err)
		}
		var result resolver.Result
		if e := json.Unmarshal([]byte(out), &result); e != nil {
			t.Fatal(e)
		}
		want := 0
		if project == "commerce-api" {
			want = 1
		}
		if len(result.RequiredRules) != want {
			t.Fatal(out)
		}
		if strings.Contains(out, root) || strings.Contains(out, "null") {
			t.Fatal(out)
		}
	}
	_, out, _ = invoke(t, root, "resolve", "--json", "--project", "other")
	if strings.TrimSpace(out) != "{\n  \"required_rules\": [],\n  \"optional_rules\": [],\n  \"skills\": [],\n  \"references\": [],\n  \"decisions\": []\n}" {
		t.Fatal(out)
	}
	after, _ := os.ReadFile(configFile)
	if !bytes.Equal(before, after) {
		t.Fatal("config mutated")
	}
	code, out, err = invoke(t, t.TempDir(), "validate", "--config", configFile, "--json")
	if code != 0 || err != "" {
		t.Fatal(code, out, err)
	}
	for idx, status := range []string{"draft", "archived"} {
		put(t, filepath.Join(root, "knowledge", status+".md"), fmt.Sprintf("---\nid: %s\nname: Inactive\ndescription: Inactive rule\nkind: rule\nstatus: %s\n---\nContent %d\n", status, status, idx))
	}
	code, out, err = invoke(t, root, "list", "--json")
	if code != 0 || !strings.Contains(out, `"status": "draft"`) || !strings.Contains(out, `"status": "archived"`) || err != "" {
		t.Fatal(code, out, err)
	}
	for _, file := range []string{"/absolute", `C:\private`, "../private"} {
		code, out, err = invoke(t, root, "resolve", "--file", file)
		if code != 1 || out != "" || err == "" {
			t.Fatal(code, out, err)
		}
	}
	put(t, filepath.Join(root, "knowledge", "invalid.md"), "invalid")
	for _, command := range []string{"validate", "list", "resolve"} {
		code, out, err = invoke(t, root, command, "--json")
		if code != 1 || out != "" || !strings.Contains(err, "invalid.md") {
			t.Fatal(code, out, err)
		}
	}
}

func TestInitializeSafety(t *testing.T) {
	for _, target := range []string{"nexora.yaml", "knowledge/rules/project.md", "knowledge", "knowledge/rules"} {
		t.Run(target, func(t *testing.T) {
			root := t.TempDir()
			file := filepath.Join(root, filepath.FromSlash(target))
			put(t, file, "owned by user")
			code, out, err := invoke(t, root, "init")
			if code != 1 || out != "" || err == "" {
				t.Fatal(code, out, err)
			}
			data, _ := os.ReadFile(file)
			if string(data) != "owned by user" {
				t.Fatal("existing file modified")
			}
			other := "nexora.yaml"
			if target == "nexora.yaml" {
				other = "knowledge/rules/project.md"
			}
			if _, e := os.Stat(filepath.Join(root, filepath.FromSlash(other))); !errors.Is(e, os.ErrNotExist) {
				t.Fatal("partial creation", e)
			}
		})
	}
	for _, target := range []string{"nexora.yaml", "knowledge", "knowledge/rules", "knowledge/rules/project.md"} {
		t.Run("symlink-"+target, func(t *testing.T) {
			root := t.TempDir()
			link := filepath.Join(root, filepath.FromSlash(target))
			if e := os.MkdirAll(filepath.Dir(link), 0755); e != nil {
				t.Fatal(e)
			}
			if e := os.Symlink(t.TempDir(), link); e != nil {
				t.Skip(e)
			}
			code, out, err := invoke(t, root, "init")
			if code != 1 || out != "" || !strings.Contains(err, "symlink") {
				t.Fatal(code, out, err)
			}
		})
	}
	root := t.TempDir()
	for _, name := range []string{"AGENTS.md", "CLAUDE.md", "knowledge/unrelated.txt"} {
		put(t, filepath.Join(root, filepath.FromSlash(name)), "preserve me")
	}
	code, out, err := invoke(t, root, "init")
	if code != 0 {
		t.Fatal(code, out, err)
	}
	for _, name := range []string{"AGENTS.md", "CLAUDE.md", "knowledge/unrelated.txt"} {
		data, _ := os.ReadFile(filepath.Join(root, filepath.FromSlash(name)))
		if string(data) != "preserve me" {
			t.Fatal("unrelated file changed")
		}
	}
	if code, _, _ := invoke(t, root, "init"); code != 1 {
		t.Fatal("reinit succeeded")
	}
	code, out, err = invoke(t, root, "resolve", "--json")
	if code != 0 || !strings.Contains(out, `"project/conventions"`) {
		t.Fatal(code, out, err)
	}
	if code, _, _ := invoke(t, t.TempDir(), "init", "--project", ""); code != 1 {
		t.Fatal("empty project accepted")
	}
	t.Run("write failure cleanup", func(t *testing.T) {
		root := t.TempDir()
		calls := 0
		err := initializeWith(root, "api", func(file string, data []byte) error {
			calls++
			if calls == 2 {
				return errors.New("simulated disk error")
			}
			return writeExclusive(file, data)
		})
		if err == nil {
			t.Fatal("expected failure")
		}
		if _, e := os.Stat(filepath.Join(root, "nexora.yaml")); !errors.Is(e, os.ErrNotExist) {
			t.Fatal("owned file not removed")
		}
	})
	t.Run("exclusive creation", func(t *testing.T) {
		file := filepath.Join(t.TempDir(), "owned")
		put(t, file, "keep")
		if e := writeExclusive(file, []byte("replace")); e == nil {
			t.Fatal("overwrite succeeded")
		}
		data, _ := os.ReadFile(file)
		if string(data) != "keep" {
			t.Fatal("file overwritten")
		}
	})
}

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) { return 0, errors.New("write failed") }
func TestOutputFailureAndCancellation(t *testing.T) {
	var err bytes.Buffer
	if code := run(context.Background(), []string{"help"}, t.TempDir(), failingWriter{}, &err); code != 1 || !strings.Contains(err.String(), "write failed") {
		t.Fatal(code, err.String())
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var out bytes.Buffer
	err.Reset()
	if code := run(ctx, []string{"validate"}, t.TempDir(), &out, &err); code != 1 || out.Len() != 0 {
		t.Fatal(code, out.String(), err.String())
	}
}
