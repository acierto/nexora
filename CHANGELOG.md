# Changelog

## v0.3.1 — 2026-10-29

- Explain curl certificate trust failures in the shared installer while retaining HTTPS and checksum verification.
- Document one installation shared by projects, direct MCP startup, client PATH, and warning-only handling of missing setup or partial context.
- Test the installer against real release archives and HTTP/HTTPS servers, including checksum and certificate failures.


## v0.3.0 — 2026-10-26

Nexora now works alongside an existing documentation MCP without requiring a
second index or documentation credentials.

- Add `delivery.documentation` to route agents to their existing documentation
  service. Guidance appears in native files and context responses. External-only
  configurations omit Nexora's documentation search from MCP tool discovery.
- Add `resolve --native` and MCP `get_context`'s `native` option. Current, unedited
  synced files are verified before rule and skill bodies are replaced with
  source/version citations and match explanations. Full delivery remains default.
- Preserve selected knowledge when documentation retrieval fails. Context JSON
  includes `documentation.status`; CLI emits partial context and exits 1, and MCP
  retains `structuredContent` with `isError: true`.
- Collapse exact linked passage copies with alternative source attribution and
  prefer live table content over a keyword snapshot for the same passage ID.
- Add a runnable existing-docs integration example, expand practical commerce
  knowledge and handbook fixtures, and clarify freshness and permission boundaries.

Upgrade: run sync again before using native mode. Native verification checks
files next to the config; use full context when syncing elsewhere. Context
consumers should inspect documentation status and handle partial results.
Nexora emits external lookup guidance; it does not proxy another MCP or implement
its authentication. Existing configs and default full-content requests remain valid.
