# Knowledge format

Configuration is explicit: read commands default to `./nexora.yaml`; `--config`
selects another file. Relative source paths resolve against that file, not the
working directory. Sources can point to shared local directories outside a
project. Roots must be readable, distinct, and non-overlapping. Source root paths
are canonicalized, but symlink entries inside source trees are rejected.

```yaml
schema_version: 1
project: commerce-api
sources:
  - id: local
    path: ./knowledge
```

`project`, source `id`, and source `path` must be nonempty. Source IDs must be
unique. Unknown fields, duplicate keys, extra YAML documents, explicit nulls,
and values with incorrect types fail validation. Optional fields should be
omitted rather than set to null. Configuration does not expand environment
variables or `~` and does not discover parent configurations.

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
README files outside those directories. A typical source contains `rules/`,
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
