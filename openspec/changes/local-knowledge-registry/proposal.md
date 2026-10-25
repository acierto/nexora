# Proposal

## Why

Nexora needs a useful local foundation before adding agent integrations or semantic retrieval. Developers should be able to keep knowledge in Markdown and reliably determine which rules apply to a project and its files without external services.

## What Changes

- Introduce a storage-independent knowledge model for rules, skills, references, and decisions with validated metadata and provenance.
- Load explicitly configured local Markdown sources into an in-memory registry on each invocation.
- Resolve mandatory rules and applicable optional knowledge using project, agent, and path constraints with stable ordering and selection explanations.
- Provide `init`, `validate`, `list`, and `resolve` commands with human-readable and JSON output.
- Provide automated behavior tests, a user quickstart, authoring and CLI guides, troubleshooting, and runnable Go-project and shared-knowledge examples with verified commands and expected results.
- Defer persisted indexing, keyword/semantic search, overrides, context budgets, full Agent Skills interoperability, MCP, native synchronization, databases, embedding providers, external plugins, HTTP services, and multi-user authorization.

## Capabilities

### New Capabilities

- `local-knowledge-loading`: Versioned configuration, canonical metadata, safe filesystem discovery, provenance, and validated registry construction.
- `context-resolution`: Deterministic applicability, mandatory inclusion, optional grouping, ordering, and explanations.
- `local-cli`: Safe project initialization and offline validation, listing, and resolution through a standalone CLI.

### Modified Capabilities

None. The repository has no existing capability specifications or implementation.

## Impact

- Add `cmd/nexora` and focused `internal` packages; no public Go SDK contracts yet.
- Add a Go module, one justified YAML parser dependency, test fixtures, example knowledge, and usage documentation.
- No runtime network services, credentials, databases, or Node.js dependencies. OpenSpec remains development tooling only.
- No existing files are migrated. Native agent instruction files remain under user control.
