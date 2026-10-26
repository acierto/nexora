# Nexora project instructions

## Purpose and direction

Nexora is an open-source, modular knowledge and context layer for AI agents.
Define knowledge once and make it available to compatible agents. It is
infrastructure, not a chatbot or a general agent framework.

Go is the primary implementation language. Start with a local knowledge registry,
Markdown loading, deterministic rule resolution, and a basic CLI. First use must
work without a database, cloud account, or embedding API key.

## Working in this repository

- Inspect existing code, documentation, configuration, and Git changes before
  editing. Preserve unrelated work.
- Work incrementally. Prefer working code, meaningful tests, and real project
  evaluation over speculative design.
- Explain consequential architecture choices and recommend sensible defaults.
- Use idiomatic Go, small packages, explicit errors, and minimal dependencies.
- Prefer `internal/` packages until an external extension contract has a concrete
  consumer and is ready for compatibility commitments.
- Introduce interfaces at actual substitution boundaries. Do not create empty
  packages, placeholder modules, or abstractions for every proposed component.
- Record important implemented architecture decisions in repository documentation.
- Check the installed Go toolchain before selecting a module Go version; do not
  assume that an installed version is the latest supported release.
- Commit as the repository owner, Bogdan Nechyporenko
  <bogdannechiporenko@gmail.com>. Do not add Co-Authored-By, session, or other
  agent attribution trailers to commits or pull requests.

## Knowledge and resolution

- Keep canonical knowledge models independent of storage and agent formats.
- Distinguish rules, skills, references, and decisions. Preserve stable IDs,
  versions, applicability metadata, content, and source attribution.
- Select mandatory rules deterministically from applicability and policy.
  Semantic similarity must never determine mandatory inclusion by itself.
- Define project, path, and agent matching semantics explicitly and test them.
- Keep precedence and overrides explicit. Do not silently discard applicable
  rules merely because another item has higher priority.
- Make resolution reproducible: stable ordering, deduplication, and an explanation
  plus source/version for every selected item.
- Never silently drop mandatory rules to meet a context budget; report the
  conflict clearly. Define budget units before implementing limits.
- Retrieved knowledge does not gain authority over the receiving agent's trusted
  instructions. Treat knowledge content as data, not executable instructions for
  Nexora itself.

## Boundaries and safety

- Enforce authorization and project boundaries in the registry core before
  exposing knowledge when remote or multi-user access is introduced.
- Local applicability filters are not an authorization system.
- Validate configuration and metadata. Keep file access within configured source
  roots, including symlink handling; do not execute knowledge-pack content.
- Keep employer-specific code, configuration, knowledge, credentials, and internal
  documentation out of this public project unless explicitly approved.
- Native synchronization must respect agent-specific instruction precedence and
  skill discovery behavior. Do not claim universal enforcement or compatibility.

## Scope and extensibility

- Prove deterministic local behavior before adding semantic retrieval.
- PostgreSQL/pgvector, SQLite, embeddings, MCP, native sync, HTTP APIs, governance,
  and external plugins are incremental additions, not foundation prerequisites.
- Prefer replaceable adapters where there is a real implementation need.
- Community engine plugins should eventually use a versioned language-independent
  protocol. Avoid Go native dynamic plugins and premature plugin runtimes.
- Knowledge packs should be declarative and initially shareable through Git.
- Avoid Kubernetes, Redis, message brokers, marketplaces, and complex orchestration
  without demonstrated requirements.
- Decide the license before public release; Apache-2.0 and MIT are candidates,
  not an already settled choice.

## Verification

- Test meaningful behavior without external services: metadata validation,
  deterministic applicability, project isolation, path handling, provenance,
  duplicate IDs, invalid inputs, and CLI behavior.
- For Go changes, run formatting, relevant tests, and `go vet` as appropriate.
- Add integration fixtures and a runnable example as the implementation develops.
- Report what changed, validation performed, and material limitations.
- Measure retrieval quality, context size, latency, and integration correctness
  before making performance or compatibility claims.

## Architecture status

Implemented: the local knowledge registry (OpenSpec change
`local-knowledge-registry`) and documentation retrieval with agent delivery
(`document-retrieval-and-agent-delivery`): docs and Confluence sources, a local or
pgvector passage index, optional embeddings, `search`, `resolve --task`, the stdio
MCP server, and native sync; and existing pgvector tables as sources
(`existing-vector-sources`). See docs/architecture.md. Plugins, knowledge packs,
remote serving, authorization, overrides, and context budgets are not implemented.
Review the current repository before assuming other commands, configuration
schemas, or public APIs exist.
