# Team docs: rules plus searchable documentation

This collection represents a commerce API team that keeps two kinds of knowledge.
Rules and a skill live in `knowledge/` with front matter, exactly like the
[local example](../local/README.md). The team handbook in `handbook/` is ordinary
documentation: Markdown, one HTML runbook, no Nexora metadata. Nexora indexes the
handbook and returns the passages that match a task, next to the rules.

```text
team-docs/
├── nexora.yaml
├── knowledge/
│   ├── rules/           payments/idempotency (src/payments/**), observability/logging
│   └── skills/          issue-refund/SKILL.md
└── handbook/            docs source: no front matter needed
    ├── architecture/overview.md
    ├── auth/token-rotation.md
    ├── payments/refunds.md
    └── runbooks/deploy.html
```

The handbook source is a `docs` source; the index uses the defaults: a local
file at `.nexora/index.json`, keyword search, and no embeddings. Nothing here
needs a network connection or an API key.

From the Nexora repository root, with `nexora` built and on PATH:

```sh
nexora validate --config examples/team-docs/nexora.yaml
nexora index --config examples/team-docs/nexora.yaml
nexora search --config examples/team-docs/nexora.yaml --query "rollback production deploy" --limit 1
nexora search --config examples/team-docs/nexora.yaml --query "revoked refresh token reused" --limit 1
nexora resolve --config examples/team-docs/nexora.yaml --agent codex --file src/payments/refund.go --task "add partial refunds" --limit 2
```

Validation reports 3 knowledge items and 1 documentation source. Indexing reports
4 documents and 13 passages. Expected results:

| Request | Result |
| --- | --- |
| Search "rollback production deploy" | `handbook:runbooks/deploy.html#3`, heading `Deploying commerce-api › Rolling back`, matched `rollback`, `production`, `deploy` |
| Search "revoked refresh token reused" | `handbook:auth/token-rotation.md#2`, heading `Token rotation › Rotating refresh tokens` |
| Resolve with the task | Required rules `payments/idempotency`, `observability/logging`; skill `skills/issue-refund`; passages `handbook:payments/refunds.md#3` (Partial refunds) and `#4` (Provider limits) |

The required rules and the skill are the same with or without `--task`: they are
selected by project and file, never by search. The passages are extra reference
material, with their location and the terms that matched.

Things to try:

- Ask `--query "how long can customers get a refund"`. Keyword search finds the
  refund policy but not the most precise section, because the words differ from
  the text ("30 days of delivery"). Configure an embedding provider in
  `nexora.yaml` (see [Docs and Confluence](../../docs/retrieval.md)) and run
  `nexora index` again to match by meaning.
- Run `nexora sync --agent codex --config examples/team-docs/nexora.yaml --dir /tmp/team-docs-sync`
  and open the generated `AGENTS.md` and `.agents/skills/issue-refund/SKILL.md`
  (see [Connect agents](../../docs/agents.md)).

The index is written to `examples/team-docs/.nexora/`, which is ignored by Git.
