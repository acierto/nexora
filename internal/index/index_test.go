package index

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/acierto/nexora/internal/passage"
)

func fixture() map[string][]passage.Passage {
	mk := func(id, source, heading, text string, projects []string, vec []float32) passage.Passage {
		return passage.Passage{ID: id, SourceID: source, SourceType: "docs", DocumentID: "d", Title: "Doc", Heading: heading, Text: text, Location: "d.md", Digest: "x", Projects: projects, Embedding: vec}
	}
	return map[string][]passage.Passage{
		"handbook": {
			mk("handbook:refunds.md#1", "handbook", "Refund window", "Orders can be refunded within 30 days of delivery.", nil, []float32{1, 0, 0}),
			mk("handbook:refunds.md#2", "handbook", "Partial refunds", "A partial refund never exceeds the captured amount.", nil, []float32{0.9, 0.1, 0}),
			mk("handbook:auth.md#1", "handbook", "Tokens", "Refresh tokens rotate on every use.", nil, []float32{0, 1, 0}),
		},
		"billing": {
			mk("billing:invoices.md#1", "billing", "Invoices", "Refund invoices are created by finance.", []string{"billing-api"}, []float32{0.8, 0, 0.2}),
		},
	}
}

func TestTokenize(t *testing.T) {
	got := Tokenize("How do customers get Refunds? Rotating, rotated & the API's limits")
	want := []string{"customer", "get", "refund", "rotat", "rotat", "api", "limit"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %q", got)
	}
	if !reflect.DeepEqual(Terms("refund refunds refunded"), []string{"refund"}) {
		t.Fatal(Terms("refund refunds refunded"))
	}
}

func TestLocalStore(t *testing.T) {
	ctx := context.Background()
	store := &Local{File: filepath.Join(t.TempDir(), "nested", "index.json")}
	if _, err := store.Meta(ctx); !errors.Is(err, ErrNoIndex) {
		t.Fatalf("expected ErrNoIndex, got %v", err)
	}
	meta := Meta{Model: "ollama:test", Dimensions: 3}
	if err := store.Replace(ctx, fixture(), meta, true); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(store.File)
	if err != nil || info.Mode().Perm() != 0o644 {
		t.Fatalf("index file: %v %v", info, err)
	}
	got, err := store.Keyword(ctx, Query{Terms: Terms("partial refund amount"), Project: "commerce-api", Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].Passage.ID != "handbook:refunds.md#2" {
		t.Fatalf("keyword: %+v", got)
	}
	for _, c := range got {
		if c.Passage.SourceID == "billing" {
			t.Fatal("billing passage leaked into another project")
		}
	}
	got, _ = store.Keyword(ctx, Query{Terms: Terms("refund"), Project: "billing-api", Limit: 10})
	if len(got) != 3 {
		t.Fatalf("billing project should see its own and unrestricted passages: %d", len(got))
	}
	sem, err := store.Semantic(ctx, Query{Vector: []float32{0, 1, 0}, Project: "commerce-api", Limit: 1})
	if err != nil || len(sem) != 1 || sem[0].Passage.ID != "handbook:auth.md#1" {
		t.Fatalf("semantic: %+v %v", sem, err)
	}
	p, ok, err := store.Get(ctx, "handbook:auth.md#1")
	if err != nil || !ok || p.Heading != "Tokens" || p.Embedding != nil {
		t.Fatalf("get: %+v %v %v", p, ok, err)
	}
	// Partial replace keeps other sources; a model change needs a full rebuild.
	if err := store.Replace(ctx, map[string][]passage.Passage{"billing": {}}, meta, false); err != nil {
		t.Fatal(err)
	}
	stats, _ := store.Stats(ctx)
	if !reflect.DeepEqual(stats, map[string]int{"handbook": 3}) {
		t.Fatalf("stats: %v", stats)
	}
	err = store.Replace(ctx, map[string][]passage.Passage{"billing": {}}, Meta{}, false)
	if err == nil || !strings.Contains(err.Error(), "without --source") {
		t.Fatalf("model mismatch not detected: %v", err)
	}
	if err := store.Replace(ctx, map[string][]passage.Passage{"handbook": {}}, Meta{}, true); err != nil {
		t.Fatal(err)
	}
	if m, _ := store.Meta(ctx); m.Model != "" {
		t.Fatalf("full rebuild should change the model: %+v", m)
	}
}

// TestPgvectorStore runs against a real PostgreSQL with pgvector when
// NEXORA_TEST_DATABASE_URL is set, for example:
// NEXORA_TEST_DATABASE_URL="host=/tmp port=5432 user=postgres dbname=nexora_test sslmode=disable"
func TestPgvectorStore(t *testing.T) {
	dsn := os.Getenv("NEXORA_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("NEXORA_TEST_DATABASE_URL not set")
	}
	ctx := context.Background()
	store, err := OpenPgvector(dsn, "nexora_test_passages")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	meta := Meta{Model: "ollama:test", Dimensions: 3}
	if err := store.Replace(ctx, fixture(), meta, true); err != nil {
		t.Fatal(err)
	}
	got, err := store.Keyword(ctx, Query{Terms: Terms("partial refund amount"), Project: "commerce-api", Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) == 0 || got[0].Passage.ID != "handbook:refunds.md#2" {
		t.Fatalf("keyword: %+v", got)
	}
	for _, c := range got {
		if c.Passage.SourceID == "billing" {
			t.Fatal("billing passage leaked into another project")
		}
	}
	sem, err := store.Semantic(ctx, Query{Vector: []float32{0, 1, 0}, Project: "commerce-api", Limit: 1})
	if err != nil || len(sem) != 1 || sem[0].Passage.ID != "handbook:auth.md#1" {
		t.Fatalf("semantic: %+v %v", sem, err)
	}
	if m, err := store.Meta(ctx); err != nil || m != meta {
		t.Fatalf("meta: %+v %v", m, err)
	}
	if err := store.Replace(ctx, map[string][]passage.Passage{"billing": {}}, meta, false); err != nil {
		t.Fatal(err)
	}
	stats, _ := store.Stats(ctx)
	if !reflect.DeepEqual(stats, map[string]int{"handbook": 3}) {
		t.Fatalf("stats: %v", stats)
	}
	if err := store.Replace(ctx, map[string][]passage.Passage{"billing": {}}, Meta{}, false); err == nil {
		t.Fatal("model mismatch not detected")
	}
	p, ok, err := store.Get(ctx, "handbook:refunds.md#1")
	if err != nil || !ok || p.Heading != "Refund window" || !reflect.DeepEqual(p.Projects, []string{}) {
		t.Fatalf("get: %+v %v %v", p, ok, err)
	}
}
