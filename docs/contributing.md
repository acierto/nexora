# Contributing

[Read online](https://acierto.github.io/nexora/#contributing) · [Documentation home](https://acierto.github.io/nexora/)

This guide is for people changing Nexora itself. To use Nexora, start with
[Install](install.md) and the [Quickstart](quickstart.md).

Project rules for humans and coding agents are in [AGENTS.md](../AGENTS.md):
deterministic selection of mandatory rules, minimal dependencies, `internal/`
packages until there is an external consumer, and tests without external
services. Read it before larger changes.

## Set up

You need Go 1.23 or newer and Git. Python 3 is used by one documentation script
and by an optional CLI recipe.

```shell
git clone https://github.com/acierto/nexora.git
cd nexora
go build -o bin/nexora ./cmd/nexora
./bin/nexora version
```

`bin/` and `dist/` are ignored by Git. The module floor is `go 1.23.0`; do not use
standard-library APIs added in later Go releases.

## Repository layout

| Path | Contents |
| --- | --- |
| `cmd/nexora` | Process entry point only |
| `internal/cli` | Command parsing, output, exit codes |
| `internal/app` | Loading and coordination shared by the CLI and MCP server |
| `internal/config`, `internal/yamlutil` | Strict configuration and YAML decoding |
| `internal/knowledge`, `internal/markdown`, `internal/registry`, `internal/resolver` | Knowledge model, loading, snapshots, deterministic resolution |
| `internal/passage`, `internal/confluence`, `internal/embed`, `internal/index`, `internal/retrieval` | Documentation passages, connectors, stores, ranking |
| `internal/mcp`, `internal/nativesync` | Agent delivery |
| `tests/` | Executable-level integration tests, including the documented recipes |
| `examples/` | Runnable sample projects, also used as test fixtures |
| `docs/` | Guides and the documentation website (`docs/index.html`) |
| `openspec/` | Change proposals, designs, specs, and task lists |
| `scripts/` | Release build, install script, documentation sync |

See [architecture decisions](architecture.md) for the dependency graph and the
reasoning behind it.

## Checks

Run these before opening a pull request. CI runs the same steps.

```shell
gofmt -l .                     # must print nothing
go vet ./...
go test ./...
python3 scripts/sync_docs.py --check
```

The tests need no network or services. They build the binary, run the examples,
and execute the `sh` code blocks in `docs/quickstart.md`, `docs/cli.md`,
`docs/troubleshooting.md`, and the example READMEs, so documentation that drifts
from the implementation fails the build. Use `shell` instead of `sh` for commands
that must not run in tests (for example ones that need an agent installed).

The pgvector store has an integration test that runs only when
`NEXORA_TEST_DATABASE_URL` is set:

```shell
docker run -d --name nexora-pg -e POSTGRES_PASSWORD=nexora -e POSTGRES_DB=nexora_test -p 5432:5432 pgvector/pgvector:pg16
NEXORA_TEST_DATABASE_URL="postgres://postgres:nexora@localhost:5432/nexora_test?sslmode=disable" go test ./internal/index/
```

Embedding providers and Confluence are tested against local HTTP test servers;
never call real services from tests.

## Making a change

Larger changes go through OpenSpec, which keeps a proposal, design, specs, and a
task list under `openspec/changes/<name>/`. The repository includes OpenSpec
skills for Claude Code (`.claude/`) and Codex (`.agents/`): `propose`, `apply`,
`update`, `sync`, and `archive`. Validate a change with:

```shell
npx @fission-ai/openspec validate <change-name> --strict --no-interactive
```

Mark tasks done only after their verification passes, and record what you checked
in [the verification record](verification.md). Small fixes and documentation
edits do not need a change proposal.

Keep pull requests focused, include tests for behavior changes, and update the
user guides in the same pull request when behavior or output changes.

## Documentation

User guides live in `docs/*.md` and are written for people using Nexora:
install, quickstart, task guides, and reference. Contributor material goes in
this guide, [architecture.md](architecture.md), and [verification.md](verification.md).

The website at https://acierto.github.io/nexora/ is the single file
`docs/index.html`, published by GitHub Pages from the `docs/` folder on `main`
(`.nojekyll` keeps it a plain static site). It embeds copies of the Markdown
guides. After editing a guide, refresh and check the copies:

```shell
python3 scripts/sync_docs.py
python3 scripts/sync_docs.py --check
```

To add a new guide, add it to `GUIDES` in `scripts/sync_docs.py` and add a
matching `<script type="text/markdown" id="md-<key>">` block, page entry, and
sidebar link in `docs/index.html`.

Preview locally with `python3 -m http.server 8000 --directory docs` and open
http://localhost:8000/. The page loads fonts, Three.js, and the Markdown renderer
from CDNs, so previewing needs an internet connection.

## Releases

Releases are built by GitHub Actions when a version tag is pushed:

```shell
git tag v0.1.0
git push origin v0.1.0
```

Alternatively, run the **Release** workflow from the Actions tab (or
`gh workflow run release.yml -f version=v0.1.0`); it creates the tag on the
selected commit.

The [release workflow](../.github/workflows/release.yml) runs the tests, builds
static binaries for macOS (arm64, amd64) and Linux (amd64, arm64) with
`scripts/release-build.sh`, and publishes `nexora_<os>_<arch>.tar.gz` archives
plus `checksums.txt` to a GitHub release with generated notes. The version is
compiled in and shown by `nexora version`. Archive names carry no version, so the
install script can always download from `releases/latest`.

To try a release build locally:

```shell
scripts/release-build.sh v0.0.0-local
ls dist/
```

To test the install script against that build, serve `dist/` and point the script
at it:

```shell
python3 -m http.server 8765 --directory dist &
NEXORA_DOWNLOAD_BASE=http://127.0.0.1:8765 NEXORA_INSTALL_DIR=/tmp/nexora-bin sh scripts/install.sh
```

License selection is still pending; choose one before the first public release.
