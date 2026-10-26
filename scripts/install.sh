#!/bin/sh
# Install the Nexora CLI from GitHub Releases on macOS or Linux.
#
#   curl -fsSL https://raw.githubusercontent.com/acierto/nexora/main/scripts/install.sh | sh
#
# Environment variables:
#   NEXORA_VERSION      release tag to install, e.g. v0.1.0 (default: latest)
#   NEXORA_INSTALL_DIR  target directory (default: $HOME/.local/bin)
#   NEXORA_DOWNLOAD_BASE  download from this URL instead of GitHub (mirrors, tests)
#
# The script downloads nexora_<os>_<arch>.tar.gz and checksums.txt, verifies
# the SHA-256 checksum, and installs the nexora binary. It never uses sudo.
set -eu

repo="acierto/nexora"
version=${NEXORA_VERSION:-latest}
install_dir=${NEXORA_INSTALL_DIR:-"$HOME/.local/bin"}

fail() {
	echo "nexora install: $*" >&2
	exit 1
}

case "$(uname -s)" in
Darwin) os=darwin ;;
Linux) os=linux ;;
*) fail "no prebuilt binary for $(uname -s); build from source: https://acierto.github.io/nexora/#install" ;;
esac
case "$(uname -m)" in
x86_64 | amd64) arch=amd64 ;;
arm64 | aarch64) arch=arm64 ;;
*) fail "no prebuilt binary for $(uname -m); build from source: https://acierto.github.io/nexora/#install" ;;
esac

if [ "$version" = latest ]; then
	base="https://github.com/$repo/releases/latest/download"
else
	base="https://github.com/$repo/releases/download/$version"
fi
base=${NEXORA_DOWNLOAD_BASE:-$base}
asset="nexora_${os}_${arch}.tar.gz"

download() {
	if command -v curl >/dev/null 2>&1; then
		curl -fsSL "$1" -o "$2"
	elif command -v wget >/dev/null 2>&1; then
		wget -q "$1" -O "$2"
	else
		fail "curl or wget is required"
	fi
}

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT INT TERM

echo "Downloading $asset ($version)"
download "$base/$asset" "$tmp/$asset" || fail "could not download $base/$asset"
download "$base/checksums.txt" "$tmp/checksums.txt" || fail "could not download checksums.txt"

expected=$(awk -v f="$asset" '$2 == f { print $1 }' "$tmp/checksums.txt")
[ -n "$expected" ] || fail "checksums.txt has no entry for $asset"
if command -v sha256sum >/dev/null 2>&1; then
	actual=$(sha256sum "$tmp/$asset" | awk '{ print $1 }')
else
	actual=$(shasum -a 256 "$tmp/$asset" | awk '{ print $1 }')
fi
[ "$expected" = "$actual" ] || fail "checksum mismatch for $asset"

tar -xzf "$tmp/$asset" -C "$tmp" nexora
mkdir -p "$install_dir"
install -m 0755 "$tmp/nexora" "$install_dir/nexora"

echo "Installed $("$install_dir/nexora" version) to $install_dir/nexora"
case ":$PATH:" in
*":$install_dir:"*) ;;
*) echo "Add $install_dir to your PATH, for example: export PATH=\"$install_dir:\$PATH\"" ;;
esac
