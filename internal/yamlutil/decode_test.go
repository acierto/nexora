package yamlutil

import "testing"

func TestStrictDecode(t *testing.T) {
	type entry struct {
		Name    string   `yaml:"name"`
		Count   int      `yaml:"count"`
		Enabled *bool    `yaml:"enabled"`
		Values  []string `yaml:"values"`
	}
	for _, data := range []string{"name: [", "name: 42", "count: 1.2", "enabled: yes", "name: null", "values: [null]", "extra: value", "name: a\nname: b", "name: a\n---\nname: b", "values: scalar", "[]", "name: !!binary YQ==", "values: &a [*a]", ""} {
		t.Run(data, func(t *testing.T) {
			var value entry
			if Decode([]byte(data), &value) == nil {
				t.Fatalf("accepted %q", data)
			}
		})
	}
	var got entry
	if err := Decode([]byte("name: demo\ncount: 0\nenabled: false\nvalues: [go]"), &got); err != nil {
		t.Fatal(err)
	}
	if got.Enabled == nil || *got.Enabled || got.Count != 0 || got.Name != "demo" {
		t.Fatalf("%+v", got)
	}
	if err := Decode([]byte("name: &name demo\nvalues: [*name]"), &got); err != nil {
		t.Fatal(err)
	}
}
