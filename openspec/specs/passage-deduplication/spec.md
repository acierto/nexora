# passage-deduplication Specification

## Purpose

Avoid repeated documentation passages from overlapping sources while preserving source attribution, distinct sections, and current live table content.

## Requirements

### Requirement: Exact linked duplicate removal
Retrieval SHALL collapse ranked passages with the same absolute HTTP(S) location, heading, trimmed body, and digest before applying the result limit. The highest-ranked representative SHALL retain its score and list the alternative IDs, sources, titles, locations, and digests. Distinct text, digests, and local paths SHALL remain separate.

#### Scenario: Two sources contain an exact linked copy
- **WHEN** both sources return identical linked passages
- **THEN** one result slot contains the highest-ranked representative and the other citation is preserved in duplicates

#### Scenario: Same URL with a different version
- **WHEN** bodies or digests differ at the same URL
- **THEN** both results remain available

### Requirement: Live table content preference
When a passage ID occurs in both a live table semantic result and a keyword snapshot, retrieval SHALL return the live passage text and metadata while retaining both match explanations.

#### Scenario: Keyword snapshot predates a table edit
- **WHEN** a semantic table hit has newer content than the keyword candidate with the same ID
- **THEN** the fused result contains live text, title, and digest
