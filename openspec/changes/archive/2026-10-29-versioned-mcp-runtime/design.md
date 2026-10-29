# Design

The installed CLI acts as the bootstrap executable. Only mcp --version activates runtime selection; the child receives mcp --config without a selector to prevent recursion. Existing mcp calls remain direct. Exact pins and latest are explicit; latest queries release metadata every startup, and network failures never silently use stale latest state. No automatic fallback, runtime authentication or tool permission gate is introduced.

An internal runtimeversion package owns platform mapping, HTTPS requests, bounded release/checksum reads, regular-file extraction, SHA-256 validation and atomic cache installation. It returns a verified executable path and resolved tag. The CLI owns subprocess streams and exit codes. Runtime cache integrity is checked even for offline reuse. HTTP/cache injection exists only at actual test boundaries. The normal Go system trust store remains enabled; no TLS bypass or new dependency is added.

Nexora owns startup failure warnings, including invalid configuration. A missing launcher executable is necessarily reported by the MCP client; projects need no hook implementation. Crystal declares an exact release selector in its MCP config and may explicitly choose latest. The loader does not claim sandboxing of malicious local cache owners.

Grill: ownership, selector validation, offline pins, fresh latest resolution, checksum failure, archive entry boundaries, cache corruption, concurrent installs, stdout protocol purity, subprocess exit propagation and rollback reviewed. No open product decisions. Version selectors were explicitly requested by the user; exact pin is the project default. Rollback removes the optional flag and uses an explicitly installed compatible binary.
