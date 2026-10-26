# Keep your existing documentation MCP

This example reuses the 25-item commerce knowledge collection and delegates
lookup to an existing, fictional `team-docs` MCP already connected to the agent.
The guidance is text for the agent; no remote service is required to run this
example. It does not configure that service or grant access to it.

From the repository root with `nexora` on PATH:

```shell
nexora validate --config examples/existing-docs/nexora.yaml
nexora sync --config examples/existing-docs/nexora.yaml --agent codex
nexora resolve --config examples/existing-docs/nexora.yaml --agent codex --native --file src/orders/checkout.go --task "recover an authorization timeout" --json
```

For Claude, use `--agent claude` in both sync and resolve. Sync writes native
files next to this example's config. Use full context without `--native` if your
agent does not load those files, or if you synced into a different directory.

Expected result:

- `documentation.status` is `external`, with the configured lookup guidance.
- `passages` is empty; Nexora does not query or copy the existing service's index.
- `native_knowledge` cites tenant isolation, order transitions, Go conventions,
  the agent reminder, and the endpoint and checkout skills, including versions,
  source digests, match explanations, and native file paths.
- `optional_rules`, `references`, and `decisions` retain their full bodies.
- Editing a source or generated block makes native mode fail until sync is current.

The receiving agent reads the native rules and skills, asks `team-docs` about the
authorization timeout, and cites the authoritative page. Only that service owns
its authentication, document permissions, ranking, and freshness. The example
cannot verify a real service's permissions or force an agent to call it.

`go test ./tests -run TestComplementaryDelivery -v` runs this workflow for both
agents in temporary directories, including stale-sync and hand-edit rejection.
