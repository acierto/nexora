# CLI guide

[Read online](https://acierto.github.io/nexora/#cli) · [Documentation home](https://acierto.github.io/nexora/)

Commands use flags after the command name; positional arguments are rejected.
`nexora help` and `nexora <command> --help` describe available flags.

| Command | Flags | Result |
| --- | --- | --- |
| `init` | `--project ID` | Creates configuration and a scoped example rule; default ID is directory basename |
| `validate` | `--config PATH`, `--json` | Loads every source; reports validity and item count |
| `list` | `--config PATH`, `--json` | Lists active, draft, and archived items in ID order |
| `resolve` | `--config PATH`, `--json`, `--project ID`, `--agent ID`, repeatable `--file PATH` | Returns active applicable knowledge |

The default config is `./nexora.yaml`; Nexora never searches parent directories.
Relative sources resolve against the config's directory. Project defaults to its
configured value; an invocation override never modifies that file. An absent
agent excludes agent-restricted items. No files means path-restricted items are
excluded. A supplied empty project is distinct from omitting the flag.

## Recipes

Run these from the Nexora repository root with the installed binary on PATH:

```sh
nexora validate --config examples/local/nexora.yaml
nexora list --config examples/local/nexora.yaml
nexora resolve --config examples/local/nexora.yaml --agent codex --file src/auth/service.go
nexora resolve --config examples/local/nexora.yaml --agent claude --file src/auth/service.go
nexora resolve --config examples/local/nexora.yaml --agent codex --file src/auth/service.go --file src/catalog/service.go
nexora resolve --config examples/local/nexora.yaml --project unrelated --agent codex --file src/auth/service.go
```

Validation reports 9 items. Codex selects required IDs `security/authentication`,
`go/conventions`, `agents/codex`; Claude substitutes `agents/claude` for the last
ID. Repeated file flags select the union of file matches without duplicate items.
The unrelated project selects only the global `go/conventions` and agent-specific
`agents/codex` rules; the authentication rule is explicitly project-restricted.

`--agent` is a filter value, not a connection to an agent. Agent integration,
skill execution, automatic instruction synchronization, and instruction
enforcement are not implemented.

## JSON and shell scripts

Validation produces `{"valid":true,"item_count":9}`. List produces
`{"items":[...]}` with full knowledge values. Resolution produces these arrays:

```json
{"required_rules":[],"optional_rules":[],"skills":[],"references":[],"decisions":[]}
```

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
Configuration, source validation, path validation, and I/O failures exit 1.
Failures use stderr and do not return a partial JSON registry. Human-readable
output also includes content, provenance, and selection reasons. Output is data;
do not execute knowledge content as shell commands.
