package markdown

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/acierto/nexora/internal/config"
	"github.com/acierto/nexora/internal/knowledge"
)

const document = "---\nid: go/testing\nname: Testing\ndescription: Testing conventions\nkind: rule\n---\nWrite useful tests.\n"

func TestParse(t *testing.T) {
	for _, kind := range []string{"rule", "skill", "reference", "decision"} {
		t.Run(kind, func(t *testing.T) {
			item, err := Parse([]byte(strings.Replace(document, "kind: rule", "kind: "+kind, 1)), knowledge.Source{ID: "local", Path: "test.md"})
			if err != nil {
				t.Fatal(err)
			}
			if item.Required != (kind == "rule") || item.Status != knowledge.Active || item.Priority != 0 || item.Content != "Write useful tests.\n" || !strings.HasPrefix(item.Version, "sha256:") {
				t.Fatalf("%+v", item)
			}
		})
	}
	crlf := strings.ReplaceAll(document, "\n", "\r\n")
	item, err := Parse([]byte(crlf), knowledge.Source{})
	if err != nil || item.Content != "Write useful tests.\r\n" {
		t.Fatalf("%+v %v", item, err)
	}
	custom := strings.Replace(document, "kind: rule", "kind: rule\nrequired: false\nversion: \"1.0\"\npriority: -1\nstatus: draft\ntags: [go]\napplies_to:\n  projects: [api]\n  agents: [codex]\n  paths: [\"src/**/*.go\"]", 1)
	item, err = Parse([]byte(custom), knowledge.Source{})
	if err != nil || item.Required || item.Version != "1.0" || item.Priority != -1 {
		t.Fatalf("%+v %v", item, err)
	}
	for name, data := range map[string]string{
		"no front matter": "plain markdown", "unclosed": "---\nid: a\n", "invalid YAML": "---\nid: [\n---\nbody",
		"unknown": strings.Replace(document, "kind: rule", "kind: rule\nunknown: x", 1), "unknown nested": strings.Replace(document, "kind: rule", "kind: rule\napplies_to:\n  typo: x", 1),
		"missing id": strings.Replace(document, "id: go/testing\n", "", 1), "empty name": strings.Replace(document, "name: Testing", "name: \"\"", 1), "kind": strings.Replace(document, "kind: rule", "kind: other", 1),
		"skill required": strings.Replace(document, "kind: rule", "kind: skill\nrequired: false", 1), "empty body": strings.Replace(document, "Write useful tests.", " ", 1),
		"invalid UTF8": document + string([]byte{255}), "numeric version": strings.Replace(document, "kind: rule", "kind: rule\nversion: 1.0", 1), "null": strings.Replace(document, "kind: rule", "kind: rule\nrequired: null", 1),
		"status": strings.Replace(document, "kind: rule", "kind: rule\nstatus: nope", 1), "invalid pattern": strings.Replace(document, "kind: rule", "kind: rule\napplies_to:\n  paths: [\"[ab]\"]", 1),
	} {
		t.Run(name, func(t *testing.T) {
			if _, e := Parse([]byte(data), knowledge.Source{}); e == nil {
				t.Fatal("accepted invalid document")
			}
		})
	}
}

func TestUnreadableSourceEntries(t *testing.T) {
	if os.Geteuid() == 0 || runtime.GOOS == "windows" {
		t.Skip("requires POSIX permission enforcement")
	}
	for _, directory := range []bool{false, true} {
		t.Run(fmt.Sprintf("directory=%v", directory), func(t *testing.T) {
			root := t.TempDir()
			file := filepath.Join(root, "locked.md")
			if directory {
				if err := os.Mkdir(file, 0000); err != nil {
					t.Fatal(err)
				}
			} else {
				if err := os.WriteFile(file, []byte(document), 0000); err != nil {
					t.Fatal(err)
				}
			}
			defer os.Chmod(file, 0755)
			items, err := Load(context.Background(), config.Source{ID: "local", Root: root})
			if err == nil || items != nil || !strings.Contains(err.Error(), "locked.md") {
				t.Fatalf("expected path-specific permission failure: %v %v", items, err)
			}
		})
	}
}

func TestLoad(t *testing.T) {
	root := t.TempDir()
	if e := os.Mkdir(filepath.Join(root, "nested"), 0755); e != nil {
		t.Fatal(e)
	}
	file := filepath.Join(root, "nested", "rule.md")
	if e := os.WriteFile(file, []byte(document), 0644); e != nil {
		t.Fatal(e)
	}
	if e := os.WriteFile(filepath.Join(root, "ignore.txt"), []byte("ignored"), 0644); e != nil {
		t.Fatal(e)
	}
	source := config.Source{ID: "local", Root: root}
	first, err := Load(context.Background(), source)
	if err != nil || len(first) != 1 {
		t.Fatalf("%v %v", first, err)
	}
	second, err := Load(context.Background(), source)
	if err != nil || !reflect.DeepEqual(first, second) {
		t.Fatal("nonrepeatable load", err)
	}
	if first[0].Source.Path != "nested/rule.md" || len(first[0].Source.Digest) != 64 {
		t.Fatalf("%+v", first[0].Source)
	}
	if e := os.WriteFile(file, []byte(document+"Changed."), 0644); e != nil {
		t.Fatal(e)
	}
	changed, e := Load(context.Background(), source)
	if e != nil || changed[0].Source.Digest == first[0].Source.Digest {
		t.Fatal("digest unchanged", e)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if items, e := Load(ctx, source); !errors.Is(e, context.Canceled) || items != nil {
		t.Fatalf("%v %v", items, e)
	}
	for _, name := range []string{"link.txt", "linked-dir"} {
		t.Run(name, func(t *testing.T) {
			target := file
			if name == "linked-dir" {
				target = root
			}
			link := filepath.Join(root, name)
			if e := os.Symlink(target, link); e != nil {
				t.Skip(e)
			}
			defer os.Remove(link)
			if items, e := Load(context.Background(), source); e == nil || items != nil || !strings.Contains(e.Error(), name) {
				t.Fatalf("%v %v", items, e)
			}
		})
	}
	if e := os.WriteFile(filepath.Join(root, "invalid.md"), []byte("invalid"), 0644); e != nil {
		t.Fatal(e)
	}
	if items, e := Load(context.Background(), source); e == nil || items != nil || !strings.Contains(e.Error(), "local, invalid.md") {
		t.Fatalf("%v %v", items, e)
	}
	if items, e := Load(context.Background(), config.Source{ID: "empty", Root: t.TempDir()}); e != nil || len(items) != 0 || items == nil {
		t.Fatalf("%v %v", items, e)
	}
	if items, e := Load(context.Background(), config.Source{ID: "missing", Root: filepath.Join(root, "missing")}); e == nil || items != nil {
		t.Fatalf("%v %v", items, e)
	}
}
