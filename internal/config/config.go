// Package config loads and validates local Nexora configuration.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/acierto/nexora/internal/yamlutil"
)

type Source struct {
	ID   string `yaml:"id"`
	Path string `yaml:"path"`
	Root string `yaml:"-"`
}

type Config struct {
	SchemaVersion int      `yaml:"schema_version"`
	Project       string   `yaml:"project"`
	Sources       []Source `yaml:"sources"`
}

func Load(file string) (Config, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return Config{}, fmt.Errorf("config %s: %w; run nexora init or supply --config <path>", file, err)
	}
	var cfg Config
	if err = yamlutil.Decode(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("config %s: %w", file, err)
	}
	if err = cfg.validate(filepath.Dir(file)); err != nil {
		return Config{}, fmt.Errorf("config %s: %w", file, err)
	}
	return cfg, nil
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
		if strings.TrimSpace(source.ID) == "" || strings.TrimSpace(source.Path) == "" {
			return fmt.Errorf("source id and path must not be empty")
		}
		if seen[source.ID] {
			return fmt.Errorf("duplicate source id %q", source.ID)
		}
		seen[source.ID] = true
		root := source.Path
		if !filepath.IsAbs(root) {
			root = filepath.Join(base, root)
		}
		root, err := filepath.Abs(root)
		if err != nil {
			return fmt.Errorf("source %s: %w", source.ID, err)
		}
		root, err = filepath.EvalSymlinks(root)
		if err != nil {
			return fmt.Errorf("source %s: %w", source.ID, err)
		}
		dir, err := os.Open(root)
		if err != nil {
			return fmt.Errorf("source %s: %w", source.ID, err)
		}
		_, readErr := dir.ReadDir(0)
		closeErr := dir.Close()
		if readErr != nil {
			return fmt.Errorf("source %s must be a readable directory: %w", source.ID, readErr)
		}
		if closeErr != nil {
			return fmt.Errorf("source %s: %w", source.ID, closeErr)
		}
		for prev := 0; prev < idx; prev++ {
			other := c.Sources[prev]
			if within(other.Root, root) || within(root, other.Root) {
				return fmt.Errorf("sources %s and %s have overlapping roots", other.ID, source.ID)
			}
		}
		source.Root = root
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
