# Local Knowledge Loading

## Purpose

Let developers maintain validated knowledge in local Markdown files with portable configuration and traceable sources, without external services.

## ADDED Requirements

### Requirement: Versioned local configuration
The system SHALL load YAML configuration with `schema_version: 1`, a nonempty default `project`, and one or more local sources with unique IDs and paths. Relative source paths SHALL resolve against the configuration directory. Invalid configuration, unknown fields, unsupported versions, and duplicate source IDs MUST fail before knowledge is exposed.

#### Scenario: Configuration independent of working directory
- **WHEN** an explicit configuration path names a file whose source is `./knowledge`
- **THEN** the source resolves relative to that file, regardless of the invocation directory

#### Scenario: Invalid configuration
- **WHEN** configuration contains an unsupported schema version, an unknown field, or duplicate source IDs
- **THEN** loading fails with the configuration location and reason

### Requirement: Canonical Markdown metadata
The system SHALL load UTF-8 `.md` files with YAML front matter containing nonempty `id`, `name`, `description`, and `kind` plus a nonempty Markdown body. Kinds SHALL be `rule`, `skill`, `reference`, or `decision`. Optional fields SHALL be `version`, `status`, `tags`, `priority`, `required`, and `applies_to` containing `projects`, `agents`, and `paths`. Unknown fields and invalid field types MUST fail.

#### Scenario: Valid knowledge of each kind
- **WHEN** valid documents of all four kinds are loaded
- **THEN** their metadata and Markdown bodies are available without rendering or executing the content

#### Scenario: Malformed document
- **WHEN** a Markdown file has missing front matter, an empty required field, an invalid kind, invalid UTF-8, or malformed YAML
- **THEN** the entire load fails with the source ID, relative file path, and reason

### Requirement: Metadata defaults and constraints
The system SHALL default status to `active`, priority to zero, and rule `required` to true. Valid statuses SHALL be `active`, `draft`, and `archived`. Explicit `required` metadata on non-rules MUST fail. IDs SHALL be unique case-sensitive strings matching `[a-z0-9][a-z0-9._/-]*`, without empty, `.` or `..` slash segments. Applicability entries MUST be nonempty, and invalid path patterns MUST fail loading.

#### Scenario: Default mandatory rule
- **WHEN** a rule omits status, priority, and required metadata
- **THEN** it loads as active, priority zero, and required

#### Scenario: Invalid metadata combination
- **WHEN** a skill declares `required: true` or an item declares an invalid status or path pattern
- **THEN** loading fails with a field-specific error

### Requirement: Safe source traversal
The system SHALL recursively discover `.md` files under explicitly configured source directories, ignore other regular files, and reject missing or unreadable sources and symlink entries. Source roots SHALL be resolved to canonical directories; overlapping roots MUST be rejected. The system MUST NOT fetch network content or execute file content.

#### Scenario: Nested local knowledge
- **WHEN** a source contains nested Markdown files and unrelated non-Markdown files
- **THEN** all nested Markdown files are loaded and other regular files are ignored

#### Scenario: Symlink or overlapping source
- **WHEN** a source tree contains a symlink or two configured roots overlap
- **THEN** loading fails with the offending source or path and exposes no partial registry

### Requirement: Validated registry with provenance
Each loaded item SHALL include its source ID, slash-separated source-relative path, and a SHA-256 digest of its original file bytes. Explicit versions SHALL be preserved; absent versions SHALL be represented as `sha256:<digest>`. Duplicate IDs across or within sources MUST reject the complete registry, naming both locations. Repeated loads of unchanged files SHALL produce the same item values and ID ordering.

#### Scenario: Duplicate knowledge ID
- **WHEN** two documents declare the same ID
- **THEN** registry construction fails and identifies both source locations

#### Scenario: Reproducible provenance
- **WHEN** unchanged source files are loaded in different filesystem traversal orders
- **THEN** items retain identical versions, digests, source locations, and ascending ID order
