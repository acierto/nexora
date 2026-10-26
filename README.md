# Nexora

One knowledge layer. Every AI agent.

**[Documentation](https://acierto.github.io/nexora/)** ·
[Install](https://acierto.github.io/nexora/#install) ·
[Quickstart](https://acierto.github.io/nexora/#start) ·
[Interactive playground](https://acierto.github.io/nexora/#playground)

Your team's know-how is scattered across agent files (`CLAUDE.md`, `AGENTS.md`,
Cursor rules), skill folders, internal docs, and Confluence. Nexora gives every
coding agent one place to get it from:

- **Your rules and skills**, written once in Markdown, reach Claude Code, Codex,
  and other agents. Each task gets exactly the rules that apply to its project,
  agent, and files.
- **Your docs and Confluence pages** are indexed, so the agent gets the passages
  that match what it is doing, with a link to the source.
- **Your agents connect directly** through an MCP server, or through generated
  `AGENTS.md`/`CLAUDE.md` sections and `SKILL.md` files.

Every result says why it was selected and where it came from. Nexora is one
executable: no Docker, database, cloud account, or API key needed to start.

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
- [Docs and Confluence](docs/retrieval.md): index documentation, embeddings, pgvector.
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
Confluence indexing with optional embeddings and pgvector, the MCP server, and
native sync for Claude Code and Codex. Planned: plugins and shareable knowledge
packs, a remote server with permission-aware access, more connectors (Notion,
Google Docs), overrides, and context budgets. License selection is pending before
the first public release.
