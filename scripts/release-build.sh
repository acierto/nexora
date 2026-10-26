#!/bin/sh
# Build release archives for the supported platforms into dist/.
#
#   scripts/release-build.sh v0.1.0
#
# Produces nexora_<os>_<arch>.tar.gz (binary, README.md) for each target and
# checksums.txt with SHA-256 sums. Archive names carry no version so that
# https://github.com/acierto/nexora/releases/latest/download/<name> always
# points at the newest release. Binaries are static (CGO_ENABLED=0).
set -eu

version=${1:?usage: scripts/release-build.sh <version, e.g. v0.1.0>}
targets=${NEXORA_TARGETS:-"darwin/arm64 darwin/amd64 linux/amd64 linux/arm64"}
root=$(cd "$(dirname "$0")/.." && pwd)
dist="$root/dist"

rm -rf "$dist"
mkdir -p "$dist"
cd "$root"

for target in $targets; do
	os=${target%/*}
	arch=${target#*/}
	name="nexora_${os}_${arch}"
	work="$dist/$name"
	mkdir -p "$work"
	echo "building $target"
	CGO_ENABLED=0 GOOS=$os GOARCH=$arch go build -trimpath \
		-ldflags "-s -w -X github.com/acierto/nexora/internal/cli.Version=$version" \
		-o "$work/nexora" ./cmd/nexora
	cp README.md "$work/README.md"
	tar -C "$work" -czf "$dist/$name.tar.gz" nexora README.md
	rm -rf "$work"
done

cd "$dist"
if command -v sha256sum >/dev/null 2>&1; then
	sha256sum ./*.tar.gz | sed 's# \./# #' > checksums.txt
else
	shasum -a 256 ./*.tar.gz | sed 's# \./# #' > checksums.txt
fi
cat checksums.txt
