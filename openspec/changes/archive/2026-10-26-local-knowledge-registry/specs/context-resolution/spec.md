# Context Resolution

## Purpose

Select applicable local knowledge predictably and explain every selection, preserving mandatory rules without depending on semantic similarity.

## ADDED Requirements

### Requirement: Explicit applicability matching
The system SHALL select only active items matching every restricted applicability dimension. Project and agent values SHALL use exact case-sensitive matching; values within a dimension SHALL use OR. An omitted or empty list SHALL impose no restriction. A missing request value SHALL not match a restricted dimension. Local filters MUST NOT be presented as authorization.

#### Scenario: All dimensions must match
- **WHEN** an active rule restricts project to `commerce-api` and agent to `codex`, and the request supplies those values
- **THEN** the rule matches these dimensions

#### Scenario: Project boundary or absent agent
- **WHEN** the project differs or the request omits an agent required by the rule
- **THEN** the rule is excluded

#### Scenario: Inactive knowledge
- **WHEN** a draft or archived item otherwise matches the request
- **THEN** it is excluded

### Requirement: Portable file applicability
The system SHALL match patterns against normalized repository-relative paths. `/` SHALL delimit segments; `*` and `?` SHALL match within one segment, and a whole `**` segment SHALL match zero or more segments. Other glob syntax MUST be rejected. Absolute paths and paths containing `..` segments MUST fail. Backslashes SHALL normalize to `/`. A path-restricted item SHALL require at least one matching request file.

#### Scenario: Recursive match including zero directories
- **WHEN** a rule has `paths: ["src/**/*.go"]` and files include `src/main.go` or `src/auth/service.go`
- **THEN** the rule matches the file dimension

#### Scenario: No files provided
- **WHEN** a rule has path restrictions and the request contains no files
- **THEN** it is excluded while otherwise matching unrestricted rules remain eligible

#### Scenario: Invalid request file
- **WHEN** a request contains `/etc/passwd`, `C:\\secrets.txt`, or `src/../private.go`
- **THEN** resolution fails with a path-validation error rather than silently ignoring that file

### Requirement: Mandatory and optional grouping
The result SHALL expose `required_rules`, `optional_rules`, `skills`, `references`, and `decisions` as arrays of selections. Every applicable required rule MUST appear exactly once in `required_rules`; other applicable items SHALL appear exactly once in their corresponding group. Priority MUST NOT suppress another rule. The MVP SHALL return all matches without truncation or inferred overrides.

#### Scenario: Conflicting priorities preserve mandatory rules
- **WHEN** two applicable mandatory rules have different priorities
- **THEN** both are returned as required rules

#### Scenario: Optional knowledge does not gain mandatory authority
- **WHEN** an optional rule, a skill, a reference, and a decision match
- **THEN** each appears only in its corresponding optional group

### Requirement: Stable ordering and explanations
Within each result group, selections SHALL sort by descending priority then ascending ID. Each selection SHALL contain the full item, source, version, and structured reasons covering active status and every applicability dimension, including unrestricted dimensions. Duplicate request files MUST NOT duplicate selections or reasons.

#### Scenario: Deterministic result
- **WHEN** registry items and duplicate request files arrive in different orders
- **THEN** resolution produces identical ordered selections and explanations

#### Scenario: Traceable selection
- **WHEN** a rule matches a project and a file pattern
- **THEN** its result identifies the matched project, pattern and file, source location, version, and active status
