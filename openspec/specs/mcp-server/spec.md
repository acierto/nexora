# mcp-server Specification

## Purpose

Serve Nexora knowledge to MCP-capable agents over stdio.

## Requirements

### Requirement: Stdio lifecycle
`nexora mcp` SHALL speak newline-delimited JSON-RPC 2.0 on stdin/stdout, answer `initialize` with a supported protocol version, server info, and the tools capability, accept `notifications/initialized`, answer `ping`, and write diagnostics only to stderr. Unknown methods SHALL return error -32601 and malformed JSON error -32700.

#### Scenario: Handshake
- **WHEN** a client sends `initialize` with protocol version `2025-06-18`
- **THEN** the server responds with that version, `serverInfo.name` `nexora`, and a `tools` capability

### Requirement: Knowledge tools
`tools/list` SHALL describe `get_context`, `search_knowledge`, `get_knowledge`, and `list_skills` with JSON Schema inputs. `get_context` SHALL return the same groups as `resolve --json` for the given project, agent, files, and task. `search_knowledge` SHALL return passage hits. `get_knowledge` SHALL return a knowledge item or passage by ID. `list_skills` SHALL list active applicable skills.

#### Scenario: Context for a file
- **WHEN** `get_context` is called with agent `codex` and file `src/auth/service.go` on the local example
- **THEN** required rules are `security/authentication`, `go/conventions`, `agents/codex`

### Requirement: Tool results and errors
Tool results SHALL include the JSON value as text content and as `structuredContent`. Invalid arguments and failures SHALL return a result with `isError: true` and a message; unknown tool names SHALL return JSON-RPC error -32602.

#### Scenario: Unknown item
- **WHEN** `get_knowledge` is called with an ID that does not exist
- **THEN** the result has `isError: true` and names the ID

#### Scenario: Unknown argument
- **WHEN** a tool receives an argument it does not define
- **THEN** the result has `isError: true`

### Requirement: Fresh data per call
The server SHALL reload configuration, knowledge, and the index for every tool call so that edits are visible without restarting.

#### Scenario: Edited rule
- **WHEN** a rule file changes between two calls
- **THEN** the second call returns the new content and version

### Requirement: Explicit runtime selection
The mcp command SHALL accept an optional --version selector. An exact release tag SHALL select only that release. latest SHALL resolve the current published release to an immutable tag on each startup without silently falling back to an older cached version. Omitting the selector SHALL retain direct installed-binary behavior.

#### Scenario: Exact pin
- **WHEN** a caller selects v0.4.0
- **THEN** the launcher serves using only the verified v0.4.0 runtime, regardless of the launcher version

#### Scenario: Latest
- **WHEN** latest resolves to v0.4.0
- **THEN** v0.4.0 is selected and reported on stderr, and a later startup resolves latest again

#### Scenario: Unsupported selector
- **WHEN** the selector is empty or not latest or a release tag
- **THEN** argument validation rejects it before network access or server startup

### Requirement: Verified shared runtime cache
Nexora SHALL verify release archive checksums before installing an executable into a shared per-user cache. Exact cached versions SHALL work offline after verifying cached executable integrity. Only supported platform archives and the regular executable entry SHALL be installed atomically. Failed downloads, checksums or cache integrity checks SHALL never launch a different version.

#### Scenario: Verified install and offline reuse
- **WHEN** a verified exact release has been cached and network access is unavailable
- **THEN** the same intact executable is used without contacting the release service

#### Scenario: Corrupted artifact
- **WHEN** a checksum or cached executable does not match
- **THEN** startup reports the failure and no executable is launched

#### Scenario: Missing platform or archive executable
- **WHEN** the platform is unsupported or the archive lacks a regular executable
- **THEN** startup reports the failure without installing a runtime

### Requirement: Upstream startup diagnostics
MCP startup failures SHALL emit an actionable warning on stderr and exit nonzero without writing non-protocol text to stdout. The warning SHALL explain that repository work can continue under trusted local instructions. No agent permission gate SHALL be installed by this runtime selection feature.

#### Scenario: Invalid setup
- **WHEN** configuration loading or runtime acquisition fails
- **THEN** stderr identifies the issue and recovery guidance, stdout stays empty, and the server exits without denying other agent tools
