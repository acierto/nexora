# Repeatable CLI installation

## Why

Projects need reliable initial setup without maintaining their own installer logic.

## What Changes

Publish install.sh in each release and add --ensure for a stable minimum version.
Validate the staged executable and use atomic replacement to preserve existing
installations when download or validation fails.

## Capabilities

### New Capabilities

- cli-installation: release-owned, repeatable local CLI installation.

### Modified Capabilities

None.

## Impact

POSIX installer, release packaging and documentation. No dependencies or CLI
runtime changes. Windows installers and globally serialized concurrent upgrades
are outside this increment.
