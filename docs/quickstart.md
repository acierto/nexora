# Quickstart

[Read online](https://acierto.github.io/nexora/#start) · [Documentation home](https://acierto.github.io/nexora/)

In about ten minutes you will have a project rule and your team's docs available
to Claude Code or Codex. When the agent works on a file, it gets the rules that
apply to that file and the documentation passages that match its task, with links
back to the source.

You need Nexora installed ([Install](install.md)) and a project, for example a Git
repository. Run every command from the project root.

## 1. Create the configuration

```sh
nexora init --project commerce-api
```

This creates two files and changes nothing else:

```text
nexora.yaml                  # which project this is and where knowledge lives
knowledge/rules/project.md   # an example rule for this project
```

Use your own project name instead of `commerce-api`.

## 2. Write a rule

A rule is a Markdown file with a short header that says when it applies. Save this
as `knowledge/rules/authentication.md`:

```markdown
---
id: security/authentication
name: Authentication rules
description: Requirements for authentication code
kind: rule
priority: 100
applies_to:
  projects: [commerce-api]
  paths: ["src/auth/**/*.go"]
---
Never log credentials. Test expired and invalid credentials.
```

This rule applies only to Go files under `src/auth/`. Rules without `paths` apply
to every file. [Write knowledge](knowledge-format.md) explains every field and the
other kinds: skills, references, and decisions.

## 3. Add your docs

Point Nexora at a folder of documentation you already have. Markdown, plain text,
and HTML work as they are, with no special headers. This example creates a small
folder and adds it to `nexora.yaml`:

```sh
mkdir -p docs/handbook
printf '# Token rotation\n\nRefresh tokens rotate on every use. Reusing a revoked refresh token revokes the whole token family.\n' > docs/handbook/tokens.md
cat >> nexora.yaml <<'EOF'
  - id: handbook
    type: docs
    path: ./docs/handbook
EOF
```

For a real project, use your own folder in `path`, such as `./docs`. To add
Confluence spaces, see [Docs and Confluence](retrieval.md).

## 4. Index and check

```sh
nexora validate
nexora index
nexora resolve --file src/auth/service.go --task "rotate refresh tokens"
```

You should see:

- `Valid: 2 knowledge items` and one documentation source;
- `handbook (docs): 1 documents, 1 passages`;
- under **Required rules**, `security/authentication` and then `project/conventions`,
  each with the reason it was selected;
- under **Passages**, `Token rotation` from `tokens.md`, with the matched words.

The file does not have to exist: `--file` describes what you are about to change.
Add `.nexora/index.json` to `.gitignore`; it is the local index.

## 5. Connect your agent

Give the agent access to Nexora through MCP. Use the absolute path of your
`nexora.yaml`.

Claude Code:

```shell
claude mcp add nexora -- nexora mcp --config "$PWD/nexora.yaml"
```

Codex, in `~/.codex/config.toml`:

```toml
[mcp_servers.nexora]
command = "nexora"
args = ["mcp", "--config", "/absolute/path/to/your/project/nexora.yaml"]
```

Optionally, also write the rules that always apply into the agent's own
instruction file, so it sees them before calling any tool:

```shell
nexora sync --agent claude    # CLAUDE.md and .claude/skills/
nexora sync --agent codex     # AGENTS.md and .agents/skills/
```

## What you get

Ask the agent to work on authentication, for example "add refresh token rotation
to src/auth/service.go". It can call the `get_context` tool and receive:

- the required rules for that file (`security/authentication`, `project/conventions`);
- skills, references, and decisions that apply;
- documentation passages about token rotation, with their file or page link.

When you edit a rule or a doc, the next call sees the change; run `nexora index`
again after changing documentation. Nothing leaves your machine unless you
configure Confluence or an embedding service.

## Next steps

- [Write knowledge](knowledge-format.md): rules, skills, references, and decisions.
- [Docs and Confluence](retrieval.md): Confluence spaces, embeddings, and pgvector.
- [Connect agents](agents.md): MCP tools and native sync in detail.
- [Examples](../examples/README.md): complete sample projects you can run.
