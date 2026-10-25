# Local MVP verification

[Read online](https://acierto.github.io/nexora/#verification) · [Documentation home](https://acierto.github.io/nexora/)

Verified on 2026-10-25 with Go 1.23.3 on macOS arm64. This records the initial
implementation check, not a promise of ongoing compatibility on every platform.

## Checks performed

| Check | Result |
| --- | --- |
| `gofmt -l cmd internal tests` | No unformatted files |
| `go test ./...` | All package and executable integration tests passed |
| `go vet ./...` | Passed |
| `go build -o bin/nexora ./cmd/nexora` | Native standalone binary built |
| `GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bin/nexora-linux-amd64 ./cmd/nexora` | Built ELF amd64 executable |
| `GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -o bin/nexora-darwin-arm64 ./cmd/nexora` | Built Mach-O arm64 executable |
| `GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o bin/nexora-windows-amd64.exe ./cmd/nexora` | Built PE32+ amd64 executable |
| `openspec validate local-knowledge-registry --strict --no-interactive` | Passed |
| `git diff --check` | Passed |

Linux and Windows binaries were cross-compiled, not executed on those operating
systems. Native tests include Windows-style request paths. Generated binaries
are ignored under bin/ and are not release artifacts.

## Offline execution

Run the native executable with network access denied by the macOS sandbox:

```sh
/usr/bin/sandbox-exec -p '(version 1) (allow default) (deny network*)' ./bin/nexora validate --config examples/local/nexora.yaml
/usr/bin/sandbox-exec -p '(version 1) (allow default) (deny network*)' ./bin/nexora resolve --config examples/local/nexora.yaml --agent codex --file src/auth/service.go --json
```

Both commands passed. All four commands (init, validate, list, resolve) also
passed under the same profile in a fresh temporary project. The local example
selected `security/authentication`, `go/conventions`, and `agents/codex`.
The shared commerce example selected `security/authentication`, `go/conventions`,
and `commerce/errors`; billing selected `go/conventions` and `billing/errors`.

The sandbox command is a macOS verification technique, not a runtime dependency.
Other operating systems can run the same example binary in their own network
isolation environment. No external services, keys, OpenSpec process, or agent
runtime participates in CLI execution.

## Specification coverage

| Scenarios | Verification source |
| --- | --- |
| Configuration location, invalid schemas/fields, independent and overlapping roots | [Configuration tests](../internal/config/config_test.go) |
| Strict scalar types, duplicate keys, nulls, extra documents | [YAML tests](../internal/yamlutil/decode_test.go) |
| All knowledge kinds, defaults, malformed documents, safe discovery, unreadable files, provenance | [Markdown tests](../internal/markdown/load_test.go) |
| Authoring-guide samples load or fail as documented | [Documentation sample tests](../internal/markdown/docs_test.go) |
| Duplicate IDs, deterministic listing, immutable snapshot values | [Registry tests](../internal/registry/registry_test.go) |
| Portable path patterns and invalid request paths | [Knowledge tests](../internal/knowledge/knowledge_test.go) |
| Applicability, mandatory preservation, optional grouping, order, stable evidence | [Resolver tests](../internal/resolver/resolve_test.go) |
| Non-destructive initialization, I/O cleanup, help, read commands, JSON envelopes and exit codes | [CLI tests](../internal/cli/run_test.go) |
| Fresh-project walkthrough, local/shared examples, project isolation, documented recipes and local links | [Executable integration tests](../tests/integration_test.go) |

The first release remains local and single-user. Native agent integration,
full Agent Skills imports, databases, semantic retrieval, and remote authorization
remain deferred as specified in the proposal. License selection and publication
are separate work.

## Documentation retrieval and agent delivery

Verified on 2026-10-25 for the `document-retrieval-and-agent-delivery` change,
with Go 1.24.7 on linux/amd64 building the module at its `go 1.23.0` floor,
PostgreSQL 16.15, and pgvector 0.6.0.

| Check | Result |
| --- | --- |
| `gofmt -l .` | No unformatted files |
| `go vet ./...` | Passed |
| `go test ./...` with `NEXORA_TEST_DATABASE_URL` set | All packages passed, including the pgvector store test against a real database |
| `GOOS=linux/darwin/windows … CGO_ENABLED=0 go build ./cmd/nexora` | ELF amd64, Mach-O arm64, and PE32+ amd64 executables built |
| `openspec validate document-retrieval-and-agent-delivery --strict --no-interactive` | Valid |
| `openspec validate local-knowledge-registry --strict --no-interactive` | Valid |

Coverage added by this change:

| Behavior | Verification source |
| --- | --- |
| Source types, Confluence fields, index and chunk settings | [Source configuration tests](../internal/config/sources_test.go) |
| Docs discovery, HTML and Confluence storage conversion, chunking | [Passage tests](../internal/passage/passage_test.go) |
| Confluence pagination, basic and bearer auth, errors | [Confluence tests](../internal/confluence/confluence_test.go) |
| Ollama and OpenAI-compatible embeddings | [Embedding tests](../internal/embed/embed_test.go) |
| Local and pgvector stores, project filtering, model checks | [Index tests](../internal/index/index_test.go) |
| Rank fusion, explanations, determinism, model mismatch | [Retrieval tests](../internal/retrieval/retrieval_test.go) |
| MCP handshake, tools, errors, fresh data per call | [MCP tests](../internal/mcp/server_test.go) |
| Managed block, skills, manifest conflicts, stale files | [Sync tests](../internal/nativesync/nativesync_test.go) |
| Team docs example, Confluence indexing, MCP and sync binaries | [Retrieval integration tests](../tests/retrieval_test.go) |

The keyless workflow (`validate`, `index`, `search`, `resolve --task`, `sync`,
and an MCP `initialize`) was run with the team docs example inside a network
namespace with no network access (`unshare -rn`) and passed.

Embedding providers and Confluence were exercised against local HTTP test servers
that follow their documented request and response formats, not against a live
Confluence instance, Ollama, or the OpenAI API. Retrieval quality was checked on
the bundled example only; measure it on your own documentation before relying on
ranking. The MCP server was tested with JSON-RPC exchanges over stdio; check your
agent's MCP client configuration with your own setup.

