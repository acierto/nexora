# Verification — 2026-10-30

- `go test ./tests -run TestInstaller -count=1`: passed, using real versioned binaries and local HTTP/TLS archives. Covers equal/newer stable reuse without network, fresh/older/development/prerelease installation, explicit downgrade, invalid ensure pins, checksum/certificate/version mismatch, preservation and staging cleanup, custom paths and directory rejection.
- `go test ./...`: passed.
- `go vet ./...`: passed.
- `gofmt` applied to changed Go test; `git diff --check`: passed.
- `scripts/release-build.sh v0.4.1`: all four supported targets built.
- `shasum -a 256 -c checksums.txt`: all four archives and install.sh passed; packaged installer matches scripts/install.sh.
- `openspec validate repeatable-cli-installation --strict`: passed.

Implementation and durable decisions are recorded in docs/install.md and docs/architecture.md. No external service credentials or employer content are used in tests or documentation.
