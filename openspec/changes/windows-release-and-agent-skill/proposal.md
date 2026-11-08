# Windows releases and agent skill

## Why

Windows users must install Go to use Nexora, and agents working in a Nexora
project have no packaged guidance telling them to resolve rules before editing
and to validate and sync after changing knowledge. Both block a 1.0.0 release that
works the same way on every supported desktop platform.

## What Changes

- Build `windows/amd64` and `windows/arm64` on the existing Linux release runner
  and publish `nexora_windows_<arch>.zip` (`nexora.exe`, `README.md`, `LICENSE`)
  with version-free names and checksums.
- Add `scripts/install.ps1`, a PowerShell installer mirroring `install.sh`
  (`NEXORA_VERSION`, `NEXORA_INSTALL_DIR`, `NEXORA_DOWNLOAD_BASE`), verifying
  SHA-256 and updating only the user `PATH`. Publish it in each release.
- Let `nexora mcp --version` select Windows runtimes from the new zip archives.
- Verify the Windows build on a real Windows 11 ARM64 VM (native arm64 and
  emulated amd64) and fix any platform defects found.
- Add the `nexora` agent skill in `skills/nexora/SKILL.md`, checked against the
  real CLI, and `scripts/package-skill.sh` producing `dist/nexora-skill.zip`.
- Document Windows installation, the agent skill and MIT contribution terms in
  the guides, README and docs site; record the release in the changelog as v1.0.0.

MVP boundary: unsigned Windows binaries (SmartScreen may warn), user-scope
installation only, no `--ensure` equivalent in the PowerShell installer, no
Windows CI runner. Code signing, winget/Scoop packages, an MSI and a marketplace
URL for the skill are deferred.

## Capabilities

### New Capabilities

- agent-skill: a packaged agent skill that drives the Nexora CLI workflow.

### Modified Capabilities

- cli-installation: Windows archives and a PowerShell installer.
- mcp-server: Windows runtime selection from zip archives.

## Impact

Release script and workflow, a new installer script, `internal/runtimeversion`,
tests, guides, README, docs site and changelog. No new Go dependencies; `zip` is
required on the release runner.
