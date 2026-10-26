// Package retrieval finds documentation passages for a task and explains why.
// It never touches deterministic knowledge selection.
package retrieval

import (
	"context"
	"fmt"
	"math"
	"net/url"
	"sort"
	"strings"

	"github.com/acierto/nexora/internal/embed"
	"github.com/acierto/nexora/internal/index"
	"github.com/acierto/nexora/internal/passage"
)

// Reason explains how one method found a hit.
type Reason struct {
	Method string `json:"method"`
	// Source names the pgvector source whose own vectors found the hit.
	Source string   `json:"source,omitempty"`
	Rank   int      `json:"rank"`
	Score  float64  `json:"score"`
	Terms  []string `json:"terms"`
}

// Hit is one retrieved passage with its fused score.
type Hit struct {
	Passage    passage.Passage `json:"passage"`
	Score      float64         `json:"score"`
	Reasons    []Reason        `json:"reasons"`
	Duplicates []Duplicate     `json:"duplicates,omitempty"`
}

// Duplicate preserves attribution when identical linked passages are returned
// by more than one source. Different bodies or document digests remain separate.
type Duplicate struct {
	ID       string `json:"id"`
	SourceID string `json:"source_id"`
	Title    string `json:"title"`
	Location string `json:"location"`
	Digest   string `json:"digest"`
}

// Request describes a search.
type Request struct {
	Text    string
	Project string
	Limit   int
	// Vectors are existing tables searched by meaning with their own vectors.
	Vectors []VectorSource
}

// VectorSource is a pgvector source: its table and the embedder that produced
// the stored vectors, used here only for the query.
type VectorSource struct {
	ID       string
	Table    index.VectorTable
	Embedder embed.Embedder
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
	lists := map[string][]index.Candidate{"keyword": keyword, "semantic": semantic}
	if req.Text != "" {
		// Sources that share a model share one query embedding.
		cache := map[string][]float32{}
		if q.Vector != nil {
			cache[embedder.Model()] = q.Vector
		}
		for _, src := range req.Vectors {
			vector, ok := cache[src.Embedder.Model()]
			if !ok {
				vectors, err := src.Embedder.Embed(ctx, []string{req.Text})
				if err != nil {
					return nil, fmt.Errorf("source %s: %w", src.ID, err)
				}
				vector = vectors[0]
				cache[src.Embedder.Model()] = vector
			}
			found, err := src.Table.Nearest(ctx, vector, pool)
			if err != nil {
				return nil, err
			}
			allowed := []index.Candidate{}
			for _, c := range found {
				if c.Passage.AllowedFor(req.Project) {
					allowed = append(allowed, c)
				}
			}
			lists["semantic/"+src.ID] = allowed
		}
	}
	return Fuse(lists, terms, req.Limit), nil
}

// Fuse combines ranked candidate lists with reciprocal rank fusion. Ties break
// by passage ID so results are reproducible. A list named "semantic/<source>"
// is reported as method semantic with that source.
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
			// An existing vector table is read live. Prefer its current text and
			// citation over the keyword snapshot for the same passage ID.
			if strings.HasPrefix(method, "semantic/") {
				h.Passage = c.Passage.Clone()
			}
			h.Score += 1 / float64(rrfK+rank+1)
			name, source, _ := strings.Cut(method, "/")
			reason := Reason{Method: name, Source: source, Rank: rank + 1, Score: round(c.Score), Terms: []string{}}
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
	// Collapse only exact linked copies. A shared title, URL, or similar text
	// alone must never erase a distinct section or a conflicting version.
	unique := make([]Hit, 0, len(hits))
	linked := map[string]int{}
	for _, h := range hits {
		p := h.Passage
		u, err := url.Parse(p.Location)
		key := ""
		if err == nil && (u.Scheme == "http" || u.Scheme == "https") && u.Host != "" && strings.TrimSpace(p.Text) != "" {
			key = p.Location + "\x00" + p.Heading + "\x00" + p.Digest + "\x00" + strings.TrimSpace(p.Text)
		}
		if idx, ok := linked[key]; key != "" && ok {
			unique[idx].Duplicates = append(unique[idx].Duplicates, Duplicate{ID: p.ID, SourceID: p.SourceID, Title: p.Title, Location: p.Location, Digest: p.Digest})
			continue
		}
		if key != "" {
			linked[key] = len(unique)
		}
		unique = append(unique, h)
	}
	hits = unique
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
