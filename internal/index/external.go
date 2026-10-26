package index

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/lib/pq"

	"github.com/acierto/nexora/internal/config"
	"github.com/acierto/nexora/internal/passage"
)

// VectorTable is an existing table of embedded passages that Nexora reads but
// never writes. Rows supplies text for the keyword index; Nearest answers
// semantic queries in place with the table's own vectors.
type VectorTable interface {
	Rows(ctx context.Context) ([]passage.Passage, error)
	Nearest(ctx context.Context, vector []float32, limit int) ([]Candidate, error)
	Close() error
}

// External reads a pgvector table described by a pgvector source.
type External struct {
	db     *sql.DB
	source config.Source
}

// OpenExternal connects lazily. Identifiers were validated by config and are
// quoted here as well.
func OpenExternal(source config.Source, dsn string) (*External, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("source %s: %w", source.ID, err)
	}
	return &External{db: db, source: source}, nil
}

// OpenVectorTable opens the table of a pgvector source using its dsn_env.
func OpenVectorTable(source config.Source, getenv func(string) string) (VectorTable, error) {
	dsn := getenv(source.DSNEnv)
	if strings.TrimSpace(dsn) == "" {
		return nil, fmt.Errorf("source %s: environment variable %s is not set", source.ID, source.DSNEnv)
	}
	return OpenExternal(source, dsn)
}

func (e *External) Close() error { return e.db.Close() }

func (e *External) table() string {
	parts := strings.Split(e.source.Table, ".")
	for i, p := range parts {
		parts[i] = pq.QuoteIdentifier(p)
	}
	return strings.Join(parts, ".")
}

func textColumn(name string) string {
	if name == "" {
		return "''"
	}
	return "COALESCE(" + pq.QuoteIdentifier(name) + "::text, '')"
}

func (e *External) selectList() string {
	c := e.source.Columns
	return strings.Join([]string{textColumn(c.ID), textColumn(c.Text), textColumn(c.Title), textColumn(c.Location), textColumn(c.Document)}, ", ")
}

// operator returns the pgvector distance operator and how to turn a distance
// into a score where higher is better.
func (e *External) operator() (op, score string) {
	switch e.source.Distance {
	case config.DistanceL2:
		return "<->", "-(%s)"
	case config.DistanceInnerProduct:
		return "<#>", "-(%s)"
	default:
		return "<=>", "1 - (%s)"
	}
}

func (e *External) wrap(err error) error {
	return fmt.Errorf("source %s (%s): %w", e.source.ID, e.source.Table, err)
}

func (e *External) scan(rows *sql.Rows, extra ...any) (passage.Passage, error) {
	var id, text, title, location, document string
	if err := rows.Scan(append([]any{&id, &text, &title, &location, &document}, extra...)...); err != nil {
		return passage.Passage{}, err
	}
	return e.passage(id, text, title, location, document), nil
}

// passage maps a row without re-chunking, so passage IDs stay tied to rows.
func (e *External) passage(id, text, title, location, document string) passage.Passage {
	if document == "" {
		document = id
	}
	if location == "" {
		location = document
	}
	heading := title
	if title == "" {
		title = document
	}
	sum := sha256.Sum256([]byte(text))
	return passage.Passage{
		ID:         e.source.ID + ":" + id,
		SourceID:   e.source.ID,
		SourceType: config.SourcePgvector,
		DocumentID: document,
		Title:      title,
		Heading:    heading,
		Text:       text,
		Location:   location,
		Digest:     hex.EncodeToString(sum[:]),
		Projects:   append([]string{}, e.source.Projects...),
	}
}

// Rows returns every row with text, in ID order.
func (e *External) Rows(ctx context.Context) ([]passage.Passage, error) {
	idCol := pq.QuoteIdentifier(e.source.Columns.ID)
	rows, err := e.db.QueryContext(ctx, `SELECT `+e.selectList()+` FROM `+e.table()+`
		WHERE `+idCol+` IS NOT NULL ORDER BY `+idCol+`::text`)
	if err != nil {
		return nil, e.wrap(err)
	}
	defer rows.Close()
	out := []passage.Passage{}
	for rows.Next() {
		p, err := e.scan(rows)
		if err != nil {
			return nil, e.wrap(err)
		}
		if strings.TrimSpace(p.Text) != "" {
			out = append(out, p)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, e.wrap(err)
	}
	return out, nil
}

// Nearest returns the rows closest to the query vector.
func (e *External) Nearest(ctx context.Context, vector []float32, n int) ([]Candidate, error) {
	if len(vector) == 0 {
		return []Candidate{}, nil
	}
	vec := pq.QuoteIdentifier(e.source.Columns.Vector)
	op, score := e.operator()
	distance := vec + " " + op + " $1::vector"
	rows, err := e.db.QueryContext(ctx, `SELECT `+e.selectList()+`, (`+fmt.Sprintf(score, distance)+`)::float8 AS score
		FROM `+e.table()+`
		WHERE `+vec+` IS NOT NULL
		ORDER BY `+distance+`, `+pq.QuoteIdentifier(e.source.Columns.ID)+`::text LIMIT $2`, vectorLiteral(vector), limit(n))
	if err != nil {
		if strings.Contains(err.Error(), "different vector dimensions") {
			return nil, e.wrap(fmt.Errorf("the query embedding has %d dimensions but the stored vectors do not; embeddings must use the model that produced the table: %w", len(vector), err))
		}
		return nil, e.wrap(err)
	}
	defer rows.Close()
	out := []Candidate{}
	for rows.Next() {
		var s float64
		p, err := e.scan(rows, &s)
		if err != nil {
			return nil, e.wrap(err)
		}
		out = append(out, Candidate{Passage: p, Score: s})
	}
	if err := rows.Err(); err != nil {
		return nil, e.wrap(err)
	}
	return out, nil
}
