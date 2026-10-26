package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConfig(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "knowledge", "nested"), 0755); err != nil {
		t.Fatal(err)
	}
	base := "schema_version: 1\nproject: commerce-api\nsources:\n  - id: local\n    path: ./knowledge\n"
	file := filepath.Join(root, "nexora.yaml")
	write := func(data string) {
		t.Helper()
		if e := os.WriteFile(file, []byte(data), 0644); e != nil {
			t.Fatal(e)
		}
	}
	write(base)
	cfg, err := Load(file)
	if err != nil {
		t.Fatal(err)
	}
	want, _ := filepath.EvalSymlinks(filepath.Join(root, "knowledge"))
	if cfg.Sources[0].Root != want {
		t.Fatalf("root: %s", cfg.Sources[0].Root)
	}
	for name, data := range map[string]string{
		"version": strings.Replace(base, "version: 1", "version: 2", 1), "unknown": base + "other: true\n", "missing project": strings.Replace(base, "project: commerce-api\n", "", 1),
		"duplicate": base + "  - id: local\n    path: ./knowledge\n", "overlap": base + "  - id: nested\n    path: ./knowledge/nested\n", "identical": base + "  - id: other\n    path: ./knowledge\n",
		"missing root": strings.Replace(base, "./knowledge", "./absent", 1), "file root": strings.Replace(base, "./knowledge", "./nexora.yaml", 1), "empty sources": "schema_version: 1\nproject: x\nsources: []\n",
		"unknown nested": base + "    root: /private\n",
	} {
		t.Run(name, func(t *testing.T) {
			write(data)
			if _, e := Load(file); e == nil {
				t.Fatal("accepted invalid config")
			}
		})
	}
	if e := os.Mkdir(filepath.Join(root, "shared"), 0755); e != nil {
		t.Fatal(e)
	}
	write(base + "  - id: shared\n    path: ./shared\n")
	if _, e := Load(file); e != nil {
		t.Fatal(e)
	}
	if _, e := Load(filepath.Join(root, "missing.yaml")); e == nil || !strings.Contains(e.Error(), "nexora init") {
		t.Fatalf("%v", e)
	}
	t.Run("unreadable", func(t *testing.T) {
		if os.Geteuid() == 0 {
			t.Skip("root bypasses permissions")
		}
		dir := filepath.Join(root, "locked")
		if e := os.Mkdir(dir, 0000); e != nil {
			t.Fatal(e)
		}
		defer os.Chmod(dir, 0755)
		write(strings.Replace(base, "./knowledge", "./locked", 1))
		if _, e := Load(file); e == nil {
			t.Fatal("accepted unreadable root")
		}
	})
	t.Run("case alias overlap", func(t *testing.T) {
		if _, err := os.Stat(filepath.Join(root, "KNOWLEDGE")); err != nil {
			t.Skip("case-sensitive filesystem")
		}
		write(base + "  - id: other\n    path: ./KNOWLEDGE/nested\n")
		if _, err := Load(file); err == nil {
			t.Fatal("accepted overlapping directory identities")
		}
	})
}

func TestDeliveryGuidanceValidation(t *testing.T) {
	root := t.TempDir()
	base := "schema_version: 1\nproject: test\nsources:\n  - id: local\n    path: .\n"
	file := filepath.Join(root, "nexora.yaml")
	for _, tc := range []struct {
		value string
		valid bool
	}{
		{"documentation: '  Use the connected docs MCP.  '", true},
		{"documentation: ''", true},
		{"documentation: 42", false},
		{"documentation: null", false},
		{"documentation: '" + strings.Repeat("x", 2001) + "'", false},
		{"endpoint: https://example.test", false},
	} {
		if err := os.WriteFile(file, []byte(base+"delivery:\n  "+tc.value+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		cfg, err := Load(file)
		if (err == nil) != tc.valid {
			t.Fatalf("%s: %v", tc.value, err)
		}
		if tc.valid && cfg.Delivery.Documentation != strings.TrimSpace(cfg.Delivery.Documentation) {
			t.Fatal("untrimmed guidance")
		}
	}
}
