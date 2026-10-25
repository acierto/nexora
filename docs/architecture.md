# Initial architecture decisions

[Read online](https://acierto.github.io/nexora/#architecture) · [Documentation home](https://acierto.github.io/nexora/)

Nexora runs as a standalone Go binary. `cmd/nexora` only handles process startup;
`internal/cli` coordinates configuration, loading, registry construction, and
resolution. `internal/knowledge` and `internal/resolver` have no YAML or I/O
dependencies. There is no public SDK yet.

Configuration and metadata share `internal/yamlutil`. The sole production
dependency is the YAML organization's maintained `go.yaml.in/yaml/v3`, pinned to
v3.0.5. Its [Node API](https://pkg.go.dev/go.yaml.in/yaml/v3#Node) lets us validate
types and duplicate/unknown keys before decoding, avoiding YAML-to-string
coercion and partially accepted metadata. We check all nested fields, reject
nulls and extra documents, and preserve explicit false/zero values.

Standard-library command parsing keeps the command layer small. Markdown is
never rendered or executed. Every invocation rebuilds an in-memory snapshot;
persistent indexing and extension interfaces await actual consumers and measured
needs.

## Documentation retrieval and agent delivery

The second milestone adds documentation passages, retrieval, and delivery to
agents without changing deterministic resolution:

```text
cmd/nexora -> internal/cli -> app, mcp, nativesync
app -> config, markdown, registry, resolver, passage, confluence, embed, index, retrieval
retrieval -> index, embed;  index -> passage;  confluence -> passage
```

- `passage` converts docs (Markdown, text, HTML) and Confluence storage XHTML into
  heading-aware passages with provenance. HTML is tokenized with the standard
  library XML decoder in non-strict mode; nothing is rendered or executed.
  Passages are a separate type from knowledge items, so documentation can never be
  selected as a mandatory rule.
- `confluence` and `embed` are small standard-library HTTP clients (Confluence REST
  content API; Ollama `/api/embed` and OpenAI-compatible `/embeddings`).
- `index.Store` is the one new interface, because there are two real
  implementations: a local JSON file (default, written atomically, BM25 at query
  time) and PostgreSQL with pgvector. The pgvector store uses `github.com/lib/pq`,
  a pure-Go driver with no transitive dependencies; it is the only dependency added.
- `retrieval` merges keyword and semantic candidates with reciprocal rank fusion
  in Go, so both stores rank and explain hits identically. The index records the
  embedding model, and mismatches fail instead of silently degrading.
- `app` coordinates loading for the CLI and MCP server and reloads everything on
  each call, keeping the no-stale-snapshot behavior of the first milestone.
- `mcp` is a hand-written stdio JSON-RPC loop. The protocol surface needed
  (initialize, ping, tools/list, tools/call) is small enough that an SDK
  dependency is not justified yet.
- `nativesync` owns only a marked block in `AGENTS.md`/`CLAUDE.md` and the skill
  files it generated, tracked by digest in `.nexora/sync-<agent>.json`.

Network access is limited to `nexora index` (Confluence, embeddings) and to query
embedding when a provider is configured. `validate`, `list`, and `resolve` without
`--task` remain offline.

See the [verification record](verification.md) for tested behavior and
platform-build limitations.
