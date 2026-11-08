# cli-installation Specification

## Purpose

Provide a reusable, verified installation of the Nexora CLI for local users and
project setup scripts, preserving working installations when upgrades fail.

## Requirements

### Requirement: Repeatable pinned setup

The installer SHALL accept --ensure with a stable NEXORA_VERSION and reuse a stable
equal-or-newer version from the configured installation directory without network.
It SHALL install the requested release for missing, older, development or prerelease
executables. Without --ensure, explicit installation SHALL permit a downgrade.

#### Scenario: Compatible offline installation
- **WHEN** --ensure finds the same or a newer stable version
- **THEN** it succeeds without a download or modifying the executable

#### Scenario: Fresh or outdated installation
- **WHEN** --ensure finds no compatible executable
- **THEN** it installs the requested release in the configured directory

#### Scenario: Explicit downgrade
- **WHEN** an older version is explicitly installed without --ensure
- **THEN** the requested version replaces the newer executable

#### Scenario: Invalid ensure pin
- **WHEN** --ensure receives latest, development or prerelease as its pin
- **THEN** it fails before modifying the installation directory

### Requirement: Verified publication

The installer SHALL keep TLS verification enabled, bound download durations,
verify archive checksums and the staged executable version, and publish using a
same-filesystem rename. Failure SHALL preserve an existing executable and clean
staging. A directory destination SHALL be refused.

#### Scenario: Valid archive
- **WHEN** the archive checksum and executable version match
- **THEN** the installed CLI runs and reports the requested version

#### Scenario: Verification failure
- **WHEN** checksum, certificate trust or executable-version validation fails
- **THEN** installation fails with diagnostics and leaves an existing binary unchanged

#### Scenario: Directory destination
- **WHEN** the destination named nexora is a directory
- **THEN** installation refuses to publish inside it

### Requirement: Downloadable installer

Every release SHALL publish install.sh beside binary archives and checksums.txt.

#### Scenario: Release packaging
- **WHEN** release-build.sh builds a release
- **THEN** dist includes the installer and its checksum and the release workflow uploads it

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
