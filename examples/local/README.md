# Go API: select context for authentication work

This collection represents a commerce API. A developer changing authentication
needs a general Go baseline, a security rule, and an optional implementation
procedure. A retired rule should remain inspectable without affecting selection.

```text
local/
├── nexora.yaml
└── knowledge/
    ├── rules/           Go, authentication, agent reminders, optional, retired
    ├── skills/          add-endpoint/SKILL.md
    ├── references/      Authentication architecture
    └── decisions/       Why knowledge is kept in Git
```

From the Nexora repository root, with `nexora` built and on PATH:

```sh
nexora validate --config examples/local/nexora.yaml
nexora list --config examples/local/nexora.yaml
nexora resolve --config examples/local/nexora.yaml --agent codex --file src/auth/service.go
nexora resolve --config examples/local/nexora.yaml --agent claude --file src/auth/service.go --json
nexora resolve --config examples/local/nexora.yaml --file src/catalog/service.go --json
```

Validation reports 9 items. Expected IDs:

| Request | Required rules, in order | Other selected knowledge |
| --- | --- | --- |
| Codex + auth file | `security/authentication`, `go/conventions`, `agents/codex` | `go/naming`, `skills/add-endpoint`, `references/authentication`, `decisions/local-knowledge` |
| Claude + auth file | `security/authentication`, `go/conventions`, `agents/claude` | Same optional knowledge |
| No agent + catalog file | `go/conventions` | `go/naming`, `skills/add-endpoint`, `decisions/local-knowledge` |

`legacy/authentication` is always excluded because it is archived. The other
agent's reminder is excluded. With a catalog file, the authentication rule and
reference are excluded because their path restriction does not match.

The authentication selection explains project `commerce-api` and the pair
`src/auth/**/*.go` / `src/auth/service.go`. Provenance names source `local`,
path `rules/authentication.md`, version `1.0.0`, and the actual file digest.
Unrestricted dimensions are identified explicitly.

Try changing the authentication rule's paths to `src/**/*.go` and rerun the
catalog request. It will now select that rule. Restore the original pattern
afterward. This changes applicability; priority never overrides another rule.

The SKILL.md uses Nexora metadata; it is retrieved as text, not executed. Agent
filter examples do not configure or connect Codex or Claude Code.
