# Tasks

## 1. Go foundation and canonical model

- [x] 1.1 Initialize `github.com/acierto/nexora` with the Go compatibility floor from design.md; verify `go list -m` and record the initial package boundaries and build prerequisites in README.md without creating placeholder packages.
- [x] 1.2 Add internal knowledge kinds, statuses, items, applicability, and provenance with explicit JSON field names; verify table-driven tests cover the four kinds, required fields, ID grammar, statuses, priorities, and rule-only required metadata.
- [x] 1.3 Define and implement the portable path-pattern parser and request-path normalization used by validation and resolution; verify tests cover `*`, `?`, zero-or-more `**` segments, Windows separators, drive/UNC/absolute paths, traversal, and rejected glob syntax; document the grammar in docs/knowledge-format.md.

## 2. Configuration and strict YAML decoding

- [x] 2.1 Check official documentation for a maintained Go YAML parser, pin one dependency with strict fields and duplicate-key handling, and document the choice; verify malformed/type-mismatched YAML, duplicate keys, extra documents, and explicit nulls are rejected by focused tests.
- [x] 2.2 Implement configuration v1 decoding and validation for project and unique source IDs; verify tests reject unknown fields, unsupported versions, missing values, and duplicate IDs and resolve relative paths against the config directory rather than the working directory.
- [x] 2.3 Resolve and validate canonical source directories, rejecting missing, unreadable, identical, or nested roots; verify temporary-directory tests cover overlapping roots and valid independent shared roots, and document explicit configuration and shared-source behavior in docs/knowledge-format.md.

## 3. Markdown source loading

- [x] 3.1 Implement UTF-8 front matter extraction and metadata decoding without rendering Markdown; verify tests cover all four kinds, LF/CRLF, content preservation, defaults, explicit optional rules, malformed delimiters, missing metadata, invalid scalar types, unknown nested fields, and empty bodies.
- [x] 3.2 Implement deterministic recursive `.md` discovery under configured roots, rejecting symlink entries and propagating source/path-specific I/O errors; verify nested fixtures, ignored non-Markdown files, symlink files/directories, empty sources, and cancellation without returning partial success.
- [x] 3.3 Attach source IDs, slash-separated relative paths, raw-file SHA-256 digests, and explicit or digest-derived versions; verify repeat-load tests preserve identical values and changed bytes change the digest.
- [x] 3.4 Document configuration and front matter defaults, supported metadata, strict validation, symlink policy, and the current SKILL.md compatibility limitation in docs/knowledge-format.md; verify every valid document example loads and every invalid example fails as described.

## 4. Validated registry snapshot

- [x] 4.1 Implement registry construction across all sources with global duplicate-ID rejection naming both locations; verify tests cover duplicates within and across sources and prevent partially valid registries.
- [x] 4.2 Implement deterministic Get and List behavior with defensive copies of slice fields; verify tests cover found/missing IDs, empty snapshots, ascending ID order, and caller mutation isolation.

## 5. Deterministic resolution

- [x] 5.1 Implement active-status filtering and project/agent applicability with AND across dimensions and OR within each; verify table-driven tests cover unrestricted and empty lists, missing request values, exact case matching, different projects, and inactive items.
- [x] 5.2 Match normalized request files against validated patterns using the shared path semantics; verify tests cover missing files, repeated files, nonexistent but valid file names, nested and zero-depth matches, and invalid paths failing the entire request.
- [x] 5.3 Return required_rules, optional_rules, skills, references, and decisions with all matches and priority-descending/ID-ascending ordering; verify tests preserve multiple mandatory rules regardless of priority, exclude implicit overrides, and serialize empty groups as arrays.
- [x] 5.4 Add structured, stable status/project/agent/path selection reasons with source/version on every selected item; verify permutation tests produce identical results and sorted, deduplicated path evidence.
- [x] 5.5 Document applicability, grouping, priority, explanations, and the single-user trust boundary in docs/resolution.md; verify the documented examples against resolver fixtures and state that filters are not authorization and optional matches are not semantic recommendations.

## 6. Read-only CLI workflow

- [x] 6.1 Add a testable CLI runner and thin executable entry point using standard-library command parsing; verify runner tests cover help, unknown commands/flags, rejected positional arguments, stdout/stderr separation, and exit codes 0/1/2.
- [x] 6.2 Wire configuration, all source loaders, and registry validation into validate and list with --config; verify CLI tests cover explicit paths from another working directory, missing-config guidance, all statuses in list, item counts, and no partial output on invalid input.
- [x] 6.3 Wire resolve with --project, --agent, and repeated --file flags; verify CLI tests cover configured project defaults, invocation-only overrides, expected rule IDs and groups, and unchanged source/configuration bytes.
- [x] 6.4 Implement readable text and --json for read commands; verify JSON decoding and focused golden tests cover the specified envelopes, snake_case fields, empty arrays, full selected content/reasons, and absence of progress messages or absolute paths in successful provenance.
- [x] 6.5 Document command syntax, flags, JSON envelopes, and exit codes in README.md; verify help and documented command invocations match the executable behavior.

## 7. Safe initialization

- [x] 7.1 Implement init with an optional --project and directory-basename default, generating nexora.yaml and a scoped project/conventions rule; verify a temporary fresh project immediately passes validate and returns that rule through resolve.
- [x] 7.2 Add conflict and destination-component preflight, exclusive file creation, and cleanup limited to files created by the invocation; verify tests cover each existing target, symlink destinations, a write failure, and byte-for-byte preservation of native agent files and unrelated content.
- [x] 7.3 Document the init-generated layout and non-overwrite behavior in README.md; verify the documented fresh-project sequence succeeds and a repeated init fails without changing files.

## 8. User documentation and practical examples

- [x] 8.1 Create a clear README entry point and docs/getting-started.md covering Nexora's purpose, prerequisites, building/installing the current executable from source, and a copyable init -> edit a rule -> validate -> list -> resolve walkthrough with expected output; verify the full walkthrough in a fresh temporary project and link the reference guides and example catalog from README.md.
- [x] 8.2 Expand docs/knowledge-format.md into a user authoring guide with complete rule, optional rule, skill, reference, and decision examples, metadata/default tables, configuration/source layout, version/provenance behavior, and project/agent/path scoping; verify every valid sample loads and document that Nexora-formatted skills do not imply full Agent Skills import or execution support.
- [x] 8.3 Add docs/cli.md with command/flag tables and practical recipes for resolving a Go file, comparing codex/claude agent filters, overriding the project, using an explicit config path, and consuming --json in a shell script; verify recipes against the built binary, show representative outputs, and explain that agent filtering does not connect to or enforce instructions in an agent.
- [x] 8.4 Add examples/local as a realistic Go API knowledge collection containing a global rule, a project-and-path authentication rule, agent-specific rules, optional knowledge of each kind, and an inactive item; provide a README explaining the scenario, directory tree, commands, expected selected/excluded IDs, reasons, and a small change the reader can try; verify all commands and expected results through executable tests.
- [x] 8.5 Add examples/shared-knowledge with a shared standards directory and two project configurations using explicit non-overlapping local/shared sources; show reuse, project isolation, and additive project rules without override claims, including a README with copyable commands and expected IDs for both projects; verify both configurations validate, shared rules appear where applicable, and project-restricted rules do not leak into the other project's result.
- [x] 8.6 Add docs/troubleshooting.md covering missing config, malformed front matter, duplicate IDs, invalid patterns, symlink/overlap errors, init conflicts, and unexpectedly missing rules due to status or project/agent/file filters; verify each diagnosis and fix against a reproducible fixture and state current limitations, trusted-instruction boundaries, and deferred integrations in plain language.
- [x] 8.7 Add examples/README.md as a short catalog with learning goals and links, and reconcile the user guides with the feature-level documentation from earlier sections; verify local links resolve and run documented commands against their real example files in automated example checks so documentation does not drift from implementation.

## 9. Integrated acceptance

- [x] 9.1 Add executable-level integration tests for init -> validate -> list -> resolve and both documented example collections, including a second-project negative case and invalid-document failure; verify the built binary returns expected JSON and exit codes without an agent runtime, OpenSpec, or service credentials.
- [x] 9.2 Run formatting checks, `go test ./...`, `go vet ./...`, and `go build ./cmd/nexora`; verify linux/amd64, darwin/arm64, and windows/amd64 cross-builds and exercise the native example binary in a network-disabled environment, recording commands and outcomes.
- [x] 9.3 Review implementation against every scenario in the three capability specs, reconcile any documented limitations, and run `openspec validate local-knowledge-registry --strict --no-interactive`; mark only verified tasks complete and leave publication/license selection and future integrations outside this milestone.
