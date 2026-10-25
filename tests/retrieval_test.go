package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

type hit struct {
	Passage struct {
		ID       string `json:"id"`
		SourceID string `json:"source_id"`
		Location string `json:"location"`
		Heading  string `json:"heading"`
	} `json:"passage"`
	Reasons []struct {
		Method string   `json:"method"`
		Terms  []string `json:"terms"`
	} `json:"reasons"`
}

func copyTree(t *testing.T, from, to string) {
	t.Helper()
	err := filepath.WalkDir(from, func(file string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(from, file)
		if strings.HasPrefix(rel, ".nexora") {
			return nil
		}
		target := filepath.Join(to, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, err := os.ReadFile(file)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o644)
	})
	if err != nil {
		t.Fatal(err)
	}
}

func commandEnv(t *testing.T, dir string, env []string, stdin string, args ...string) (int, string, string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, executable, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), env...)
	cmd.Stdin = strings.NewReader(stdin)
	var out, errOut bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errOut
	err := cmd.Run()
	if ctx.Err() != nil {
		t.Fatal(ctx.Err())
	}
	if exit, ok := err.(*exec.ExitError); ok {
		return exit.ExitCode(), out.String(), errOut.String()
	}
	if err != nil {
		t.Fatal(err)
	}
	return 0, out.String(), errOut.String()
}

func TestTeamDocsExample(t *testing.T) {
	dir := t.TempDir()
	copyTree(t, filepath.Join(repository, "examples", "team-docs"), dir)

	code, _, errOut := command(t, dir, "resolve", "--task", "partial refunds")
	if code != 1 || !strings.Contains(errOut, "nexora index") {
		t.Fatalf("resolve --task before indexing: %d %s", code, errOut)
	}
	code, out, errOut := command(t, dir, "index", "--json")
	if code != 0 || errOut != "" {
		t.Fatal(code, out, errOut)
	}
	var report struct {
		Store   string `json:"store"`
		Model   string `json:"model"`
		Sources []struct {
			ID        string `json:"id"`
			Documents int    `json:"documents"`
			Passages  int    `json:"passages"`
		} `json:"sources"`
	}
	if err := json.Unmarshal([]byte(out), &report); err != nil || report.Store != "local" || report.Model != "" || len(report.Sources) != 1 || report.Sources[0].Documents != 4 || report.Sources[0].Passages != 13 {
		t.Fatalf("index report: %s %v", out, err)
	}
	if _, err := os.Stat(filepath.Join(dir, ".nexora", "index.json")); err != nil {
		t.Fatal(err)
	}

	search := func(args ...string) []hit {
		t.Helper()
		code, out, errOut := command(t, dir, append([]string{"search", "--json"}, args...)...)
		if code != 0 || errOut != "" {
			t.Fatal(code, out, errOut)
		}
		var r struct {
			Hits []hit `json:"hits"`
		}
		if err := json.Unmarshal([]byte(out), &r); err != nil {
			t.Fatal(err)
		}
		return r.Hits
	}
	hits := search("--query", "rollback production deploy", "--limit", "1")
	if len(hits) != 1 || hits[0].Passage.ID != "handbook:runbooks/deploy.html#3" || hits[0].Reasons[0].Method != "keyword" || !reflect.DeepEqual(hits[0].Reasons[0].Terms, []string{"rollback", "production", "deploy"}) {
		t.Fatalf("deploy search: %+v", hits)
	}
	hits = search("--query", "revoked refresh token reused", "--limit", "1")
	if len(hits) != 1 || hits[0].Passage.Location != "auth/token-rotation.md" || hits[0].Passage.Heading != "Token rotation › Rotating refresh tokens" {
		t.Fatalf("token search: %+v", hits)
	}

	args := []string{"resolve", "--json", "--agent", "codex", "--file", "src/payments/refund.go"}
	code, plain, _ := command(t, dir, args...)
	code2, withTask, errOut := command(t, dir, append(args, "--task", "add partial refunds", "--limit", "2")...)
	if code != 0 || code2 != 0 || errOut != "" {
		t.Fatal(code, code2, errOut)
	}
	var a, b map[string]json.RawMessage
	json.Unmarshal([]byte(plain), &a)
	json.Unmarshal([]byte(withTask), &b)
	for _, group := range []string{"required_rules", "optional_rules", "skills", "references", "decisions"} {
		if !bytes.Equal(a[group], b[group]) {
			t.Fatalf("%s changed when a task was added", group)
		}
	}
	var passages []hit
	json.Unmarshal(b["passages"], &passages)
	if string(a["passages"]) != "[]" || len(passages) != 2 || passages[0].Passage.ID != "handbook:payments/refunds.md#3" {
		t.Fatalf("passages: %s", b["passages"])
	}
	var required []struct {
		Item struct {
			ID string `json:"id"`
		} `json:"item"`
	}
	json.Unmarshal(b["required_rules"], &required)
	if len(required) != 2 || required[0].Item.ID != "payments/idempotency" || required[1].Item.ID != "observability/logging" {
		t.Fatalf("required rules: %s", b["required_rules"])
	}

	for _, bad := range [][]string{{"search"}, {"search", "--query", "x", "--limit", "0"}, {"resolve", "--limit", "51"}, {"sync"}, {"sync", "--agent", "cursor"}} {
		if code, _, _ := command(t, dir, bad...); code != 2 {
			t.Fatalf("%v: expected usage error, got %d", bad, code)
		}
	}
	if code, _, errOut := command(t, dir, "index", "--source", "rules"); code != 1 || !strings.Contains(errOut, "not indexed") {
		t.Fatalf("indexing a knowledge source: %d %s", code, errOut)
	}
}

func TestConfluenceIndexing(t *testing.T) {
	var fail atomic.Bool
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if fail.Load() || !ok || user != "me@example.com" || pass != "token-123" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		space := r.URL.Query().Get("spaceKey")
		page := map[string]any{"id": "42", "title": "Chargeback handling", "version": map[string]any{"number": 7},
			"body":   map[string]any{"storage": map[string]any{"value": "<h1>Chargebacks</h1><h2>Evidence</h2><p>Submit delivery evidence within seven days of a chargeback notice.</p>"}},
			"_links": map[string]any{"webui": "/spaces/" + space + "/pages/42"}}
		json.NewEncoder(w).Encode(map[string]any{"results": []any{page}, "_links": map[string]any{"base": server.URL + "/wiki"}})
	}))
	defer server.Close()

	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "knowledge"), 0o755)
	config := "schema_version: 1\nproject: commerce-api\nsources:\n  - id: local\n    path: ./knowledge\n  - id: wiki\n    type: confluence\n    url: " + server.URL + "/wiki\n    spaces: [PAY]\n    user_env: TEST_CONFLUENCE_USER\n    token_env: TEST_CONFLUENCE_TOKEN\n    projects: [commerce-api]\n"
	os.WriteFile(filepath.Join(dir, "nexora.yaml"), []byte(config), 0o644)
	env := []string{"TEST_CONFLUENCE_USER=me@example.com", "TEST_CONFLUENCE_TOKEN=token-123"}

	if code, _, errOut := commandEnv(t, dir, nil, "", "index"); code != 1 || !strings.Contains(errOut, "TEST_CONFLUENCE_TOKEN") {
		t.Fatalf("missing token: %d %s", code, errOut)
	}
	if code, out, errOut := commandEnv(t, dir, env, "", "index"); code != 0 || !strings.Contains(out, "wiki (confluence): 1 documents, 1 passages") {
		t.Fatal(code, out, errOut)
	}
	searchFor := func(project string) []hit {
		code, out, errOut := commandEnv(t, dir, env, "", "search", "--json", "--query", "chargeback evidence", "--project", project)
		if code != 0 {
			t.Fatal(code, out, errOut)
		}
		var r struct {
			Hits []hit `json:"hits"`
		}
		json.Unmarshal([]byte(out), &r)
		return r.Hits
	}
	hits := searchFor("commerce-api")
	if len(hits) != 1 || hits[0].Passage.ID != "wiki:page-42#1" || hits[0].Passage.Location != server.URL+"/wiki/spaces/PAY/pages/42" || hits[0].Passage.Heading != "Chargebacks › Evidence" {
		t.Fatalf("confluence hit: %+v", hits)
	}
	if hits := searchFor("billing-api"); len(hits) != 0 {
		t.Fatalf("project-restricted passage leaked: %+v", hits)
	}
	fail.Store(true)
	if code, _, errOut := commandEnv(t, dir, env, "", "index"); code != 1 || !strings.Contains(errOut, "401") || !strings.Contains(errOut, "space PAY") {
		t.Fatalf("auth failure: %d %s", code, errOut)
	}
	if hits := searchFor("commerce-api"); len(hits) != 1 {
		t.Fatal("a failed index run must keep the previous index")
	}
}

func TestMCPAndSyncExecutables(t *testing.T) {
	input := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18"}}` + "\n" +
		`{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"get_context","arguments":{"agent":"claude","files":["src/auth/service.go"]}}}` + "\n"
	code, out, errOut := commandEnv(t, repository, nil, input, "mcp", "--config", "examples/local/nexora.yaml")
	if code != 0 || errOut != "" {
		t.Fatal(code, out, errOut)
	}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) != 2 || !strings.Contains(lines[0], `"protocolVersion":"2025-06-18"`) || !strings.Contains(lines[1], `agents/claude`) {
		t.Fatalf("mcp output: %s", out)
	}

	dir := t.TempDir()
	copyTree(t, filepath.Join(repository, "examples", "team-docs"), dir)
	if code, out, errOut := command(t, dir, "sync", "--agent", "claude"); code != 0 || !strings.Contains(out, "create           CLAUDE.md") || !strings.Contains(out, ".claude/skills/issue-refund/SKILL.md") {
		t.Fatal(code, out, errOut)
	}
	if code, out, _ := command(t, dir, "sync", "--agent", "claude", "--check"); code != 0 || !strings.Contains(out, "up to date") {
		t.Fatal(code, out)
	}
	rule := filepath.Join(dir, "knowledge", "rules", "logging.md")
	data, _ := os.ReadFile(rule)
	os.WriteFile(rule, bytes.Replace(data, []byte("Never log card data"), []byte("Never log card numbers"), 1), 0o644)
	if code, out, _ := command(t, dir, "sync", "--agent", "claude", "--check"); code != 1 || !strings.Contains(out, "would update     CLAUDE.md") {
		t.Fatal(code, out)
	}
	if code, out, _ := command(t, dir, "sync", "--agent", "claude", "--check", "--json"); code != 1 || !strings.Contains(out, `"action": "update"`) {
		t.Fatal(code, out)
	}
}
