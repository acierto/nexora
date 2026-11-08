# agent-skill Specification

## Purpose
Give coding agents a packaged skill that drives the Nexora CLI workflow, resolving
applicable knowledge before edits and validating and syncing after knowledge
changes, using only commands the CLI provides.

## Requirements

### Requirement: Nexora agent skill
The repository SHALL provide `skills/nexora/SKILL.md` with `name` and `description`
front matter. It SHALL instruct agents to check `nexora version`, recognise a
project by `nexora.yaml`, resolve applicable knowledge before editing a file,
and validate and sync after changing knowledge. It SHALL forbid installing
software or editing agent configuration without the user's consent. Every
command and flag it names SHALL exist in the CLI.

#### Scenario: Commands run as written
- **WHEN** each command in the skill runs in a fresh `nexora init --project demo` folder
- **THEN** every command succeeds (exit 0), except that `nexora mcp` serves until stdin closes

#### Scenario: No invented commands
- **WHEN** the skill is compared with `nexora --help` and the CLI reference
- **THEN** it names no command or flag that the CLI rejects

### Requirement: Skill package
`scripts/package-skill.sh` SHALL produce `dist/nexora-skill.zip` with the `nexora/`
folder at the archive root. `dist/` SHALL be ignored by Git.

#### Scenario: Package layout
- **WHEN** the script runs
- **THEN** the zip contains `nexora/SKILL.md` and no entries outside `nexora/`
