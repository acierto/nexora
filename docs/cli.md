# CLI reference

[Read online](https://acierto.github.io/nexora/#cli) · [Documentation home](https://acierto.github.io/nexora/)

Commands use flags after the command name; positional arguments are rejected.
`nexora help` and `nexora <command> --help` describe available flags.

| Command | Flags | Result |
| --- | --- | --- |
| `init` | `--project ID` | Creates configuration and a scoped example rule; default ID is directory basename |
| `validate` | `--config PATH`, `--json` | Loads every source; reports validity and item count |
| `list` | `--config PATH`, `--json` | Lists active, draft, and archived items in ID order |
| `resolve` | `--config PATH`, `--json`, `--project ID`, `--agent ID`, repeatable `--file PATH`, `--task TEXT`, `--limit N`, `--native` | Returns active applicable knowledge; with `--task`, also documentation passages |
| `index` | `--config PATH`, `--json`, `--source ID` | Indexes every docs, confluence, and pgvector source, or one |
| `search` | `--config PATH`, `--json`, `--query TEXT` (required), `--project ID`, `--limit N` | Returns documentation passages with explanations |
| `mcp` | `--config PATH` | Serves knowledge over MCP on stdin/stdout until EOF |
| `sync` | `--config PATH`, `--json`, `--agent codex\|claude` (required), `--project ID`, `--dir PATH`, `--check` | Writes the managed block and generated skills; `--check` writes nothing |
| `version` | none (also `--version`) | Prints the version, for example `nexora v0.1.0` |

The default config is `./nexora.yaml`; Nexora never searches parent directories.
Relative sources resolve against the config's directory. Project defaults to its
configured value; an invocation override never modifies that file. An absent
agent excludes agent-restricted items. No files means path-restricted items are
excluded. A supplied empty project is distinct from omitting the flag.
`--limit` accepts 1 to 50 and defaults to 5.

## Recipes

These use the sample projects in the repository. Clone it and run them from its root
with `nexora` on your PATH:

```sh
nexora validate --config examples/local/nexora.yaml
nexora list --config examples/local/nexora.yaml
nexora resolve --config examples/local/nexora.yaml --agent codex --file src/auth/service.go
nexora resolve --config examples/local/nexora.yaml --agent claude --file src/auth/service.go
nexora resolve --config examples/local/nexora.yaml --agent codex --file src/auth/service.go --file src/catalog/service.go
nexora resolve --config examples/local/nexora.yaml --project unrelated --agent codex --file src/auth/service.go
```

Validation reports 25 items. Codex selects required IDs `security/authentication`,
`go/conventions`, `agents/codex`; Claude substitutes `agents/claude` for the last
ID. Repeated file flags select the union of file matches without duplicate items.
The unrelated project selects only the global `go/conventions` and agent-specific
`agents/codex` rules; the authentication rule is explicitly project-restricted.

`--agent` is a filter value, not a connection to an agent. To deliver knowledge
to an agent, use the MCP server or native sync described in
[Connect agents](agents.md). Nexora never executes skills or enforces instructions.

## Documentation recipes

These use the [team docs example](../examples/team-docs/README.md) and write its
index to `examples/team-docs/.nexora/`:

```sh
nexora index --config examples/team-docs/nexora.yaml
nexora search --config examples/team-docs/nexora.yaml --query "rollback production deploy" --limit 1
nexora resolve --config examples/team-docs/nexora.yaml --agent codex --file src/payments/refund.go --task "add partial refunds" --limit 2
```

Indexing reports 12 documents and 37 passages for the `handbook` source. The search
returns the runbook passage `handbook:runbooks/deploy.html#3`. The resolve call
returns the required rules `payments/idempotency` and `observability/logging`, the
skill `skills/issue-refund`, and two refund-policy passages.

To serve the same knowledge to an agent, or to write it into native files:

```shell
nexora mcp --config "$PWD/nexora.yaml"
nexora sync --agent codex
nexora sync --agent claude --check
```

`sync --check` exits 1 when files would change or when a generated file was edited
by hand, which makes it suitable for CI.

## JSON and shell scripts

Validation produces `{"valid":true,"item_count":25}`. List produces
`{"items":[...]}` with full knowledge values. Resolution produces these arrays:

```json
{"required_rules":[],"optional_rules":[],"skills":[],"references":[],"decisions":[],"passages":[],"documentation":{"status":"not_requested"}}
```

`documentation.status` is `not_requested` without a task, `not_configured` when
no document sources or external guidance exist, `external` when lookup belongs
to the configured existing service, `complete` after successful Nexora retrieval,
or `unavailable` on retrieval failure. `complete` with an empty passage array
means the query had no matches, not an outage. Guidance and failure messages are
included when applicable.

With `--native`, selected required rules and skills are represented in
`native_knowledge`, with ID, kind, required status, source/version, reasons, and
native file path. Full optional rules, references, and decisions remain in their
normal arrays. This requires `--agent codex|claude` and verified up-to-date sync
next to the config. See [native context delivery](agents.md#avoid-repeating-native-knowledge).

`passages` is empty unless `--task` is given. Search produces `{"hits":[...]}`.
A hit is `{"passage":{...},"score":...,"reasons":[...]}`; passage fields are `id`,
`source_id`, `source_type`, `document_id`, `title`, `heading`, `text`, `location`,
`digest`, and `projects`, and each reason has `method` (`keyword` or `semantic`),
`rank`, `score`, and the matched `terms`. Exact linked copies can include
`duplicates` citations (`id`, `source_id`, `title`, `location`, `digest`). Index produces
`{"store":...,"model":...,"sources":[{"id","type","documents","passages","embedded"}]}`.
Sync produces `{"agent","project","changes":[{"path","action"}],"conflicts":[]}`.

Each selection is `{"item":{...},"reasons":[...]}`. Item fields include `id`,
`name`, `description`, `kind`, `content`, `version`, `status`, `tags`, `priority`,
`required`, `applies_to`, and `source`. Source holds `id`, relative `path`, and
`digest`. Reasons hold `dimension`, `unrestricted`, `value`, and `matches`;
path matches contain `pattern` and `file`. Empty collections are `[]`, not null.

This POSIX recipe uses Python 3, an optional scripting helper, to print required
IDs. The intermediate command checks Nexora's status before parsing JSON:

```sh
context=$(nexora resolve --config examples/local/nexora.yaml --agent codex --file src/auth/service.go --json) || exit 1
printf '%s\n' "$context" | python3 -c 'import json,sys; print("\n".join(s["item"]["id"] for s in json.load(sys.stdin)["required_rules"]))'
```

Expected output:

```text
security/authentication
go/conventions
agents/codex
```

Success exits 0; unknown commands, flags, and invalid CLI usage exit 2.
Configuration, source validation, path validation, I/O, missing index, embedding,
Confluence, and sync conflict failures exit 1.
Configuration and knowledge failures use stderr and do not return partial
context. A documentation retrieval failure during `resolve --task` is different:
stdout retains selected knowledge with `documentation.status: unavailable`, stderr
reports the problem, and the command exits 1. Consumers must inspect exit status
and documentation status. Standalone `search` still fails without partial hits. Human-readable
output also includes content, provenance, and selection reasons. Output is data;
do not execute knowledge content as shell commands.
