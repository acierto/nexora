---
id: references/webhook-events
name: Payment webhook event mapping
description: Payment webhook event mapping for the fictional commerce API
kind: reference
version: "1.0.0"
applies_to:
  projects: [commerce-api]
  paths: ["src/webhooks/**/*.go"]
---
# Inbox contract

Synthetic event fields are event_id, type, occurred_at, payment_id, and operation_id.
The provider/event_id pair deduplicates delivery; operation_id associates the event
with a locally recorded operation. An unknown payment never creates an order.

| Event | Effect | Duplicate or late event |
| --- | --- | --- |
| authorization.succeeded | Record authorization result | Preserve original result |
| capture.succeeded | Mark recorded capture settled | No second capture |
| refund.succeeded | Mark recorded refund settled | No second credit |
| refund.failed | Record failure if still pending | Never regress settled to pending |

Persist raw verified events under the sample retention policy of seven days;
retain event IDs and processing outcomes for 90 days. Raw payloads are not log fields.
