# complementary-delivery Specification

## Purpose

Deliver project knowledge alongside an existing documentation service while avoiding repeated native content and preserving valid knowledge during retrieval outages.

## Requirements

### Requirement: External documentation guidance
Nexora SHALL accept optional delivery.documentation text up to 2000 UTF-8 bytes and emit it in native instructions and context responses without contacting the named service.

#### Scenario: Existing documentation service
- **WHEN** a task is resolved with guidance and no Nexora document sources
- **THEN** documentation status is external, guidance is included, passages are empty, and no index or credentials are required

#### Scenario: Invalid guidance
- **WHEN** guidance is null, non-string, or over the size limit
- **THEN** configuration validation fails

### Requirement: External search ownership
With external guidance and no Nexora document sources, Nexora SHALL omit search_knowledge from MCP tool discovery and direct explicit search requests to the configured service through an actionable error.

#### Scenario: External-only tool discovery
- **WHEN** an agent lists tools for an external-only configuration
- **THEN** Nexora exposes knowledge lookup tools without advertising its own documentation search

### Requirement: Verified native citations
Opt-in native context SHALL require current unedited synced files and a manifest for the requested project and supported agent next to the configuration. Selected required rules and skills SHALL be cited with ID, kind, required status, source, version, match explanations, and native path instead of repeating bodies. Other knowledge SHALL retain full bodies.

#### Scenario: Current native files
- **WHEN** native context is requested after a successful sync for the same agent and project
- **THEN** every selected required rule and skill has a citation to a file containing its body, and optional rules, references, and decisions remain intact

#### Scenario: Missing or stale sync
- **WHEN** files are missing, knowledge changed, the project differs, or generated content was edited
- **THEN** native context fails with resync or full-context recovery guidance and suppresses no new mandatory content

### Requirement: Explicit partial context
A documentation retrieval failure SHALL preserve deterministic selected knowledge with documentation status unavailable. Resolve SHALL emit context and exit 1. MCP SHALL return the partial context in structuredContent with isError true. Partial messages SHALL avoid raw provider connection details. Invalid knowledge and configuration SHALL remain hard failures.

#### Scenario: Missing index
- **WHEN** context retrieval is requested before indexing
- **THEN** required rules and applicable skills remain available with an explicit unavailable status and indexing guidance

#### Scenario: Invalid knowledge
- **WHEN** mandatory knowledge fails validation
- **THEN** the request fails without partial context

#### Scenario: Empty successful search
- **WHEN** retrieval succeeds with no matching passages
- **THEN** documentation status is complete and passages is empty
