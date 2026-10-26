# Design

## Context

The inspected repository contains project instructions and OpenSpec tooling only, with no Go module or existing APIs. The installed Go toolchain is 1.23.3. See proposal.md for motivation and specs/ for observable behavior. This design establishes an initial format rather than claiming compatibility with an existing Nexora contract.

## Goals / Non-Goals

**Goals:** Keep domain logic independent of parsing and terminal output; make selection reproducible and testable with plain Go values; deliver one offline workflow useful in a personal repository.

**Non-Goals:** No public SDK stability promise, long-running process, plugin lifecycle, persistent index, or interpretation of knowledge content as executable policy. Markdown files named SKILL.md can use the Nexora schema, but importing arbitrary Agent Skills metadata is a later adapter.

## Decisions

### 1. Internal packages with a one-way dependency graph

```text
cmd/nexora -> internal/cli -> config, markdown, registry, resolver
markdown, registry, resolver -> knowledge
```

`knowledge` owns Item, Kind, Status, Applicability, Source and field validation. `resolver` owns Request, Selection, Reason and Result. Neither imports YAML, filesystem loaders, or CLI code. `registry` is a validated in-memory snapshot with Get and List operations. `markdown.Load(ctx, source)` reads one source; the CLI coordinates sources and constructs the registry. `resolver.Resolve(items, request)` is pure. The CLI runner accepts arguments and stdout/stderr writers and returns an exit code; main handles process exit only.

Alternative: publish `pkg` interfaces and module registration now. Rejected because there are no external consumers or alternate implementations to validate those contracts. Extract interfaces when substitution becomes concrete.

### 2. Small Go and YAML dependency footprint

Use module path `github.com/acierto/nexora`, matching the inspected origin. Begin with a `go 1.23.0` compatibility floor so the installed toolchain can build it; reassess the supported release policy before public release. Use standard-library flag sets per command and no command framework initially. Select one maintained YAML parser during implementation after checking its official documentation for strict-field and duplicate-key behavior, pin its version, and record the choice. Do not add dependencies for Markdown rendering or embeddings.

Alternative: Cobra immediately. Defer until command count, completion, or shared flag needs justify it.

### 3. Versioned configuration and front matter

Configuration v1:

```yaml
schema_version: 1
project: commerce-api
sources:
  - id: local
    path: ./knowledge
```

No implicit environment substitution, home expansion, network URLs, or parent-directory discovery. Paths can be absolute or relative to the config directory, allowing explicitly configured shared knowledge outside the current repository. Canonical roots must not overlap.

Document v1:

```markdown
---
id: security/authentication
name: Authentication rules
description: Requirements for authentication changes
kind: rule
version: "1.0.0"
status: active
required: true
priority: 100
tags: [security, go]
applies_to:
  projects: [commerce-api]
  agents: [codex, claude]
  paths: ["src/**/*.go"]
---
Validate credentials and return consistent authentication errors.
```

Required scalar metadata and defaults follow local-knowledge-loading. Front matter begins at the start of the file and uses `---` delimiter lines; accept LF and CRLF. Reject duplicate mapping keys, additional YAML documents, null values for explicitly supplied fields, and unknown fields at every schema level. Missing optional metadata uses defaults. Use pointer/raw representations while decoding to distinguish absent values from explicit false or zero. Preserve body text after the closing delimiter and reject whitespace-only bodies. Require integer priority and string versions; reject coercion from another YAML scalar type.

Alternative: infer metadata from directories or import arbitrary Markdown. Rejected for the foundation because silent inference hides mistakes in mandatory rules.

### 4. Transactional snapshot loading and provenance

Walk each configured canonical directory, inspect symlinks before traversal, and reject symlink entries even when their targets remain inside the root. Read only regular `.md` files, with deterministic traversal and errors ordered by source ID and relative path. Reject I/O failures without publishing partial items. Validate all items and global ID uniqueness before constructing the snapshot. Keep files read-only.

Compute SHA-256 over original file bytes, store a lowercase hexadecimal digest and normalized source-relative path, and preserve an explicit version or derive `sha256:<digest>`. Do not include load time or filesystem timestamps in results. List returns copies, including copied slice fields, so callers cannot mutate registry state.

Alternative: persisted index. Deferred until repeated-load performance is measured; rereading avoids stale mandatory instructions.

### 5. Portable deterministic matching

Use exact project/agent matching and AND across dimensions. With no restrictions, a dimension matches; with a restriction and no request value, it does not. Parse path patterns once during validation. Support literals, `*`, `?`, and full `**` segments only, rejecting bracket, brace, negation, embedded `**`, empty segments and traversal syntax. Patterns use `/`; input file backslashes normalize to `/`. Reject absolute/drive-qualified/UNC paths, `..` segments, and empty file entries; strip leading `./` and normalize `.` segments. Files need not exist: requests describe applicability, not filesystem reads.

Implement segment matching without OS-dependent glob behavior. A `**` segment can consume zero segments, so `src/**/*.go` matches `src/main.go`. Deduplicate and sort normalized request files. Evaluate all active items, group by required/optional rule or other kind, and sort by descending priority then ID. Include all matches without implicit overrides.

Represent reasons as ordered structured entries: active status, project matching, agent matching, then path matching. Each dimension records whether it is unrestricted; restricted path evidence lists all matching pattern/file pairs in sorted, deduplicated order. This avoids explanation changes caused by map or input ordering.

Alternative: semantic ranking or priority-based replacement. Rejected because neither defines whether mandatory instructions apply.

### 6. CLI output is a deliberately small contract

Implement init, validate, list, resolve, and help. Each read command accepts `--config` and `--json`; resolve also accepts `--project`, `--agent`, and repeatable `--file`. Reject positional arguments and unknown flags. Help exits 0. Decode/validate the entire source collection on each read command.

JSON validation returns `{ "valid": true, "item_count": N }`; listing returns `{ "items": [...] }`; resolution returns the five arrays named in context-resolution. Selections use `{ "item": ..., "reasons": [...] }`, with provenance on the item. Use explicit snake_case JSON tags, non-null arrays, no timestamps, and no absolute local paths in successful source attribution. Errors use stderr and exit codes specified in local-cli, leaving stdout empty. Human output summarizes these same values. Treat this as an initial CLI contract, separate from any future public Go API.

For init, preflight both target files and every existing destination component before writing. Create targets exclusively and remove only files created by this invocation if a later write fails. Preserve unrelated directories and files; never add a force-overwrite flag in this milestone. The example rule ID is `project/conventions` and is scoped to the initialized project. Native AGENTS.md and CLAUDE.md are untouched.

### 7. User documentation and examples stay executable

README.md is the user entry point, linking docs/getting-started.md, docs/knowledge-format.md, docs/resolution.md, docs/cli.md, docs/troubleshooting.md, and examples/README.md. Keep feature-level documentation alongside implementation tasks; the dedicated documentation section turns these references into a connected learning path and adds practical recipes rather than postponing all documentation until the end.

Use examples/local for a realistic Go API scenario and examples/shared-knowledge for two projects reusing an explicitly configured shared source alongside separate local sources. Each example includes a short explanation, directory tree, runnable commands, expected selected and excluded IDs, and an exercise. Shared/project rules are additive; examples must not imply override support. Agent-specific examples demonstrate matching only, without claiming native agent integration.

Use example files themselves as executable acceptance fixtures, compare stable IDs and reason fields rather than brittle complete terminal transcripts, and check local documentation links. Run the fresh-project quickstart and troubleshooting reproductions against temporary directories. Document source installation until a release distribution actually exists; no documentation website or publishing pipeline is needed for this milestone.

## Risks / Trade-offs

- Strict metadata is less convenient for existing libraries → provide a small valid example and actionable errors; add import adapters later.
- Rereading all Markdown limits scale → measure with real knowledge collections before adding caching.
- Custom path semantics can surprise users → document the small grammar and test zero-depth recursive matching and Windows inputs.
- Local filters cannot protect confidential data from someone with filesystem access → document the single-user trust boundary and defer remote serving.
- Symlink rejection reduces convenience → allow explicitly configured canonical source roots; document concurrent hostile filesystem mutation as outside this local MVP's security guarantee.
- Initialization can fail during I/O → preflight conflicts, use exclusive creates, and test cleanup without deleting pre-existing content.
- A compatibility floor does not establish a production toolchain support policy → decide supported Go releases and license before publication.

## Migration Plan

There is no existing data or executable to migrate. Add packages only when used, then a runnable example and README. Verify the complete workflow before marking tasks done. Reverting implementation files leaves user-authored Markdown intact; no index or remote state needs rollback. Do not archive this change until implementation and its acceptance checks are complete.
