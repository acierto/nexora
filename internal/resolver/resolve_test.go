package resolver

import (
	"encoding/json"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/acierto/nexora/internal/knowledge"
)

func rule(id string) knowledge.Item {
	return knowledge.Item{ID: id, Name: id, Description: id, Content: "Rule content", Kind: knowledge.Rule, Status: knowledge.Active, Required: true, Version: "v1", Source: knowledge.Source{ID: "local", Path: id + ".md", Digest: "digest"}}
}

func TestApplicability(t *testing.T) {
	for _, tc := range []struct {
		name    string
		scope   knowledge.Applicability
		request Request
		status  knowledge.Status
		want    bool
	}{
		{"unrestricted", knowledge.Applicability{}, Request{}, knowledge.Active, true},
		{"matching", knowledge.Applicability{Projects: []string{"api", "other"}, Agents: []string{"codex", "claude"}, Paths: []string{"src/**/*.go"}}, Request{Project: "api", Agent: "codex", Files: []string{"src/main.go"}}, knowledge.Active, true},
		{"wrong project", knowledge.Applicability{Projects: []string{"api"}}, Request{Project: "other"}, knowledge.Active, false},
		{"missing project", knowledge.Applicability{Projects: []string{"api"}}, Request{}, knowledge.Active, false},
		{"missing agent", knowledge.Applicability{Agents: []string{"codex"}}, Request{}, knowledge.Active, false},
		{"case", knowledge.Applicability{Agents: []string{"codex"}}, Request{Agent: "Codex"}, knowledge.Active, false},
		{"missing files", knowledge.Applicability{Paths: []string{"**"}}, Request{}, knowledge.Active, false},
		{"wrong extension", knowledge.Applicability{Paths: []string{"src/**/*.go"}}, Request{Files: []string{"src/test.ts"}}, knowledge.Active, false},
		{"windows", knowledge.Applicability{Paths: []string{"src/**/*.go"}}, Request{Files: []string{`src\auth\service.go`}}, knowledge.Active, true},
		{"draft", knowledge.Applicability{}, Request{}, knowledge.Draft, false}, {"archived", knowledge.Applicability{}, Request{}, knowledge.Archived, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			item := rule("rule")
			item.Status = tc.status
			item.AppliesTo = tc.scope
			got, e := Resolve([]knowledge.Item{item}, tc.request)
			if e != nil {
				t.Fatal(e)
			}
			if (len(got.RequiredRules) == 1) != tc.want {
				t.Fatalf("%+v", got)
			}
		})
	}
	for _, file := range []string{"/etc/passwd", `C:\secrets.txt`, "src/../private.go", ""} {
		if _, e := Resolve(nil, Request{Files: []string{file}}); e == nil {
			t.Errorf("accepted %q", file)
		}
	}
}

func TestGroupingOrderingReasons(t *testing.T) {
	a, b, c := rule("a"), rule("b"), rule("c")
	c.Priority = 100
	a.AppliesTo = knowledge.Applicability{Projects: []string{"api"}, Agents: []string{"codex"}, Paths: []string{"src/**/*.go", "**/*.go", "src/**/*.go"}}
	optional := rule("optional")
	optional.Required = false
	items := []knowledge.Item{b, c, a, optional}
	for _, kind := range []knowledge.Kind{knowledge.Skill, knowledge.Reference, knowledge.Decision} {
		i := rule(string(kind))
		i.Kind = kind
		i.Required = false
		items = append(items, i)
	}
	request := Request{Project: "api", Agent: "codex", Files: []string{"src/z.go", "src/a.go", "./src/a.go"}}
	got, e := Resolve(items, request)
	if e != nil {
		t.Fatal(e)
	}
	if len(got.RequiredRules) != 3 || got.RequiredRules[0].Item.ID != "c" || got.RequiredRules[1].Item.ID != "a" || len(got.OptionalRules) != 1 || len(got.Skills) != 1 || len(got.References) != 1 || len(got.Decisions) != 1 {
		t.Fatalf("%+v", got)
	}
	reasons := got.RequiredRules[1].Reasons
	if len(reasons) != 4 || reasons[1].Value != "api" || reasons[2].Value != "codex" || len(reasons[3].Matches) != 4 || reasons[3].Matches[0] != (PathMatch{"**/*.go", "src/a.go"}) {
		t.Fatalf("%+v", reasons)
	}
	if got.RequiredRules[1].Item.Version != "v1" || got.RequiredRules[1].Item.Source.ID != "local" {
		t.Fatal("lost provenance")
	}
	slices.Reverse(items)
	slices.Reverse(request.Files)
	again, e := Resolve(items, request)
	if e != nil || !reflect.DeepEqual(got, again) {
		t.Fatalf("unstable output %v", e)
	}
	empty, e := Resolve(nil, Request{})
	if e != nil {
		t.Fatal(e)
	}
	data, _ := json.Marshal(empty)
	if string(data) != `{"required_rules":[],"optional_rules":[],"skills":[],"references":[],"decisions":[]}` {
		t.Fatal(string(data))
	}
	data, _ = json.Marshal(got)
	if strings.Contains(string(data), "null") {
		t.Fatal("null array", string(data))
	}
	if _, e := Resolve([]knowledge.Item{a, a}, request); e == nil {
		t.Fatal("duplicate accepted")
	}
}
