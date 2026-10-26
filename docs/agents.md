# Connect agents

[Read online](https://acierto.github.io/nexora/#agents) · [Documentation home](https://acierto.github.io/nexora/)

Nexora delivers deterministic project knowledge through native files, its CLI,
and a stdio MCP server. Choose one primary route for rule and skill bodies.
Documentation retrieval is optional; an existing documentation service can keep
owning search, freshness, authentication, and permissions.

## Keep an existing documentation MCP

Start with this arrangement when your agent already has documentation access:

```text
Git knowledge --> Nexora sync --> Native rules and skills --> Agent
                                                              |
                                                              +--> Existing docs MCP
```

Configure only your knowledge source and agent-facing lookup guidance:

```yaml
schema_version: 1
project: commerce-api
sources:
  - id: local
    path: ./knowledge
delivery:
  documentation: Use the connected team-docs MCP to search documentation and open authoritative pages.
```

Then run:

```shell
nexora sync --agent codex
nexora resolve --agent codex --native --file src/orders/checkout.go --task "recover an authorization timeout" --json
```

Use `claude` for both commands when syncing Claude's files. The generated native
block tells the agent where to look up documentation. Context responses also
include that guidance. With a task and no Nexora document sources,
`documentation.status` is `external` and `passages` is empty. Nexora does not call,
proxy, authenticate to, or configure the other MCP. Connect that service through
its own supported setup; let it enforce the caller's document permissions.

No production database connection, embedding provider, Confluence token, or second
index is required by Nexora for this arrangement. You can still add distinct
local docs as Nexora sources; those are searched independently and external
guidance remains available. Give each corpus one primary search owner to avoid
asking two services for the same documents.

Try the [runnable existing-docs example](../examples/existing-docs/README.md).
When external guidance is configured without Nexora document sources, MCP tool
listing omits `search_knowledge` so the agent sees only the existing service's
search tool. Direct Nexora search requests point to the configured service.
Restart the MCP connection after changing this setup so clients refresh their
tool discovery. A combined call that delegates search is future work.

## Avoid repeating native knowledge

Full context remains the default. Use `resolve --native` or MCP `get_context`
with `native: true` only when the receiving agent loads Nexora's synced files.
Supply `agent: codex` or `agent: claude`.

Before suppressing any body, Nexora computes a read-only sync plan for the
requested project and agent next to `nexora.yaml`. All generated files and the
manifest must be current and unedited. Missing files, changed knowledge, a
different project, or hand edits fail the request. Resync or omit native mode to
get full context. Syncing to a different `--dir` is supported, but native context
verification currently checks only the config directory.

Selected required rules and skills move into `native_knowledge` citations with
ID, kind, required status, version, source/digest, applicability reasons, and the
native file path. Their normal arrays are empty because their bodies are already
in the verified files. Optional rule bodies, references, and decisions remain in
the normal arrays: native sync does not deliver their complete content.

Verification proves the files are current; it cannot prove an agent loaded or
followed them. Native mode is explicit and never selected automatically.

## MCP server

`nexora mcp` speaks the Model Context Protocol over stdio (newline-delimited
JSON-RPC 2.0). It supports protocol revisions `2025-06-18`, `2025-03-26`, and
`2024-11-05`, writes diagnostics only to stderr, and reloads configuration,
knowledge, and the index on every call, so edits show up without a restart.

Pass an absolute `--config` path, because agents may start the server from
another directory.

Claude Code:

```shell
claude mcp add nexora -- nexora mcp --config /path/to/repo/nexora.yaml
```

Codex (`~/.codex/config.toml`):

```toml
[mcp_servers.nexora]
command = "nexora"
args = ["mcp", "--config", "/path/to/repo/nexora.yaml"]
```

If a documentation source needs credentials (Confluence, OpenAI embeddings, a
pgvector DSN), the server process needs the same environment variables as the
CLI. Pass them through your client's MCP environment settings.

### Tools

All tools are read-only.

| Tool | Arguments | Returns |
| --- | --- | --- |
| `get_context` | `project`, `agent`, `files`, `task`, `limit`, `native` (all optional) | The same groups as `nexora resolve --json`, including `passages` when `task` is set |
| `search_knowledge` | `query` (required), `project`, `limit` | `{ "hits": [...] }` documentation passages with explanations |
| `get_knowledge` | `id` (required) | `{ "kind": "knowledge", "item": ... }` or `{ "kind": "passage", "passage": ... }` |
| `list_skills` | `project`, `agent` | Active skills that apply, with their path restrictions |

`project` defaults to the configured project. Results include the JSON as text
content and as `structuredContent`. Invalid arguments and failures, such as an
unknown ID or a missing index, come back as tool errors (`isError: true`) with a
message the agent can act on. `get_context` also preserves selected knowledge
when documentation fails: its result has `isError: true`, usable
`structuredContent`, and `documentation.status: unavailable`. This is a partial
result, not a successful complete lookup. Fix the retrieval problem before
assuming documentation is covered. Configuration and knowledge validation errors
remain hard errors without partial context.

Initialization guidance explains native delivery, external documentation lookup,
and partial results. Required knowledge remains subject to the receiving agent's
trusted instructions; passages are reference material.

## Native sync

```shell
nexora sync --agent codex            # AGENTS.md and .agents/skills/<name>/SKILL.md
nexora sync --agent claude           # CLAUDE.md and .claude/skills/<name>/SKILL.md
nexora sync --agent codex --check    # write nothing; exit 1 if anything is out of date
nexora sync --agent codex --dir ../service --project billing-api
```

Files are written relative to `--dir`, which defaults to the directory of the
configuration file.

**Instruction file.** Nexora owns one block between `<!-- nexora:begin -->` and
`<!-- nexora:end -->`. Everything outside the markers is preserved byte for byte;
a file without markers gets the block appended. The block contains active
required rules that apply to the project and agent:

- rules without path restrictions under **Always apply**, with their content;
- path-restricted rules under **When you change matching files**, with their
  patterns and content;
- optional rules as a one-line summary each;
- the generated skills;
- a pointer to verified native context lookup and either configured external
  documentation guidance or Nexora task retrieval.

Each rule shows its source and version. Rules restricted to another agent are
left out.

**Skills.** Every active skill that applies to the project and agent becomes an
[Agent Skills](https://agentskills.io) compatible `SKILL.md` with `name` and
`description` front matter. The name is the ID without a leading `skills/`
segment, with other separators turned into hyphens, so `skills/issue-refund`
becomes `issue-refund`.

**Protecting your edits.** Sync records what it wrote, with digests, in
`.nexora/sync-<agent>.json`. It stops without writing anything when:

- a skill file it would write already exists and was not generated by Nexora;
- a generated skill file or the managed block was edited by hand since the last
  sync.

Generated skills that no longer apply are removed only if they are unchanged.
Edit the knowledge files, not the generated output, then sync again. Commit the
generated files and the manifest together. In CI, `nexora sync --check` fails when
someone changed knowledge without syncing.

## What this does not do

Nexora delivers text. It does not control how an agent weighs that text against
its own instructions, and it cannot force an agent to call a tool. Native files
follow each agent's own discovery and precedence rules. `--agent` values match
knowledge metadata; they are not authentication. The server is a local, single-user
stdio process. A remote HTTP server with authorization is future work.
