# Proposal

## Why

Teams with an established documentation MCP should be able to use Nexora's deterministic project knowledge without operating a second document search stack. Native sync and MCP currently repeat knowledge, and retrieval errors hide otherwise valid rule selections.

## What Changes

- Add configurable agent-facing guidance for an existing documentation service, with no connector, credentials, or network calls.
- Add opt-in native context delivery that verifies generated files before returning selected rule and skill citations instead of repeated bodies.
- Return explicit partial context when documentation retrieval fails, preserving selected knowledge and reporting failure through CLI status and MCP isError.
- Collapse exact linked passage copies across sources while retaining duplicate attribution; prefer live table content over an older keyword snapshot of the same ID.
- Update integration guidance, public positioning, runnable examples, and release notes.
- Keep local indexing optional and unchanged as the offline path. Defer delegated retrieval, MCP proxying, hosted service integration, and remote authorization.

## Capabilities

### New Capabilities

- `complementary-delivery`: external lookup guidance, verified native citations, and partial context delivery.
- `passage-deduplication`: conservative linked-copy deduplication and live-content preference.

### Modified Capabilities

None in the main spec inventory. This extends the implemented MCP, native-sync, and task-retrieval behavior recorded in earlier unarchived changes.

## Impact

Configuration gains delivery.documentation; resolve gains --native; get_context gains native. Context JSON gains documentation status and optional native_knowledge citations. Retrieval hits may include duplicates. No new dependencies or external service requirements. Default full knowledge delivery remains available; retrieval failure now returns usable partial JSON with a non-success status.
