---
id: security/webhook-verification
name: Authenticated webhook ingestion
description: Authenticated webhook ingestion for the fictional commerce API
kind: rule
version: "1.0.0"
priority: 100
applies_to:
  projects: [commerce-api]
  paths: ["src/webhooks/**/*.go"]
---
# Verify before decoding effects

Bound the request body and verify the provider signature against the exact raw
bytes using the provider's documented algorithm and constant-time comparison.
Reject timestamps outside a five-minute clock-skew window. Accept a previous
signing key only during the explicitly configured rotation overlap.

Persist a verified event under a unique provider/event_id key before acknowledging
it. Duplicate deliveries receive the original successful acknowledgement and do
not repeat business effects. Unknown event types are recorded and acknowledged
without changing orders. Failed verification must never enqueue an event.

# Acceptance cases

Test tampered bytes, a stale timestamp, a valid previous key during rotation,
duplicate delivery, out-of-order delivery, and a persistence failure before ack.
Never log the signature, signing secret, or raw payload containing personal data.
