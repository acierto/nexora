## 1. Windows release binaries

- [x] 1.1 Add windows/amd64 and windows/arm64 to release-build.sh, packaging `nexora.exe`, README.md and LICENSE as version-free zips listed in checksums.txt.
- [x] 1.2 Add scripts/install.ps1 (env vars, arch detection, SHA-256 and version verification, user-PATH only, `irm | iex` safe) and copy it into dist with a checksum.
- [x] 1.3 Support Windows zips and `nexora.exe` in runtimeversion for `mcp --version`; unit-test zip extraction, invalid entries and cache naming.
- [x] 1.4 Update release.yml: install zip if missing, upload zips and install.ps1, mention Windows and the SmartScreen note in release notes.
- [x] 1.5 Add tests/release_test.go running release-build.sh for Windows targets and checking zip contents and checksums. Verify: `go test ./tests -run Release`.

## 2. Real Windows verification

- [ ] 2.1 Build `scripts/release-build.sh v0.0.0-test`, boot the Freeboard Windows 11 ARM64 VM, and install the arm64 zip via install.ps1 with a local NEXORA_DOWNLOAD_BASE.
- [ ] 2.2 Run version, init, validate, sync --agent claude, resolve --agent claude --file README.md and a brief mcp start on arm64 and amd64; fix failures.
- [ ] 2.3 Verify install.ps1 checksum-mismatch failure and `irm | iex` style invocation in the VM.
- [x] 2.4 Add scripts/windows/README.md describing the repeatable VM test without copying VM scripts.

## 3. Install docs

- [x] 3.1 Add a Windows section to docs/install.md (one-liner, manual zip, checksum verification, PATH, SmartScreen); keep go install for other systems; drop the "not yet exercised" sentence only if 2.2 passed.
- [x] 3.2 Update the README install section with a Windows line.

## 4. Agent skill

- [x] 4.1 Add skills/nexora/SKILL.md and correct every command, flag and path against the CLI and guides.
- [x] 4.2 Add scripts/package-skill.sh producing dist/nexora-skill.zip with nexora/ at the root; confirm dist/ is ignored.
- [x] 4.3 Run every skill command as written in a temporary `nexora init --project demo` folder outside the repository.

## 5. Guides and site

- [x] 5.1 Add an "Agent skill" section to docs/agents.md (behavior, Claude Code install paths, Agensi listing TODO) and a README section.
- [x] 5.2 State MIT contribution terms in docs/contributing.md and CONTRIBUTING.md; mention Windows in the release section.
- [x] 5.3 Run scripts/sync_docs.py; add "MIT licensed" by the brand and a linked footer to docs/index.html; confirm search covers new guide text.
- [x] 5.4 Check `sync_docs.py --check` and the rendered Install, Connect agents and Contributing pages and hash links in a browser.

## 6. Checks, changelog and release

- [x] 6.1 Run gofmt, go vet, go test ./... (including tests/), sync_docs --check and the CI release build.
- [x] 6.2 Add the changelog entries and record the release as v1.0.0.
- [ ] 6.3 Validate the change strictly, sync specs and archive the change.
- [ ] 6.4 After approval, merge to main and publish v1.0.0; verify release assets and checksums.
