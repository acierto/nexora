# Proposal

## Why

The local registry selects rules, skills, references, and decisions deterministically, but most of what a team knows lives elsewhere: internal documentation and Confluence spaces. Agents cannot see that knowledge, so engineers paste it in by hand. Agents also still need the selected knowledge delivered to them without copy and paste. This change turns Nexora into the knowledge layer described in the project direction: one place that indexes team documentation, retrieves the passages relevant to a task, and serves rules plus passages to Claude Code, Codex, and other MCP clients.

## What Changes

- Add two source types beside the existing front-matter `knowledge` sources: `docs` (local Markdown, text, and HTML documentation without Nexora metadata) and `confluence` (pages from configured Confluence spaces read through the REST API).
- Split documentation into heading-aware passages that keep source, location, title, heading trail, and digest.
- Add an explicit `nexora index` step that builds a passage index. The default index is a local file and needs no database, network service, or API key; keyword (BM25) retrieval works offline.
- Add optional embeddings through Ollama or an OpenAI-compatible endpoint, and an optional PostgreSQL + pgvector store. Search fuses keyword and semantic rankings with reciprocal rank fusion and explains every hit.
- Add `nexora search` and a `--task` flag for `nexora resolve` that adds a `passages` group. Mandatory and optional knowledge selection stays deterministic; passages never become mandatory and never replace a rule.
- Add `nexora mcp`, a stdio Model Context Protocol server with `get_context`, `search_knowledge`, `get_knowledge`, and `list_skills` tools.
- Add `nexora sync --agent codex|claude`, which writes a managed block into `AGENTS.md` or `CLAUDE.md` and generates `SKILL.md` files, guarded by a manifest so manual edits are never overwritten.
- Defer: plugins and knowledge packs, remote/multi-user serving, authorization that mirrors Confluence permissions per viewer, HTTP MCP transport, more connectors (Notion, Google Docs), context budgets, and overrides.

## Capabilities

### New Capabilities

- `document-sources`: Docs and Confluence source configuration, safe discovery, conversion, and passage chunking with provenance.
- `knowledge-index`: Local and pgvector passage stores, optional embedding providers, model consistency checks, and hybrid retrieval with explanations.
- `task-retrieval`: `search` and `resolve --task` behavior, project isolation for passages, and the guarantee that retrieval never affects mandatory selection.
- `mcp-server`: Stdio MCP lifecycle and tool contracts.
- `native-sync`: Managed instruction blocks, generated skills, manifest-based conflict detection, and check mode.

### Modified Capabilities

None. The local registry capabilities from `local-knowledge-registry` are extended, not changed: knowledge sources, front matter, and the five deterministic groups behave as before.

## Impact

- New internal packages for passages, Confluence, embeddings, index stores, retrieval, MCP, and sync; new CLI commands and flags.
- One new runtime dependency, `github.com/lib/pq`, used only when the pgvector store is configured. Embedding and Confluence clients use the standard library.
- Network access happens only in `nexora index` (Confluence, embeddings) and in queries that embed the task text. `validate`, `list`, and `resolve` without `--task` stay offline.
- Generated files (`AGENTS.md`/`CLAUDE.md` blocks, `SKILL.md`, `.nexora/`) are written only by `nexora sync` and `nexora index`.
