# Design

## Context

`local-knowledge-registry` delivered strict front-matter knowledge, deterministic resolution, and a four-command CLI. The repository rules require that first use works without a database or embedding key, that mandatory selection never depends on similarity, that knowledge content is data, and that integrations are incremental. The installed toolchain is Go 1.23+, and the module floor stays `go 1.23.0`.

## Goals / Non-Goals

**Goals:** index team documentation and Confluence pages; retrieve passages relevant to a task with explanations and source links; serve rules and passages to agents through MCP; generate native instruction files without destroying manual edits; keep the offline, keyless path fully functional.

**Non-Goals:** per-viewer authorization, multi-user servers, background crawling, write-back to Confluence, plugin runtime, ranking that can drop or demote mandatory rules, or claims about how a given agent prioritizes the delivered text.

## Decisions

### 1. Source types extend configuration v1

Sources gain an optional `type`: `knowledge` (default, unchanged), `docs`, or `confluence`. `knowledge` and `docs` require `path` and keep the canonical, non-overlapping root rules. `confluence` requires `url`, a non-empty unique `spaces` list, and `token_env`; `user_env` is optional. Credentials are read from the named environment variables at index time only; configuration never stores secrets and still performs no variable substitution. `docs` and `confluence` sources accept `projects`, restricting their passages to those projects. Adding optional fields keeps `schema_version: 1`; older binaries reject them as unknown fields, which is the intended strict behavior.

A new optional `index` mapping configures `store` (`local` default, or `pgvector`), `path` (default `.nexora/index.json` relative to the config), `dsn_env`, `table`, `embeddings` (`provider` none/ollama/openai, `model`, `url`, `api_key_env`, `dimensions`), and `chunk` (`max_chars` 1200, `overlap_chars` 150).

### 2. Passages, not items

Documentation does not become knowledge items: it has no applicability metadata and must never be eligible for mandatory selection. A `passage.Passage` holds a stable ID (`<source>:<document>#<n>`), source ID and type, document ID, title, heading trail, text, location (relative path or page URL), digest of the original document, and allowed projects. The chunker keeps Markdown headings as boundaries, packs paragraphs up to `max_chars`, splits oversized paragraphs on sentence or word boundaries, and repeats up to `overlap_chars` from the previous passage. Leading YAML front matter in docs is skipped, not parsed. HTML and Confluence storage XHTML are converted to Markdown-like text with the standard library XML tokenizer in non-strict mode (headings, lists, tables, code macros), never executed or rendered.

`docs` discovery reuses the knowledge loader policy: deterministic walk, regular files only, symlink entries rejected, extensions `.md`, `.markdown`, `.txt`, `.html`, `.htm`, UTF-8 required.

### 3. Confluence through REST, explicit and read-only

The connector pages through `GET {url}/rest/api/content?spaceKey=…&type=page&status=current&expand=body.storage,version`, following `_links.next`. Basic auth is used when `user_env` is set (Cloud email + API token), otherwise a bearer token (Data Center personal access token). Requests honor context cancellation and fail on non-2xx with the status and space. Anything the token can read is indexed; this is documented as a single-user trust boundary.

### 4. Store interface with two implementations

`index.Store` has `Replace(sourceID, passages, meta)`, `Search(query)`, `Get(id)`, and `Stats()`. It is a real substitution boundary with two implementations:

- **local**: one JSON file written atomically (temp file + rename) holding passages, embeddings, and the embedding model. BM25 is computed at query time; cosine similarity is used when both the query and the passages have embeddings.
- **pgvector**: `lib/pq` (pure Go, no transitive dependencies) against a table with a generated `tsvector` column, a GIN index, an optional HNSW cosine index when `dimensions` is set, and a `<table>_meta` table storing the model. `Replace` deletes and inserts in one transaction. Keyword candidates use `websearch_to_tsquery('simple', …)` and `ts_rank_cd`; semantic candidates use `<=>`.

Both stores return keyword and semantic candidate lists; `retrieval.Fuse` combines them with reciprocal rank fusion (k = 60) in Go, so ranking behavior is identical across stores. Every hit carries its methods, per-method rank, score, and matched terms. Embedding model mismatches between index and query fail with a re-index instruction.

### 5. Retrieval is additive to deterministic resolution

`resolve --task` runs the unchanged resolver first, then retrieves passages filtered by the request project. The JSON envelope gains a sixth array, `passages`, which is always present (empty without `--task`). Passages are context with sources, not instructions. If documentation sources are configured but no index exists, `--task` fails with `run nexora index`; without documentation sources the group is empty.

### 6. MCP over stdio, hand-written

The protocol surface needed is small: `initialize`, `notifications/initialized`, `ping`, `tools/list`, `tools/call`. A hand-written newline-delimited JSON-RPC 2.0 loop avoids an SDK dependency and keeps the contract testable with pipes. The server negotiates `2025-06-18`, `2025-03-26`, or `2024-11-05`. Tool results include JSON text content and `structuredContent`; tool failures return `isError: true`, protocol failures return JSON-RPC errors. The registry and configuration are reloaded on every call, matching the CLI's no-stale-snapshot behavior. Logs go to stderr only.

### 7. Native sync writes only what it owns

`sync --agent codex` targets `AGENTS.md` and `.agents/skills/<name>/SKILL.md`; `--agent claude` targets `CLAUDE.md` and `.claude/skills/<name>/SKILL.md`, relative to `--dir` (default: the config directory). The instruction file gets one block between `<!-- nexora:begin -->` and `<!-- nexora:end -->` markers containing active required rules applicable to the project and agent: unrestricted rules under "Always apply", path-restricted rules under "When you change matching files" with their patterns, optional rules summarized, and a pointer to `nexora resolve` and the MCP server. Text outside the markers is preserved byte for byte; a file without markers gets the block appended. Skills become Agent Skills-compatible files whose `name` is the slugged ID. A manifest at `.nexora/sync-<agent>.json` records generated files and digests; a skill file that exists without a manifest entry, or whose digest changed, is a conflict and stops the sync without writing anything. Stale generated skills are removed only when unmodified. `--check` reports whether files are current and exits 1 when they are not.

## Risks / Trade-offs

- Keyword-only retrieval misses paraphrases → embeddings are one configuration block away; the default stays keyless.
- JSON index loads fully into memory → acceptable for thousands of passages; pgvector covers larger corpora.
- Confluence storage format is rich → conversion keeps text, headings, lists, tables, and code; macros without text are dropped.
- Indexed Confluence content is readable by anyone with the index file or database → documented; per-viewer authorization is deferred.
- Agents may ignore or reprioritize synced text → documentation states that sync delivers text and does not enforce it.

## Migration Plan

Existing configurations and knowledge load unchanged. The only visible contract change is the additional `passages` array in `resolve --json`. Index files and generated agent files are new and can be deleted to roll back.
