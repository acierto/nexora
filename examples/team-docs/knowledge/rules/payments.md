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
