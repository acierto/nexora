# Testing Nexora on Windows

Release builds for Windows are cross-compiled on Linux, so CI never runs them on
Windows. Before a release that changes the Windows build or `install.ps1`, test the
zips on a real Windows machine. This uses the disposable Windows 11 ARM64 VM from
the Freeboard project ([`acierto/freeboard`](https://github.com/acierto/freeboard),
`scripts/windows/vm/`); the VM scripts are not copied into this repository.

Requirements (macOS on Apple Silicon): Homebrew QEMU, a checkout of Freeboard next
to this repository, and the ISOs that Freeboard's `start.sh` expects in
`~/soft/vms/freeboard-windows/` (override the parent folder with `VMS_DIR`):
`windows.iso`, Microsoft's Windows 11 Arm64 ISO, and `virtio-win.iso`. The first
start installs Windows unattended and takes about an hour; later starts boot the
existing disk. If the installer's "Press any key to boot from CD" prompt times out
and the UEFI shell appears, reset the VM and press a key in its window during boot.
`ssh.sh` answers once first-logon setup has installed OpenSSH. That step downloads
OpenSSH from Windows Update and can take an hour or more while the screen stays
black; do not reset the VM during it, or Windows shows "Why did my PC restart?"
on the next boot. The VM is long-lived and shared with other projects: reuse it,
never delete or recreate it.

## 1. Build and start

```sh
scripts/release-build.sh v0.0.0-test
../freeboard/scripts/windows/vm/start.sh &        # SSH on 127.0.0.1:2223
VM=../freeboard/scripts/windows/vm/ssh.sh          # runs PowerShell in the VM
$VM 'Get-Date'                                     # wait until this answers
```

## 2. Install with install.ps1

Copy the arm64 zip, `checksums.txt` and the installer, then install from that local
folder:

```sh
$VM 'New-Item -ItemType Directory -Force C:\nexora-test | Out-Null'
scp -P 2223 -i ~/.ssh/freeboard_vm -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null \
  dist/nexora_windows_arm64.zip dist/nexora_windows_amd64.zip dist/checksums.txt dist/install.ps1 \
  dev@127.0.0.1:C:/nexora-test/
$VM '$env:NEXORA_DOWNLOAD_BASE="C:\nexora-test"; Get-Content C:\nexora-test\install.ps1 -Raw | iex'
```

`Get-Content ... | iex` runs the installer the same way as the documented
`irm ... | iex` one-liner. Expect `Installed nexora to
C:\Users\dev\AppData\Local\Programs\nexora\nexora.exe`, then `nexora v0.0.0-test`.
A second run must not add the directory to the user PATH again.

## 3. Exercise the CLI

The SSH server's environment predates the PATH change, so load the user PATH
first, and run test scripts with `-ExecutionPolicy Bypass` (Windows blocks `.ps1`
files by default; `irm | iex` is not affected):

```sh
$VM '$env:Path += ";" + [Environment]::GetEnvironmentVariable("Path","User"); nexora version'
$VM '$env:Path += ";" + [Environment]::GetEnvironmentVariable("Path","User")
     $d = Join-Path $env:TEMP "nexora-demo"; Remove-Item -Recurse -Force $d -ErrorAction SilentlyContinue
     New-Item -ItemType Directory $d | Out-Null; Set-Location $d; Set-Content README.md "# demo"
     nexora init --project demo; nexora validate; nexora sync --agent claude
     nexora resolve --agent claude --file README.md; nexora sync --agent claude --check'
$VM '$env:Path += ";" + [Environment]::GetEnvironmentVariable("Path","User")
     Set-Location (Join-Path $env:TEMP "nexora-demo")
     $init = "{""jsonrpc"":""2.0"",""id"":1,""method"":""initialize"",""params"":{""protocolVersion"":""2025-06-18"",""capabilities"":{},""clientInfo"":{""name"":""test"",""version"":""1""}}}"
     $init | nexora mcp'
```

Each command must exit 0; `nexora mcp` must answer the initialize request with a
JSON-RPC result and exit when its input ends.

## 4. The amd64 build

Windows 11 on ARM emulates x64. Run the amd64 binary once directly:

```sh
$VM 'Expand-Archive -Force C:\nexora-test\nexora_windows_amd64.zip C:\nexora-test\amd64
     C:\nexora-test\amd64\nexora.exe version
     Set-Location (Join-Path $env:TEMP "nexora-demo"); C:\nexora-test\amd64\nexora.exe resolve --agent claude --file README.md'
```

## 5. Failure path

Serve a copy with a wrong checksum and confirm the installer refuses it and keeps
the installed binary:

```sh
$VM '$env:Path += ";" + [Environment]::GetEnvironmentVariable("Path","User")
     $bad = "C:\nexora-test\bad"; New-Item -ItemType Directory -Force $bad | Out-Null
     Copy-Item C:\nexora-test\nexora_windows_arm64.zip $bad
     ("0" * 64) + "  nexora_windows_arm64.zip" | Set-Content "$bad\checksums.txt"
     $env:NEXORA_DOWNLOAD_BASE = $bad
     try { Get-Content C:\nexora-test\install.ps1 -Raw | iex } catch { $_.Exception.Message }
     nexora version'
```

Expect `nexora install: checksum mismatch for nexora_windows_arm64.zip`, followed by
the previously installed version.

Stop the VM by closing its QEMU window, or `$VM 'Stop-Computer -Force'`.
