# Tasks

## 1. Configuration

- [x] 1.1 Add source `type` (`knowledge`, `docs`, `confluence`), Confluence fields (`url`, `spaces`, `user_env`, `token_env`), and source `projects`; verify tests accept a mixed configuration and reject misplaced fields, duplicate spaces, non-HTTP URLs, invalid environment variable names, and overlapping docs roots.
- [x] 1.2 Add the optional `index` mapping (store, path, dsn_env, table, embeddings, chunk) with defaults; verify tests cover local and pgvector defaults, provider URL defaults, explicit zero overlap, and every invalid combination.

## 2. Documentation sources and passages

- [x] 2.1 Implement docs discovery for `.md`, `.markdown`, `.txt`, `.html`, `.htm` with deterministic order, symlink rejection, UTF-8 validation, and front matter skipping; verify with nested fixtures, ignored files, invalid UTF-8, and a symlink.
- [x] 2.2 Implement HTML and Confluence storage conversion (headings, lists, tables, code macros; scripts, head, and macro parameters dropped); verify with HTML and storage-format fixtures.
- [x] 2.3 Implement heading-aware chunking with size limits, overlap, heading trails, stable IDs, and provenance; verify determinism, maximum length, and overlap markers.
- [x] 2.4 Implement the Confluence REST client with pagination, basic or bearer authentication, loop detection, and actionable HTTP errors; verify against an httptest server.

## 3. Index and retrieval

- [x] 3.1 Implement optional Ollama and OpenAI-compatible embedding clients with batching, response validation, and dimension checks; verify against httptest servers.
- [x] 3.2 Implement the local JSON store with atomic writes, BM25 keyword candidates, cosine candidates, project filtering, partial and full replacement, and model consistency checks; verify with unit tests.
- [x] 3.3 Implement the pgvector store with lib/pq, generated tsvector, GIN and HNSW indexes, a meta table, and transactional replacement; verify the same scenarios against PostgreSQL 16 with pgvector 0.6 (test gated by `NEXORA_TEST_DATABASE_URL`).
- [x] 3.4 Implement reciprocal rank fusion with per-method explanations and deterministic tie-breaking, and model checks at query time; verify ordering, reasons, limits, and mismatch errors.

## 4. CLI

- [x] 4.1 Add `index`, `search`, and `resolve --task/--limit`, always emitting the `passages` array; verify that the five deterministic groups are identical with and without `--task`, that a missing index fails with guidance, and that usage errors exit 2.
- [x] 4.2 Verify end to end with the team docs example (offline, keyless) and a fake Confluence server: indexing, search results, project isolation of passages, and keeping the previous index when indexing fails.

## 5. Agent delivery

- [x] 5.1 Implement the stdio MCP server (initialize, ping, tools/list, tools/call) with `get_context`, `search_knowledge`, `get_knowledge`, and `list_skills`; verify handshake, tool results, tool errors, protocol errors, and fresh data per call, in-process and through the built binary.
- [x] 5.2 Implement native sync for codex and claude: managed block, generated skills, manifest, conflict detection, stale skill removal, and `--check`; verify preservation of hand-written text, idempotency, manual-edit conflicts, untracked files, broken markers, and CLI exit codes.

## 6. Documentation and verification

- [x] 6.1 Add docs/retrieval.md and docs/agents.md, update README, getting started, knowledge format, resolution, CLI, troubleshooting, architecture, and the example catalog; add the runnable `examples/team-docs` collection with executed recipes.
- [x] 6.2 Run gofmt, `go vet ./...`, `go test ./...` (including pgvector), cross-builds for linux/amd64, darwin/arm64, and windows/amd64, and the keyless workflow with networking disabled; record results in docs/verification.md.
- [x] 6.3 Run `openspec validate document-retrieval-and-agent-delivery --strict --no-interactive`.
