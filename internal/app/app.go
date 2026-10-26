// Package app coordinates configuration, knowledge, the documentation index,
// and retrieval for the CLI and the MCP server. Every call reloads from disk,
// so results never come from a stale snapshot.
package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"sort"
	"strings"

	"github.com/acierto/nexora/internal/config"
	"github.com/acierto/nexora/internal/confluence"
	"github.com/acierto/nexora/internal/embed"
	"github.com/acierto/nexora/internal/index"
	"github.com/acierto/nexora/internal/knowledge"
	"github.com/acierto/nexora/internal/markdown"
	"github.com/acierto/nexora/internal/nativesync"
	"github.com/acierto/nexora/internal/passage"
	"github.com/acierto/nexora/internal/registry"
	"github.com/acierto/nexora/internal/resolver"
	"github.com/acierto/nexora/internal/retrieval"
)

type App struct {
	ConfigPath string
	Getenv     func(string) string
	HTTP       *http.Client
	// OpenStore overrides store construction in tests.
	OpenStore func(config.Config) (index.Store, error)
	// OpenTable overrides opening pgvector source tables in tests.
	OpenTable func(config.Source) (index.VectorTable, error)
}

// Loaded is a validated configuration plus its knowledge snapshot.
type Loaded struct {
	Config config.Config
	Items  []knowledge.Item
	byID   *registry.Registry
}

func (a App) Load(ctx context.Context) (Loaded, error) {
	cfg, err := config.Load(a.ConfigPath)
	if err != nil {
		return Loaded{}, err
	}
	items := []knowledge.Item{}
	for _, source := range cfg.KnowledgeSources() {
		loaded, err := markdown.Load(ctx, source)
		if err != nil {
			return Loaded{}, err
		}
		items = append(items, loaded...)
	}
	snapshot, err := registry.New(items)
	if err != nil {
		return Loaded{}, err
	}
	return Loaded{Config: cfg, Items: snapshot.List(), byID: snapshot}, nil
}

// Project returns the explicit project or the configured default.
func Project(cfg config.Config, project string, explicit bool) string {
	if explicit {
		return project
	}
	return cfg.Project
}

type ContextRequest struct {
	Project         string
	ProjectExplicit bool
	Agent           string
	Files           []string
	Task            string
	Limit           int
	Native          bool
}

// ContextResult is the deterministic resolution plus retrieved passages.
type ContextResult struct {
	resolver.Result
	Passages        []retrieval.Hit      `json:"passages"`
	Documentation   DocumentationContext `json:"documentation"`
	NativeKnowledge []NativeSelection    `json:"native_knowledge,omitempty"`
}

type DocumentationContext struct {
	Status   string `json:"status"` // not_requested, not_configured, external, complete, unavailable
	Guidance string `json:"guidance,omitempty"`
	Message  string `json:"message,omitempty"`
}

// NativeSelection identifies selected knowledge whose body is already available
// in verified native files. Optional rule bodies, references, and decisions remain
// in the normal result because sync does not deliver their full content.
type NativeSelection struct {
	ID       string            `json:"id"`
	Kind     knowledge.Kind    `json:"kind"`
	Required bool              `json:"required"`
	Version  string            `json:"version"`
	Source   knowledge.Source  `json:"source"`
	Reasons  []resolver.Reason `json:"reasons"`
	File     string            `json:"file"`
}

func (a App) Context(ctx context.Context, req ContextRequest) (ContextResult, error) {
	loaded, err := a.Load(ctx)
	if err != nil {
		return ContextResult{}, err
	}
	project := Project(loaded.Config, req.Project, req.ProjectExplicit)
	result, err := resolver.Resolve(loaded.Items, resolver.Request{Project: project, Agent: req.Agent, Files: req.Files})
	if err != nil {
		return ContextResult{}, err
	}
	out := ContextResult{Result: result, Passages: []retrieval.Hit{}, Documentation: DocumentationContext{Status: "not_requested", Guidance: loaded.Config.Delivery.Documentation}}
	if req.Native {
		plan, err := nativesync.Plan(nativesync.Options{Agent: req.Agent, Project: project, Dir: loaded.Config.Dir, Items: loaded.Items, Documentation: loaded.Config.Delivery.Documentation})
		if err != nil {
			return ContextResult{}, fmt.Errorf("native delivery: %w", err)
		}
		if plan.Changed() || len(plan.Conflicts) > 0 {
			return ContextResult{}, fmt.Errorf("native delivery requires up-to-date, unedited synced files next to nexora.yaml; run nexora sync --agent %s, or omit native mode for full context", req.Agent)
		}
		target := nativesync.Targets[req.Agent]
		for _, group := range [][]resolver.Selection{result.RequiredRules, result.Skills} {
			for _, s := range group {
				file := target.Instructions
				if s.Item.Kind == knowledge.Skill {
					file = target.SkillsDir + "/" + nativesync.SkillName(s.Item.ID) + "/SKILL.md"
				}
				out.NativeKnowledge = append(out.NativeKnowledge, NativeSelection{ID: s.Item.ID, Kind: s.Item.Kind, Required: s.Item.Required, Version: s.Item.Version, Source: s.Item.Source, Reasons: s.Reasons, File: file})
			}
		}
		out.RequiredRules = []resolver.Selection{}
		out.Skills = []resolver.Selection{}
	}
	if strings.TrimSpace(req.Task) == "" {
		return out, nil
	}
	if len(loaded.Config.DocumentSources()) == 0 {
		out.Documentation.Status = "not_configured"
		if out.Documentation.Guidance != "" {
			out.Documentation.Status = "external"
		}
		return out, nil
	}
	out.Passages, err = a.search(ctx, loaded.Config, req.Task, project, req.Limit)
	if err != nil {
		if ctx.Err() != nil {
			return ContextResult{}, ctx.Err()
		}
		out.Passages = []retrieval.Hit{}
		out.Documentation.Status = "unavailable"
		// Provider errors can include URLs or credentials. Keep the partial
		// context safe; search remains the explicit diagnostic command.
		out.Documentation.Message = "Documentation retrieval unavailable; run nexora index if needed, or nexora search to diagnose configured sources. Selected knowledge is preserved."
		return out, nil
	}
	out.Documentation.Status = "complete"
	return out, nil
}

type SearchRequest struct {
	Query           string
	Project         string
	ProjectExplicit bool
	Limit           int
}

func (a App) Search(ctx context.Context, req SearchRequest) ([]retrieval.Hit, error) {
	cfg, err := config.Load(a.ConfigPath)
	if err != nil {
		return nil, err
	}
	if len(cfg.DocumentSources()) == 0 {
		if cfg.Delivery.Documentation != "" {
			return nil, fmt.Errorf("documentation lookup belongs to an existing service: %s Nexora does not call that service; use it directly", cfg.Delivery.Documentation)
		}
		return nil, fmt.Errorf("no docs, confluence, or pgvector sources are configured; add one to %s and run nexora index", a.ConfigPath)
	}
	return a.search(ctx, cfg, req.Query, Project(cfg, req.Project, req.ProjectExplicit), req.Limit)
}

func (a App) search(ctx context.Context, cfg config.Config, text, project string, limit int) ([]retrieval.Hit, error) {
	store, err := a.store(cfg)
	if err != nil {
		return nil, err
	}
	defer store.Close()
	embedder, err := embed.New(cfg.Index.Embeddings, a.getenv, a.HTTP)
	if err != nil {
		return nil, err
	}
	req := retrieval.Request{Text: text, Project: project, Limit: limit}
	for _, source := range cfg.VectorSources() {
		table, err := a.table(source)
		if err != nil {
			return nil, err
		}
		defer table.Close()
		queryEmbedder, err := embed.New(*source.Embeddings, a.getenv, a.HTTP)
		if err != nil {
			return nil, fmt.Errorf("source %s: %w", source.ID, err)
		}
		req.Vectors = append(req.Vectors, retrieval.VectorSource{ID: source.ID, Table: table, Embedder: queryEmbedder})
	}
	hits, err := retrieval.Search(ctx, store, embedder, req)
	if errors.Is(err, index.ErrNoIndex) {
		return nil, fmt.Errorf("%w (configured documentation sources have not been indexed)", index.ErrNoIndex)
	}
	return hits, err
}

func (a App) store(cfg config.Config) (index.Store, error) {
	if a.OpenStore != nil {
		return a.OpenStore(cfg)
	}
	return index.Open(cfg, a.getenv)
}

func (a App) table(source config.Source) (index.VectorTable, error) {
	if a.OpenTable != nil {
		return a.OpenTable(source)
	}
	return index.OpenVectorTable(source, a.getenv)
}

func (a App) getenv(name string) string {
	if a.Getenv == nil {
		return ""
	}
	return a.Getenv(name)
}

// SourceReport summarizes indexing of one source.
type SourceReport struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	Documents int    `json:"documents"`
	Passages  int    `json:"passages"`
	Embedded  int    `json:"embedded"`
	// Vectors is "in place" for pgvector sources, whose own vectors are
	// searched at query time instead of being copied.
	Vectors string `json:"vectors,omitempty"`
}

type IndexReport struct {
	Store   string         `json:"store"`
	Model   string         `json:"model"`
	Sources []SourceReport `json:"sources"`
}

// Index rebuilds passages for every documentation source, or only one. A
// failure leaves the stored index unchanged.
func (a App) Index(ctx context.Context, only string) (IndexReport, error) {
	cfg, err := config.Load(a.ConfigPath)
	if err != nil {
		return IndexReport{}, err
	}
	sources := cfg.DocumentSources()
	if only != "" {
		idx := slices.IndexFunc(cfg.Sources, func(s config.Source) bool { return s.ID == only })
		if idx < 0 {
			return IndexReport{}, fmt.Errorf("unknown source %q", only)
		}
		if cfg.Sources[idx].Type == config.SourceKnowledge {
			return IndexReport{}, fmt.Errorf("source %s holds knowledge items, which are resolved deterministically and not indexed", only)
		}
		sources = []config.Source{cfg.Sources[idx]}
	}
	if len(sources) == 0 {
		return IndexReport{}, fmt.Errorf("nothing to index: no docs, confluence, or pgvector sources are configured")
	}
	embedder, err := embed.New(cfg.Index.Embeddings, a.getenv, a.HTTP)
	if err != nil {
		return IndexReport{}, err
	}
	report := IndexReport{Store: cfg.Index.Store, Sources: []SourceReport{}}
	batch := map[string][]passage.Passage{}
	meta := index.Meta{}
	if embedder != nil {
		meta.Model = embedder.Model()
		report.Model = meta.Model
	}
	opt := passage.Options{MaxChars: cfg.Index.Chunk.MaxChars, OverlapChars: cfg.Index.Chunk.OverlapChars}
	for _, source := range sources {
		if source.Type == config.SourcePgvector {
			list, err := a.rows(ctx, source)
			if err != nil {
				return IndexReport{}, err
			}
			batch[source.ID] = list
			report.Sources = append(report.Sources, SourceReport{ID: source.ID, Type: source.Type, Documents: documentCount(list), Passages: len(list), Vectors: "in place"})
			continue
		}
		docs, err := a.documents(ctx, source)
		if err != nil {
			return IndexReport{}, err
		}
		list := []passage.Passage{}
		seen := map[string]bool{}
		for _, doc := range docs {
			for _, p := range passage.Chunk(doc, source.ID, source.Type, source.Projects, opt) {
				if seen[p.ID] {
					return IndexReport{}, fmt.Errorf("source %s: duplicate passage id %s", source.ID, p.ID)
				}
				seen[p.ID] = true
				list = append(list, p)
			}
		}
		sr := SourceReport{ID: source.ID, Type: source.Type, Documents: len(docs), Passages: len(list)}
		if embedder != nil && len(list) > 0 {
			texts := make([]string, len(list))
			for i, p := range list {
				texts[i] = p.SearchText()
			}
			vectors, err := embedder.Embed(ctx, texts)
			if err != nil {
				return IndexReport{}, fmt.Errorf("source %s: %w", source.ID, err)
			}
			for i := range list {
				list[i].Embedding = vectors[i]
			}
			meta.Dimensions = len(vectors[0])
			sr.Embedded = len(list)
		}
		batch[source.ID] = list
		report.Sources = append(report.Sources, sr)
	}
	store, err := a.store(cfg)
	if err != nil {
		return IndexReport{}, err
	}
	defer store.Close()
	if err := store.Replace(ctx, batch, meta, only == ""); err != nil {
		return IndexReport{}, err
	}
	return report, nil
}

// rows copies the text of a pgvector source for keyword search. Its vectors stay
// in the table and are never re-embedded.
func (a App) rows(ctx context.Context, source config.Source) ([]passage.Passage, error) {
	table, err := a.table(source)
	if err != nil {
		return nil, err
	}
	defer table.Close()
	list, err := table.Rows(ctx)
	if err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	for _, p := range list {
		if seen[p.ID] {
			return nil, fmt.Errorf("source %s: duplicate row id %s; columns.id must be unique", source.ID, strings.TrimPrefix(p.ID, source.ID+":"))
		}
		seen[p.ID] = true
	}
	return list, nil
}

func documentCount(list []passage.Passage) int {
	docs := map[string]bool{}
	for _, p := range list {
		docs[p.DocumentID] = true
	}
	return len(docs)
}

func (a App) documents(ctx context.Context, source config.Source) ([]passage.Document, error) {
	if source.Type == config.SourceDocs {
		return passage.LoadDocuments(ctx, source)
	}
	token := a.getenv(source.TokenEnv)
	if strings.TrimSpace(token) == "" {
		return nil, fmt.Errorf("source %s: environment variable %s is not set", source.ID, source.TokenEnv)
	}
	client := confluence.Client{BaseURL: source.URL, Token: token, HTTP: a.HTTP}
	if source.UserEnv != "" {
		client.User = a.getenv(source.UserEnv)
		if strings.TrimSpace(client.User) == "" {
			return nil, fmt.Errorf("source %s: environment variable %s is not set", source.ID, source.UserEnv)
		}
	}
	docs := []passage.Document{}
	for _, space := range source.Spaces {
		pages, err := client.Pages(ctx, space)
		if err != nil {
			return nil, fmt.Errorf("source %s: %w", source.ID, err)
		}
		for _, page := range pages {
			docs = append(docs, page.Document())
		}
	}
	sort.Slice(docs, func(i, j int) bool { return docs[i].ID < docs[j].ID })
	return docs, nil
}

// Status reports passage counts per indexed source.
func (a App) Status(ctx context.Context) (map[string]int, index.Meta, error) {
	cfg, err := config.Load(a.ConfigPath)
	if err != nil {
		return nil, index.Meta{}, err
	}
	store, err := a.store(cfg)
	if err != nil {
		return nil, index.Meta{}, err
	}
	defer store.Close()
	meta, err := store.Meta(ctx)
	if err != nil {
		return nil, index.Meta{}, err
	}
	stats, err := store.Stats(ctx)
	return stats, meta, err
}

// Found is either a knowledge item or a passage.
type Found struct {
	Kind    string           `json:"kind"`
	Item    *knowledge.Item  `json:"item,omitempty"`
	Passage *passage.Passage `json:"passage,omitempty"`
}

func (a App) Get(ctx context.Context, id string) (Found, error) {
	loaded, err := a.Load(ctx)
	if err != nil {
		return Found{}, err
	}
	if item, ok := loaded.byID.Get(id); ok {
		return Found{Kind: "knowledge", Item: &item}, nil
	}
	if len(loaded.Config.DocumentSources()) > 0 {
		store, err := a.store(loaded.Config)
		if err != nil {
			return Found{}, err
		}
		defer store.Close()
		p, ok, err := store.Get(ctx, id)
		if err != nil && !errors.Is(err, index.ErrNoIndex) {
			return Found{}, err
		}
		if ok {
			return Found{Kind: "passage", Passage: &p}, nil
		}
	}
	return Found{}, fmt.Errorf("no knowledge item or passage with id %q", id)
}

// Skills lists active skills applicable to the project and agent. Path
// restrictions are kept on the items for the caller to show.
func (a App) Skills(ctx context.Context, project string, explicit bool, agent string) ([]knowledge.Item, error) {
	loaded, err := a.Load(ctx)
	if err != nil {
		return nil, err
	}
	project = Project(loaded.Config, project, explicit)
	out := []knowledge.Item{}
	for _, item := range loaded.Items {
		if item.Kind == knowledge.Skill && Applies(item, project, agent) {
			out = append(out, item)
		}
	}
	return out, nil
}

// Applies checks status, project, and agent with the resolver's semantics,
// ignoring path restrictions.
func Applies(item knowledge.Item, project, agent string) bool {
	if item.Status != knowledge.Active {
		return false
	}
	match := func(allowed []string, value string) bool {
		return len(allowed) == 0 || value != "" && slices.Contains(allowed, value)
	}
	return match(item.AppliesTo.Projects, project) && match(item.AppliesTo.Agents, agent)
}
