// Package markdown loads local Markdown knowledge without rendering or execution.
package markdown

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/acierto/nexora/internal/config"
	"github.com/acierto/nexora/internal/knowledge"
	"github.com/acierto/nexora/internal/yamlutil"
)

type metadata struct {
	ID          string            `yaml:"id"`
	Name        string            `yaml:"name"`
	Description string            `yaml:"description"`
	Kind        knowledge.Kind    `yaml:"kind"`
	Version     *string           `yaml:"version"`
	Status      *knowledge.Status `yaml:"status"`
	Tags        []string          `yaml:"tags"`
	Priority    int               `yaml:"priority"`
	Required    *bool             `yaml:"required"`
	AppliesTo   struct {
		Projects []string `yaml:"projects"`
		Agents   []string `yaml:"agents"`
		Paths    []string `yaml:"paths"`
	} `yaml:"applies_to"`
}

func Parse(data []byte, source knowledge.Source) (knowledge.Item, error) {
	if !utf8.Valid(data) {
		return knowledge.Item{}, fmt.Errorf("invalid UTF-8")
	}
	lines := bytes.SplitAfter(data, []byte("\n"))
	delimiter := func(line []byte) bool {
		return string(bytes.TrimSuffix(bytes.TrimSuffix(line, []byte("\n")), []byte("\r"))) == "---"
	}
	if len(lines) < 3 || !delimiter(lines[0]) {
		return knowledge.Item{}, fmt.Errorf("front matter must start with --- on its own line")
	}
	end := 1
	for end < len(lines) && !delimiter(lines[end]) {
		end++
	}
	if end == len(lines) {
		return knowledge.Item{}, fmt.Errorf("front matter is missing closing --- delimiter")
	}
	var m metadata
	if err := yamlutil.Decode(bytes.Join(lines[1:end], nil), &m); err != nil {
		return knowledge.Item{}, fmt.Errorf("front matter: %w", err)
	}
	if m.Required != nil && m.Kind != knowledge.Rule {
		return knowledge.Item{}, fmt.Errorf("required: only valid for rules")
	}
	i := knowledge.Item{ID: m.ID, Name: m.Name, Description: m.Description, Kind: m.Kind, Content: string(bytes.Join(lines[end+1:], nil)), Status: knowledge.Active, Tags: m.Tags, Priority: m.Priority, Required: m.Kind == knowledge.Rule, AppliesTo: knowledge.Applicability{Projects: m.AppliesTo.Projects, Agents: m.AppliesTo.Agents, Paths: m.AppliesTo.Paths}, Source: source}
	i.Source.Digest = fmt.Sprintf("%x", sha256.Sum256(data))
	i.Version = "sha256:" + i.Source.Digest
	if m.Version != nil {
		if strings.TrimSpace(*m.Version) == "" {
			return knowledge.Item{}, fmt.Errorf("version: must not be empty")
		}
		i.Version = *m.Version
	}
	if m.Status != nil {
		i.Status = *m.Status
	}
	if m.Required != nil {
		i.Required = *m.Required
	}
	if err := i.Validate(); err != nil {
		return knowledge.Item{}, err
	}
	return i.Clone(), nil
}

func Load(ctx context.Context, source config.Source) ([]knowledge.Item, error) {
	items := []knowledge.Item{}
	err := filepath.WalkDir(source.Root, func(file string, entry fs.DirEntry, walkErr error) error {
		rel, err := filepath.Rel(source.Root, file)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		wrap := func(err error) error { return fmt.Errorf("source %s, %s: %w", source.ID, rel, err) }
		if err = ctx.Err(); err != nil {
			return wrap(err)
		}
		if walkErr != nil {
			return wrap(walkErr)
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return wrap(fmt.Errorf("symlink entries are not allowed"))
		}
		if entry.IsDir() {
			return nil
		}
		if !entry.Type().IsRegular() || filepath.Ext(file) != ".md" {
			return nil
		}
		data, err := os.ReadFile(file)
		if err != nil {
			return wrap(err)
		}
		if err = ctx.Err(); err != nil {
			return wrap(err)
		}
		item, err := Parse(data, knowledge.Source{ID: source.ID, Path: rel})
		if err != nil {
			return wrap(err)
		}
		items = append(items, item)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return items, nil
}
