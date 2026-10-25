# MCP Server

## Purpose

Serve Nexora knowledge to MCP-capable agents over stdio.

## ADDED Requirements

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
