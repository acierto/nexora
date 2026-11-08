## ADDED Requirements

### Requirement: Windows release archives
Every release SHALL publish `nexora_windows_amd64.zip` and `nexora_windows_arm64.zip`,
each containing `nexora.exe`, `README.md` and `LICENSE`, built on the Linux release
runner with version-free names. `checksums.txt` SHALL list both zips and both installers.

#### Scenario: Release packaging includes Windows
- **WHEN** release-build.sh builds a release with default targets
- **THEN** dist contains both Windows zips with `nexora.exe` and checksum entries for each zip, install.sh and install.ps1

#### Scenario: Windows executable reports its version
- **WHEN** the packaged `nexora.exe` runs on Windows
- **THEN** `nexora version` prints the release version passed to release-build.sh

### Requirement: PowerShell installer
The release SHALL publish `install.ps1`. It SHALL honor `NEXORA_VERSION` (default
latest), `NEXORA_INSTALL_DIR` (default `$env:LOCALAPPDATA\Programs\nexora`) and
`NEXORA_DOWNLOAD_BASE`, detect amd64 or arm64, verify the archive SHA-256 against
`checksums.txt`, verify the staged executable version, install `nexora.exe`, add the
install directory to the user PATH only, and run without administrator rights.
It SHALL work when piped to `iex`.

#### Scenario: Verified Windows install
- **WHEN** the installer runs with a download base serving a valid zip and checksums
- **THEN** `nexora.exe` is installed, the user PATH contains the install directory and `nexora version` is printed

#### Scenario: Checksum mismatch
- **WHEN** the zip does not match its checksum entry
- **THEN** the installer fails with "checksum mismatch" and leaves any existing `nexora.exe` unchanged

#### Scenario: Unsupported architecture
- **WHEN** the processor is neither x64 nor ARM64
- **THEN** the installer fails and points to building from source
