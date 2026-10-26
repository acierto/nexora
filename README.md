# Nexora

One knowledge layer. Every AI agent.

**[Documentation](https://acierto.github.io/nexora/)** ·
[Install](https://acierto.github.io/nexora/#install) ·
[Quickstart](https://acierto.github.io/nexora/#start) ·
[Interactive playground](https://acierto.github.io/nexora/#playground)

Nexora turns the docs and Confluence pages your team already maintains into
context for coding agents. It indexes them once. When Claude Code, Codex, or any
MCP client starts a task, Nexora returns your rules for those files and the
passages that matter, with sources. Nobody rewrites that knowledge into
`CLAUDE.md`, `AGENTS.md`, or Cursor rules again.

## How it works

```text
Index time   Confluence spaces ┐
             docs/ folders     ┴─ nexora index ─▶ passages: BM25 + optional embeddings
                                                  (local file or PostgreSQL + pgvector)
             existing pgvector table ─────────▶ text copied for BM25; vectors stay in place

Task time    agent ─ get_context(task, files) ─▶ Nexora ─▶ required rules  from knowledge/, matched by project, agent, file
             (MCP)                                        + top passages   from the index, keyword + vector, rank-fused
                                                          each with its source and the reason it was picked
```

- **Rules are never ranked away.** Required rules are selected from metadata;
  search only adds passages next to them.
- **Every item is traceable** to its file or page, version or digest, and the
  reason it matched.
- **Starts offline.** Keyword search needs no API key; add embeddings (Ollama or
  OpenAI-compatible) and pgvector when you want them.
- **Reuses what you embedded.** Point Nexora at an existing pgvector table and it
  searches those vectors in place, embedding only the query.
- **Also native.** `nexora sync` writes always-on rules into `AGENTS.md` or
  `CLAUDE.md` and generates `SKILL.md` files, without touching your own text.

Nexora is one executable: no Docker, database, cloud account, or API key needed
to start.

## Install

macOS and Linux:

```shell
curl -fsSL https://raw.githubusercontent.com/acierto/nexora/main/scripts/install.sh | sh
```

Windows and other systems: `go install github.com/acierto/nexora/cmd/nexora@latest`
with Go 1.23 or newer. Details, manual downloads, and checksums are in
[Install](docs/install.md).

## Connect it in five steps

```shell
nexora init --project my-api                 # 1. create nexora.yaml and knowledge/
$EDITOR knowledge/rules/authentication.md     # 2. write a rule (see the quickstart)
# 3. add your docs folder to nexora.yaml as a "docs" source
nexora index                                  # 4. index your docs
claude mcp add nexora -- nexora mcp --config "$PWD/nexora.yaml"   # 5. connect your agent
```

Now when the agent works on `src/auth/service.go`, it can ask Nexora for context
and get your authentication rule plus the doc passages about the task. The
[Quickstart](docs/quickstart.md) walks through each step with the exact files and
expected output.

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
