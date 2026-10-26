# task-retrieval Specification

## Purpose

Deliver documentation passages next to deterministically selected knowledge without weakening mandatory selection.

## Requirements

### Requirement: Search command
`nexora search --query TEXT` SHALL return up to `--limit` (default 5) passages for the configured or overridden project, as text or `--json` `{ "hits": [...] }`.

#### Scenario: Empty query
- **WHEN** `--query` is missing or blank
- **THEN** the command exits 2 with usage guidance

### Requirement: Resolve with a task
`nexora resolve --task TEXT` SHALL return the unchanged deterministic groups plus a `passages` array of retrieved hits. The `passages` array SHALL always be present and empty when no task is given. Passages MUST NOT add, remove, reorder, or demote any rule, skill, reference, or decision.

#### Scenario: Rules unaffected by retrieval
- **WHEN** the same request runs with and without `--task`
- **THEN** the five deterministic groups are identical

#### Scenario: Missing index
- **WHEN** documentation sources are configured, no index exists, and `--task` is given
- **THEN** resolve exits 1 asking to run `nexora index`

### Requirement: Passage project isolation
Passages from a source with `projects` SHALL be returned only for requests whose project is listed. Sources without `projects` SHALL be available to every project. This filter MUST NOT be described as authorization.

#### Scenario: Other project
- **WHEN** a passage belongs to a source restricted to `commerce-api` and the request project is `billing-api`
- **THEN** the passage is not returned
