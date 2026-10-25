# How resolution works

[Read online](https://acierto.github.io/nexora/#resolution) · [Documentation home](https://acierto.github.io/nexora/)

Nexora selects active knowledge by explicit metadata. Every restricted dimension
must match: project AND agent AND files. Within a dimension values are
alternatives. Omitted or empty lists are unrestricted. A missing request value
does not match a restricted dimension. Files describe a task; they need not exist.

For example, a rule restricted to project `commerce-api`, agent `codex`, and
pattern `src/**/*.go` matches a request for `src/main.go` with that project and
agent. It is excluded if the agent is missing, the project differs, no files are
supplied, or the item is draft/archived. See the [path grammar](knowledge-format.md).

All applicable required rules are returned, even when priorities differ. Higher
priority sorts first; ties sort by ID. Priority does not replace or override a
rule. Results contain `required_rules`, `optional_rules`, `skills`, `references`,
and `decisions`; each selected item appears once, with full content, source,
version, and reasons for active status and project/agent/path matching. Path
evidence records matching patterns and normalized file names in stable order.
Duplicate request files do not duplicate knowledge or evidence.

Optional matches are applicability matches, not semantic recommendations. There
is no token budget, truncation, or inferred conflict resolution. Inspect both
rules if their content conflicts and edit your knowledge explicitly.

## Passages from documentation

With `--task`, resolution also searches indexed documentation and adds a sixth
array, `passages`. It is always present in JSON and empty without a task.
Passages come from `docs` and `confluence` sources, are ranked by keyword search
and optional embeddings, and carry their own explanations (see
[Docs and Confluence](retrieval.md)). They are reference material, never rules:
the five knowledge groups are computed first and are identical with or without
`--task`. A passage cannot add, remove, reorder, or demote a required rule.
Passages from a source with `projects` appear only for those projects.

This is a single-user local registry. Project and agent filters are not
authorization. Anyone with access to source files can read them. Selected text
remains retrieved knowledge; it does not outrank the trusted instructions of the
agent receiving it. `--agent codex` or `--agent claude` only supplies a matching
value; it does not connect to either agent. To deliver results to an agent, see
[Connect agents](agents.md).
