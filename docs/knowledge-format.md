# Knowledge format

[Read online](https://acierto.github.io/nexora/#format) · [Documentation home](https://acierto.github.io/nexora/)

Nexora reads one configuration file, `nexora.yaml`, and two kinds of content:
**knowledge items** (rules, skills, references, and decisions written as Markdown
with front matter, selected deterministically) and **documentation** (ordinary
docs folders and Confluence spaces, indexed into passages and retrieved for a
task). This page covers the configuration and the knowledge format. Retrieval is
described in [Docs and Confluence](retrieval.md).

## Configuration

Configuration is explicit: commands default to `./nexora.yaml`, and `--config`
selects another file. Relative paths resolve against that file, not the working
directory. Sources can point to shared local directories outside a project.

```yaml
schema_version: 1
project: commerce-api
sources:
  - id: local                  # knowledge source (type defaults to knowledge)
    path: ./knowledge
  - id: handbook
    type: docs                 # Markdown, text, and HTML; no front matter needed
    path: ./handbook
    projects: [commerce-api]   # optional
  - id: wiki
    type: confluence
    url: https://your-company.atlassian.net/wiki
    spaces: [PAY, ARCH]
    user_env: CONFLUENCE_USER
    token_env: CONFLUENCE_TOKEN
index:                         # optional; these are the defaults plus an embedding provider
  store: local
  path: .nexora/index.json
  embeddings:
    provider: ollama
    model: nomic-embed-text
  chunk:
    max_chars: 1200
    overlap_chars: 150
```

Only `schema_version`, `project`, and one source are required. The smallest
valid file, which `nexora init` writes, has a single knowledge source.

| Field | Applies to | Meaning / default |
| --- | --- | --- |
| `schema_version` | file | Required; must be `1` |
| `project` | file | Required project ID used when `--project` is omitted |
| `sources[].id` | all sources | Required, unique; appears in provenance and passage IDs |
| `sources[].type` | all sources | `knowledge` (default), `docs`, or `confluence` |
| `sources[].path` | knowledge, docs | Required directory; readable, canonical, not overlapping another root |
| `sources[].projects` | docs, confluence | Only return this source's passages for these projects; omitted means every project |
| `sources[].url` | confluence | Required base URL, for example `https://x.atlassian.net/wiki` |
| `sources[].spaces` | confluence | Required list of unique space keys |
| `sources[].token_env` | confluence | Required name of the environment variable holding the API token |
| `sources[].user_env` | confluence | Name of the variable holding the user (Cloud email); enables basic auth, otherwise bearer |
| `index.store` | index | `local` (default) or `pgvector` |
| `index.path` | local store | Index file, default `.nexora/index.json` |
| `index.dsn_env` / `index.table` | pgvector store | Variable holding the connection string; table, default `nexora_passages` |
| `index.embeddings.*` | index | `provider` (`none`, `ollama`, `openai`), `model`, `url`, `api_key_env`, `dimensions` |
| `index.chunk.*` | index | `max_chars` (200-20000, default 1200), `overlap_chars` (default 150) |

Unknown fields, duplicate keys, extra YAML documents, explicit nulls, and values
with incorrect types fail validation, at every level. Optional fields should be
omitted rather than set to null. Fields that belong to another source type are
rejected: a knowledge source cannot set `url` or `projects`, and a confluence
source cannot set `path`. Configuration never contains secrets and does not
expand environment variables or `~`; fields ending in `_env` name a variable that
is read only when Nexora needs it. Parent directories are not searched.

Source root paths are canonicalized, but symlink entries inside source trees are
rejected.

## Path patterns

File applicability uses repository-relative, case-sensitive paths with `/`
separators. `*` matches zero or more characters within one segment, `?` matches
one character, and a whole `**` segment matches zero or more directories.
`src/**/*.go` matches both `src/main.go` and `src/auth/service.go`.
Brackets, braces, negation, embedded `**`, empty segments, `.`/`..` segments,
absolute patterns, and backslashes in patterns are rejected. Request files may
use Windows separators; absolute, drive-qualified, UNC, and traversal paths fail.

## Markdown metadata

Every `.md` file in a configured source must be UTF-8 and begin with YAML front
matter between `---` delimiter lines, followed by a nonempty Markdown body.
LF and CRLF are accepted and body text is preserved. Other regular files are
ignored. One invalid document rejects the entire load; files are never executed.

| Field | Meaning / default |
| --- | --- |
| `id` | Required stable unique ID; lowercase letters, numbers, `.`, `_`, `/`, `-`; starts with a letter/number; no empty, `.` or `..` slash segments |
| `name` | Required nonempty display name |
| `description` | Required nonempty discovery description |
| `kind` | Required: `rule`, `skill`, `reference`, `decision` |
| `version` | Optional nonempty string; quote numeric-looking versions |
| `status` | `active` (default), `draft`, or `archived` |
| `required` | Rules only; defaults to `true`; `false` makes a rule optional |
| `priority` | Integer, default `0`; controls order, never overrides another rule |
| `tags` | Optional list of strings |
| `applies_to.projects` | Exact case-sensitive project IDs; unrestricted if omitted/empty |
| `applies_to.agents` | Exact case-sensitive agent IDs; unrestricted if omitted/empty |
| `applies_to.paths` | Path patterns; unrestricted if omitted/empty |

### Rule

```markdown
---
id: security/authentication
name: Authentication
description: Authentication requirements for the API
kind: rule
version: "1.0.0"
priority: 100
tags: [go, security]
applies_to:
  projects: [commerce-api]
  agents: [codex, claude]
  paths: ["src/**/*.go"]
---
Never log credentials. Test invalid and expired credentials.
```

### Optional rule

```markdown
---
id: go/naming
name: Naming guidance
description: Suggestions for descriptive names
kind: rule
required: false
---
Prefer domain-specific names for exported operations.
```

### Skill

```markdown
---
id: skills/add-endpoint
name: Add an endpoint
description: A repeatable API implementation procedure
kind: skill
---
1. Define request and response types.
2. Implement validation and the handler.
3. Test success and failure behavior.
```

A file named `SKILL.md` can use this Nexora schema. Arbitrary Agent Skills files
are not imported automatically, and Nexora does not execute skills.

### Reference

```markdown
---
id: references/api-layout
name: API layout
description: Where API code belongs
kind: reference
---
Place HTTP handlers in the transport package and business logic in the service.
```

### Decision

```markdown
---
id: decisions/local-first
name: Local-first storage
description: Why this project keeps knowledge in Git
kind: decision
status: active
---
Store knowledge as reviewable Markdown alongside the project.
```

## Sources and versions

Each item retains the source ID and a `/`-separated path relative to that source,
plus a lowercase SHA-256 digest of the original file bytes. Explicit versions
are retained; absent versions become `sha256:<digest>`. Even a line-ending change
changes the digest. Load times and filesystem modification times are not versions.
Duplicate IDs fail with both locations; source order does not override knowledge.

Use dedicated source directories for metadata-bearing knowledge. Keep ordinary
README files outside those directories, or put ordinary documentation in a
separate `type: docs` source, which needs no front matter. A typical source contains `rules/`,
`skills/`, `references/`, and `decisions/`, but directories do not infer item kinds.

Example of invalid metadata (must fail validation):

```markdown
---
id: skills/invalid
name: Invalid skill
description: Required is only valid on rules
kind: skill
required: true
---
This document is invalid even though its body is present.
```

## Documentation sources

Files in a `docs` source need no Nexora metadata. Nexora reads `.md`,
`.markdown`, `.txt`, `.html`, and `.htm` files in stable order, requires UTF-8,
rejects symlink entries, and skips leading YAML front matter without parsing it,
so existing docs (including Jekyll or Docusaurus pages) work as they are. HTML
and Confluence pages are converted to text that keeps headings, lists, tables,
and code blocks; nothing is rendered or executed.

`nexora index` splits each document into passages at headings and paragraphs.
A passage records its source, document, heading trail (for example
`Refund policy › Partial refunds`), location, and the SHA-256 digest of the
document. Passage IDs have the form `<source>:<document>#<n>`, such as
`handbook:payments/refunds.md#3` or `wiki:page-42#1`.

Documentation never becomes a knowledge item. It has no `kind`, `required`, or
`applies_to`, and it cannot be selected as a required rule. Put instructions an
agent must follow in knowledge sources, and reference material in docs. See
[Docs and Confluence](retrieval.md) for indexing, search, embeddings, and pgvector.
