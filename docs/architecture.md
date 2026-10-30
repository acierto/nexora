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
- `index.VectorTable` reads an existing pgvector table (a `pgvector` source)
  without writing to it: `Rows` copies text into the store for keyword search and
  `Nearest` runs the semantic query in place. Rows are not re-chunked or
  re-embedded, so passage IDs map to row keys and only the query is embedded, with
  the model the source declares. Identifiers are validated in `config` and quoted
  with `pq.QuoteIdentifier`.
- `retrieval` merges keyword and semantic candidates, including one semantic list
  per pgvector source, with reciprocal rank fusion in Go, so every store ranks and
  explains hits identically. The index records the
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

## Complementary agent delivery

Nexora owns deterministic project knowledge. When a team already has a
documentation service, `delivery.documentation` supplies agent-facing lookup
guidance through native files and context responses. It does not add a connector,
copy an index, or forward credentials. Documentation authentication and permissions
stay with the existing service. External-only configurations omit the redundant
MCP search tool and direct explicit search attempts to the configured service.
Optional Nexora indexing remains useful for local
or separately owned corpora; an MCP proxy and hosted composition are deferred.

Explicit native context requests reuse `nativesync.Plan` as a read-only freshness
check in the config directory. Only current, unchanged generated files permit
required rule and skill bodies to be replaced by `native_knowledge` citations.
Those citations preserve selection explanations and source/version. Optional
rules, references, and decisions retain bodies. This avoids a second native
tracking mechanism, while requiring the caller to know that its agent loads the
files. No public extension contract or new dependency is introduced.

Documentation retrieval failure does not invalidate deterministic selection.
Context results include documentation status and a sanitized message; the CLI
emits the partial context with exit 1, and MCP sets isError while retaining
structuredContent. Invalid configuration or knowledge still fails completely.
Cancellation also remains a hard failure. Standalone search reports diagnostic
errors rather than partial hits.

Rank fusion prefers live table content when a keyword snapshot returns the same
ID. After ranking, exact linked copies share a slot while alternative citations
remain in duplicates. Matching requires the same HTTP(S) URL, heading, trimmed
body, and digest. Different versions and local paths are preserved. Snapshot
refresh and live authorization are independent concerns, not solved by dedup.

## Version-selected MCP runtimes

The installed CLI can bootstrap an exact release or explicitly select latest
through `mcp --version`. `internal/runtimeversion` owns immutable tag resolution,
platform archives, bounded HTTP reads, SHA-256 checksums, regular-file extraction,
atomic per-user caching and cached executable integrity. The CLI owns child
streams, startup diagnostics and exit propagation. Children receive a direct MCP
invocation without a selector, preventing recursion. No new dependency or public
extension interface is introduced.

A cached exact pin needs no network. Latest resolution is fresh per invocation
and never silently falls back. Project configuration selects policy without
embedding a downloader or warning hook. Setup warnings originate in Nexora;
missing bootstrap executables are necessarily client launch errors. Neither
failure establishes a permission gate for repository work. Local cache owners
remain trusted, and runtime selection does not add remote authorization.

## Repeatable CLI installation

The release-owned POSIX installer supports opt-in `--ensure` with a stable minimum
version. Compatibility applies only to the configured installation directory;
projects can launch that deterministic path without duplicating version selection.
The installer checks release checksums and the staged executable's reported version
before same-filesystem rename. Download, extraction or validation errors leave the
previous binary in place. Rechecking compatibility before publication allows a run
to reuse an installation completed while downloading; concurrent upgrades to
different pins are not serialized. Exact MCP runtime selection remains independent.
The installer is also published as a release asset, alongside archives and checksums.
