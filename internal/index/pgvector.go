package index

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/lib/pq"

	"github.com/acierto/nexora/internal/passage"
)

// Pgvector stores passages in PostgreSQL with the pgvector extension.
type Pgvector struct {
	db    *sql.DB
	table string
}

// OpenPgvector connects lazily; table names are validated by config.
func OpenPgvector(dsn, table string) (*Pgvector, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("pgvector: %w", err)
	}
	return &Pgvector{db: db, table: table}, nil
}

func (p *Pgvector) Close() error { return p.db.Close() }

func (p *Pgvector) ident(suffix string) string { return pq.QuoteIdentifier(p.table + suffix) }

func (p *Pgvector) exists(ctx context.Context) (bool, error) {
	var name sql.NullString
	err := p.db.QueryRowContext(ctx, `SELECT to_regclass($1)::text`, p.table).Scan(&name)
	if err != nil {
		return false, fmt.Errorf("pgvector: %w", err)
	}
	return name.Valid, nil
}

func (p *Pgvector) Meta(ctx context.Context) (Meta, error) {
	ok, err := p.exists(ctx)
	if err != nil {
		return Meta{}, err
	}
	if !ok {
		return Meta{}, ErrNoIndex
	}
	var m Meta
	err = p.db.QueryRowContext(ctx, `SELECT model, dimensions FROM `+p.ident("_meta")+` WHERE id = 1`).Scan(&m.Model, &m.Dimensions)
	if errors.Is(err, sql.ErrNoRows) {
		return Meta{}, ErrNoIndex
	}
	if err != nil {
		return Meta{}, fmt.Errorf("pgvector: %w", err)
	}
	return m, nil
}

func (p *Pgvector) create(ctx context.Context, tx *sql.Tx, meta Meta) error {
	vectorType := "vector"
	if meta.Dimensions > 0 {
		vectorType = fmt.Sprintf("vector(%d)", meta.Dimensions)
	}
	statements := []string{
		`CREATE EXTENSION IF NOT EXISTS vector`,
		`DROP TABLE IF EXISTS ` + p.ident("") + `, ` + p.ident("_meta"),
		`CREATE TABLE ` + p.ident("") + ` (
			id text PRIMARY KEY,
			source_id text NOT NULL,
			source_type text NOT NULL,
			document_id text NOT NULL,
			title text NOT NULL,
			heading text NOT NULL,
			content text NOT NULL,
			location text NOT NULL,
			digest text NOT NULL,
			projects text[] NOT NULL DEFAULT '{}',
			embedding ` + vectorType + `,
			tsv tsvector GENERATED ALWAYS AS (to_tsvector('english', title || ' ' || heading || ' ' || content)) STORED
		)`,
		`CREATE INDEX ON ` + p.ident("") + ` USING gin (tsv)`,
		`CREATE INDEX ON ` + p.ident("") + ` (source_id)`,
		`CREATE TABLE ` + p.ident("_meta") + ` (id int PRIMARY KEY, model text NOT NULL, dimensions int NOT NULL)`,
	}
	if meta.Dimensions > 0 {
		statements = append(statements, `CREATE INDEX ON `+p.ident("")+` USING hnsw (embedding vector_cosine_ops)`)
	}
	for _, s := range statements {
		if _, err := tx.ExecContext(ctx, s); err != nil {
			return fmt.Errorf("pgvector: %w", err)
		}
	}
	return nil
}

func (p *Pgvector) Replace(ctx context.Context, sources map[string][]passage.Passage, meta Meta, full bool) error {
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("pgvector: %w", err)
	}
	defer tx.Rollback()
	ok, err := p.exists(ctx)
	if err != nil {
		return err
	}
	if full || !ok {
		if err := p.create(ctx, tx, meta); err != nil {
			return err
		}
	} else {
		var existing Meta
		if err := tx.QueryRowContext(ctx, `SELECT model, dimensions FROM `+p.ident("_meta")+` WHERE id = 1`).Scan(&existing.Model, &existing.Dimensions); err != nil {
			return fmt.Errorf("pgvector: %w", err)
		}
		ids := []string{}
		for id := range sources {
			ids = append(ids, id)
		}
		var others int
		if err := tx.QueryRowContext(ctx, `SELECT count(*) FROM `+p.ident("")+` WHERE NOT (source_id = ANY($1))`, pq.Array(ids)).Scan(&others); err != nil {
			return fmt.Errorf("pgvector: %w", err)
		}
		if err := checkMeta(existing, meta, others > 0); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `DELETE FROM `+p.ident("")+` WHERE source_id = ANY($1)`, pq.Array(ids)); err != nil {
			return fmt.Errorf("pgvector: %w", err)
		}
	}
	insert, err := tx.PrepareContext(ctx, `INSERT INTO `+p.ident("")+` (id, source_id, source_type, document_id, title, heading, content, location, digest, projects, embedding)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11::vector)`)
	if err != nil {
		return fmt.Errorf("pgvector: %w", err)
	}
	defer insert.Close()
	for _, list := range sources {
		for _, ps := range list {
			var vec any
			if len(ps.Embedding) > 0 {
				vec = vectorLiteral(ps.Embedding)
			}
			if _, err := insert.ExecContext(ctx, ps.ID, ps.SourceID, ps.SourceType, ps.DocumentID, ps.Title, ps.Heading, ps.Text, ps.Location, ps.Digest, pq.Array(nonNil(ps.Projects)), vec); err != nil {
				return fmt.Errorf("pgvector: insert %s: %w", ps.ID, err)
			}
		}
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO `+p.ident("_meta")+` (id, model, dimensions) VALUES (1, $1, $2)
		ON CONFLICT (id) DO UPDATE SET model = EXCLUDED.model, dimensions = EXCLUDED.dimensions`, meta.Model, meta.Dimensions); err != nil {
		return fmt.Errorf("pgvector: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("pgvector: %w", err)
	}
	return nil
}

const columns = `id, source_id, source_type, document_id, title, heading, content, location, digest, projects`

func scanPassage(rows interface{ Scan(...any) error }, extra ...any) (passage.Passage, error) {
	var ps passage.Passage
	var projects pq.StringArray
	dest := append([]any{&ps.ID, &ps.SourceID, &ps.SourceType, &ps.DocumentID, &ps.Title, &ps.Heading, &ps.Text, &ps.Location, &ps.Digest, &projects}, extra...)
	if err := rows.Scan(dest...); err != nil {
		return passage.Passage{}, err
	}
	ps.Projects = append([]string{}, projects...)
	return ps, nil
}

func (p *Pgvector) query(ctx context.Context, sqlText string, args ...any) ([]Candidate, error) {
	ok, err := p.exists(ctx)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrNoIndex
	}
	rows, err := p.db.QueryContext(ctx, sqlText, args...)
	if err != nil {
		return nil, fmt.Errorf("pgvector: %w", err)
	}
	defer rows.Close()
	out := []Candidate{}
	for rows.Next() {
		var score float64
		ps, err := scanPassage(rows, &score)
		if err != nil {
			return nil, fmt.Errorf("pgvector: %w", err)
		}
		out = append(out, Candidate{Passage: ps, Score: score})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("pgvector: %w", err)
	}
	return top(out, 0), nil
}

const projectFilter = `(cardinality(projects) = 0 OR $2 = ANY(projects))`

// Keyword uses PostgreSQL full-text search over the generated tsvector.
func (p *Pgvector) Keyword(ctx context.Context, q Query) ([]Candidate, error) {
	if len(q.Terms) == 0 {
		return []Candidate{}, nil
	}
	// Terms come from Tokenize and contain only letters and digits.
	tsquery := strings.Join(q.Terms, " | ")
	return p.query(ctx, `SELECT `+columns+`, ts_rank_cd(tsv, query, 32)::float8 AS score
		FROM `+p.ident("")+`, to_tsquery('english', $1) AS query
		WHERE tsv @@ query AND `+projectFilter+`
		ORDER BY score DESC, id LIMIT $3`, tsquery, q.Project, limit(q.Limit))
}

// Semantic uses cosine distance through pgvector.
func (p *Pgvector) Semantic(ctx context.Context, q Query) ([]Candidate, error) {
	if len(q.Vector) == 0 {
		return []Candidate{}, nil
	}
	return p.query(ctx, `SELECT `+columns+`, (1 - (embedding <=> $1::vector))::float8 AS score
		FROM `+p.ident("")+`
		WHERE embedding IS NOT NULL AND `+projectFilter+`
		ORDER BY embedding <=> $1::vector, id LIMIT $3`, vectorLiteral(q.Vector), q.Project, limit(q.Limit))
}

func (p *Pgvector) Get(ctx context.Context, id string) (passage.Passage, bool, error) {
	ok, err := p.exists(ctx)
	if err != nil {
		return passage.Passage{}, false, err
	}
	if !ok {
		return passage.Passage{}, false, ErrNoIndex
	}
	ps, err := scanPassage(p.db.QueryRowContext(ctx, `SELECT `+columns+` FROM `+p.ident("")+` WHERE id = $1`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return passage.Passage{}, false, nil
	}
	if err != nil {
		return passage.Passage{}, false, fmt.Errorf("pgvector: %w", err)
	}
	return ps, true, nil
}

func (p *Pgvector) Stats(ctx context.Context) (map[string]int, error) {
	if _, err := p.Meta(ctx); err != nil {
		return nil, err
	}
	rows, err := p.db.QueryContext(ctx, `SELECT source_id, count(*) FROM `+p.ident("")+` GROUP BY source_id`)
	if err != nil {
		return nil, fmt.Errorf("pgvector: %w", err)
	}
	defer rows.Close()
	out := map[string]int{}
	for rows.Next() {
		var id string
		var n int
		if err := rows.Scan(&id, &n); err != nil {
			return nil, fmt.Errorf("pgvector: %w", err)
		}
		out[id] = n
	}
	return out, rows.Err()
}

func vectorLiteral(v []float32) string {
	parts := make([]string, len(v))
	for i, f := range v {
		parts[i] = strconv.FormatFloat(float64(f), 'g', -1, 32)
	}
	return "[" + strings.Join(parts, ",") + "]"
}

func nonNil(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}

func limit(n int) int {
	if n <= 0 {
		return 50
	}
	return n
}
