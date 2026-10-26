---
id: skills/replay-webhook
name: Reproduce webhook delivery locally
description: Reproduce webhook delivery locally for the fictional commerce API
kind: skill
version: "1.0.0"
applies_to:
  projects: [commerce-api]
  paths: ["src/webhooks/**/*.go"]
---
# Procedure

1. Create a synthetic event with a fresh event_id and fictional order reference.
2. Sign the exact bytes with a test-only key and an injected fixed clock.
3. Send the event twice to the in-process handler. Expect two acknowledgements,
   one inbox row, and one business transition.
4. Alter one byte without re-signing; expect rejection and no new inbox row.
5. Deliver an older event after a newer one; assert the order does not regress.
6. Simulate a crash after persistence and resume processing the same inbox row.

Keep fixtures free of customer payloads and production signing material. This
skill describes a local test, not permission to replay traffic to a provider.
