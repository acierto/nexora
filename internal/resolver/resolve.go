// Package resolver selects knowledge using deterministic applicability metadata.
package resolver

import (
	"fmt"
	"slices"
	"sort"

	"github.com/acierto/nexora/internal/knowledge"
)

type Request struct {
	Project string   `json:"project"`
	Agent   string   `json:"agent"`
	Files   []string `json:"files"`
}

type PathMatch struct {
	Pattern string `json:"pattern"`
	File    string `json:"file"`
}
type Reason struct {
	Dimension    string      `json:"dimension"`
	Unrestricted bool        `json:"unrestricted"`
	Value        string      `json:"value"`
	Matches      []PathMatch `json:"matches"`
}
type Selection struct {
	Item    knowledge.Item `json:"item"`
	Reasons []Reason       `json:"reasons"`
}
type Result struct {
	RequiredRules []Selection `json:"required_rules"`
	OptionalRules []Selection `json:"optional_rules"`
	Skills        []Selection `json:"skills"`
	References    []Selection `json:"references"`
	Decisions     []Selection `json:"decisions"`
}

func Resolve(items []knowledge.Item, request Request) (Result, error) {
	result := Result{RequiredRules: []Selection{}, OptionalRules: []Selection{}, Skills: []Selection{}, References: []Selection{}, Decisions: []Selection{}}
	files := []string{}
	for _, file := range request.Files {
		normalized, err := knowledge.NormalizeFile(file)
		if err != nil {
			return Result{}, err
		}
		files = append(files, normalized)
	}
	sort.Strings(files)
	files = slices.Compact(files)
	seen := map[string]bool{}
	for _, item := range items {
		if err := item.Validate(); err != nil {
			return Result{}, fmt.Errorf("item %s: %w", item.ID, err)
		}
		if seen[item.ID] {
			return Result{}, fmt.Errorf("duplicate id %q", item.ID)
		}
		seen[item.ID] = true
		if item.Status != knowledge.Active {
			continue
		}
		project, ok := dimension("project", item.AppliesTo.Projects, request.Project)
		if !ok {
			continue
		}
		agent, ok := dimension("agent", item.AppliesTo.Agents, request.Agent)
		if !ok {
			continue
		}
		paths := Reason{Dimension: "paths", Unrestricted: len(item.AppliesTo.Paths) == 0, Matches: []PathMatch{}}
		patterns := append([]string{}, item.AppliesTo.Paths...)
		sort.Strings(patterns)
		patterns = slices.Compact(patterns)
		for _, raw := range patterns {
			pattern, err := knowledge.ParsePattern(raw)
			if err != nil {
				return Result{}, err
			}
			for _, file := range files {
				if pattern.Match(file) {
					paths.Matches = append(paths.Matches, PathMatch{raw, file})
				}
			}
		}
		if !paths.Unrestricted && len(paths.Matches) == 0 {
			continue
		}
		selection := Selection{Item: item.Clone(), Reasons: []Reason{{Dimension: "status", Value: string(item.Status), Matches: []PathMatch{}}, project, agent, paths}}
		switch item.Kind {
		case knowledge.Rule:
			if item.Required {
				result.RequiredRules = append(result.RequiredRules, selection)
			} else {
				result.OptionalRules = append(result.OptionalRules, selection)
			}
		case knowledge.Skill:
			result.Skills = append(result.Skills, selection)
		case knowledge.Reference:
			result.References = append(result.References, selection)
		case knowledge.Decision:
			result.Decisions = append(result.Decisions, selection)
		}
	}
	for _, group := range [][]Selection{result.RequiredRules, result.OptionalRules, result.Skills, result.References, result.Decisions} {
		sort.Slice(group, func(i, j int) bool {
			if group[i].Item.Priority != group[j].Item.Priority {
				return group[i].Item.Priority > group[j].Item.Priority
			}
			return group[i].Item.ID < group[j].Item.ID
		})
	}
	return result, nil
}

func dimension(name string, allowed []string, value string) (Reason, bool) {
	reason := Reason{Dimension: name, Unrestricted: len(allowed) == 0, Matches: []PathMatch{}}
	if len(allowed) == 0 {
		return reason, true
	}
	if value != "" && slices.Contains(allowed, value) {
		reason.Value = value
		return reason, true
	}
	return reason, false
}
