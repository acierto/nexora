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

## Check the installation

```shell
nexora version
nexora help
```

`nexora version` prints the release, for example `nexora v0.1.0`.

## Windows, other systems, and other architectures

There are no prebuilt binaries for Windows, FreeBSD, or 32-bit systems yet.
Nexora is plain Go with no C dependencies, so it builds anywhere Go runs:

1. Install Go 1.23 or newer from [go.dev/dl](https://go.dev/dl/).
2. Install Nexora:

```shell
go install github.com/acierto/nexora/cmd/nexora@latest
```

This puts `nexora` (`nexora.exe` on Windows) in Go's binary directory: run
`go env GOPATH` and look in its `bin` folder, typically `~/go/bin` or
`%USERPROFILE%\go\bin`. Add that folder to your `PATH`. On Windows: Settings →
System → About → Advanced system settings → Environment Variables, then edit
`Path` for your user. From a checkout of the repository, `go build -o nexora ./cmd/nexora`
builds the binary in place instead.

On Windows, run the commands in PowerShell or Command Prompt as shown in the
guides, replacing `nexora` with `nexora.exe` if your shell does not add the
extension. Configuration paths use `/` and work on Windows too. The Windows build
is compiled and tested on Linux and macOS but not yet exercised on Windows itself;
please report problems in the [issue tracker](https://github.com/acierto/nexora/issues).

## Upgrade and uninstall

Run the install script again to upgrade, or repeat the manual download. With Go,
run the same `go install` command. To uninstall, delete the `nexora` binary.
Nexora writes nothing outside your projects except where you configure it: the
`.nexora/` folder next to `nexora.yaml` holds the local index and sync manifests.

Next: [connect Nexora to your agent](quickstart.md).
