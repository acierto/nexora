# Design

## Context

Retrieval already fuses keyword and semantic candidate lists with reciprocal rank fusion. The passage store holds Nexora's own passages and, optionally, their embeddings. An existing team table has its own chunking, keys, and vectors, produced by a model Nexora did not choose.

## Decisions

### A source, not a store

The table is configured as a source, beside `docs` and `confluence`, because it is one collection among several and must combine with them. Making it the index store would force every other source into the team's schema and require write access.

### Copy text, query vectors in place

`nexora index` copies `id`, `text`, `title`, `location`, and `document` into the passage store, so BM25 keyword search, `get_knowledge`, and explanations work exactly as for other sources. Vectors are never copied or recomputed: queries go to the table with its own index. Vector results are therefore always current; the keyword copy is as fresh as the last `nexora index`.

Alternatives considered: keyword search through PostgreSQL full-text on the team table (needs a text index the team may not have and ranks differently from other sources), and copying vectors (duplicates storage and goes stale).

### Rows are passages

Rows are not re-chunked. Passage IDs are `<source>:<row id>`, so a semantic hit from the table and a keyword hit from the copy are the same passage and fuse into one result.

### The query uses the declared model

`embeddings` is required on the source and names the model that produced the vectors. Nexora cannot detect a different model with the same dimensions, so the documentation asks for the exact model; a dimension mismatch is reported with that explanation. Sources with the same model share one query embedding per request.

### Safety

Table and column names are validated against an identifier pattern in configuration and quoted with `pq.QuoteIdentifier`; values travel as query parameters. Nexora only issues `SELECT` statements. The connection string comes from an environment variable. `projects` remains an applicability filter, not authorization.

## Risks

- Large tables make `nexora index` read every row; acceptable for documentation-sized collections. Incremental refresh is deferred.
- Distance functions must match how the vectors were built; the default is cosine, the most common choice.
