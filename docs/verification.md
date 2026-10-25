# Local MVP verification

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
