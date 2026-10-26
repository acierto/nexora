# Design

## Context

See proposal.md for motivation. Deterministic resolution is already independent of documentation retrieval. Native sync uses digest-tracked generated files; the existing vector source has live semantic reads and a keyword snapshot.

## Goals / Non-Goals

Goals: compose native project knowledge with existing documentation access, remove repeated bodies only after verification, and retain knowledge through retrieval outages.
Non-goals: hosted deployment, delegated retrieval, an MCP proxy, new authorization, upstream crawling, or semantic duplicate detection.

## Decisions

- Add delivery.documentation as bounded agent-facing text in validated configuration. Emit it through sync and context without interpreting it as a connection. This avoids another credential and protocol boundary.
- Native delivery is an explicit request, not a server default. Reuse the read-only sync plan to verify files, manifest, project, agent, and current knowledge. Require config-directory sync. Return citations for selected required rules and skills; retain optional rules, references, and decisions in full. A client acknowledging versions would avoid disk reads but would not establish the presence of generated files.
- Report retrieval unavailable as a structured partial result. The CLI emits context and exits 1; MCP retains structuredContent and sets isError. Invalid configuration or knowledge still fails completely. Use a generic partial error message to avoid propagating provider connection details; standalone search remains diagnostic.
- Deduplicate only exact linked copies with matching URL, heading, trimmed body, and digest after fusion. Preserve alternative citations without adding scores. Do not merge approximate text or different versions. Prefer live semantic/table content when the same passage ID also has a stale keyword hit.

## Risks / Trade-offs

- Native files can exist without being loaded by an agent: require explicit native mode and document discovery limitations.
- Full read-only sync planning costs filesystem reads: acceptable for the current small local registry; no performance claims.
- Conservative dedup misses aliases and different chunking: avoid data loss and assign one primary owner per corpus.
- Live semantic reads and keyword snapshots can diverge: explain freshness boundaries and refresh explicitly.

## Migration Plan

Existing configs and full-body calls remain valid. After upgrading, resync generated files to receive new guidance before requesting native mode. Context consumers must recognize documentation status and partial results; release as v0.3.0. Roll back the binary and remove new config fields when reverting to v0.2.0; preserve original knowledge and native manifests.
