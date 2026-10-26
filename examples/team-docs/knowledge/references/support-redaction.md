---
id: references/support-redaction
name: Support export redaction
description: Support export redaction for the fictional commerce API
kind: reference
version: "1.0.0"
applies_to:
  projects: [commerce-api]
  paths: ["src/support/**"]
---
Support exports contain order IDs, state, currency, amounts, and event timestamps.
Exclude access tokens, card data, raw provider payloads, street addresses, and email
addresses. Replace customer identifiers with export-scoped pseudonyms when possible.

An export belongs to one tenant and has a short-lived download link. Record the
requester, reason, and expiry. The deletion runbook describes how legal holds affect
erasure; a support request alone does not authorize deletion.
