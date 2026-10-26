---
id: payments/idempotency
name: Idempotent payment operations
description: Payment and refund calls must be safe to retry
kind: rule
version: "1.0.0"
priority: 100
applies_to:
  projects: [commerce-api]
  paths: ["src/payments/**/*.go"]
---
Every payment and refund request carries an idempotency key derived from the order
and operation. Retries reuse the key; never create a second provider request.

Include a request fingerprint so reuse of a key with a different amount or currency
returns a conflict. Persist the operation before contacting the provider. Treat a
timeout as unknown and reconcile the original key before retrying. Store amounts
in integer minor units and forbid cumulative refunds above the captured amount.

Test a duplicate request, a changed amount under the same key, simultaneous partial
refunds, and a provider timeout followed by a successful reconciliation. A webhook
can settle the recorded operation but cannot create a second refund.
