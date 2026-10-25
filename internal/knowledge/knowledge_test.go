package knowledge

import "testing"

func TestValidation(t *testing.T) {
	valid := Item{ID: "go/testing", Name: "Testing", Description: "Tests", Kind: Rule, Status: Active, Required: true, Content: "Test behavior"}
	for _, kind := range []Kind{Rule, Skill, Reference, Decision} {
		item := valid
		item.Kind = kind
		item.Required = kind == Rule
		if err := item.Validate(); err != nil {
			t.Fatal(err)
		}
	}
	for _, status := range []Status{Active, Draft, Archived} {
		item := valid
		item.Status = status
		if err := item.Validate(); err != nil {
			t.Fatal(err)
		}
	}
	for name, mutate := range map[string]func(*Item){
		"id": func(i *Item) { i.ID = "A" }, "segment": func(i *Item) { i.ID = "a/../b" }, "empty segment": func(i *Item) { i.ID = "a//b" },
		"name": func(i *Item) { i.Name = " " }, "description": func(i *Item) { i.Description = "" }, "content": func(i *Item) { i.Content = "" },
		"kind": func(i *Item) { i.Kind = "other" }, "status": func(i *Item) { i.Status = "other" }, "required": func(i *Item) { i.Kind = Skill },
		"empty applicability": func(i *Item) { i.AppliesTo.Agents = []string{""} }, "pattern": func(i *Item) { i.AppliesTo.Paths = []string{"a/**b"} },
	} {
		t.Run(name, func(t *testing.T) {
			item := valid
			mutate(&item)
			if item.Validate() == nil {
				t.Fatal("accepted invalid item")
			}
		})
	}
	valid.Priority = -10
	if err := valid.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestPatterns(t *testing.T) {
	for _, tc := range []struct {
		pattern, file string
		want          bool
	}{
		{"src/**/*.go", "src/main.go", true}, {"src/**/*.go", "src/auth/service.go", true}, {"src/*.go", "src/auth/service.go", false},
		{"**/test?.go", "test1.go", true}, {"**/test?.go", "x/test12.go", false}, {"**", "main.go", true}, {"src/?.go", "src/é.go", true},
		{"a/**/b/**/c", "a/b/c", true}, {"a/**/b/**/c", "a/x/b/y/c", true}, {"a/*", "b/a", false},
	} {
		t.Run(tc.pattern+tc.file, func(t *testing.T) {
			p, e := ParsePattern(tc.pattern)
			if e != nil {
				t.Fatal(e)
			}
			if p.Match(tc.file) != tc.want {
				t.Fatalf("match %q %q", tc.pattern, tc.file)
			}
		})
	}
	for _, p := range []string{"", "/a", "a//b", "../a", "./a", "a/**b", "[ab]", "{a,b}", "!a", `a\b`, "C:/a"} {
		if _, e := ParsePattern(p); e == nil {
			t.Errorf("accepted %q", p)
		}
	}
	for input, want := range map[string]string{`src\auth\a.go`: "src/auth/a.go", "./src/./main.go": "src/main.go"} {
		got, e := NormalizeFile(input)
		if e != nil || got != want {
			t.Errorf("%q: %q %v", input, got, e)
		}
	}
	for _, p := range []string{"", " ", ".", "/a", `C:\a`, "C:a", `\\server\a`, "a/../b", "../a"} {
		if _, e := NormalizeFile(p); e == nil {
			t.Errorf("accepted file %q", p)
		}
	}
}
