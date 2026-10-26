# Configuration

[Read online](https://acierto.github.io/nexora/#configuration) · [Documentation home](https://acierto.github.io/nexora/)

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

## Environment variables

Nexora reads only the variables you name in the configuration, and only when it
needs them:

| Field | Read by | Example |
| --- | --- | --- |
| `sources[].token_env` | `nexora index` | `CONFLUENCE_TOKEN` |
| `sources[].user_env` | `nexora index` | `CONFLUENCE_USER` |
| `index.embeddings.api_key_env` | `index`, `search`, `resolve --task`, MCP | `OPENAI_API_KEY` |
| `index.dsn_env` | `index`, `search`, `resolve --task`, MCP | `NEXORA_DATABASE_URL` |

When Nexora runs as an MCP server, these variables must be set in the agent's MCP
server environment, not only in your shell.

## Files Nexora writes

| Path | Written by | Commit it? |
| --- | --- | --- |
| `nexora.yaml`, `knowledge/rules/project.md` | `nexora init` (never overwrites) | Yes |
| `.nexora/index.json` | `nexora index` (local store) | No, add it to `.gitignore` |
| `.nexora/sync-<agent>.json` | `nexora sync` | Yes, together with the generated files |
| `AGENTS.md` / `CLAUDE.md` block, `.agents/skills/`, `.claude/skills/` | `nexora sync` | Yes |

Ignore only `.nexora/index.json`, not the whole `.nexora/` folder, so the sync
manifests stay in version control next to the files they protect.
