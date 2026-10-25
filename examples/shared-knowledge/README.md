# Share standards across two projects

Two local projects reuse the same standards directory while loading their own
knowledge independently. No server or registry database is required.

```text
shared-knowledge/
├── standards/
│   ├── go.md
│   └── authentication.md
└── projects/
    ├── commerce-api/
    │   ├── nexora.yaml
    │   └── knowledge/errors.md
    └── billing-api/
        ├── nexora.yaml
        └── knowledge/errors.md
```

Each config loads `../../standards` as source `shared` and `./knowledge` as source
`project`. These paths resolve against that config, independent of your working
directory. From the Nexora repository root with `nexora` on PATH:

```sh
nexora validate --config examples/shared-knowledge/projects/commerce-api/nexora.yaml
nexora validate --config examples/shared-knowledge/projects/billing-api/nexora.yaml
nexora resolve --config examples/shared-knowledge/projects/commerce-api/nexora.yaml --file src/auth/service.go --json
nexora resolve --config examples/shared-knowledge/projects/billing-api/nexora.yaml --file src/auth/service.go --json
nexora resolve --config examples/shared-knowledge/projects/commerce-api/nexora.yaml --project billing-api --file src/auth/service.go --json
```

Both configurations validate with 3 items. Expected required IDs:

| Request | IDs in order |
| --- | --- |
| Commerce config, default project | `security/authentication`, `go/conventions`, `commerce/errors` |
| Billing config, default project | `go/conventions`, `billing/errors` |
| Commerce config, project overridden to billing | `go/conventions` |

All optional groups are empty. Both projects get the shared baseline. The
authentication rule is physically readable in both registries but applies only
to commerce. The override does not load billing's private source automatically;
it changes selection within the configured registry only. Local filters are not
authorization or protection from someone who can read the files.

Project rules add to shared rules; they do not override them. Reusing an ID in
two sources causes a duplicate-ID error instead of replacing a rule.

Try adding a new global Markdown rule under `standards/`, with a fresh ID. Both
configurations will list and resolve it on their next invocation. No indexing
step is necessary. Keep employer-specific standards outside public checkouts.
