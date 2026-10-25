// Package knowledge defines storage-independent knowledge values and validation.
package knowledge

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
)

type Kind string

const (
	Rule      Kind = "rule"
	Skill     Kind = "skill"
	Reference Kind = "reference"
	Decision  Kind = "decision"
)

type Status string

const (
	Active   Status = "active"
	Draft    Status = "draft"
	Archived Status = "archived"
)

type Applicability struct {
	Projects []string `json:"projects"`
	Agents   []string `json:"agents"`
	Paths    []string `json:"paths"`
}

type Source struct {
	ID     string `json:"id"`
	Path   string `json:"path"`
	Digest string `json:"digest"`
}

type Item struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Kind        Kind          `json:"kind"`
	Content     string        `json:"content"`
	Version     string        `json:"version"`
	Status      Status        `json:"status"`
	Tags        []string      `json:"tags"`
	Priority    int           `json:"priority"`
	Required    bool          `json:"required"`
	AppliesTo   Applicability `json:"applies_to"`
	Source      Source        `json:"source"`
}

var identifier = regexp.MustCompile(`^[a-z0-9][a-z0-9._/-]*$`)

func (i Item) Validate() error {
	if !identifier.MatchString(i.ID) {
		return fmt.Errorf("id: invalid identifier %q", i.ID)
	}
	for _, segment := range strings.Split(i.ID, "/") {
		if segment == "" || segment == "." || segment == ".." {
			return fmt.Errorf("id: invalid segment in %q", i.ID)
		}
	}
	for _, field := range []struct{ name, value string }{{"name", i.Name}, {"description", i.Description}, {"content", i.Content}} {
		if strings.TrimSpace(field.value) == "" {
			return fmt.Errorf("%s: must not be empty", field.name)
		}
	}
	if !slices.Contains([]Kind{Rule, Skill, Reference, Decision}, i.Kind) {
		return fmt.Errorf("kind: invalid value %q", i.Kind)
	}
	if !slices.Contains([]Status{Active, Draft, Archived}, i.Status) {
		return fmt.Errorf("status: invalid value %q", i.Status)
	}
	if i.Kind != Rule && i.Required {
		return fmt.Errorf("required: only valid for rules")
	}
	for _, field := range []struct {
		name   string
		values []string
	}{{"projects", i.AppliesTo.Projects}, {"agents", i.AppliesTo.Agents}, {"paths", i.AppliesTo.Paths}} {
		for _, value := range field.values {
			if strings.TrimSpace(value) == "" {
				return fmt.Errorf("applies_to.%s: entries must not be empty", field.name)
			}
			if field.name == "paths" {
				if _, err := ParsePattern(value); err != nil {
					return fmt.Errorf("applies_to.paths: %w", err)
				}
			}
		}
	}
	return nil
}

// Clone returns a value with independent, non-nil slices for safe snapshots/JSON.
func (i Item) Clone() Item {
	i.Tags = append([]string{}, i.Tags...)
	i.AppliesTo.Projects = append([]string{}, i.AppliesTo.Projects...)
	i.AppliesTo.Agents = append([]string{}, i.AppliesTo.Agents...)
	i.AppliesTo.Paths = append([]string{}, i.AppliesTo.Paths...)
	return i
}
