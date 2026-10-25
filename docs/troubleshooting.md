# Troubleshooting

[Read online](https://acierto.github.io/nexora/#troubleshooting) · [Documentation home](https://acierto.github.io/nexora/)

Start with `nexora validate --config path/to/nexora.yaml`. Every configured
Markdown document must be valid before any command exposes registry data.
Errors name the configuration or source and file. Commands return exit 1 for
these failures; unknown commands/flags return 2.

| Symptom / reproduction | Why | Fix |
| --- | --- | --- |
| Run `nexora validate` in a directory without nexora.yaml | Config is explicit; parents are not searched | Run init there or pass `--config` |
| Put plain text or unclosed `---` in a source `.md` | Missing/invalid front matter | Add the required metadata and closing delimiter from the authoring guide |
| Copy a knowledge file without changing `id` | IDs are unique across the entire registry | Give the copy a fresh ID, or remove the redundant file |
| Set `applies_to.paths: ["src/**.go"]` | `**` must be a whole path segment | Use `src/**/*.go` |
| Link a file or directory inside a source | Symlink entries are rejected | Store real files there or configure a separate canonical root |
| Configure both `knowledge` and `knowledge/rules` as sources | Roots overlap | Configure the parent once, or use disjoint roots |
| Run init a second time | Init refuses existing targets | Edit existing configuration and rules; do not delete user knowledge just to reinitialize |
| Set `version: 1.0` | YAML numeric value where a string is required | Quote it: `version: "1.0"` |
| Use `required: true` on a skill | Only rules can be mandatory | Remove `required`; skills are applicable optional procedures |
| `no documentation index found; run nexora index` | Docs or confluence sources are configured but not indexed yet | Run `nexora index` |
| `index was built with embeddings "X" but the configuration uses "Y"` | The embedding provider or model changed | Run `nexora index` without `--source` to rebuild every source |
| `environment variable CONFLUENCE_TOKEN is not set` | Credentials are read from the variables named in `token_env`/`user_env` | Export the variable in the shell or MCP client that runs Nexora |
| Confluence `HTTP 401` or `HTTP 403` for a space | Wrong user/token pair, or the token cannot read that space | Cloud: set `user_env` to your email and use an API token. Data Center: omit `user_env` and use a personal access token |
| `embeddings (...): HTTP 404: model not found` | The model is not available at the provider URL | Pull the model (`ollama pull nomic-embed-text`) or fix `model`/`url` |
| pgvector `extension "vector" is not available` | The PostgreSQL server lacks pgvector | Install pgvector on the server, or have an administrator run `CREATE EXTENSION vector` |
| A search finds nothing useful | Keyword search needs shared words with the text | Rephrase with the document's terms, or configure embeddings and reindex |
| `sync stopped; nothing was written` | A generated file or the managed block was edited by hand, or an existing skill file has the same name | Move your edits into the knowledge files, delete or rename the conflicting file, then sync again |
| `found <!-- nexora:begin --> without <!-- nexora:end -->` | The managed block markers were damaged | Restore both markers or remove the partial block, then sync again |

These cases are reproduced in configuration, Markdown, initialization, and CLI
tests. For permission failures, ensure the current user can read every source
and file; do not run the CLI as a privileged user to bypass the diagnosis.

## Why did a rule disappear?

List includes inactive and nonmatching items; resolve selects only active matches.
Inspect the item's metadata in `nexora list --json` and check each dimension:

| Check | Exclusion | Correction |
| --- | --- | --- |
| Status | `draft` or `archived` | Set `active` if the rule is ready to apply |
| Project | Config/default override differs from `applies_to.projects` | Supply the intended `--project` or correct the metadata |
| Agent | Restricted agent missing or mismatched, including case | Supply the exact `--agent` value |
| Files | No `--file`, wrong extension, or no matching pattern | Supply relevant repository-relative paths with repeatable flags |

From the repository root, this request has no file, so authentication is excluded:

```sh
nexora resolve --config examples/local/nexora.yaml --agent codex --json
```

This corrected request includes the required file evidence:

```sh
nexora resolve --config examples/local/nexora.yaml --agent codex --file src/auth/service.go --json
```

The first result's required IDs are `go/conventions`, `agents/codex`; the second
adds `security/authentication` first. Setting high priority cannot force a rule
to match. Duplicate IDs cannot be used as overrides. See [resolution](resolution.md).

## Current boundaries

Nexora serves knowledge over a local stdio MCP server and writes native
instruction files, but it cannot make an agent call a tool or follow a rule. It
does not import arbitrary Agent Skills libraries, execute procedures, truncate to
a token budget, serve MCP over HTTP, or provide remote authorization. Indexed
Confluence content is readable by anyone who can read the index file or database.
Retrieved text does not outrank trusted instructions.

Source files are trusted local data. Filters help select relevant items, not
hide secrets from someone with filesystem access. Concurrent malicious changes
to the filesystem are outside this MVP's security guarantee. Keep proprietary
workplace knowledge and private configuration outside the public repository.
