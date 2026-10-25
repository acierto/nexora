package retrieval

import (
	"context"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/acierto/nexora/internal/index"
	"github.com/acierto/nexora/internal/passage"
)

func cand(id, text string, score float64) index.Candidate {
	return index.Candidate{Passage: passage.Passage{ID: id, Text: text, Projects: []string{}}, Score: score}
}

func TestFuse(t *testing.T) {
	terms := index.Terms("refund window")
	keyword := []index.Candidate{cand("a", "refund window rules", 3), cand("b", "refund only", 2)}
	semantic := []index.Candidate{cand("c", "money back period", 0.9), cand("a", "refund window rules", 0.8)}
	hits := Fuse(map[string][]index.Candidate{"semantic": semantic, "keyword": keyword}, terms, 10)
	ids := []string{}
	for _, h := range hits {
		ids = append(ids, h.Passage.ID)
	}
	if !reflect.DeepEqual(ids, []string{"a", "c", "b"}) {
		t.Fatalf("order: %v", ids)
	}
	a := hits[0]
	if len(a.Reasons) != 2 || a.Reasons[0].Method != "keyword" || a.Reasons[0].Rank != 1 || !reflect.DeepEqual(a.Reasons[0].Terms, []string{"refund", "window"}) || a.Reasons[1].Method != "semantic" || a.Reasons[1].Rank != 2 {
		t.Fatalf("reasons: %+v", a.Reasons)
	}
	again := Fuse(map[string][]index.Candidate{"keyword": keyword, "semantic": semantic}, terms, 10)
	if !reflect.DeepEqual(hits, again) {
		t.Fatal("fusion is not deterministic")
	}
	if got := Fuse(map[string][]index.Candidate{"keyword": keyword}, terms, 1); len(got) != 1 {
		t.Fatal("limit ignored")
	}
}

type fakeEmbedder struct{ model string }

func (f fakeEmbedder) Model() string { return f.model }
func (f fakeEmbedder) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	out := make([][]float32, len(texts))
	for i := range texts {
		out[i] = []float32{1, 0}
	}
	return out, nil
}

func TestSearchModelChecks(t *testing.T) {
	ctx := context.Background()
	store := &index.Local{File: filepath.Join(t.TempDir(), "index.json")}
	p := passage.Passage{ID: "s:a#1", SourceID: "s", Text: "refund window", Projects: []string{}, Embedding: []float32{1, 0}}
	if err := store.Replace(ctx, map[string][]passage.Passage{"s": {p}}, index.Meta{Model: "ollama:m", Dimensions: 2}, true); err != nil {
		t.Fatal(err)
	}
	hits, err := Search(ctx, store, fakeEmbedder{"ollama:m"}, Request{Text: "refund", Project: "x", Limit: 3})
	if err != nil || len(hits) != 1 || len(hits[0].Reasons) != 2 {
		t.Fatalf("hybrid search: %+v %v", hits, err)
	}
	if _, err := Search(ctx, store, fakeEmbedder{"openai:other"}, Request{Text: "refund"}); err == nil || !strings.Contains(err.Error(), "run nexora index") {
		t.Fatalf("model mismatch not reported: %v", err)
	}
	if _, err := Search(ctx, store, nil, Request{Text: "refund"}); err == nil {
		t.Fatal("index with embeddings searched without an embedder")
	}
}
