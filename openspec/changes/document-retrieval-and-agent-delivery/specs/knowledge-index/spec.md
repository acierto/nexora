# Knowledge Index

## Purpose

Store passages and retrieve the most relevant ones with explanations, offline by default and with optional embeddings and pgvector.

## ADDED Requirements

### Requirement: Explicit indexing
`nexora index` SHALL rebuild passages for every docs and confluence source, or one source selected with `--source`, and replace that source's passages atomically. Knowledge sources SHALL NOT be indexed. Without documentation sources the command SHALL report that there is nothing to index.

#### Scenario: Rebuild one source
- **WHEN** `nexora index --source handbook` runs
- **THEN** only passages of `handbook` are replaced and other sources keep theirs

#### Scenario: Failure keeps previous index
- **WHEN** reading a source or calling the embedding provider fails
- **THEN** the command exits 1 and the stored passages for that source are unchanged

### Requirement: Local default store
The default store SHALL be a JSON file at `index.path` written atomically. It SHALL require no database, network, or API key, and SHALL support keyword retrieval with BM25.

#### Scenario: Keyless first use
- **WHEN** a configuration has a docs source and no index or embeddings settings
- **THEN** index and search work offline using keyword retrieval

### Requirement: Optional embeddings
When `index.embeddings.provider` is `ollama` or `openai`, passages and queries SHALL be embedded with the configured model. The index SHALL record the model; searching with a different model MUST fail with an instruction to re-index.

#### Scenario: Model change
- **WHEN** the index was built with model A and the configuration now names model B
- **THEN** search fails asking to run `nexora index`

### Requirement: Optional pgvector store
When `index.store` is `pgvector`, passages SHALL be stored in PostgreSQL with the pgvector extension using the DSN from `dsn_env`, a generated full-text column, and an HNSW cosine index when dimensions are configured.

#### Scenario: Same ranking contract
- **WHEN** the same passages and query are searched in the local and pgvector stores
- **THEN** both return fused hits with the same explanation structure

### Requirement: Hybrid ranking with explanations
Search SHALL combine keyword and semantic candidates with reciprocal rank fusion and SHALL explain every hit with the methods that found it, per-method rank, score, and matched query terms. Results SHALL be deterministic for identical inputs, with ties broken by passage ID.

#### Scenario: Keyword-only explanation
- **WHEN** no embeddings are configured and a passage matches `refund` and `window`
- **THEN** the hit lists method `keyword` with its rank and the matched terms
