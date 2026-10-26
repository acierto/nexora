# Tasks

## 1. Configuration

- [x] 1.1 Add the `pgvector` source type with `dsn_env`, `table`, `columns`, `distance`, and `embeddings`; share embedding validation with `index.embeddings`; verify valid and invalid configurations in config tests.

## 2. Table access

- [x] 2.1 Implement `index.VectorTable` for PostgreSQL: quoted identifiers, row mapping with NULL handling, nearest rows for cosine, L2, and inner product, and a dimension mismatch message; verify against PostgreSQL 16 with pgvector (`TestExternalTable`).

## 3. Retrieval and delivery

- [x] 3.1 Add one semantic list per pgvector source to retrieval, embed the query once per model, filter by project, and report the source in reasons; verify with fake tables and embedders.
- [x] 3.2 Index pgvector sources as a text copy without embedding, and wire sources into search, `resolve --task`, and MCP; verify end to end with the executable (`TestExistingVectorTable`).

## 4. Documentation

- [x] 4.1 Document the source in the retrieval guide, configuration reference, architecture, README, and verification record; refresh the website copies.
