# vector-sources Specification

## Purpose

Use documentation that a team already embedded in a PostgreSQL pgvector table, without re-embedding it, through the same search and context paths as every other source.

## Requirements

### Requirement: pgvector source configuration
A source with `type: pgvector` SHALL require `dsn_env`, `table` (a name or `schema.name`), `columns.id`, `columns.text`, `columns.vector`, and `embeddings` with a provider and model. It MAY set `columns.title`, `columns.location`, `columns.document`, `distance` (`cosine` by default, `l2`, `inner_product`), and `projects`. Identifiers that are not plain SQL identifiers, and fields that belong to other source types, MUST be rejected.

#### Scenario: Valid source
- **WHEN** a configuration declares a pgvector source with the required fields
- **THEN** it loads, distance defaults to cosine, and the source is listed among documentation sources

#### Scenario: Unsafe identifier
- **WHEN** `table` or a column contains characters outside letters, digits, and underscores
- **THEN** loading fails naming the field

### Requirement: Read-only keyword copy
`nexora index` SHALL read the mapped text columns of every row with non-empty text into the passage store, one passage per row with ID `<source>:<row id>`, without calling an embedding provider for those rows and without writing to the table.

#### Scenario: No re-embedding
- **WHEN** a pgvector source is indexed with an embedding provider configured
- **THEN** the provider receives no request during indexing and the report says the vectors are searched in place

### Requirement: In-place semantic search
For search, `resolve --task`, and MCP context, Nexora SHALL embed only the query with the source's model, retrieve the nearest rows from the table with the configured distance, apply the source's project filter, and fuse them with other candidates by reciprocal rank fusion. Each such reason SHALL report method `semantic` and the source ID.

#### Scenario: Meaning without shared words
- **WHEN** the query shares no keywords with a row whose vector is closest to the query embedding
- **THEN** that row is returned with a semantic reason naming its source

#### Scenario: Shared query embedding
- **WHEN** two pgvector sources use the same model
- **THEN** the query is embedded once per request

### Requirement: Model consistency
A query embedding whose dimensions differ from the stored vectors MUST fail with an error stating that the source must use the model that produced the table.

#### Scenario: Wrong model size
- **WHEN** the configured model returns vectors of a different size than the table
- **THEN** the command exits 1 with that explanation
