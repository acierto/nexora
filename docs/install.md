# Install

[Read online](https://acierto.github.io/nexora/#install) · [Documentation home](https://acierto.github.io/nexora/)

Nexora is a single executable called `nexora`. It has no runtime dependencies:
no Docker, database, or language runtime is needed to run it.

## macOS and Linux

Prebuilt binaries are published for every release. These links always download
the [latest release](https://github.com/acierto/nexora/releases/latest):

| System | Archive |
| --- | --- |
| macOS, Apple Silicon (M1 and later) | [`nexora_darwin_arm64.tar.gz`](https://github.com/acierto/nexora/releases/latest/download/nexora_darwin_arm64.tar.gz) |
| macOS, Intel | [`nexora_darwin_amd64.tar.gz`](https://github.com/acierto/nexora/releases/latest/download/nexora_darwin_amd64.tar.gz) |
| Linux, x86-64 | [`nexora_linux_amd64.tar.gz`](https://github.com/acierto/nexora/releases/latest/download/nexora_linux_amd64.tar.gz) |
| Linux, arm64 (Graviton, Raspberry Pi 4/5 with a 64-bit OS) | [`nexora_linux_arm64.tar.gz`](https://github.com/acierto/nexora/releases/latest/download/nexora_linux_arm64.tar.gz) |
| SHA-256 checksums for all archives | [`checksums.txt`](https://github.com/acierto/nexora/releases/latest/download/checksums.txt) |

The Linux binaries are statically linked, so they run on any distribution with
one of these architectures (Ubuntu, Debian, Fedora, Alpine, Arch, and others).

### Install script

```shell
curl -fsSL https://raw.githubusercontent.com/acierto/nexora/main/scripts/install.sh | sh
```

The script picks the archive for your system, downloads it from the latest
GitHub release, verifies its SHA-256 checksum, and installs `nexora` into
`~/.local/bin`. It never uses `sudo`. If that directory is not on your `PATH`,
the script prints the line to add to your shell profile.

Options, set as environment variables before `sh`:

```shell
curl -fsSL https://raw.githubusercontent.com/acierto/nexora/main/scripts/install.sh | NEXORA_VERSION=v0.1.0 sh
curl -fsSL https://raw.githubusercontent.com/acierto/nexora/main/scripts/install.sh | NEXORA_INSTALL_DIR=/usr/local/bin sh
```

### Automated project setup

Releases include `install.sh`. Projects can download a pinned copy and invoke it
without maintaining their own version checks, platform detection or installer:

```shell
curl -fsSL --connect-timeout 10 --max-time 60 \
  https://github.com/acierto/nexora/releases/download/v0.4.1/install.sh \
  -o /tmp/nexora-install.sh
NEXORA_VERSION=v0.4.1 sh /tmp/nexora-install.sh --ensure
```

`--ensure` requires a stable version pin. It reuses that version or a newer stable
release in `NEXORA_INSTALL_DIR` (default `~/.local/bin`) without downloading a
binary. It does not search PATH. Missing, older, broken, development and prerelease
installations are replaced. Fetching the script itself still requires network
access; a previously downloaded script can be used offline with a compatible
installation. Without `--ensure`, the installer installs the requested version,
including an explicit downgrade.

Downloads have time limits and retain TLS verification. Installation verifies the
archive checksum and executable version in a temporary directory on the target
filesystem, then renames the executable into place. Failed downloads or validation
preserve an existing executable and clean temporary files. No shell profile is
changed. Projects decide whether an installation failure should stop their setup.

For MCP clients whose PATH omits the install directory, use a shell only to expand
the user-specific location; no project launcher script is needed:

```json
{
  "command": "sh",
  "args": ["-c", "exec \"${NEXORA_INSTALL_DIR:-$HOME/.local/bin}/nexora\" \"$@\"", "nexora", "mcp", "--version", "v0.4.0", "--config", "nexora.yaml"]
}
```

Start the client in the project root, or provide an absolute config path. Set the
same `NEXORA_INSTALL_DIR` for setup and the client when customizing it. Reconnect
an already-running client after setup. MCP launch itself never invokes the installer.

### Reuse the installation across projects

Install Nexora once and let each project keep its own `nexora.yaml` and knowledge
files. MCP clients can run the installed binary directly:

```json
{
  "mcpServers": {
    "Project Knowledge": {
      "command": "nexora",
      "args": ["mcp", "--version", "v0.4.0", "--config", "nexora.yaml"]
    }
  }
}
```

Start the client from the project root, or pass an absolute config path. Ensure
its process inherits the install directory in PATH; desktop clients may need an
absolute binary path. Restart/reconnect after installing. Projects do not need
their own downloader, archive extractor or Python runtime. The upstream installer
owns release selection, platform detection and checksum verification.

The example pins the project runtime independently of the installed bootstrap.
Use `"latest"` as the selector to follow published releases. Runtime selection
requires a v0.4.0 or newer bootstrap and never silently falls back to another
version; see [the CLI reference](cli.md#runtime-selection-for-mcp).

No certificate export is part of normal installation. If installation fails with
an HTTPS trust error, see [troubleshooting](troubleshooting.md#installation-and-certificate-trust).

### Manual download

1. Download your archive and `checksums.txt` from the table above or the
   [releases page](https://github.com/acierto/nexora/releases/latest).
2. Check the archive: `shasum -a 256 -c checksums.txt --ignore-missing` on macOS,
   or `sha256sum -c checksums.txt --ignore-missing` on Linux.
3. Unpack it and move the binary onto your `PATH`:

```shell
tar -xzf nexora_darwin_arm64.tar.gz
mkdir -p ~/.local/bin && mv nexora ~/.local/bin/
```

On macOS, a binary downloaded with a web browser is quarantined by Gatekeeper
and refuses to start ("cannot be opened because the developer cannot be
verified"). Remove the quarantine flag once:

```shell
xattr -d com.apple.quarantine ~/.local/bin/nexora
```

Downloads made with `curl`, including the install script, are not quarantined.

## Windows

Prebuilt binaries for Windows 10 and 11 are published for every release:

| System | Archive |
| --- | --- |
| Windows, x64 (most PCs) | [`nexora_windows_amd64.zip`](https://github.com/acierto/nexora/releases/latest/download/nexora_windows_amd64.zip) |
| Windows on Arm (Snapdragon, Copilot+ PCs) | [`nexora_windows_arm64.zip`](https://github.com/acierto/nexora/releases/latest/download/nexora_windows_arm64.zip) |
| SHA-256 checksums for all archives | [`checksums.txt`](https://github.com/acierto/nexora/releases/latest/download/checksums.txt) |

### PowerShell install script

```powershell
irm https://raw.githubusercontent.com/acierto/nexora/main/scripts/install.ps1 | iex
```

The script picks the archive for your processor, downloads it from the latest
GitHub release, verifies its SHA-256 checksum, and installs `nexora.exe` into
`%LOCALAPPDATA%\Programs\nexora`. It adds that folder to your user `PATH` (never
the system `PATH`) and needs no administrator rights. Open a new terminal
afterwards so the updated `PATH` applies. It works in Windows PowerShell 5.1 and
PowerShell 7.

Options, set as environment variables in the same PowerShell session first:

```powershell
$env:NEXORA_VERSION = "v1.0.0"                        # default: latest
$env:NEXORA_INSTALL_DIR = "$HOME\bin"                 # default: %LOCALAPPDATA%\Programs\nexora
irm https://raw.githubusercontent.com/acierto/nexora/main/scripts/install.ps1 | iex
```

Releases include `install.ps1`, so automation can download a pinned copy from
`https://github.com/acierto/nexora/releases/download/<version>/install.ps1`.
`NEXORA_DOWNLOAD_BASE` points the script at a mirror URL or a local folder holding
the zip and `checksums.txt`. Failed downloads or verification leave an existing
`nexora.exe` unchanged. The PowerShell script has no `--ensure` mode.

### Manual download on Windows

1. Download your zip and `checksums.txt` from the table above or the
   [releases page](https://github.com/acierto/nexora/releases/latest).
2. Check the zip: run `Get-FileHash nexora_windows_amd64.zip` in PowerShell and
   compare the hash (case-insensitive) with that file's line in `checksums.txt`.
3. Extract `nexora.exe` into a folder of your choice, for example
   `%LOCALAPPDATA%\Programs\nexora`, and add that folder to your `PATH`:
   Settings → System → About → Advanced system settings → Environment Variables,
   then edit `Path` for your user.

### SmartScreen warning

The Windows binaries are not code-signed yet. The first time you run a
browser-downloaded `nexora.exe`, Microsoft Defender SmartScreen may show "Windows
protected your PC". Choose **More info**, then **Run anyway**. Verify the checksum
first if you want to be sure the file is the published one.

Commands in the guides work the same in PowerShell and Command Prompt.
Configuration paths use `/` and work on Windows too, and `--file` accepts either
`/` or `\`. Knowledge files checked out with CRLF line endings have the same
versions as on macOS and Linux. The Windows build is compiled and tested on Linux
and macOS but not yet exercised on Windows itself; please report problems in the
[issue tracker](https://github.com/acierto/nexora/issues).

## Check the installation

```shell
nexora version
nexora help
```

`nexora version` prints the release, for example `nexora v0.1.0`.

## Other systems and architectures

There are no prebuilt binaries for FreeBSD or 32-bit systems. Nexora is plain Go
with no C dependencies, so it builds anywhere Go runs:

1. Install Go 1.23 or newer from [go.dev/dl](https://go.dev/dl/).
2. Install Nexora:

```shell
go install github.com/acierto/nexora/cmd/nexora@latest
```

This puts `nexora` in Go's binary directory: run `go env GOPATH` and look in its
`bin` folder, typically `~/go/bin`. Add that folder to your `PATH`. From a checkout
of the repository, `go build -o nexora ./cmd/nexora` builds the binary in place
instead. The same works on Windows (`%USERPROFILE%\go\bin\nexora.exe`) if you
prefer building from source.

## Upgrade and uninstall

Run the install script again to upgrade, or repeat the manual download. With Go,
run the same `go install` command. To uninstall, delete the `nexora` binary
(`nexora.exe` on Windows, and remove its folder from your user `PATH`).
Nexora writes nothing outside your projects except where you configure it: the
`.nexora/` folder next to `nexora.yaml` holds the local index and sync manifests.

Next: [connect Nexora to your agent](quickstart.md).
