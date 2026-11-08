#!/bin/sh
# Package the Nexora agent skill for upload to skill marketplaces.
#
#   scripts/package-skill.sh
#
# Writes dist/nexora-skill.zip with the nexora/ folder at the top level.
set -eu

root=$(cd "$(dirname "$0")/.." && pwd)
command -v zip >/dev/null 2>&1 || { echo "package-skill: zip is required" >&2; exit 1; }
[ -f "$root/skills/nexora/SKILL.md" ] || { echo "package-skill: skills/nexora/SKILL.md is missing" >&2; exit 1; }
mkdir -p "$root/dist"
rm -f "$root/dist/nexora-skill.zip"
cd "$root/skills"
zip -q -r -X "$root/dist/nexora-skill.zip" nexora -x '*/.DS_Store'
echo "$root/dist/nexora-skill.zip"
