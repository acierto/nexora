package tests

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/acierto/nexora/internal/app"
)

func TestComplementaryDelivery(t *testing.T) {
	for _, agent := range []string{"codex", "claude"} {
		t.Run(agent, func(t *testing.T) {
			workspace := t.TempDir()
			dir := filepath.Join(workspace, "existing-docs")
			copyTree(t, filepath.Join(repository, "examples", "existing-docs"), dir)
			copyTree(t, filepath.Join(repository, "examples", "local", "knowledge"), filepath.Join(workspace, "local", "knowledge"))
			guidance := "Use the connected team-docs MCP to search documentation and open authoritative pages."
			args := []string{"resolve", "--agent", agent, "--file", "src/orders/checkout.go", "--task", "ambiguous authorization", "--json"}
			read := func(native bool) (int, app.ContextResult, string) {
				t.Helper()
				a := append([]string{}, args...)
				if native {
					a = append(a, "--native")
				}
				code, out, errOut := command(t, dir, a...)
				var r app.ContextResult
				if out != "" {
					if err := json.Unmarshal([]byte(out), &r); err != nil {
						t.Fatal(out, err)
					}
				}
				return code, r, errOut
			}
			if code, _, errOut := read(true); code != 1 || !strings.Contains(errOut, "up-to-date") {
				t.Fatal(code, errOut)
			}
			if code, out, errOut := command(t, dir, "sync", "--agent", agent); code != 0 {
				t.Fatal(code, out, errOut)
			}
			instruction := "AGENTS.md"
			if agent == "claude" {
				instruction = "CLAUDE.md"
			}
			generated, err := os.ReadFile(filepath.Join(dir, instruction))
			if err != nil || !strings.Contains(string(generated), guidance) || strings.Contains(string(generated), "--task") {
				t.Fatal(string(generated), err)
			}
			code, full, errOut := read(false)
			if code != 0 || errOut != "" || full.Documentation.Status != "external" || full.Documentation.Guidance != guidance || len(full.Passages) != 0 {
				t.Fatal(code, full, errOut)
			}
			code, native, errOut := read(true)
			if code != 0 || errOut != "" || len(native.RequiredRules) != 0 || len(native.Skills) != 0 {
				t.Fatal(code, native, errOut)
			}
			if !reflect.DeepEqual(full.OptionalRules, native.OptionalRules) || !reflect.DeepEqual(full.References, native.References) || !reflect.DeepEqual(full.Decisions, native.Decisions) {
				t.Fatal("native mode lost knowledge not fully synced")
			}
			selected := append(full.RequiredRules, full.Skills...)
			if len(native.NativeKnowledge) != len(selected) {
				t.Fatal("selected native knowledge missing")
			}
			for i, s := range selected {
				n := native.NativeKnowledge[i]
				if n.ID != s.Item.ID || n.Version != s.Item.Version || n.Source != s.Item.Source || n.Required != s.Item.Required || !reflect.DeepEqual(n.Reasons, s.Reasons) {
					t.Fatalf("native citation mismatch: %+v", n)
				}
				body, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(n.File)))
				if err != nil || !strings.Contains(string(body), strings.TrimSpace(s.Item.Content)) {
					t.Fatalf("native citation does not point to delivered body: %+v %v", n, err)
				}
			}
			input := `{"jsonrpc":"2.0","id":1,"method":"tools/list"}` + "\n" +
				`{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"get_context","arguments":{"agent":"` + agent + `","native":true,"files":["src/orders/checkout.go"],"task":"ambiguous authorization"}}}` + "\n"
			code, out, errOut := commandEnv(t, dir, nil, input, "mcp")
			lines := strings.Split(strings.TrimSpace(out), "\n")
			if code != 0 || errOut != "" || len(lines) != 2 || strings.Contains(lines[0], "search_knowledge") {
				t.Fatal("external-only MCP should not advertise a duplicate documentation search", code, out, errOut)
			}
			var response struct {
				Result struct {
					IsError    bool              `json:"isError"`
					Structured app.ContextResult `json:"structuredContent"`
				} `json:"result"`
			}
			if err := json.Unmarshal([]byte(lines[1]), &response); err != nil || response.Result.IsError || !reflect.DeepEqual(response.Result.Structured, native) {
				t.Fatal("MCP and CLI native delivery differ", out, err)
			}
			if code, _, errOut := command(t, dir, "search", "--query", "authorization"); code != 1 || !strings.Contains(errOut, guidance) {
				t.Fatal("standalone search did not point to the existing service", code, errOut)
			}
			other := "claude"
			if agent == other {
				other = "codex"
			}
			if code, _, _ := command(t, dir, append(args, "--native", "--agent", other)...); code != 1 {
				t.Fatal("accepted sync for another agent")
			}
			// Switching project or agent must not reuse the other sync context.
			if code, _, _ := command(t, dir, append(args, "--native", "--project", "billing-api")...); code != 1 {
				t.Fatal("accepted sync for a different project")
			}
			// A knowledge edit must invalidate suppression until sync runs again.
			rule := filepath.Join(workspace, "local", "knowledge", "rules", "tenant-isolation.md")
			body, _ := os.ReadFile(rule)
			if err := os.WriteFile(rule, append(body, []byte("\nNew tenant constraint.\n")...), 0o644); err != nil {
				t.Fatal(err)
			}
			if code, _, _ := read(true); code != 1 {
				t.Fatal("stale sync suppressed new mandatory rule")
			}
			if code, _, _ := read(false); code != 0 {
				t.Fatal("full context unavailable with stale sync")
			}
			if code, out, errOut := command(t, dir, "sync", "--agent", agent); code != 0 {
				t.Fatal(code, out, errOut)
			}
			if code, _, _ := read(true); code != 0 {
				t.Fatal("resync did not restore native mode")
			}
			// Hand edits in a generated file must also invalidate suppression.
			current, err := os.ReadFile(filepath.Join(dir, instruction))
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(dir, instruction), []byte(strings.Replace(string(current), "## Team knowledge", "## Edited team knowledge", 1)), 0o644); err != nil {
				t.Fatal(err)
			}
			if code, _, _ := read(true); code != 1 {
				t.Fatal("accepted an edited native block")
			}
		})
	}
}

func TestPartialContextKeepsMandatoryKnowledge(t *testing.T) {
	dir := t.TempDir()
	copyTree(t, filepath.Join(repository, "examples", "team-docs"), dir)
	args := []string{"resolve", "--file", "src/payments/refund.go", "--task", "partial refunds", "--json"}
	code, out, errOut := command(t, dir, args...)
	var result app.ContextResult
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatal(out, err)
	}
	if code != 1 || !strings.Contains(errOut, "nexora index") || result.Documentation.Status != "unavailable" || len(result.Passages) != 0 {
		t.Fatal(code, result, errOut)
	}
	expectIDs(t, result.RequiredRules, "payments/idempotency", "observability/logging")
	expectIDs(t, result.Skills, "skills/issue-refund")
	// MCP must carry the same partial result, not replace it with an error string.
	input := `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"get_context","arguments":{"files":["src/payments/refund.go"],"task":"partial refunds"}}}` + "\n"
	code, out, errOut = commandEnv(t, dir, nil, input, "mcp")
	var response struct {
		Result struct {
			IsError    bool              `json:"isError"`
			Structured app.ContextResult `json:"structuredContent"`
		} `json:"result"`
	}
	if err := json.Unmarshal([]byte(out), &response); err != nil {
		t.Fatal(out, err)
	}
	if code != 0 || errOut != "" || !response.Result.IsError || !reflect.DeepEqual(response.Result.Structured, result) {
		t.Fatal(code, out, errOut)
	}
	// An invalid knowledge item remains a hard error, not a partial success.
	if err := os.WriteFile(filepath.Join(dir, "knowledge", "invalid.md"), []byte("invalid"), 0o644); err != nil {
		t.Fatal(err)
	}
	if code, out, _ := command(t, dir, args...); code != 1 || out != "" {
		t.Fatal("invalid mandatory knowledge returned a partial result", code, out)
	}
}
