# cli-installation Specification

## Purpose
TBD - created by archiving change repeatable-cli-installation. Update Purpose after archive.

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
