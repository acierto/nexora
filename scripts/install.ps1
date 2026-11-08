# Install the Nexora CLI from GitHub Releases on Windows.
#
#   irm https://raw.githubusercontent.com/acierto/nexora/main/scripts/install.ps1 | iex
#
# Environment variables:
#   NEXORA_VERSION        release tag to install, e.g. v1.0.0 (default: latest)
#   NEXORA_INSTALL_DIR    target directory (default: $env:LOCALAPPDATA\Programs\nexora)
#   NEXORA_DOWNLOAD_BASE  download from this URL or local folder instead of GitHub
#                         (mirrors, tests)
#
# The script downloads nexora_windows_<arch>.zip and checksums.txt, verifies the
# SHA-256 checksum and the executable version, installs nexora.exe, and adds the
# install directory to the user PATH. It never changes the machine PATH and never
# needs administrator rights. Works in Windows PowerShell 5.1 and PowerShell 7.

function Install-Nexora {
	$ErrorActionPreference = 'Stop'
	$ProgressPreference = 'SilentlyContinue'
	$repo = 'acierto/nexora'

	function Fail([string] $message) { throw "nexora install: $message" }

	$version = if ($env:NEXORA_VERSION) { $env:NEXORA_VERSION } else { 'latest' }
	$installDir = if ($env:NEXORA_INSTALL_DIR) { $env:NEXORA_INSTALL_DIR } else { Join-Path $env:LOCALAPPDATA 'Programs\nexora' }
	$installDir = [System.IO.Path]::GetFullPath($installDir)

	$arch = $null
	try {
		$arch = switch ([System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture.ToString()) {
			'X64' { 'amd64' } 'Arm64' { 'arm64' } default { $null }
		}
	} catch {}
	if (-not $arch) {
		$reported = if ($env:PROCESSOR_ARCHITEW6432) { $env:PROCESSOR_ARCHITEW6432 } else { $env:PROCESSOR_ARCHITECTURE }
		$arch = switch ($reported) { 'AMD64' { 'amd64' } 'ARM64' { 'arm64' } default { $null } }
		if (-not $arch) { Fail "no prebuilt binary for $reported; build from source: https://acierto.github.io/nexora/#install" }
	}

	if ($version -eq 'latest') {
		$base = "https://github.com/$repo/releases/latest/download"
	} else {
		$base = "https://github.com/$repo/releases/download/$version"
	}
	if ($env:NEXORA_DOWNLOAD_BASE) { $base = $env:NEXORA_DOWNLOAD_BASE }
	$base = $base.TrimEnd('/', '\')
	$asset = "nexora_windows_$arch.zip"

	# Windows PowerShell 5.1 may default to TLS 1.0; GitHub requires TLS 1.2.
	try {
		[Net.ServicePointManager]::SecurityProtocol = [Net.ServicePointManager]::SecurityProtocol -bor [Net.SecurityProtocolType]::Tls12
	} catch {}

	function Get-Asset([string] $name, [string] $destination) {
		if ($base -match '^https?://') {
			try {
				Invoke-WebRequest -UseBasicParsing -TimeoutSec 120 -Uri "$base/$name" -OutFile $destination
			} catch {
				Fail "could not download $base/$name ($($_.Exception.Message))"
			}
		} else {
			$folder = if ($base -match '^file://') { ([Uri] $base).LocalPath } else { $base }
			$source = Join-Path $folder $name
			if (-not (Test-Path -LiteralPath $source -PathType Leaf)) { Fail "could not find $source" }
			Copy-Item -LiteralPath $source -Destination $destination
		}
	}

	$target = Join-Path $installDir 'nexora.exe'
	New-Item -ItemType Directory -Force -Path $installDir | Out-Null
	if (Test-Path -LiteralPath $target -PathType Container) { Fail "destination is a directory: $target" }
	# Executables replaced while running are renamed first; remove leftovers.
	Get-ChildItem -LiteralPath $installDir -Filter '.nexora-old-*.exe' -Force -ErrorAction SilentlyContinue |
		Remove-Item -Force -ErrorAction SilentlyContinue

	$staging = Join-Path $installDir ('.nexora-install.' + [Guid]::NewGuid().ToString('N').Substring(0, 8))
	New-Item -ItemType Directory -Path $staging | Out-Null
	try {
		Write-Host "Downloading $asset ($version)"
		$archive = Join-Path $staging $asset
		$checksums = Join-Path $staging 'checksums.txt'
		Get-Asset $asset $archive
		Get-Asset 'checksums.txt' $checksums

		$expected = $null
		foreach ($line in Get-Content -LiteralPath $checksums) {
			$fields = -split $line
			if ($fields.Count -eq 2 -and $fields[1].TrimStart('*') -eq $asset) { $expected = $fields[0].ToLowerInvariant() }
		}
		if (-not $expected) { Fail "checksums.txt has no entry for $asset" }
		$actual = (Get-FileHash -LiteralPath $archive -Algorithm SHA256).Hash.ToLowerInvariant()
		if ($expected -ne $actual) { Fail "checksum mismatch for $asset" }

		$unpacked = Join-Path $staging 'unpacked'
		Expand-Archive -LiteralPath $archive -DestinationPath $unpacked
		$staged = Join-Path $unpacked 'nexora.exe'
		$item = Get-Item -LiteralPath $staged -Force -ErrorAction SilentlyContinue
		if (-not $item -or $item.PSIsContainer -or ($item.Attributes -band [IO.FileAttributes]::ReparsePoint)) {
			Fail 'archive must contain a regular nexora.exe'
		}
		$reported = (& $staged version | Out-String).Trim()
		if ($LASTEXITCODE -ne 0) { Fail 'downloaded executable cannot run' }
		if (-not $reported.StartsWith('nexora v')) { Fail 'downloaded executable has no release version' }
		if ($version -ne 'latest' -and $reported -ne "nexora $version") {
			Fail "downloaded executable reports $reported; expected nexora $version"
		}

		if (Test-Path -LiteralPath $target -PathType Container) { Fail "destination is a directory: $target" }
		if (Test-Path -LiteralPath $target) {
			# Windows cannot overwrite a running executable but can rename it.
			$old = Join-Path $installDir ('.nexora-old-' + [Guid]::NewGuid().ToString('N').Substring(0, 8) + '.exe')
			Move-Item -LiteralPath $target -Destination $old
			try {
				Move-Item -LiteralPath $staged -Destination $target
			} catch {
				Move-Item -LiteralPath $old -Destination $target
				throw
			}
			Remove-Item -LiteralPath $old -Force -ErrorAction SilentlyContinue
		} else {
			Move-Item -LiteralPath $staged -Destination $target
		}
	} finally {
		Remove-Item -LiteralPath $staging -Recurse -Force -ErrorAction SilentlyContinue
	}

	# Update only the user PATH, keeping unexpanded entries such as %USERPROFILE%.
	$key = [Microsoft.Win32.Registry]::CurrentUser.OpenSubKey('Environment', $true)
	try {
		$userPath = [string] $key.GetValue('Path', '', [Microsoft.Win32.RegistryValueOptions]::DoNotExpandEnvironmentNames)
		$entries = @($userPath -split ';' | Where-Object { $_ })
		$normalized = { param($p) [Environment]::ExpandEnvironmentVariables($p).TrimEnd('\').ToLowerInvariant() }
		$present = $entries | Where-Object { (& $normalized $_) -eq (& $normalized $installDir) }
		if (-not $present) {
			$kind = if ($userPath) { $key.GetValueKind('Path') } else { [Microsoft.Win32.RegistryValueKind]::ExpandString }
			$key.SetValue('Path', (($entries + $installDir) -join ';'), $kind)
			# Tell Explorer and new terminals that the environment changed.
			[Environment]::SetEnvironmentVariable('NEXORA_INSTALL_REFRESH', '1', 'User')
			[Environment]::SetEnvironmentVariable('NEXORA_INSTALL_REFRESH', $null, 'User')
			Write-Host "Added $installDir to your user PATH. Open a new terminal to use nexora."
		}
	} finally {
		$key.Close()
	}
	$sessionEntries = $env:Path -split ';' | ForEach-Object { $_.TrimEnd('\') }
	if ($sessionEntries -notcontains $installDir.TrimEnd('\')) { $env:Path = "$env:Path;$installDir" }

	Write-Host "Installed nexora to $target"
	& $target version
}

Install-Nexora
