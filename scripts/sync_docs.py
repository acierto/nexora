#!/usr/bin/env python3
"""Refresh the supplied static site's embedded Markdown from the user guides."""

import argparse
from pathlib import Path
import re
import sys


ROOT = Path(__file__).resolve().parents[1]
GUIDES = {
    "start": "getting-started.md",
    "format": "knowledge-format.md",
    "resolution": "resolution.md",
    "cli": "cli.md",
    "troubleshooting": "troubleshooting.md",
    "architecture": "architecture.md",
    "verification": "verification.md",
}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--check", action="store_true", help="fail if embedded guides are out of date")
    args = parser.parse_args()
    site = ROOT / "docs" / "index.html"
    original = site.read_text(encoding="utf-8")
    updated = original
    for key, filename in GUIDES.items():
        markdown = (ROOT / "docs" / filename).read_text(encoding="utf-8")
        if "</script" in markdown.lower():
            raise ValueError(f"{filename}: cannot embed a closing script tag")
        pattern = re.compile(rf'(<script type="text/markdown" id="md-{key}">).*?(</script>)', re.S)
        updated, count = pattern.subn(lambda match: match[1] + markdown + match[2], updated)
        if count != 1:
            raise ValueError(f"Expected one embedded guide for {filename}; found {count}")
    if args.check:
        if original != updated:
            print("Embedded guides are stale; run python3 scripts/sync_docs.py", file=sys.stderr)
            return 1
        print("All seven embedded guides match their Markdown sources.")
    else:
        site.write_text(updated, encoding="utf-8")
        print("Updated all seven embedded guides in docs/index.html.")
    return 0


if __name__ == "__main__":
    sys.exit(main())
