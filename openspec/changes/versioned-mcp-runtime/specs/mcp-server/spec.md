## ADDED Requirements

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
