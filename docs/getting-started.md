# Your first knowledge registry

[Read online](https://acierto.github.io/nexora/#start) · [Documentation home](https://acierto.github.io/nexora/)

Nexora is a standalone command-line binary. No Docker, server, database, API key,
or agent installation is needed. You need Go 1.23 or newer to build from source.
An internet connection may be needed once to download the build dependency;
the built executable works offline. No release binaries are published yet.

From a Nexora checkout:

```sh
go build -o bin/nexora ./cmd/nexora
go install ./cmd/nexora
```

Use the executable in `bin/`, or add Go's install directory to PATH:
`$(go env GOPATH)/bin`, unless you set GOBIN. On Windows, use `nexora.exe` and
add that directory to PATH through your shell settings. The following walkthrough
uses a POSIX shell and an installed `nexora`.

## Initialize an existing project

In an existing project directory (or a new temporary directory), run:

```sh
nexora init --project commerce-api
nexora validate
nexora list
nexora resolve
```

Init prints the two created paths. Validation prints `Valid: 1 knowledge item`.
List shows `project/conventions`; resolve shows `Required rules (1)` and that
rule's content and provenance. The layout is:

```text
nexora.yaml
knowledge/
  rules/
    project.md
```

Running init again fails rather than overwriting your files. It also refuses
symlinked destinations and leaves native agent instructions alone.

## Add a file-specific rule

Create `knowledge/rules/authentication.md` with this complete content:

```markdown
---
id: security/authentication
name: Authentication rules
description: Requirements for authentication code
kind: rule
priority: 100
applies_to:
  projects: [commerce-api]
  paths: ["src/auth/**/*.go"]
---
Never log credentials. Test expired and invalid credentials.
```

Now run:

```sh
nexora validate
nexora resolve --file src/auth/service.go
nexora resolve --file src/catalog/service.go
nexora resolve --project another-project --file src/auth/service.go
nexora resolve --file src/auth/service.go --json
```

Validation now reports 2 items. The authentication request selects
`security/authentication` then `project/conventions`. The catalog request selects
only `project/conventions`. The other-project request selects neither: both
rules explicitly restrict the project. No source files need to exist for these
requests; paths describe the files you plan to change.

JSON exposes the same selection in `required_rules`, with full items and
structured `reasons`; the other four groups are empty arrays. For example, path
evidence includes `{"pattern":"src/auth/**/*.go","file":"src/auth/service.go"}`.

## Add your team's docs

Point a `docs` source at a documentation folder (or a `confluence` source at a
space), index it, and describe your task:

```yaml
sources:
  - id: local
    path: ./knowledge
  - id: handbook
    type: docs
    path: ./docs
```

```shell
nexora index
nexora resolve --file src/auth/service.go --task "rotate refresh tokens"
```

The result keeps the same rules and adds a `passages` group with the matching
documentation sections, their files, and the words that matched. Add `.nexora/`
(the local index) to `.gitignore`. See [Docs and Confluence](retrieval.md) for
Confluence, embeddings, and pgvector.

## Connect your agent

Register the MCP server so the agent can ask for context itself, or write the
always-on rules into its native file:

```shell
claude mcp add nexora -- nexora mcp --config "$PWD/nexora.yaml"
nexora sync --agent codex
```

Nexora delivers text and does not raise its authority above the agent's trusted
instructions. See [Connect agents](agents.md), [resolution](resolution.md), and
the [practical examples](../examples/README.md) for more.
