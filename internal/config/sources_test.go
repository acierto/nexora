package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDocumentSourcesAndIndex(t *testing.T) {
	root := t.TempDir()
	for _, dir := range []string{"knowledge", "handbook"} {
		if err := os.MkdirAll(filepath.Join(root, dir), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	file := filepath.Join(root, "nexora.yaml")
	base := `schema_version: 1
project: commerce-api
sources:
  - id: local
    path: ./knowledge
  - id: handbook
    type: docs
    path: ./handbook
    projects: [commerce-api]
  - id: wiki
    type: confluence
    url: https://example.atlassian.net/wiki/
    spaces: [PAY, ARCH]
    user_env: CONFLUENCE_USER
    token_env: CONFLUENCE_TOKEN
`
	write := func(data string) {
		t.Helper()
		if err := os.WriteFile(file, []byte(data), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write(base)
	cfg, err := Load(file)
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.KnowledgeSources()) != 1 || cfg.KnowledgeSources()[0].Type != SourceKnowledge {
		t.Fatalf("knowledge sources: %+v", cfg.KnowledgeSources())
	}
	docs := cfg.DocumentSources()
	if len(docs) != 2 || docs[0].ID != "handbook" || docs[1].ID != "wiki" || docs[1].URL != "https://example.atlassian.net/wiki" {
		t.Fatalf("document sources: %+v", docs)
	}
	if docs[1].Root != "" {
		t.Fatal("confluence source must not have a root")
	}
	if cfg.Index.Store != StoreLocal || cfg.Index.File != filepath.Join(cfg.Dir, ".nexora", "index.json") {
		t.Fatalf("index defaults: %+v", cfg.Index)
	}
	if cfg.Index.Embeddings.Provider != ProviderNone || cfg.Index.Chunk.MaxChars != 1200 || cfg.Index.Chunk.OverlapChars != 150 {
		t.Fatalf("defaults: %+v", cfg.Index)
	}

	write(base + "index:\n  embeddings:\n    provider: openai\n    model: text-embedding-3-small\n    api_key_env: OPENAI_API_KEY\n  chunk:\n    max_chars: 800\n    overlap_chars: 0\n")
	cfg, err = Load(file)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Index.Embeddings.URL != "https://api.openai.com/v1" || cfg.Index.Chunk.OverlapChars != 0 {
		t.Fatalf("openai defaults: %+v", cfg.Index)
	}
	write(base + "index:\n  store: pgvector\n  dsn_env: NEXORA_DATABASE_URL\n  embeddings:\n    provider: ollama\n    model: nomic-embed-text\n    dimensions: 768\n")
	cfg, err = Load(file)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Index.Table != "nexora_passages" || cfg.Index.Embeddings.URL != "http://localhost:11434" {
		t.Fatalf("pgvector defaults: %+v", cfg.Index)
	}

	confluence := "  - id: c\n    type: confluence\n    url: https://x.example/wiki\n    spaces: [A]\n    token_env: TOKEN\n"
	head := "schema_version: 1\nproject: p\nsources:\n"
	for name, data := range map[string]string{
		"unknown type":             head + "  - id: x\n    type: wiki\n    path: ./knowledge\n",
		"confluence without space": head + strings.Replace(confluence, "    spaces: [A]\n", "", 1),
		"confluence duplicate":     head + strings.Replace(confluence, "[A]", "[A, A]", 1),
		"confluence with path":     head + confluence + "    path: ./knowledge\n",
		"confluence bad url":       head + strings.Replace(confluence, "https://x.example/wiki", "ftp://x", 1),
		"confluence no token":      head + strings.Replace(confluence, "    token_env: TOKEN\n", "", 1),
		"confluence bad env":       head + strings.Replace(confluence, "TOKEN", "my token", 1),
		"knowledge with url":       head + "  - id: k\n    path: ./knowledge\n    url: https://x.example\n",
		"knowledge with projects":  head + "  - id: k\n    path: ./knowledge\n    projects: [p]\n",
		"docs overlap":             head + "  - id: a\n    path: ./knowledge\n  - id: b\n    type: docs\n    path: ./knowledge\n",
		"bad store":                base + "index:\n  store: sqlite\n",
		"pgvector no dsn":          base + "index:\n  store: pgvector\n",
		"pgvector with path":       base + "index:\n  store: pgvector\n  dsn_env: DSN\n  path: ./x.json\n",
		"bad table":                base + "index:\n  store: pgvector\n  dsn_env: DSN\n  table: Passages-1\n",
		"local with dsn":           base + "index:\n  dsn_env: DSN\n",
		"model without provider":   base + "index:\n  embeddings:\n    model: x\n",
		"provider without model":   base + "index:\n  embeddings:\n    provider: ollama\n",
		"openai without key":       base + "index:\n  embeddings:\n    provider: openai\n    model: m\n",
		"ollama with key":          base + "index:\n  embeddings:\n    provider: ollama\n    model: m\n    api_key_env: K\n",
		"chunk too small":          base + "index:\n  chunk:\n    max_chars: 50\n",
		"overlap too large":        base + "index:\n  chunk:\n    max_chars: 400\n    overlap_chars: 300\n",
		"unknown index field":      base + "index:\n  backend: x\n",
	} {
		t.Run(name, func(t *testing.T) {
			write(data)
			if _, err := Load(file); err == nil {
				t.Fatal("accepted invalid configuration")
			}
		})
	}
}
