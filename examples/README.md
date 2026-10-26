# Practical examples

[Install Nexora](../docs/install.md), clone this repository, and run the commands
from its root.

| Example | Learn |
| --- | --- |
| [Local Go API](local/README.md) | 25 items: HTTP contracts, tenant isolation, checkout, inventory, migrations, webhooks, agent filters |
| [Shared standards](shared-knowledge/README.md) | Shared Go/security standards and distinct commerce/billing error contracts |
| [Team docs](team-docs/README.md) | 7 knowledge items plus 12 documents: search operational scenarios and resolve rules with task passages |

Each collection is fictional public sample content and runs offline, including
the documentation index. It includes
configuration, complete Markdown knowledge, expected selections, and an exercise.
The example files also serve as executable integration fixtures.

To run the example regression checks from the repository root:

```sh
go test ./tests -run 'Test(LocalExample|PracticalLocalScenarios|SharedExamples|TeamDocsExample)$' -v
```

The suite builds the CLI, copies the retrieval corpus to a temporary directory,
and checks exact selections, source attribution, project isolation, and expected
search sections. No database, embedding provider, or external service is required.
The fictional policies are test material; the examples do not include a working
commerce or billing application.
