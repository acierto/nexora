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
# Pass --ensure with a stable NEXORA_VERSION to reuse that version or newer.
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

ensure=false
case "${1:-}" in
"") [ "$#" -eq 0 ] || fail "usage: install.sh [--ensure]" ;;
--ensure) [ "$#" -eq 1 ] || fail "usage: install.sh [--ensure]"; ensure=true ;;
*) fail "usage: install.sh [--ensure]" ;;
esac

stable() {
	printf '%s\n' "$1" | LC_ALL=C grep -Eq '^v[0-9]{1,8}\.[0-9]{1,8}\.[0-9]{1,8}$'
}
compatible() {
	[ -x "$install_dir/nexora" ] || return 1
	installed=$("$install_dir/nexora" version 2>/dev/null) || return 1
	case "$installed" in "nexora v"*) installed=${installed#nexora } ;; *) return 1 ;; esac
	stable "$installed" || return 1
	awk -v have="${installed#v}" -v need="${version#v}" 'BEGIN {
		split(have, h, "."); split(need, n, ".")
		for (i=1; i<=3; i++) {
			if (h[i]+0 > n[i]+0) exit 0
			if (h[i]+0 < n[i]+0) exit 1
		}
		exit 0
	}'
}
if "$ensure"; then
	stable "$version" || fail "--ensure requires a stable NEXORA_VERSION, for example v0.4.1"
	if compatible; then
		echo "Using nexora $installed at $install_dir/nexora"
		exit 0
	fi
fi

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
		if curl -fsSL --connect-timeout 10 --max-time 120 "$1" -o "$2"; then
			return 0
		else
			status=$?
		fi
		if [ "$status" -eq 60 ] || [ "$status" -eq 77 ]; then
			echo "nexora install: HTTPS certificate trust failed. Check your system trust store or curl CA settings; see https://acierto.github.io/nexora/#troubleshooting. Certificate verification remains enabled." >&2
		fi
		return "$status"
	elif command -v wget >/dev/null 2>&1; then
		wget -q --timeout=120 --tries=1 "$1" -O "$2"
	else
		fail "curl or wget is required"
	fi
}

mkdir -p "$install_dir"
[ ! -d "$install_dir/nexora" ] || fail "destination is a directory: $install_dir/nexora"
tmp=$(mktemp -d "$install_dir/.nexora-install.XXXXXX")
trap 'rm -rf "$tmp"' 0
trap 'exit 130' INT
trap 'exit 143' TERM

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
[ -f "$tmp/nexora" ] && [ ! -L "$tmp/nexora" ] || fail "archive must contain a regular nexora executable"
chmod 0755 "$tmp/nexora"
reported=$("$tmp/nexora" version) || fail "downloaded executable cannot run"
case "$reported" in "nexora v"*) ;; *) fail "downloaded executable has no release version" ;; esac
if [ "$version" != latest ] && [ "$reported" != "nexora $version" ]; then
	fail "downloaded executable reports $reported; expected nexora $version"
fi
if "$ensure" && compatible; then
	echo "Using nexora $installed at $install_dir/nexora"
	exit 0
fi
[ ! -d "$install_dir/nexora" ] || fail "destination is a directory: $install_dir/nexora"
mv -f "$tmp/nexora" "$install_dir/nexora"

echo "Installed $("$install_dir/nexora" version) to $install_dir/nexora"
case ":$PATH:" in
*":$install_dir:"*) ;;
*) echo "Add $install_dir to your PATH, for example: export PATH=\"$install_dir:\$PATH\"" ;;
esac
