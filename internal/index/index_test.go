package index

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/acierto/nexora/internal/config"
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

// TestExternalTable reads a table Nexora did not create, as a team's existing
// embedding pipeline would leave it: a schema, a uuid key, nullable columns,
// and vectors of a fixed size.
func TestExternalTable(t *testing.T) {
	dsn := os.Getenv("NEXORA_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("NEXORA_TEST_DATABASE_URL not set")
	}
	ctx := context.Background()
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	for _, s := range []string{
		`CREATE EXTENSION IF NOT EXISTS vector`,
		`DROP SCHEMA IF EXISTS nexora_ext CASCADE`,
		`CREATE SCHEMA nexora_ext`,
		`CREATE TABLE nexora_ext."Chunks" (chunk_id int PRIMARY KEY, body text, page_title text, url text, page_id text, emb vector(3))`,
		`INSERT INTO nexora_ext."Chunks" VALUES
			(1, 'Use the pipeline rollback action to redeploy the previous image.', 'Rolling back', 'https://docs.example/deploy#rollback', 'deploy', '[1,0,0]'),
			(2, 'Announce every production deploy in the team channel.', 'Before you deploy', NULL, 'deploy', '[0.6,0.8,0]'),
			(3, 'Refunds go back to the original payment method.', NULL, NULL, NULL, '[0,0,1]'),
			(4, 'Draft without a vector yet.', 'Draft', NULL, NULL, NULL),
			(5, '', 'Empty', NULL, NULL, '[0,1,0]')`,
	} {
		if _, err := db.ExecContext(ctx, s); err != nil {
			t.Fatal(err)
		}
	}
	source := config.Source{ID: "team", Table: "nexora_ext.Chunks", Distance: config.DistanceCosine, Projects: []string{"commerce-api"},
		Columns: config.Columns{ID: "chunk_id", Text: "body", Vector: "emb", Title: "page_title", Location: "url", Document: "page_id"}}
	table, err := OpenExternal(source, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer table.Close()
	rows, err := table.Rows(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 4 {
		t.Fatalf("rows with text: %d", len(rows))
	}
	first := rows[0]
	if first.ID != "team:1" || first.SourceType != config.SourcePgvector || first.Title != "Rolling back" || first.Heading != "Rolling back" ||
		first.Location != "https://docs.example/deploy#rollback" || first.DocumentID != "deploy" || first.Digest == "" || !reflect.DeepEqual(first.Projects, []string{"commerce-api"}) {
		t.Fatalf("row mapping: %+v", first)
	}
	if third := rows[2]; third.Title != "3" || third.Location != "3" || third.Heading != "" {
		t.Fatalf("null columns: %+v", third)
	}
	near, err := table.Nearest(ctx, []float32{0.9, 0.1, 0}, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(near) != 2 || near[0].Passage.ID != "team:1" || near[1].Passage.ID != "team:2" || near[0].Score <= near[1].Score || near[0].Score > 1 {
		t.Fatalf("nearest: %+v", near)
	}
	source.Distance = config.DistanceL2
	l2, _ := OpenExternal(source, dsn)
	defer l2.Close()
	if near, err := l2.Nearest(ctx, []float32{0, 0, 1}, 1); err != nil || near[0].Passage.ID != "team:3" || near[0].Score != 0 {
		t.Fatalf("l2: %+v %v", near, err)
	}
	if _, err := table.Nearest(ctx, []float32{1, 0}, 1); err == nil || !strings.Contains(err.Error(), "model that produced the table") {
		t.Fatalf("dimension mismatch not explained: %v", err)
	}
	source.Columns.Text = "missing_column"
	bad, _ := OpenExternal(source, dsn)
	defer bad.Close()
	if _, err := bad.Rows(ctx); err == nil || !strings.Contains(err.Error(), "source team (nexora_ext.Chunks)") {
		t.Fatalf("missing column not attributed: %v", err)
	}
}
