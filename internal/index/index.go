// Package index stores documentation passages and returns keyword and semantic
// candidates. Ranking fusion lives in package retrieval so both stores behave
// the same way.
package index

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"unicode"

	"github.com/acierto/nexora/internal/config"
	"github.com/acierto/nexora/internal/passage"
)

// ErrNoIndex reports that nothing has been indexed yet.
var ErrNoIndex = errors.New("no documentation index found; run nexora index")

// Meta describes how an index was built.
type Meta struct {
	// Model is "<provider>:<model>" or empty when built without embeddings.
	Model      string `json:"model"`
	Dimensions int    `json:"dimensions"`
}

// Query asks a store for candidates. Terms are already tokenized.
type Query struct {
	Terms   []string
	Vector  []float32
	Project string
	Limit   int
}

// Candidate is one passage found by one method.
type Candidate struct {
	Passage passage.Passage
	Score   float64
}

// Store persists passages per source.
type Store interface {
	// Replace swaps the passages of the given sources. With full set, every
	// other source is removed too, which allows changing the embedding model.
	Replace(ctx context.Context, sources map[string][]passage.Passage, meta Meta, full bool) error
	Keyword(ctx context.Context, q Query) ([]Candidate, error)
	Semantic(ctx context.Context, q Query) ([]Candidate, error)
	Get(ctx context.Context, id string) (passage.Passage, bool, error)
	Meta(ctx context.Context) (Meta, error)
	// Stats returns passage counts per source.
	Stats(ctx context.Context) (map[string]int, error)
	Close() error
}

// Open returns the configured store. It does not create anything.
func Open(cfg config.Config, getenv func(string) string) (Store, error) {
	switch cfg.Index.Store {
	case config.StorePgvector:
		dsn := getenv(cfg.Index.DSNEnv)
		if strings.TrimSpace(dsn) == "" {
			return nil, fmt.Errorf("index: environment variable %s is not set", cfg.Index.DSNEnv)
		}
		return OpenPgvector(dsn, cfg.Index.Table)
	default:
		return &Local{File: cfg.Index.File}, nil
	}
}

func checkMeta(existing, next Meta, hasOtherSources bool) error {
	if hasOtherSources && (existing.Model != next.Model || existing.Dimensions != next.Dimensions && existing.Dimensions != 0 && next.Dimensions != 0) {
		return fmt.Errorf("index was built with embeddings %q but the configuration uses %q; run nexora index without --source to rebuild every source", describe(existing.Model), describe(next.Model))
	}
	return nil
}

func describe(model string) string {
	if model == "" {
		return "none"
	}
	return model
}

var stopwords = map[string]bool{}

func init() {
	for _, w := range strings.Fields(`a an and are as at be but by can do does for from has have how i if in into is it its
		me my no not of on or our so than that the their them then there these they this to us was we were what when where
		which who why will with you your should would could about after before also any all been being more most other some such only own same too very just`) {
		stopwords[w] = true
	}
}

// Tokenize lowercases text, splits on non-alphanumerics, drops stopwords and
// one-letter tokens, and strips a plural "s" so "refunds" matches "refund".
func Tokenize(text string) []string {
	out := []string{}
	for _, raw := range strings.FieldsFunc(strings.ToLower(text), func(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsDigit(r) }) {
		if len([]rune(raw)) < 2 || stopwords[raw] {
			continue
		}
		out = append(out, stem(raw))
	}
	return out
}

// Terms returns the unique tokens of a query in first-seen order.
func Terms(text string) []string {
	seen := map[string]bool{}
	out := []string{}
	for _, t := range Tokenize(text) {
		if !seen[t] {
			seen[t] = true
			out = append(out, t)
		}
	}
	return out
}

// stem is a deliberately small suffix stripper, applied identically to
// documents and queries: refunds/refunded -> refund, rotating/rotate -> rotat.
func stem(w string) string {
	n := len(w)
	switch {
	case n > 4 && strings.HasSuffix(w, "ies"):
		w = w[:n-3] + "y"
	case n > 3 && strings.HasSuffix(w, "s") && !strings.HasSuffix(w, "ss") && !strings.HasSuffix(w, "us") && !strings.HasSuffix(w, "is"):
		w = w[:n-1]
	}
	n = len(w)
	switch {
	case n > 5 && strings.HasSuffix(w, "ing"):
		w = w[:n-3]
	case n > 4 && strings.HasSuffix(w, "ed"):
		w = w[:n-2]
	}
	if len(w) > 3 && strings.HasSuffix(w, "e") {
		w = w[:len(w)-1]
	}
	return w
}

// MatchedTerms returns query terms that occur in the passage, in query order.
func MatchedTerms(p passage.Passage, terms []string) []string {
	present := map[string]bool{}
	for _, t := range Tokenize(p.SearchText()) {
		present[t] = true
	}
	out := []string{}
	for _, t := range terms {
		if present[t] {
			out = append(out, t)
		}
	}
	return out
}
