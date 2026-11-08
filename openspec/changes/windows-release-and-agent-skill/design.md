# Design

## Packaging

Go cross-compiles with `CGO_ENABLED=0`, so the Linux runner builds Windows
binaries; no Windows runner is required. `release-build.sh` names the binary
`nexora.exe` for `windows/*` targets and packages a `.zip` instead of a `.tar.gz`.
Archive names carry no version so `releases/latest/download/<name>` keeps
working. `zip` is required; the workflow installs it if absent. `checksums.txt`
covers every archive and both installers.

## PowerShell installer

`install.ps1` mirrors `install.sh`: the same environment variables, defaulting to
`latest` and `$env:LOCALAPPDATA\Programs\nexora`. It maps
`PROCESSOR_ARCHITECTURE` (and `PROCESSOR_ARCHITEW6432` for 32-bit hosts) to
`amd64` or `arm64`, downloads the zip and `checksums.txt` into a staging folder
inside the install directory, compares `Get-FileHash` output, checks the staged
executable reports a release version (and the requested one when pinned), then
moves it into place. Failures leave an existing `nexora.exe` unchanged and remove
staging. It appends the directory to the user `PATH` only (never machine scope,
never elevation) and broadcasts nothing beyond the current session update. It
works as `irm ... | iex`, so it uses no `param()` block or `exit` that would close
the caller's shell; errors are thrown.

`--ensure` is not mirrored in this increment; project automation on Windows can
compare `nexora version` itself.

## Windows runtime selection

`runtimeversion.Resolve` accepts `windows/amd64|arm64`, downloads
`nexora_windows_<arch>.zip`, extracts the single regular `nexora.exe` entry with
`archive/zip`, and caches it as `nexora.exe` so Windows can execute it. Releases
before v1.0.0 have no Windows asset and fail with the existing "no entry" error.

## Agent skill

`skills/nexora/SKILL.md` contains only commands and flags that exist. Every
command is executed in a fresh `nexora init` project. `package-skill.sh` zips the
folder with `nexora/` at the archive root into the ignored `dist/` directory.

## Verification

A Go test runs `release-build.sh` for Windows targets and checks zip contents
and checksum entries. The Windows VM procedure is documented in
`scripts/windows/README.md` and reuses the Freeboard VM scripts rather than
copying them.
