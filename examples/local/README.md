# Go API: practical development knowledge

This collection represents a commerce API with authentication, HTTP transport,
checkout, inventory, payments, migrations, and webhook workflows. Start with the
authentication walkthrough below, then try the development scenarios. A retired
rule remains inspectable without affecting selection.

```text
local/
├── nexora.yaml
└── knowledge/
    ├── rules/           Go, security, HTTP, orders, inventory, migrations, webhooks
    ├── skills/          Endpoint, checkout, migration, HTTP test, webhook replay procedures
    ├── references/      Auth, order lifecycle, inventory, pagination, webhook contracts
    └── decisions/       Git knowledge, transactional outbox, money representation
```

From the Nexora repository root, with `nexora` built and on PATH:

```sh
nexora validate --config examples/local/nexora.yaml
nexora list --config examples/local/nexora.yaml
nexora resolve --config examples/local/nexora.yaml --agent codex --file src/auth/service.go
nexora resolve --config examples/local/nexora.yaml --agent claude --file src/auth/service.go --json
nexora resolve --config examples/local/nexora.yaml --file src/catalog/service.go --json
```

Validation reports 25 items (12 rules, 5 skills, 5 references, 3 decisions). Expected IDs:

| Request | Required rules, in order | Other selected knowledge |
| --- | --- | --- |
| Codex + auth file | `security/authentication`, `go/conventions`, `agents/codex` | `go/naming`, `skills/add-endpoint`, `references/authentication`, `decisions/local-knowledge` |
| Claude + auth file | `security/authentication`, `go/conventions`, `agents/claude` | Same optional knowledge |
| No agent + catalog file | `go/conventions` | `go/naming`, `skills/add-endpoint`, `decisions/local-knowledge` |

`legacy/authentication` is always excluded because it is archived. The other
agent's reminder is excluded. With a catalog file, the authentication rule and
reference are excluded because their path restriction does not match.

The authentication selection explains project `commerce-api` and the pair
`src/auth/**/*.go` / `src/auth/service.go`. Provenance names source `local`,
path `rules/authentication.md`, version `1.0.0`, and the actual file digest.
Unrestricted dimensions are identified explicitly.

Try changing the authentication rule's paths to `src/**/*.go` and rerun the
catalog request. It will now select that rule. Restore the original pattern
afterward. This changes applicability; priority never overrides another rule.

The SKILL.md uses Nexora metadata; it is retrieved as text, not executed. Agent
filter examples do not configure or connect Codex or Claude Code.

## Practical development scenarios

All policies describe a fictional service. The files under `src/` and `migrations/`
in these requests are applicability inputs; no commerce implementation is included.
The knowledge contains invariants, sample contracts, failure cases, and acceptance
criteria that an agent can use when implementing or reviewing such a service.

```sh
nexora resolve --config examples/local/nexora.yaml --file src/http/orders.go --json
nexora resolve --config examples/local/nexora.yaml --file src/orders/checkout.go --json
nexora resolve --config examples/local/nexora.yaml --file src/inventory/reserve.go --json
nexora resolve --config examples/local/nexora.yaml --file migrations/20260101_add_order_version.sql --json
nexora resolve --config examples/local/nexora.yaml --file src/webhooks/payment.go --json
nexora resolve --config examples/local/nexora.yaml --file src/orders/checkout.go --file src/inventory/reserve.go --json
```

Each request still includes `go/conventions`, optional `go/naming`, and decision
`decisions/local-knowledge`. Go source requests also include `skills/add-endpoint`.
Additional selections are:

| File | Required rules, in order before Go baseline | Skills | References and decisions |
| --- | --- | --- | --- |
| `src/http/orders.go` | `security/tenant-isolation`, `api/http-contracts` | `skills/test-http-endpoint` | `references/http-pagination` |
| `src/orders/checkout.go` | `security/tenant-isolation`, `orders/state-transitions` | `skills/implement-checkout` | `references/order-lifecycle`, `decisions/money-minor-units`, `decisions/transactional-outbox` |
| `src/inventory/reserve.go` | `security/tenant-isolation`, `inventory/reservations` | `skills/implement-checkout` | `references/inventory-model` |
| `migrations/20260101_add_order_version.sql` | `database/expand-contract` | `skills/add-migration` | None beyond baseline |
| `src/webhooks/payment.go` | `security/webhook-verification` | `skills/replay-webhook` | `references/webhook-events` |
| `src/payments/capture.go` | None | None beyond baseline | `decisions/money-minor-units` |

The combined checkout/inventory request selects tenant isolation and the checkout
skill once each while retaining both domain rules. A `.md` file under
`src/orders/` does not match the Go-specific order rules. Override the project with
`--project billing-api` and all commerce-scoped additions disappear.

Use these as review exercises: reject a foreign tenant's order, race two last-unit
reservations, retry a timed-out authorization without charging twice, keep old code
working during a schema migration, or replay a signed event without a second effect.
`TestPracticalLocalScenarios` in `tests/integration_test.go` verifies the selections,
normalized duplicate paths, project isolation, and source/version provenance.
