# Document Sources

## Purpose

Bring internal documentation and Confluence pages into Nexora as passages with provenance, without treating them as rules.

## ADDED Requirements

### Requirement: Typed source configuration
The system SHALL accept an optional source `type` of `knowledge` (default), `docs`, or `confluence`. `knowledge` and `docs` sources SHALL require `path` and SHALL follow the existing canonical, non-overlapping root rules. `projects` SHALL be accepted only on `docs` and `confluence` sources.

#### Scenario: Valid mixed configuration
- **WHEN** a configuration declares a knowledge source, a docs source, and a confluence source with spaces and token_env
- **THEN** it validates, and existing knowledge loading is unchanged

#### Scenario: Misplaced fields
- **WHEN** a knowledge source sets url, spaces, token_env, or projects
- **THEN** validation fails

### Requirement: Confluence source fields
`confluence` sources SHALL require `url` (http or https), a non-empty list of unique `spaces`, and `token_env`, and MUST NOT set `path`. Credentials SHALL be read only from the environment variables named by `token_env` and `user_env` and never stored in configuration.

#### Scenario: Invalid confluence source
- **WHEN** a confluence source omits spaces or token_env, repeats a space, sets path, or uses a non-HTTP URL
- **THEN** validation fails naming the source and field

#### Scenario: Missing credential at index time
- **WHEN** the variable named by token_env is not set and indexing runs
- **THEN** indexing fails naming the variable

### Requirement: Safe documentation discovery
Docs sources SHALL discover regular `.md`, `.markdown`, `.txt`, `.html`, and `.htm` files deterministically, reject symlink entries, require UTF-8, skip leading YAML front matter without parsing it, and never execute or render content.

#### Scenario: Mixed documentation tree
- **WHEN** a docs source contains Markdown, HTML, text, and unrelated files in nested folders
- **THEN** only supported files produce passages, in stable order, with slash-separated relative locations

#### Scenario: Symlink in documentation
- **WHEN** a docs source contains a symlink entry
- **THEN** indexing fails naming the source and path

### Requirement: Confluence page retrieval
The connector SHALL list current pages of every configured space through the REST content API with `body.storage` expanded, follow pagination links, use basic authentication when `user_env` is set and bearer authentication otherwise, and fail on non-success responses with the space and HTTP status. Storage XHTML SHALL convert to text preserving headings, paragraphs, lists, table cells, and code macro bodies.

#### Scenario: Paginated space
- **WHEN** a space returns pages across two responses linked by `_links.next`
- **THEN** all pages are indexed with their title, page URL, and version

#### Scenario: Authentication failure
- **WHEN** Confluence answers 401 for a space
- **THEN** indexing fails with the source, space, and status, and the previous index for that source is kept

### Requirement: Passage chunking with provenance
Documents SHALL split into passages at Markdown headings and paragraph boundaries, at most `max_chars` characters each, with up to `overlap_chars` characters repeated from the previous passage. Every passage SHALL record a stable ID `<source>:<document>#<n>`, source ID and type, title, heading trail, location, document digest, and allowed projects.

#### Scenario: Heading trail
- **WHEN** a paragraph sits under `# Payments` and `## Refunds`
- **THEN** its passage heading is `Payments › Refunds`

#### Scenario: Stable IDs
- **WHEN** the same unchanged document is indexed twice
- **THEN** passage IDs, texts, and digests are identical
