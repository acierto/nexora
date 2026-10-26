# Tasks

## 1. Complementary native delivery

- [x] 1.1 Add validated documentation guidance, verified native citations, CLI and MCP support; verify config tests and TestComplementaryDelivery for both agents, including stale knowledge and edited native files.
- [x] 1.2 Add an existing-docs example and integration documentation; verify the example through executable tests and sync the embedded guides.

## 2. Retrieval reliability

- [x] 2.1 Preserve selected knowledge in explicit partial CLI/MCP results; verify TestPartialContextKeepsMandatoryKnowledge and hard failures for invalid knowledge.
- [x] 2.2 Deduplicate exact linked copies and prefer live table content with provenance; verify TestFuseLinkedCopiesAndLiveContent and document freshness limits.

## 3. Release verification

- [x] 3.1 Run gofmt, all tests, go vet, documentation sync/link checks, OpenSpec validation, and release builds; record results in docs/verification.md.
- [ ] 3.2 Commit and publish v0.3.0 through the release workflow; verify successful workflow, published archives/checksums, and a downloaded binary's version.
