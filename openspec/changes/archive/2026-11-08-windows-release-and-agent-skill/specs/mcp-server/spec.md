## ADDED Requirements

### Requirement: Windows runtime selection
`nexora mcp --version` SHALL support windows/amd64 and windows/arm64 by downloading
`nexora_windows_<arch>.zip`, verifying its checksum, extracting only the regular
`nexora.exe` entry and caching it under that name so Windows can execute it.

#### Scenario: Windows runtime from zip
- **WHEN** a Windows host selects a release that publishes a Windows zip
- **THEN** the verified `nexora.exe` is cached and its path is returned for launch

#### Scenario: Release without a Windows archive
- **WHEN** the selected release's checksums have no entry for the Windows zip
- **THEN** selection fails with a diagnostic and nothing is launched
