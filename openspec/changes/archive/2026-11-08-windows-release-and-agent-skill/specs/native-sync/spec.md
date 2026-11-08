## ADDED Requirements

### Requirement: Line-ending-insensitive sync verification
Sync planning, `sync --check` and native-context verification SHALL treat generated files whose only difference is CRLF instead of LF line endings as unchanged and unedited. Any other difference SHALL still be reported as a change or a hand edit.

#### Scenario: CRLF checkout of synced files
- **WHEN** the managed block, generated skills and manifest are checked out with CRLF line endings
- **THEN** `sync --check` reports everything up to date and native context verification succeeds

#### Scenario: Hand edit in a CRLF file
- **WHEN** a generated skill checked out with CRLF also has edited text
- **THEN** sync reports the file as edited by hand
