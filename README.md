# Nexora

One knowledge layer. Every AI agent.

**[Documentation](https://acierto.github.io/nexora/)** ·
[Install](https://acierto.github.io/nexora/#install) ·
[Quickstart](https://acierto.github.io/nexora/#start) ·
[Interactive playground](https://acierto.github.io/nexora/#playground)

Nexora keeps project rules, skills, references, and decisions in one versioned
knowledge registry and delivers applicable knowledge to coding agents. Required
rules match project, file, and agent metadata deterministically, with provenance.

Already have a documentation MCP? Keep it. Use Nexora for project knowledge and
your existing service for documentation search, authentication, and permissions.
[Connect an existing documentation service](docs/agents.md#keep-an-existing-documentation-mcp)
without copying its index or distributing database and embedding credentials.

## How it works

```text
Git knowledge --> Nexora --> Native instructions and skills --> Agent
                     |                                         |
                     +--> CLI / MCP: applicable context         +--> Existing docs MCP
                                                                    (optional)

Optional local docs / Confluence / pgvector --> Nexora retrieval --> Agent
```

- **Required rules are deterministic.** Search never decides mandatory inclusion.
- **Traceable context.** Every selected item includes its source, version, and why
  it matched.
- **Complementary delivery.** Native sync generates agent files. Opt-in native
  context cites verified synced rules and skills instead of repeating their bodies.
- **Optional documentation retrieval.** Index local docs for offline keyword
  search, or configure Confluence, embeddings, and pgvector where needed.
- **Explicit partial results.** A retrieval outage preserves selected project
  knowledge and reports unavailable documentation.

Nexora is one Go executable. Local knowledge needs no database, cloud account,
embedding API key, or documentation index.

## Install

macOS and Linux:

```shell
curl -fsSL https://raw.githubusercontent.com/acierto/nexora/main/scripts/install.sh | sh
```

Windows and other systems: `go install github.com/acierto/nexora/cmd/nexora@latest`
with Go 1.23 or newer. Details, manual downloads, and checksums are in
[Install](docs/install.md).

## Start with project knowledge

```shell
nexora init --project my-api
$EDITOR knowledge/rules/authentication.md
nexora validate
nexora sync --agent codex                       # or: --agent claude
nexora resolve --agent codex --file src/auth/service.go
```

Use the generated native files with your agent, or connect `nexora mcp` for
context lookup. If documentation already has a search service, configure
`delivery.documentation` with lookup guidance; no indexing step is needed.
The [agent integration guide](docs/agents.md) covers both delivery modes, and the
[Quickstart](docs/quickstart.md) adds optional Nexora documentation search.

## Documentation

**Using Nexora**

- [Install](docs/install.md): prebuilt binaries, the install script, Windows and other systems.
- [Quickstart](docs/quickstart.md): from zero to a connected agent.
- [Write knowledge](docs/knowledge-format.md): rules, skills, references, and decisions.
- [Docs and Confluence](docs/retrieval.md): index documentation, reuse existing vector tables, embeddings, pgvector.
- [Connect agents](docs/agents.md): MCP setup for Claude Code and Codex, and native sync.
- [Examples](examples/README.md): sample projects you can run.

**Reference**

- [Configuration](docs/configuration.md): every `nexora.yaml` field.
- [CLI](docs/cli.md): commands, flags, JSON output, and exit codes.
- [How rules are matched](docs/resolution.md): applicability, ordering, and explanations.
- [Troubleshooting](docs/troubleshooting.md): error messages and fixes.

**Contributing**

- [Contributing](docs/contributing.md): development setup, tests, docs, and releases.
- [Architecture](docs/architecture.md) and [verification record](docs/verification.md).

## Status

Nexora is in early development. Working today: knowledge selection, docs and
Confluence indexing with optional embeddings and pgvector, existing pgvector
tables as sources, the MCP server, and native sync for Claude Code and Codex. Planned: plugins and shareable knowledge
packs, a remote server with permission-aware access, more connectors (Notion,
Google Docs), overrides, and context budgets. License selection is pending before
the first public release.
