# Nexora

One knowledge layer. Every AI agent.

Nexora is a local knowledge registry for AI development workflows. Keep rules,
skills, references, and decisions in Markdown, then select relevant context using
explicit project, agent, and file constraints.

The initial implementation is a standalone Go CLI. Docker, databases, cloud
accounts, embedding keys, and agent runtimes are not required. Agent filters
select knowledge; they do not connect to an agent or enforce instructions.

## Run locally

Build the standalone executable from this checkout:

```sh
go build -o bin/nexora ./cmd/nexora
./bin/nexora validate --config examples/local/nexora.yaml
./bin/nexora resolve --config examples/local/nexora.yaml --agent codex --file src/auth/service.go
```

The example selects three required rules: `security/authentication`,
`go/conventions`, and `agents/codex`. Each includes source, version, and matching
reasons. Knowledge stays on your machine; the CLI has no network operations.

To install into Go's binary directory, run `go install ./cmd/nexora` and put
`$(go env GOPATH)/bin` on PATH (or use your configured GOBIN). In an existing
project, `nexora init --project my-api` creates `nexora.yaml` and
`knowledge/rules/project.md`. It refuses existing targets and symlinked
destinations; it never changes AGENTS.md or CLAUDE.md.

| Command | Purpose |
| --- | --- |
| `nexora init [--project ID]` | Create local configuration and an example rule |
| `nexora validate [--config PATH] [--json]` | Validate every source and report the count |
| `nexora list [--config PATH] [--json]` | Inspect all knowledge, including inactive items |
| `nexora resolve [--config PATH] [--project ID] [--agent ID] [--file PATH ...] [--json]` | Select applicable knowledge |
| `nexora help [COMMAND]` | Explain commands and flags |

Read commands emit one JSON object with `--json`: validation returns `valid` and
`item_count`, listing returns `items`, and resolution returns five selection
arrays. Exit codes are 0 for success, 2 for usage errors, and 1 for configuration,
loading, resolution, or I/O failures. Errors go to stderr; JSON goes to stdout.

## Learn and try

- [Getting started](docs/getting-started.md): create and use your first rule.
- [Knowledge format](docs/knowledge-format.md): metadata, configuration, and examples of every kind.
- [Resolution](docs/resolution.md): applicability, ordering, provenance, and trust boundaries.
- [CLI guide](docs/cli.md): flags, recipes, JSON output, and scripting.
- [Troubleshooting](docs/troubleshooting.md): common errors and missing-rule explanations.
- [Example catalog](examples/README.md): a Go API and shared knowledge across two projects.

MCP, native instruction sync, full Agent Skills import, semantic search,
databases, remote access, overrides, and context budgets are future work.

## Development

Go 1.23 or newer is required to build from source. The initial compatibility
floor is Go 1.23; supported production toolchains will be reviewed before release.

Implementation lives under `internal/`: knowledge models and path semantics,
configuration, Markdown loading, registry snapshots, resolution, and CLI output.
The executable entry point is `cmd/nexora`. Public SDK contracts are deferred.

```sh
go test ./...
go vet ./...
```

See [architecture decisions](docs/architecture.md). This project is under
initial development. License selection is pending before public release.
