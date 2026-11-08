#!/bin/sh
# Build release archives for the supported platforms into dist/.
#
#   scripts/release-build.sh v0.1.0
#
# Produces nexora_<os>_<arch>.tar.gz (binary, README.md, LICENSE) for each macOS
# and Linux target, nexora_windows_<arch>.zip (nexora.exe, README.md, LICENSE)
# for each Windows target, the install.sh and install.ps1 installers, and
# checksums.txt with SHA-256 sums. Archive names carry no version so that
# https://github.com/acierto/nexora/releases/latest/download/<name> always
# points at the newest release. Binaries are static (CGO_ENABLED=0).
set -eu

version=${1:?usage: scripts/release-build.sh <version, e.g. v0.1.0>}
targets=${NEXORA_TARGETS:-"darwin/arm64 darwin/amd64 linux/amd64 linux/arm64 windows/amd64 windows/arm64"}
root=$(cd "$(dirname "$0")/.." && pwd)
dist="$root/dist"

case " $targets" in
*" windows/"*) command -v zip >/dev/null 2>&1 || { echo "release-build: zip is required for Windows archives" >&2; exit 1; } ;;
esac

rm -rf "$dist"
mkdir -p "$dist"
cd "$root"

for target in $targets; do
	os=${target%/*}
	arch=${target#*/}
	name="nexora_${os}_${arch}"
	work="$dist/$name"
	binary=nexora
	[ "$os" = windows ] && binary=nexora.exe
	mkdir -p "$work"
	echo "building $target"
	CGO_ENABLED=0 GOOS=$os GOARCH=$arch go build -trimpath \
		-ldflags "-s -w -X github.com/acierto/nexora/internal/cli.Version=$version" \
		-o "$work/$binary" ./cmd/nexora
	cp README.md "$work/README.md"
	cp LICENSE "$work/LICENSE"
	if [ "$os" = windows ]; then
		(cd "$work" && zip -q -X "$dist/$name.zip" "$binary" README.md LICENSE)
	else
		tar -C "$work" -czf "$dist/$name.tar.gz" "$binary" README.md LICENSE
	fi
	rm -rf "$work"
done

cp scripts/install.sh "$dist/install.sh"
cp scripts/install.ps1 "$dist/install.ps1"
cd "$dist"
set -- ./*.tar.gz ./*.zip ./install.sh ./install.ps1
files=
for file in "$@"; do
	[ -e "$file" ] && files="$files $file"
done
# shellcheck disable=SC2086 # file names contain no spaces
if command -v sha256sum >/dev/null 2>&1; then
	sha256sum $files | sed 's# \./# #' > checksums.txt
else
	shasum -a 256 $files | sed 's# \./# #' > checksums.txt
fi
cat checksums.txt
