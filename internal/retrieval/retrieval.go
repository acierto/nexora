// Package retrieval finds documentation passages for a task and explains why.
// It never touches deterministic knowledge selection.
package retrieval

import (
	"context"
	"fmt"
	"math"
	"sort"

	"github.com/acierto/nexora/internal/embed"
	"github.com/acierto/nexora/internal/index"
	"github.com/acierto/nexora/internal/passage"
)

// Reason explains how one method found a hit.
type Reason struct {
	Method string   `json:"method"`
	Rank   int      `json:"rank"`
	Score  float64  `json:"score"`
	Terms  []string `json:"terms"`
}

// Hit is one retrieved passage with its fused score.
type Hit struct {
	Passage passage.Passage `json:"passage"`
	Score   float64         `json:"score"`
	Reasons []Reason        `json:"reasons"`
}

// Request describes a search.
type Request struct {
	Text    string
	Project string
	Limit   int
}

// rrfK is the reciprocal rank fusion constant from Cormack et al. (2009).
const rrfK = 60

// Search returns up to Limit fused hits. With an embedder, the stored model must
// match; without one, an index built with embeddings is also a mismatch so that
// semantic recall never silently disappears.
func Search(ctx context.Context, store index.Store, embedder embed.Embedder, req Request) ([]Hit, error) {
	if req.Limit <= 0 {
		req.Limit = 5
	}
	meta, err := store.Meta(ctx)
	if err != nil {
		return nil, err
	}
	model := ""
	if embedder != nil {
		model = embedder.Model()
	}
	if meta.Model != model {
		return nil, fmt.Errorf("index was built with embeddings %q but the configuration uses %q; run nexora index", describe(meta.Model), describe(model))
	}
	terms := index.Terms(req.Text)
	pool := max(req.Limit*4, 20)
	q := index.Query{Terms: terms, Project: req.Project, Limit: pool}
	keyword, err := store.Keyword(ctx, q)
	if err != nil {
		return nil, err
	}
	semantic := []index.Candidate{}
	if embedder != nil && req.Text != "" {
		vectors, err := embedder.Embed(ctx, []string{req.Text})
		if err != nil {
			return nil, err
		}
		q.Vector = vectors[0]
		semantic, err = store.Semantic(ctx, q)
		if err != nil {
			return nil, err
		}
	}
	return Fuse(map[string][]index.Candidate{"keyword": keyword, "semantic": semantic}, terms, req.Limit), nil
}

// Fuse combines ranked candidate lists with reciprocal rank fusion. Ties break
// by passage ID so results are reproducible.
func Fuse(lists map[string][]index.Candidate, terms []string, limit int) []Hit {
	methods := make([]string, 0, len(lists))
	for m := range lists {
		methods = append(methods, m)
	}
	sort.Strings(methods)
	byID := map[string]*Hit{}
	for _, method := range methods {
		for rank, c := range lists[method] {
			h, ok := byID[c.Passage.ID]
			if !ok {
				h = &Hit{Passage: c.Passage.Clone(), Reasons: []Reason{}}
				byID[c.Passage.ID] = h
			}
			h.Score += 1 / float64(rrfK+rank+1)
			reason := Reason{Method: method, Rank: rank + 1, Score: round(c.Score), Terms: []string{}}
			if method == "keyword" {
				reason.Terms = index.MatchedTerms(c.Passage, terms)
			}
			h.Reasons = append(h.Reasons, reason)
		}
	}
	hits := make([]Hit, 0, len(byID))
	for _, h := range byID {
		h.Score = round(h.Score)
		hits = append(hits, *h)
	}
	sort.Slice(hits, func(i, j int) bool {
		if hits[i].Score != hits[j].Score {
			return hits[i].Score > hits[j].Score
		}
		return hits[i].Passage.ID < hits[j].Passage.ID
	})
	if limit > 0 && len(hits) > limit {
		hits = hits[:limit]
	}
	return hits
}

func round(f float64) float64 { return math.Round(f*1e6) / 1e6 }

func describe(model string) string {
	if model == "" {
		return "none"
	}
	return model
}
