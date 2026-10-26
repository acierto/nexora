// Package config loads and validates local Nexora configuration.
package config

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/acierto/nexora/internal/yamlutil"
)

// Source types. Knowledge sources hold front-matter knowledge items; docs and
// confluence sources hold documentation that is indexed into passages; pgvector
// sources are existing tables of already embedded passages, searched in place.
const (
	SourceKnowledge  = "knowledge"
	SourceDocs       = "docs"
	SourceConfluence = "confluence"
	SourcePgvector   = "pgvector"
)

// Distance functions for pgvector sources.
const (
	DistanceCosine       = "cosine"
	DistanceL2           = "l2"
	DistanceInnerProduct = "inner_product"
)

// Index stores and embedding providers.
const (
	StoreLocal    = "local"
	StorePgvector = "pgvector"

	ProviderNone   = "none"
	ProviderOllama = "ollama"
	ProviderOpenAI = "openai"
)

type Source struct {
	ID       string   `yaml:"id"`
	Type     string   `yaml:"type"`
	Path     string   `yaml:"path"`
	URL      string   `yaml:"url"`
	Spaces   []string `yaml:"spaces"`
	UserEnv  string   `yaml:"user_env"`
	TokenEnv string   `yaml:"token_env"`
	Projects []string `yaml:"projects"`
	// pgvector sources
	DSNEnv     string      `yaml:"dsn_env"`
	Table      string      `yaml:"table"`
	Columns    Columns     `yaml:"columns"`
	Distance   string      `yaml:"distance"`
	Embeddings *Embeddings `yaml:"embeddings"`
	Root       string      `yaml:"-"`
}

// Columns maps an existing table to passage fields. ID, Text, and Vector are
// required; the others are optional.
type Columns struct {
	ID       string `yaml:"id"`
	Text     string `yaml:"text"`
	Vector   string `yaml:"vector"`
	Title    string `yaml:"title"`
	Location string `yaml:"location"`
	Document string `yaml:"document"`
}

type Embeddings struct {
	Provider   string `yaml:"provider"`
	Model      string `yaml:"model"`
	URL        string `yaml:"url"`
	APIKeyEnv  string `yaml:"api_key_env"`
	Dimensions int    `yaml:"dimensions"`
}

type Chunk struct {
	MaxChars int `yaml:"max_chars"`
	// Overlap is a pointer so an explicit 0 differs from the 150 default.
	Overlap *int `yaml:"overlap_chars"`
	// OverlapChars is the effective overlap after defaults.
	OverlapChars int `yaml:"-"`
}

type Index struct {
	Store      string     `yaml:"store"`
	Path       string     `yaml:"path"`
	DSNEnv     string     `yaml:"dsn_env"`
	Table      string     `yaml:"table"`
	Embeddings Embeddings `yaml:"embeddings"`
	Chunk      Chunk      `yaml:"chunk"`
	// File is the absolute local index file (local store only).
	File string `yaml:"-"`
}

type Config struct {
	SchemaVersion int      `yaml:"schema_version"`
	Project       string   `yaml:"project"`
	Sources       []Source `yaml:"sources"`
	Index         Index    `yaml:"index"`
	// Dir is the absolute directory containing the configuration file.
	Dir string `yaml:"-"`
}

var (
	envName       = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)
	tableName     = regexp.MustCompile(`^[a-z_][a-z0-9_]{0,50}$`)
	sqlIdentifier = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_$]{0,62}$`)
)

func Load(file string) (Config, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return Config{}, fmt.Errorf("config %s: %w; run nexora init or supply --config <path>", file, err)
	}
	var cfg Config
	if err = yamlutil.Decode(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("config %s: %w", file, err)
	}
	dir, err := filepath.Abs(filepath.Dir(file))
	if err != nil {
		return Config{}, fmt.Errorf("config %s: %w", file, err)
	}
	if err = cfg.validate(dir); err != nil {
		return Config{}, fmt.Errorf("config %s: %w", file, err)
	}
	cfg.Dir = dir
	return cfg, nil
}

// KnowledgeSources returns sources holding front-matter knowledge items.
func (c Config) KnowledgeSources() []Source { return c.sourcesOf(SourceKnowledge) }

// DocumentSources returns docs, confluence, and pgvector sources, in ID order.
// All of them are indexed for keyword search.
func (c Config) DocumentSources() []Source {
	out := []Source{}
	for _, s := range c.Sources {
		if s.Type == SourceDocs || s.Type == SourceConfluence || s.Type == SourcePgvector {
			out = append(out, s)
		}
	}
	return out
}

// VectorSources returns pgvector sources, searched by meaning in place.
func (c Config) VectorSources() []Source { return c.sourcesOf(SourcePgvector) }

func (c Config) sourcesOf(kind string) []Source {
	out := []Source{}
	for _, s := range c.Sources {
		if s.Type == kind {
			out = append(out, s)
		}
	}
	return out
}

func (c *Config) validate(base string) error {
	if c.SchemaVersion != 1 {
		return fmt.Errorf("schema_version: only version 1 is supported")
	}
	if strings.TrimSpace(c.Project) == "" {
		return fmt.Errorf("project: must not be empty")
	}
	if len(c.Sources) == 0 {
		return fmt.Errorf("sources: at least one source is required")
	}
	seen := map[string]bool{}
	sort.SliceStable(c.Sources, func(i, j int) bool { return c.Sources[i].ID < c.Sources[j].ID })
	for idx := range c.Sources {
		source := &c.Sources[idx]
		if source.Type == "" {
			source.Type = SourceKnowledge
		}
		if strings.TrimSpace(source.ID) == "" {
			return fmt.Errorf("source id must not be empty")
		}
		if seen[source.ID] {
			return fmt.Errorf("duplicate source id %q", source.ID)
		}
		seen[source.ID] = true
		switch source.Type {
		case SourceKnowledge, SourceDocs:
			if err := validatePathSource(source); err != nil {
				return err
			}
		case SourceConfluence:
			if err := validateConfluence(source); err != nil {
				return err
			}
			continue
		case SourcePgvector:
			if err := validatePgvectorSource(source); err != nil {
				return err
			}
			continue
		default:
			return fmt.Errorf("source %s: type must be knowledge, docs, confluence, or pgvector", source.ID)
		}
		root, err := canonicalDir(base, source.ID, source.Path)
		if err != nil {
			return err
		}
		for prev := 0; prev < idx; prev++ {
			other := c.Sources[prev]
			if other.Root == "" {
				continue
			}
			if within(other.Root, root) || within(root, other.Root) {
				return fmt.Errorf("sources %s and %s have overlapping roots", other.ID, source.ID)
			}
		}
		source.Root = root
	}
	return c.Index.validate(base)
}

func validatePathSource(source *Source) error {
	if strings.TrimSpace(source.Path) == "" {
		return fmt.Errorf("source id and path must not be empty")
	}
	if source.URL != "" || len(source.Spaces) > 0 || source.UserEnv != "" || source.TokenEnv != "" {
		return fmt.Errorf("source %s: url, spaces, user_env, and token_env are only valid for confluence sources", source.ID)
	}
	if err := noPgvectorFields(source); err != nil {
		return err
	}
	if source.Type == SourceKnowledge && len(source.Projects) > 0 {
		return fmt.Errorf("source %s: projects is only valid for docs and confluence sources; scope knowledge with applies_to", source.ID)
	}
	return validateProjects(source)
}

func validateConfluence(source *Source) error {
	if source.Path != "" {
		return fmt.Errorf("source %s: confluence sources use url and spaces, not path", source.ID)
	}
	if err := noPgvectorFields(source); err != nil {
		return err
	}
	parsed, err := url.Parse(source.URL)
	if err != nil || (parsed.Scheme != "https" && parsed.Scheme != "http") || parsed.Host == "" {
		return fmt.Errorf("source %s: url must be an http or https Confluence base URL", source.ID)
	}
	source.URL = strings.TrimRight(source.URL, "/")
	if len(source.Spaces) == 0 {
		return fmt.Errorf("source %s: spaces must list at least one space key", source.ID)
	}
	spaces := map[string]bool{}
	for _, space := range source.Spaces {
		if strings.TrimSpace(space) == "" {
			return fmt.Errorf("source %s: spaces entries must not be empty", source.ID)
		}
		if spaces[space] {
			return fmt.Errorf("source %s: duplicate space %q", source.ID, space)
		}
		spaces[space] = true
	}
	if !envName.MatchString(source.TokenEnv) {
		return fmt.Errorf("source %s: token_env must name the environment variable holding the API token", source.ID)
	}
	if source.UserEnv != "" && !envName.MatchString(source.UserEnv) {
		return fmt.Errorf("source %s: user_env must be an environment variable name", source.ID)
	}
	return validateProjects(source)
}

func noPgvectorFields(source *Source) error {
	if source.DSNEnv != "" || source.Table != "" || source.Columns != (Columns{}) || source.Distance != "" || source.Embeddings != nil {
		return fmt.Errorf("source %s: dsn_env, table, columns, distance, and embeddings are only valid for pgvector sources", source.ID)
	}
	return nil
}

func validatePgvectorSource(source *Source) error {
	prefix := "source " + source.ID
	if source.Path != "" || source.URL != "" || len(source.Spaces) > 0 || source.UserEnv != "" || source.TokenEnv != "" {
		return fmt.Errorf("%s: pgvector sources use dsn_env, table, and columns, not path, url, spaces, user_env, or token_env", prefix)
	}
	if !envName.MatchString(source.DSNEnv) {
		return fmt.Errorf("%s: dsn_env must name the environment variable holding the PostgreSQL connection string", prefix)
	}
	parts := strings.Split(source.Table, ".")
	if len(parts) > 2 || !all(parts, sqlIdentifier.MatchString) {
		return fmt.Errorf("%s: table must be a table name or schema.table, using letters, digits, and underscores", prefix)
	}
	cols := source.Columns
	for name, value := range map[string]string{"id": cols.ID, "text": cols.Text, "vector": cols.Vector} {
		if value == "" {
			return fmt.Errorf("%s: columns.%s is required", prefix, name)
		}
	}
	for name, value := range map[string]string{"id": cols.ID, "text": cols.Text, "vector": cols.Vector, "title": cols.Title, "location": cols.Location, "document": cols.Document} {
		if value != "" && !sqlIdentifier.MatchString(value) {
			return fmt.Errorf("%s: columns.%s must be a column name using letters, digits, and underscores", prefix, name)
		}
	}
	if source.Distance == "" {
		source.Distance = DistanceCosine
	}
	switch source.Distance {
	case DistanceCosine, DistanceL2, DistanceInnerProduct:
	default:
		return fmt.Errorf("%s: distance must be cosine, l2, or inner_product", prefix)
	}
	if source.Embeddings == nil || source.Embeddings.Provider == "" || source.Embeddings.Provider == ProviderNone {
		return fmt.Errorf("%s: embeddings must name the provider and model that produced the stored vectors, so queries are embedded the same way", prefix)
	}
	if err := source.Embeddings.validate(prefix + ": embeddings"); err != nil {
		return err
	}
	return validateProjects(source)
}

func all(values []string, ok func(string) bool) bool {
	for _, v := range values {
		if !ok(v) {
			return false
		}
	}
	return true
}

func validateProjects(source *Source) error {
	for _, project := range source.Projects {
		if strings.TrimSpace(project) == "" {
			return fmt.Errorf("source %s: projects entries must not be empty", source.ID)
		}
	}
	return nil
}

func canonicalDir(base, id, path string) (string, error) {
	root := path
	if !filepath.IsAbs(root) {
		root = filepath.Join(base, root)
	}
	root, err := filepath.Abs(root)
	if err != nil {
		return "", fmt.Errorf("source %s: %w", id, err)
	}
	root, err = filepath.EvalSymlinks(root)
	if err != nil {
		return "", fmt.Errorf("source %s: %w", id, err)
	}
	dir, err := os.Open(root)
	if err != nil {
		return "", fmt.Errorf("source %s: %w", id, err)
	}
	_, readErr := dir.ReadDir(0)
	closeErr := dir.Close()
	if readErr != nil {
		return "", fmt.Errorf("source %s must be a readable directory: %w", id, readErr)
	}
	if closeErr != nil {
		return "", fmt.Errorf("source %s: %w", id, closeErr)
	}
	return root, nil
}

func (i *Index) validate(base string) error {
	if i.Store == "" {
		i.Store = StoreLocal
	}
	switch i.Store {
	case StoreLocal:
		if i.DSNEnv != "" || i.Table != "" {
			return fmt.Errorf("index: dsn_env and table are only valid for the pgvector store")
		}
		if i.Path == "" {
			i.Path = ".nexora/index.json"
		}
		file := i.Path
		if !filepath.IsAbs(file) {
			file = filepath.Join(base, file)
		}
		abs, err := filepath.Abs(file)
		if err != nil {
			return fmt.Errorf("index.path: %w", err)
		}
		i.File = abs
	case StorePgvector:
		if i.Path != "" {
			return fmt.Errorf("index.path is only valid for the local store")
		}
		if !envName.MatchString(i.DSNEnv) {
			return fmt.Errorf("index.dsn_env must name the environment variable holding the PostgreSQL connection string")
		}
		if i.Table == "" {
			i.Table = "nexora_passages"
		}
		if !tableName.MatchString(i.Table) {
			return fmt.Errorf("index.table must be lowercase letters, digits, and underscores (max 51 characters)")
		}
	default:
		return fmt.Errorf("index.store must be local or pgvector")
	}
	if err := i.Embeddings.validate("index.embeddings"); err != nil {
		return err
	}
	c := &i.Chunk
	if c.MaxChars == 0 {
		c.MaxChars = 1200
	}
	c.OverlapChars = min(150, c.MaxChars/4)
	if c.Overlap != nil {
		c.OverlapChars = *c.Overlap
	}
	if c.MaxChars < 200 || c.MaxChars > 20000 {
		return fmt.Errorf("index.chunk.max_chars must be between 200 and 20000")
	}
	if c.OverlapChars < 0 || c.OverlapChars*2 > c.MaxChars {
		return fmt.Errorf("index.chunk.overlap_chars must be between 0 and half of max_chars")
	}
	return nil
}

// validate checks a provider configuration and fills defaults. name prefixes
// error messages.
func (e *Embeddings) validate(name string) error {
	if e.Provider == "" {
		e.Provider = ProviderNone
	}
	switch e.Provider {
	case ProviderNone:
		if e.Model != "" || e.URL != "" || e.APIKeyEnv != "" || e.Dimensions != 0 {
			return fmt.Errorf("%s: set provider to ollama or openai before configuring model, url, api_key_env, or dimensions", name)
		}
	case ProviderOllama, ProviderOpenAI:
		if strings.TrimSpace(e.Model) == "" {
			return fmt.Errorf("%s.model must not be empty", name)
		}
		if e.URL == "" {
			e.URL = "http://localhost:11434"
			if e.Provider == ProviderOpenAI {
				e.URL = "https://api.openai.com/v1"
			}
		}
		parsed, err := url.Parse(e.URL)
		if err != nil || (parsed.Scheme != "https" && parsed.Scheme != "http") || parsed.Host == "" {
			return fmt.Errorf("%s.url must be an http or https URL", name)
		}
		e.URL = strings.TrimRight(e.URL, "/")
		if e.Provider == ProviderOpenAI && !envName.MatchString(e.APIKeyEnv) {
			return fmt.Errorf("%s.api_key_env must name the environment variable holding the API key", name)
		}
		if e.Provider == ProviderOllama && e.APIKeyEnv != "" {
			return fmt.Errorf("%s.api_key_env is only valid for the openai provider", name)
		}
		if e.Dimensions < 0 || e.Dimensions > 16000 {
			return fmt.Errorf("%s.dimensions must be between 1 and 16000 when set", name)
		}
	default:
		return fmt.Errorf("%s.provider must be none, ollama, or openai", name)
	}
	return nil
}

func within(root, target string) bool {
	rel, err := filepath.Rel(root, target)
	if err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && !filepath.IsAbs(rel) {
		return true
	}
	// Lexical paths can differ in case on case-insensitive filesystems. Compare
	// directory identities too, so alternate spellings cannot bypass overlap checks.
	rootInfo, err := os.Stat(root)
	if err != nil {
		return false
	}
	for current := target; ; current = filepath.Dir(current) {
		info, err := os.Stat(current)
		if err == nil && os.SameFile(rootInfo, info) {
			return true
		}
		if filepath.Dir(current) == current {
			return false
		}
	}
}
