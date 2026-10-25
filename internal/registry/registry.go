// Package registry builds validated, read-only local knowledge snapshots.
package registry

import (
	"fmt"
	"sort"

	"github.com/acierto/nexora/internal/knowledge"
)

type Registry struct {
	items []knowledge.Item
	byID  map[string]int
}

func New(items []knowledge.Item) (*Registry, error) {
	r := &Registry{items: []knowledge.Item{}, byID: map[string]int{}}
	for _, item := range items {
		if err := item.Validate(); err != nil {
			return nil, fmt.Errorf("source %s, %s: %w", item.Source.ID, item.Source.Path, err)
		}
		if index, ok := r.byID[item.ID]; ok {
			previous := r.items[index]
			return nil, fmt.Errorf("duplicate id %q in %s:%s and %s:%s", item.ID, previous.Source.ID, previous.Source.Path, item.Source.ID, item.Source.Path)
		}
		r.byID[item.ID] = len(r.items)
		r.items = append(r.items, item.Clone())
	}
	sort.Slice(r.items, func(i, j int) bool { return r.items[i].ID < r.items[j].ID })
	for idx, item := range r.items {
		r.byID[item.ID] = idx
	}
	return r, nil
}

func (r *Registry) Get(id string) (knowledge.Item, bool) {
	idx, ok := r.byID[id]
	if !ok {
		return knowledge.Item{}, false
	}
	return r.items[idx].Clone(), true
}

func (r *Registry) List() []knowledge.Item {
	items := make([]knowledge.Item, 0, len(r.items))
	for _, item := range r.items {
		items = append(items, item.Clone())
	}
	return items
}
