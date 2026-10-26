# Proposal

## Why

Many teams already embed their documentation into PostgreSQL with pgvector for a docs search or chatbot. Nexora could only search vectors it computed itself, so using such a collection meant either embedding everything a second time or bypassing Nexora with a side script. A side script is invisible to `get_context`, so agents connected over MCP kept getting keyword-only passages. Nexora should use those vectors natively.

## What Changes

- Add a `pgvector` source type that describes an existing table: connection string variable, table, column mapping (`id`, `text`, `vector`, optional `title`, `location`, `document`), distance function, and the embedding model that produced the vectors.
- `nexora index` copies only the text columns of such a source into the passage store for keyword search. It never embeds those rows and never writes to the table.
- At query time, Nexora embeds only the task text with the source's model, asks the table for its nearest rows, and fuses them with keyword hits and other sources through the existing reciprocal rank fusion. Hits carry `method: semantic` and the `source` that found them.
- The same path serves `nexora search`, `nexora resolve --task`, and the MCP `get_context` and `search_knowledge` tools.
- Defer: row filters (SQL predicates), other databases (MySQL, SQLite, vector services), and incremental keyword refresh.

## Capabilities

### New Capabilities

- `vector-sources`: configuration, read-only access, keyword copy, in-place semantic search, and model consistency for existing pgvector tables.

### Modified Capabilities

None. `knowledge-index` and `task-retrieval` gain an extra semantic candidate list; their contracts (fusion, explanations, project filtering, deterministic rule selection) are unchanged.

## Impact

- New `index.VectorTable` interface with one PostgreSQL implementation, reusing `github.com/lib/pq`. No new dependencies.
- New optional `source` field in hit reasons; existing JSON fields are unchanged.
- Network access at query time now also includes the source's embedding provider and database, only when a `pgvector` source is configured.
