---
id: decisions/transactional-outbox
name: Persist events with domain changes
description: Persist events with domain changes for the fictional commerce API
kind: decision
version: "1.0.0"
applies_to:
  projects: [commerce-api]
  paths: ["src/orders/**/*.go"]
---
# Decision

Write domain transitions and outbox entries in one database transaction. A relay
publishes committed entries and records delivery progress. Delivery is at least
once, so consumers deduplicate by event_id.

# Why

Publishing before commit can announce an order that rolls back. Publishing only
after commit without durable intent can lose the event if the process crashes.
The outbox closes this gap at the cost of a relay and duplicate delivery handling.

# Consequences

Tests cover rollback (no event), commit (one pending event), and relay restart
(possible repeated delivery with the same event_id). No broker is required for
these fixtures: use a fake publisher and explicit relay ticks.
