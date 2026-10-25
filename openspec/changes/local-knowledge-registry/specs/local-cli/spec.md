# Local CLI

## Purpose

Provide a small offline command-line workflow to initialize, validate, inspect, and resolve local knowledge in existing repositories.

## ADDED Requirements

### Requirement: Non-destructive initialization
`nexora init` SHALL create `nexora.yaml` and `knowledge/rules/project.md` in the current directory, with a source ID of `local` and a valid example rule. `--project` SHALL set the project; otherwise the directory basename SHALL be used. Existing target files or symlinked destination components MUST cause failure before writing. The command MUST NOT modify native agent files or unrelated knowledge.

#### Scenario: Fresh initialization
- **WHEN** init runs in an empty project directory
- **THEN** the generated configuration and example rule pass `nexora validate` without manual editing

#### Scenario: Existing user file
- **WHEN** either generated target already exists or a destination component is a symlink
- **THEN** init fails and leaves existing files unchanged without creating the other target

### Requirement: Explicit configuration selection
Read commands SHALL accept `--config <path>`, defaulting to `./nexora.yaml` without ancestor discovery. A missing configuration MUST fail with guidance to run init or supply a path. `resolve --project` SHALL override the configured default for that invocation only. Agent IDs SHALL be supplied through optional `--agent`; files SHALL use repeatable `--file` flags.

#### Scenario: Explicit resolve request
- **WHEN** `nexora resolve --config ./nexora.yaml --project commerce-api --agent codex --file src/auth/service.go` runs
- **THEN** the supplied request is resolved without modifying configuration

#### Scenario: Missing configuration
- **WHEN** a read command cannot find its configuration
- **THEN** it fails with actionable guidance and does not generate configuration automatically

### Requirement: Validation and inspection
`validate` SHALL load all configured sources and report validity and item count. `list` SHALL show every loaded item, including inactive items, in ascending ID order with ID, kind, name, status, version, and source. `resolve` SHALL expose the groups and explanations defined by context-resolution. Every read command MUST fail on any configuration or source error rather than use a partial registry.

#### Scenario: Inspect inactive knowledge
- **WHEN** list runs against valid active, draft, and archived items
- **THEN** all items and their statuses are visible in ascending ID order

#### Scenario: Invalid knowledge prevents partial resolution
- **WHEN** any configured document is invalid
- **THEN** validate, list, and resolve fail without returning partial success data

### Requirement: Scriptable output and exit behavior
Read commands SHALL accept `--json` and emit one JSON object on stdout on success with no progress messages. Validation SHALL return `valid` and `item_count`; list SHALL return `items`; resolve SHALL return the five selection arrays. Empty collections SHALL be `[]`. Human output SHALL remain readable. Success SHALL exit 0, usage errors 2, and configuration/loading/resolution/I/O errors 1, with failures on stderr.

#### Scenario: Machine-readable empty result
- **WHEN** resolve succeeds with no matches and `--json`
- **THEN** stdout contains one valid JSON object with five empty arrays and exit status 0

#### Scenario: Unknown command or flag
- **WHEN** a user supplies an unknown command or flag
- **THEN** the CLI prints usage guidance to stderr and exits 2

### Requirement: Offline standalone workflow
The built CLI SHALL run init, validate, list, and resolve without a database, API key, network connection, OpenSpec installation, Node.js, or an agent runtime. Help SHALL describe the supported commands and flags. The repository SHALL include a runnable local example and instructions for building and exercising it.

#### Scenario: Example workflow offline
- **WHEN** the built executable runs validate and resolve against the supplied example without network access
- **THEN** both commands succeed with the documented required rule IDs

### Requirement: Practical user documentation
The repository SHALL provide a linked quickstart, knowledge-authoring guide, CLI reference and recipes, troubleshooting, and local and shared-knowledge examples. Examples MUST contain runnable commands and expected results verified against their actual files. Documentation MUST distinguish supported MVP behavior from deferred integrations and explain that agent filters do not establish agent connections.

#### Scenario: First-time user walkthrough
- **WHEN** a user follows the documented build, init, author, validate, list, and resolve sequence in a fresh project
- **THEN** the commands succeed and produce the documented rule selections without undocumented setup

#### Scenario: Shared standards across two projects
- **WHEN** a user follows the shared-knowledge example for its two project configurations
- **THEN** applicable shared rules appear in both results and project-restricted rules appear only in their matching project's result

#### Scenario: Diagnose an excluded rule
- **WHEN** a user follows troubleshooting for a rule excluded by status, project, agent, or file applicability
- **THEN** the documented explanation and corrected request reproduce the expected exclusion and selection
