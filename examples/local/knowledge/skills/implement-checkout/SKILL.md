---
id: skills/implement-checkout
name: Implement checkout safely
description: Implement checkout safely for the fictional commerce API
kind: skill
version: "1.0.0"
applies_to:
  projects: [commerce-api]
  paths: ["src/orders/**/*.go", "src/inventory/**/*.go"]
---
# Procedure

1. Resolve tenant and customer identity; validate quantities and currency.
2. Recalculate prices from server-owned catalog data, not client totals.
3. Claim the checkout idempotency key with a hash of the normalized request.
4. Reserve all SKUs atomically. On shortage, leave no partial reservation.
5. Persist the pending order and durable intent before contacting the provider.
6. Reconcile ambiguous authorization outcomes before retrying. Confirm the order
   only when reservation and authorization both exist.
7. Commit the transition and outbox event together; publish after commit.

Test a repeated request, the same key with a different cart (409), last-unit
contention, authorization timeout, and cancellation during confirmation. Use a
fake provider and a controllable clock; no real purchases are needed.
