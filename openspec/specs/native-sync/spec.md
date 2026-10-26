# native-sync Specification

## Purpose

Write selected knowledge into the files agents load natively, without overwriting human edits.

## Requirements

### Requirement: Managed instruction block
`nexora sync --agent codex|claude` SHALL write one block between `<!-- nexora:begin -->` and `<!-- nexora:end -->` into `AGENTS.md` (codex) or `CLAUDE.md` (claude) under `--dir`. The block SHALL contain active required rules applicable to the project and agent, grouped into unrestricted rules and path-restricted rules with their patterns, plus a pointer to `nexora resolve` and the MCP server.

#### Scenario: Agent-specific rules
- **WHEN** a rule is restricted to agent `claude`
- **THEN** it appears in the CLAUDE.md block and not in the AGENTS.md block

### Requirement: Preservation of hand-written instructions
Content outside the markers SHALL be preserved byte for byte, and a file without markers SHALL get the block appended.

#### Scenario: Existing instructions preserved
- **WHEN** `AGENTS.md` contains hand-written instructions and no markers
- **THEN** sync appends the block and the original text is unchanged

#### Scenario: Idempotent
- **WHEN** sync runs twice without knowledge changes
- **THEN** the second run reports no changes and writes nothing

### Requirement: Generated skills
Active skills applicable to the project and agent SHALL be written to `.agents/skills/<name>/SKILL.md` (codex) or `.claude/skills/<name>/SKILL.md` (claude) with `name` and `description` front matter, where `name` is the ID without a leading `skills/` segment, with other separators replaced by hyphens, followed by the skill content and its Nexora source and version.

#### Scenario: Skill file
- **WHEN** skill `skills/add-endpoint` is applicable
- **THEN** `.agents/skills/add-endpoint/SKILL.md` exists with `name: add-endpoint`

### Requirement: Conflict protection and check mode
A manifest at `.nexora/sync-<agent>.json` SHALL record generated files and digests. Sync MUST stop without writing when a target skill file exists without a manifest entry or was modified since it was generated. Stale generated files SHALL be removed only when unmodified. `--check` SHALL write nothing and exit 1 when any file would change.

#### Scenario: Manual edit
- **WHEN** a generated SKILL.md was edited by hand
- **THEN** sync exits 1 naming the file and leaves every file unchanged

#### Scenario: CI check
- **WHEN** knowledge changed after the last sync and `--check` runs
- **THEN** it exits 1 listing the files that would change
