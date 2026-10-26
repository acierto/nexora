# Team docs: rules plus searchable documentation

This collection represents a commerce API team that keeps two kinds of knowledge.
Rules, skills, and a reference live in `knowledge/` with front matter, exactly like the
[local example](../local/README.md). The team handbook in `handbook/` is ordinary
documentation: Markdown, one HTML runbook, no Nexora metadata. Nexora indexes the
handbook and returns the passages that match a task, next to the rules.

```text
team-docs/
├── nexora.yaml
├── knowledge/
│   ├── rules/           payments, logging, webhook security, migrations
│   ├── skills/          issue-refund and diagnose-inventory
│   └── references/      support export redaction
└── handbook/            docs source: no front matter needed
    ├── architecture/overview.md
    ├── auth/token-rotation.md
    ├── payments/refunds.md
    ├── runbooks/deploy.html
    └── orders/, inventory/, webhooks/, database/, operations/, security/, api/, support/
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

Validation reports 7 knowledge items and 1 documentation source. Indexing reports
12 documents and 37 passages. Expected results:

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

## Retrieval evaluation scenarios

The larger handbook covers checkout recovery, stock reconciliation, webhook
verification, migration compatibility, outbox operations, erasure, pagination,
and support diagnostics. It intentionally contains related concepts in several
documents: an authorization timeout also appears in support triage, while stock
expiry appears beside checkout recovery. This gives keyword retrieval competing
passages instead of a single obvious document.

After indexing, try these queries with `--limit 1`:

| Query | Expected document | Expected section |
| --- | --- | --- |
| `ambiguous authorization` | `orders/checkout.md` | Ambiguous authorization |
| `reservation drift investigation` | `inventory/reconciliation.md` | Reservation drift investigation |
| `signature mismatch diagnosis` | `webhooks/delivery.md` | Signature mismatch diagnosis |
| `expand contract backfill` | `database/migrations.md` | Expand contract backfill |
| `relay backlog triage` | `operations/outbox.md` | Relay backlog triage |
| `erasure legal holds` | `security/data-deletion.md` | Erasure and legal holds |
| `equal timestamp regression` | `api/pagination.md` | Equal timestamp regression |
| `safe diagnostic export` | `support/incident-triage.md` | Safe diagnostic export |

```sh
nexora search --config examples/team-docs/nexora.yaml --query "reservation drift investigation" --limit 1
nexora resolve --config examples/team-docs/nexora.yaml --file src/webhooks/payment.go --task "signature mismatch diagnosis" --limit 1
nexora resolve --config examples/team-docs/nexora.yaml --file migrations/20260101_add_order_version.sql --task "expand contract backfill" --limit 1
```

The webhook request selects `security/webhook-verification` and
`observability/logging`; the migration request selects `database/expand-contract`
and `observability/logging`. Task retrieval adds reference passages without changing
those rules. Inventory files select `skills/diagnose-inventory`, and support files
select `references/support-redaction` alongside the logging baseline.

`TestTeamDocsExample` in `tests/retrieval_test.go` checks all eight query/section
pairs and an unmatched query, as well as the original refund, token, and deploy
scenarios. These are reproducible keyword regression cases, not a measurement of
semantic retrieval quality. Try paraphrases to explore where lexical matching
needs different query wording or an optional embedding provider.
