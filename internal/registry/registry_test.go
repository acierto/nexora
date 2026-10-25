package registry

import (
	"strings"
	"testing"

	"github.com/acierto/nexora/internal/knowledge"
)

func TestSnapshot(t *testing.T) {
	item := knowledge.Item{ID: "b", Kind: knowledge.Rule, Name: "B", Description: "B", Content: "B", Status: knowledge.Active, Tags: []string{"original"}, AppliesTo: knowledge.Applicability{Projects: []string{"api"}, Agents: []string{"codex"}, Paths: []string{"**"}}, Source: knowledge.Source{ID: "local", Path: "b.md"}}
	a := item.Clone()
	a.ID = "a"
	r, err := New([]knowledge.Item{item, a})
	if err != nil {
		t.Fatal(err)
	}
	item.Tags[0] = "mutated"
	if r.List()[0].ID != "a" {
		t.Fatal("not sorted")
	}
	got, ok := r.Get("b")
	if !ok || got.Tags[0] != "original" {
		t.Fatal("input mutation leaked")
	}
	got.Tags[0] = "changed"
	got.AppliesTo.Projects[0] = "changed"
	got.AppliesTo.Agents[0] = "changed"
	got.AppliesTo.Paths[0] = "changed"
	listed := r.List()
	listed[1].Tags[0] = "changed"
	again, _ := r.Get("b")
	if again.Tags[0] != "original" || again.AppliesTo.Projects[0] != "api" || again.AppliesTo.Agents[0] != "codex" || again.AppliesTo.Paths[0] != "**" {
		t.Fatal("output mutation leaked")
	}
	if _, ok := r.Get("absent"); ok {
		t.Fatal("missing item found")
	}
	for _, source := range []string{"local", "shared"} {
		dup := a.Clone()
		dup.Source = knowledge.Source{ID: source, Path: "duplicate.md"}
		if r, e := New([]knowledge.Item{a, dup}); e == nil || r != nil || !strings.Contains(e.Error(), "duplicate.md") || !strings.Contains(e.Error(), "b.md") {
			t.Fatalf("%v %v", r, e)
		}
	}
	invalid := a
	invalid.Name = ""
	if r, e := New([]knowledge.Item{a, invalid}); e == nil || r != nil {
		t.Fatal("accepted invalid item")
	}
	empty, e := New(nil)
	if e != nil || empty.List() == nil || len(empty.List()) != 0 {
		t.Fatal("empty snapshot")
	}
}
