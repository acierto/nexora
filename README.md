# Nexora

One knowledge layer. Every AI agent.

**[Read the Nexora docs online](https://acierto.github.io/nexora/)** ·
[Quickstart](https://acierto.github.io/nexora/#start) ·
[Interactive playground](https://acierto.github.io/nexora/#playground)

Your team's know-how is scattered across agent files (`CLAUDE.md`, `AGENTS.md`,
Cursor rules), skill folders, internal docs, and Confluence. Nexora is the one
place every agent gets it from:

- **Rules, skills, references, and decisions** live in Markdown with front
  matter and are selected deterministically by project, agent, and file.
- **Internal docs and Confluence pages** are indexed into passages and retrieved
  for a task with keyword search, optional embeddings (Ollama or OpenAI-compatible),
  and an optional pgvector store.
- **Agents get it directly** through an MCP server (`nexora mcp`) and through
  generated `AGENTS.md`/`CLAUDE.md` blocks and `SKILL.md` files (`nexora sync`).

Every selection explains why it was picked and where it came from. Required rules
never depend on search ranking. Nexora is a standalone Go CLI: the first run needs
no Docker, database, cloud account, or API key.

## Run locally

Build the standalone executable from this checkout:

```sh
go build -o bin/nexora ./cmd/nexora
./bin/nexora validate --config examples/local/nexora.yaml
./bin/nexora resolve --config examples/local/nexora.yaml --agent codex --file src/auth/service.go
```

The example selects three required rules: `security/authentication`,
`go/conventions`, and `agents/codex`. Each includes source, version, and matching
reasons. Then index a documentation folder and ask for a task:

```sh
./bin/nexora index --config examples/team-docs/nexora.yaml
./bin/nexora resolve --config examples/team-docs/nexora.yaml --agent codex \
  --file src/payments/refund.go --task "add partial refunds"
```

The result adds handbook passages about partial refunds, with their file and the
terms that matched, next to the payment rules. Knowledge stays on your machine.
Nexora uses the network only when `nexora index` reads Confluence or calls an
embedding provider, and when a configured provider embeds a search query.

To install into Go's binary directory, run `go install ./cmd/nexora` and put
`$(go env GOPATH)/bin` on PATH (or use your configured GOBIN). In an existing
project, `nexora init --project my-api` creates `nexora.yaml` and
`knowledge/rules/project.md`. It refuses existing targets and symlinked
destinations. Only `nexora sync` writes to `AGENTS.md` or `CLAUDE.md`, and only
inside its own marked block.

| Command | Purpose |
| --- | --- |
| `nexora init [--project ID]` | Create local configuration and an example rule |
| `nexora validate [--config PATH] [--json]` | Validate every source and report the count |
| `nexora list [--config PATH] [--json]` | Inspect all knowledge, including inactive items |
| `nexora resolve [--config PATH] [--project ID] [--agent ID] [--file PATH ...] [--task TEXT] [--limit N] [--json]` | Select applicable knowledge, plus documentation passages with `--task` |
| `nexora index [--config PATH] [--source ID] [--json]` | Index docs and Confluence sources |
| `nexora search --query TEXT [--config PATH] [--project ID] [--limit N] [--json]` | Search indexed documentation |
| `nexora mcp [--config PATH]` | Serve knowledge to agents over MCP (stdio) |
| `nexora sync --agent codex\|claude [--config PATH] [--project ID] [--dir PATH] [--check] [--json]` | Write knowledge into `AGENTS.md`/`CLAUDE.md` and skill folders |
| `nexora help [COMMAND]` | Explain commands and flags |

Read commands emit one JSON object with `--json`: validation returns `valid` and
`item_count`, listing returns `items`, resolution returns five selection arrays
plus `passages`, and search returns `hits`. Exit codes are 0 for success, 2 for usage errors, and 1 for configuration,
loading, resolution, or I/O failures. Errors go to stderr; JSON goes to stdout.

## Learn and try

- [Getting started](docs/getting-started.md): create and use your first rule.
- [Knowledge format](docs/knowledge-format.md): metadata, configuration, and examples of every kind.
- [Resolution](docs/resolution.md): applicability, ordering, provenance, and trust boundaries.
- [CLI guide](docs/cli.md): flags, recipes, JSON output, and scripting.
- [Docs and Confluence](docs/retrieval.md): documentation sources, indexing, search, embeddings, and pgvector.
- [Connect agents](docs/agents.md): the MCP server for Claude Code and Codex, and native sync.
- [Troubleshooting](docs/troubleshooting.md): common errors and missing-rule explanations.
- [Example catalog](examples/README.md): a Go API, shared knowledge across two projects, and searchable team docs.

Planned next: plugins and shareable knowledge packs, a remote MCP server with
authorization that mirrors source permissions, more connectors (Notion, Google
Docs), full Agent Skills import, overrides, and context budgets.

## Development

Go 1.23 or newer is required to build from source. The initial compatibility
floor is Go 1.23; supported production toolchains will be reviewed before release.

Implementation lives under `internal/`: knowledge models and path semantics,
configuration, Markdown loading, registry snapshots, resolution, documentation
passages, Confluence, embeddings, index stores, retrieval, MCP, native sync, and
CLI output.
The executable entry point is `cmd/nexora`. Public SDK contracts are deferred.

```sh
go test ./...
go vet ./...
```

See [architecture decisions](docs/architecture.md). This project is under
initial development. License selection is pending before public release.
