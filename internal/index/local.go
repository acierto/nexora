package index

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"

	"github.com/acierto/nexora/internal/passage"
)

// Local keeps the index in one JSON file. It needs no database or network.
type Local struct {
	File string
}

type localRecord struct {
	passage.Passage
	Embedding []float32 `json:"embedding,omitempty"`
}

type localFile struct {
	Format   int           `json:"format"`
	Meta     Meta          `json:"meta"`
	Passages []localRecord `json:"passages"`
}

func (l *Local) load() (localFile, error) {
	data, err := os.ReadFile(l.File)
	if errors.Is(err, os.ErrNotExist) {
		return localFile{}, ErrNoIndex
	}
	if err != nil {
		return localFile{}, fmt.Errorf("index %s: %w", l.File, err)
	}
	var f localFile
	if err := json.Unmarshal(data, &f); err != nil {
		return localFile{}, fmt.Errorf("index %s: %w; delete it and run nexora index", l.File, err)
	}
	if f.Format != 1 {
		return localFile{}, fmt.Errorf("index %s: unsupported format %d; run nexora index", l.File, f.Format)
	}
	return f, nil
}

func (l *Local) Replace(ctx context.Context, sources map[string][]passage.Passage, meta Meta, full bool) error {
	current := localFile{Format: 1}
	if !full {
		loaded, err := l.load()
		if err != nil && !errors.Is(err, ErrNoIndex) {
			return err
		}
		if err == nil {
			current = loaded
		}
	}
	kept := []localRecord{}
	for _, r := range current.Passages {
		if _, replaced := sources[r.SourceID]; !replaced {
			kept = append(kept, r)
		}
	}
	if err := checkMeta(current.Meta, meta, len(kept) > 0); err != nil {
		return err
	}
	for _, list := range sources {
		for _, p := range list {
			kept = append(kept, localRecord{Passage: p.Clone(), Embedding: p.Embedding})
		}
	}
	sort.Slice(kept, func(i, j int) bool { return kept[i].ID < kept[j].ID })
	if err := ctx.Err(); err != nil {
		return err
	}
	data, err := json.MarshalIndent(localFile{Format: 1, Meta: meta, Passages: kept}, "", " ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(l.File), 0o755); err != nil {
		return fmt.Errorf("index: %w", err)
	}
	tmp, err := os.CreateTemp(filepath.Dir(l.File), ".index-*.json")
	if err != nil {
		return fmt.Errorf("index: %w", err)
	}
	_, writeErr := tmp.Write(append(data, '\n'))
	closeErr := tmp.Close()
	if err := errors.Join(writeErr, closeErr); err != nil {
		os.Remove(tmp.Name())
		return fmt.Errorf("index: %w", err)
	}
	if err := os.Chmod(tmp.Name(), 0o644); err != nil {
		os.Remove(tmp.Name())
		return fmt.Errorf("index: %w", err)
	}
	if err := os.Rename(tmp.Name(), l.File); err != nil {
		os.Remove(tmp.Name())
		return fmt.Errorf("index: %w", err)
	}
	return nil
}

func (l *Local) passages(project string) ([]localRecord, error) {
	f, err := l.load()
	if err != nil {
		return nil, err
	}
	out := []localRecord{}
	for _, r := range f.Passages {
		if r.AllowedFor(project) {
			out = append(out, r)
		}
	}
	return out, nil
}

// Keyword ranks passages with BM25 (k1 = 1.2, b = 0.75).
func (l *Local) Keyword(ctx context.Context, q Query) ([]Candidate, error) {
	records, err := l.passages(q.Project)
	if err != nil || len(q.Terms) == 0 {
		return []Candidate{}, err
	}
	const k1, b = 1.2, 0.75
	docs := make([]map[string]int, len(records))
	lengths := make([]int, len(records))
	df := map[string]int{}
	total := 0
	for i, r := range records {
		tokens := Tokenize(r.SearchText())
		lengths[i] = len(tokens)
		total += len(tokens)
		tf := map[string]int{}
		for _, t := range tokens {
			tf[t]++
		}
		docs[i] = tf
		for t := range tf {
			df[t]++
		}
	}
	if len(records) == 0 {
		return []Candidate{}, nil
	}
	avg := float64(total) / float64(len(records))
	n := float64(len(records))
	out := []Candidate{}
	for i, r := range records {
		score := 0.0
		for _, t := range q.Terms {
			f := float64(docs[i][t])
			if f == 0 {
				continue
			}
			idf := math.Log(1 + (n-float64(df[t])+0.5)/(float64(df[t])+0.5))
			score += idf * f * (k1 + 1) / (f + k1*(1-b+b*float64(lengths[i])/avg))
		}
		if score > 0 {
			out = append(out, Candidate{Passage: r.Passage.Clone(), Score: score})
		}
	}
	return top(out, q.Limit), ctx.Err()
}

// Semantic ranks passages by cosine similarity to the query vector.
func (l *Local) Semantic(ctx context.Context, q Query) ([]Candidate, error) {
	if len(q.Vector) == 0 {
		return []Candidate{}, nil
	}
	records, err := l.passages(q.Project)
	if err != nil {
		return nil, err
	}
	out := []Candidate{}
	for _, r := range records {
		if len(r.Embedding) != len(q.Vector) {
			continue
		}
		out = append(out, Candidate{Passage: r.Passage.Clone(), Score: cosine(r.Embedding, q.Vector)})
	}
	return top(out, q.Limit), ctx.Err()
}

func (l *Local) Get(ctx context.Context, id string) (passage.Passage, bool, error) {
	f, err := l.load()
	if err != nil {
		return passage.Passage{}, false, err
	}
	for _, r := range f.Passages {
		if r.ID == id {
			return r.Passage.Clone(), true, nil
		}
	}
	return passage.Passage{}, false, nil
}

func (l *Local) Meta(ctx context.Context) (Meta, error) {
	f, err := l.load()
	return f.Meta, err
}

func (l *Local) Stats(ctx context.Context) (map[string]int, error) {
	f, err := l.load()
	if err != nil {
		return nil, err
	}
	out := map[string]int{}
	for _, r := range f.Passages {
		out[r.SourceID]++
	}
	return out, nil
}

func (l *Local) Close() error { return nil }

func cosine(a, b []float32) float64 {
	var dot, na, nb float64
	for i := range a {
		dot += float64(a[i]) * float64(b[i])
		na += float64(a[i]) * float64(a[i])
		nb += float64(b[i]) * float64(b[i])
	}
	if na == 0 || nb == 0 {
		return 0
	}
	return dot / (math.Sqrt(na) * math.Sqrt(nb))
}

// top sorts by descending score, then passage ID, and truncates.
func top(c []Candidate, limit int) []Candidate {
	sort.Slice(c, func(i, j int) bool {
		if c[i].Score != c[j].Score {
			return c[i].Score > c[j].Score
		}
		return c[i].Passage.ID < c[j].Passage.ID
	})
	if limit > 0 && len(c) > limit {
		c = c[:limit]
	}
	return c
}
