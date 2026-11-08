## MODIFIED Requirements

### Requirement: Validated registry with provenance
Each loaded item SHALL include its source ID, slash-separated source-relative path, and a SHA-256 digest of its file bytes after CRLF line endings are normalized to LF. Item content SHALL use the same normalized line endings. Explicit versions SHALL be preserved; absent versions SHALL be represented as `sha256:<digest>`. Duplicate IDs across or within sources MUST reject the complete registry, naming both locations. Repeated loads of unchanged files SHALL produce the same item values and ID ordering.

#### Scenario: Duplicate knowledge ID
- **WHEN** two documents declare the same ID
- **THEN** registry construction fails and identifies both source locations

#### Scenario: Reproducible provenance
- **WHEN** unchanged source files are loaded in different filesystem traversal orders
- **THEN** items retain identical versions, digests, source locations, and ascending ID order

#### Scenario: CRLF checkout
- **WHEN** a document is checked out with CRLF line endings, as Git for Windows does by default
- **THEN** its content, digest and version equal those of the LF original
