# Docs and Confluence

[Read online](https://acierto.github.io/nexora/#retrieval) · [Documentation home](https://acierto.github.io/nexora/)

Rules and skills tell an agent how to work. Most of what your team knows lives
somewhere else: architecture notes, runbooks, policies, and Confluence pages.
Nexora indexes that documentation and returns the passages that matter for a
task, each with its source and the reason it was picked.

Documentation never becomes a rule. Required rules are still selected only by
project, agent, and file metadata; retrieval cannot add, remove, or reorder them.

## Add documentation sources

Sources have a `type`. The default, `knowledge`, holds front-matter knowledge as
before. Add `docs` for local documentation folders and `confluence` for spaces:

```yaml
schema_version: 1
project: commerce-api
sources:
  - id: rules
    path: ./knowledge
  - id: handbook
    type: docs
    path: ./handbook
  - id: wiki
    type: confluence
    url: https://your-company.atlassian.net/wiki
    spaces: [PAY, ARCH]
    user_env: CONFLUENCE_USER      # Cloud: your Atlassian email
    token_env: CONFLUENCE_TOKEN    # Cloud API token or Data Center personal access token
    projects: [commerce-api]       # optional: only return these passages for this project
```

A `docs` source reads `.md`, `.markdown`, `.txt`, `.html`, and `.htm` files in
stable order. Files need no front matter; any leading YAML front matter is skipped.
Like knowledge sources, docs roots must not overlap other roots, symlink entries
are rejected, and files must be UTF-8. Nothing is rendered or executed.

A `confluence` source reads current pages of each listed space through the
Confluence REST API (`/rest/api/content` with `body.storage`). With `user_env`
set, Nexora uses basic authentication (Cloud email plus API token). Without it,
the token is sent as a bearer token (Data Center personal access tokens).
Credentials are read from the named environment variables only when indexing;
configuration never contains secrets.

`projects` limits a documentation source to the listed projects. Without it,
passages are available to every project. This is an applicability filter, not
authorization.

## Build the index

Indexing is an explicit step. It is the only time Nexora reads documentation or
calls Confluence:

```shell
nexora index                 # every docs and confluence source
nexora index --source wiki   # rebuild one source, keep the others
nexora index --json          # report as JSON
```

Each document is split into passages at Markdown headings (HTML and Confluence
headings are converted first), packed by paragraph up to `max_chars` characters
(1200 by default), with up to `overlap_chars` (150 by default) repeated from the
previous passage of the same section. A passage keeps its source, title, heading
trail such as `Refund policy › Partial refunds`, location (relative path or page
URL), and the SHA-256 digest of its document. Passage IDs look like
`handbook:payments/refunds.md#3` or `wiki:page-42#1` and stay stable while a
document is unchanged.

If reading a source or calling an embedding provider fails, the command exits 1
and the stored index is left as it was.

## Search and resolve with a task

```shell
nexora search --query "rollback production deploy" --limit 3
nexora resolve --agent codex --file src/payments/refund.go --task "add partial refunds"
```

`search` returns passages. `resolve --task` returns the usual rule, skill,
reference, and decision groups plus a `passages` group. The `passages` array is
always present in JSON and empty without `--task`. `--limit` (1-50, default 5)
caps the number of passages.

If documentation sources are configured but nothing has been indexed, `--task`
and `search` fail with `run nexora index`. Without documentation sources, the
`passages` group is simply empty.

## How passages are ranked

Every hit explains itself:

```json
{
  "passage": {
    "id": "handbook:runbooks/deploy.html#3",
    "source_id": "handbook",
    "source_type": "docs",
    "document_id": "runbooks/deploy.html",
    "title": "Deploying commerce-api",
    "heading": "Deploying commerce-api › Rolling back",
    "text": "Use the pipeline’s rollback action. It redeploys the previous image within five minutes. …",
    "location": "runbooks/deploy.html",
    "digest": "f3d8438b4070ec8f5bf9bd774536aaf14cf1dd2f6847c21b380a75b5a20b29bb",
    "projects": []
  },
  "score": 0.016393,
  "reasons": [
    { "method": "keyword", "rank": 1, "score": 5.5829, "terms": ["rollback", "production", "deploy"] }
  ]
}
```

Without embeddings, Nexora ranks with BM25 keyword scoring. Words are lowercased,
common stop words are dropped, and a small suffix rule lets `refunds` and
`refunded` match `refund`; matched terms are shown in that normalized form. With
embeddings, a second list ranks passages by cosine similarity. The two lists are
merged with reciprocal rank fusion (k = 60), so a passage found by both methods
rises to the top, and each method's rank appears in `reasons`. Ties are broken by
passage ID, so the same index and query always give the same result.

## Embeddings (optional)

Keyword search works offline with no API key. To match by meaning, configure an
embedding provider:

```yaml
index:
  embeddings:
    provider: ollama               # local, no key
    model: nomic-embed-text
    # url: http://localhost:11434  # default
```

```yaml
index:
  embeddings:
    provider: openai               # any OpenAI-compatible /embeddings endpoint
    model: text-embedding-3-small
    api_key_env: OPENAI_API_KEY
    # url: https://api.openai.com/v1
    # dimensions: 1536             # optional; must match what the model returns
```

The index records the model. Searching with a different model, or with no model
when the index has embeddings, fails and asks you to run `nexora index` again.
Rebuilding one source with `--source` also requires the same model as the rest of
the index; run `nexora index` without `--source` to switch models.

Queries are embedded too, so `search` and `resolve --task` call the provider when
embeddings are configured.

## pgvector (optional)

The default store is one JSON file at `.nexora/index.json` next to the
configuration (change it with `index.path`). Add `.nexora/index.json` to `.gitignore`.
For larger collections or a shared database, store passages in PostgreSQL with
the [pgvector](https://github.com/pgvector/pgvector) extension:

```yaml
index:
  store: pgvector
  dsn_env: NEXORA_DATABASE_URL     # e.g. postgres://user:pass@localhost:5432/nexora?sslmode=disable
  table: nexora_passages           # default
  embeddings:
    provider: ollama
    model: nomic-embed-text
    dimensions: 768                # enables an HNSW cosine index
```

Nexora runs `CREATE EXTENSION IF NOT EXISTS vector`, so the database user needs
permission for that, or the extension must already be installed. It creates the
passage table with a generated full-text column (GIN index), a `vector` column
(HNSW index when `dimensions` is set), and a `<table>_meta` table for the model.
A full `nexora index` recreates both tables; `--source` replaces one source in a
transaction. Keyword candidates come from PostgreSQL full-text search and
semantic candidates from pgvector's cosine distance; the fusion and explanations
are the same as the local store.

## Configuration reference

All source and index fields, with defaults, are listed in the
[configuration reference](configuration.md).

## Trust boundary

Everything the Confluence token can read is indexed. Anyone who can read the
index file or database can read those passages, and `projects` filters do not
change that. Use a token with access to the spaces you mean to share, and keep
index files out of version control. Retrieved passages are reference material for
an agent; they do not outrank its trusted instructions, and Nexora never executes
their content.

Try it with the [team docs example](../examples/team-docs/README.md), then
[connect your agents](agents.md).
