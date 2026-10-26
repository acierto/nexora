---
id: security/webhook-verification
name: Webhook delivery requirements
description: Webhook delivery requirements for the fictional commerce API
kind: rule
version: "1.0.0"
priority: 100
applies_to:
  projects: [commerce-api]
  paths: ["src/webhooks/**/*.go"]
---
Verify signatures over bounded raw request bytes before accepting a delivery.
Reject stale timestamps beyond five minutes. Store each verified provider/event_id
once, acknowledge after persistence, and process effects idempotently.

Test duplicate delivery, altered payloads, an unknown event type, and a crash after
the inbox write. A late failure cannot reverse an already settled operation.
Never log signatures, secrets, or full webhook payloads.
