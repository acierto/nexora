package mcp

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/acierto/nexora/internal/app"
)

func exchange(t *testing.T, s *Server, lines ...string) []map[string]any {
	t.Helper()
	var out strings.Builder
	if err := s.Serve(context.Background(), strings.NewReader(strings.Join(lines, "\n")+"\n"), &out); err != nil {
		t.Fatal(err)
	}
	responses := []map[string]any{}
	for _, line := range strings.Split(strings.TrimSpace(out.String()), "\n") {
		if line == "" {
			continue
		}
		var m map[string]any
		if err := json.Unmarshal([]byte(line), &m); err != nil {
			t.Fatalf("invalid JSON from server: %q", line)
		}
		responses = append(responses, m)
	}
	return responses
}

func call(id int, name, args string) string {
	return `{"jsonrpc":"2.0","id":` + string(rune('0'+id)) + `,"method":"tools/call","params":{"name":"` + name + `","arguments":` + args + `}}`
}

func TestProtocol(t *testing.T) {
	repo, _ := filepath.Abs("../..")
	s := &Server{App: app.App{ConfigPath: filepath.Join(repo, "examples", "local", "nexora.yaml")}, Version: "test"}
	got := exchange(t, s,
		`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-03-26","capabilities":{},"clientInfo":{"name":"t","version":"1"}}}`,
		`{"jsonrpc":"2.0","method":"notifications/initialized"}`,
		`{"jsonrpc":"2.0","id":2,"method":"tools/list"}`,
		call(3, "get_context", `{"agent":"codex","files":["src/auth/service.go"]}`),
		call(4, "get_knowledge", `{"id":"go/conventions"}`),
		call(5, "get_knowledge", `{"id":"missing/item"}`),
		call(6, "list_skills", `{}`),
		call(7, "nope", `{}`),
		`{"jsonrpc":"2.0","id":8,"method":"resources/list"}`,
		`not json`,
		call(9, "get_context", `{"unknown":true}`),
	)
	if len(got) != 10 {
		t.Fatalf("expected 10 responses (no reply to the notification), got %d", len(got))
	}
	init := got[0]["result"].(map[string]any)
	if init["protocolVersion"] != "2025-03-26" || init["serverInfo"].(map[string]any)["name"] != "nexora" || init["capabilities"].(map[string]any)["tools"] == nil {
		t.Fatalf("initialize: %v", init)
	}
	tools := got[1]["result"].(map[string]any)["tools"].([]any)
	names := []string{}
	for _, tool := range tools {
		names = append(names, tool.(map[string]any)["name"].(string))
	}
	if strings.Join(names, ",") != "get_context,search_knowledge,get_knowledge,list_skills" {
		t.Fatalf("tools: %v", names)
	}
	ctxResult := got[2]["result"].(map[string]any)
	required := ctxResult["structuredContent"].(map[string]any)["required_rules"].([]any)
	ids := []string{}
	for _, r := range required {
		ids = append(ids, r.(map[string]any)["item"].(map[string]any)["id"].(string))
	}
	if strings.Join(ids, ",") != "security/authentication,go/conventions,agents/codex" || ctxResult["isError"] != false {
		t.Fatalf("get_context: %v", ids)
	}
	if got[3]["result"].(map[string]any)["structuredContent"].(map[string]any)["kind"] != "knowledge" {
		t.Fatalf("get_knowledge: %v", got[3])
	}
	missing := got[4]["result"].(map[string]any)
	if missing["isError"] != true || !strings.Contains(missing["content"].([]any)[0].(map[string]any)["text"].(string), "missing/item") {
		t.Fatalf("missing item: %v", missing)
	}
	skills := got[5]["result"].(map[string]any)["structuredContent"].(map[string]any)["skills"].([]any)
	skillIDs := []string{}
	for _, skill := range skills {
		skillIDs = append(skillIDs, skill.(map[string]any)["id"].(string))
	}
	if strings.Join(skillIDs, ",") != "skills/add-endpoint,skills/add-migration,skills/implement-checkout,skills/replay-webhook,skills/test-http-endpoint" {
		t.Fatalf("skills: %v", skills)
	}
	for idx, code := range map[int]float64{6: -32602, 7: -32601, 8: -32700} {
		if got[idx]["error"].(map[string]any)["code"] != code {
			t.Fatalf("response %d: %v", idx, got[idx])
		}
	}
	if got[9]["result"].(map[string]any)["isError"] != true {
		t.Fatalf("unknown argument accepted: %v", got[9])
	}
}

func TestFreshDataPerCall(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "knowledge"), 0o755)
	os.WriteFile(filepath.Join(dir, "nexora.yaml"), []byte("schema_version: 1\nproject: p\nsources:\n  - id: k\n    path: ./knowledge\n"), 0o644)
	rule := func(text string) {
		os.WriteFile(filepath.Join(dir, "knowledge", "r.md"), []byte("---\nid: team/rule\nname: Rule\ndescription: A rule\nkind: rule\n---\n"+text+"\n"), 0o644)
	}
	s := &Server{App: app.App{ConfigPath: filepath.Join(dir, "nexora.yaml")}}
	content := func() string {
		r := exchange(t, s, call(1, "get_knowledge", `{"id":"team/rule"}`))
		return r[0]["result"].(map[string]any)["structuredContent"].(map[string]any)["item"].(map[string]any)["content"].(string)
	}
	rule("first version")
	if !strings.Contains(content(), "first version") {
		t.Fatal("first read")
	}
	rule("second version")
	if !strings.Contains(content(), "second version") {
		t.Fatal("edit not visible without restart")
	}
}
