---
id: skills/add-endpoint
name: Add an endpoint
description: A procedure for adding a commerce API endpoint
kind: skill
applies_to:
  projects: [commerce-api]
  paths: ["src/**/*.go"]
---
1. Define request validation and response types.
2. Implement business logic independently of HTTP transport.
3. Add tests for success, invalid input, and authorization failures.
4. Carry the request context through service and storage calls; stop work when
   cancelled and set a deadline on outbound requests.
5. Map typed domain errors to the documented HTTP contract at the transport edge.
6. For mutations, define retry behavior before adding automatic retries. Use a
   stable operation key and reject the same key with a different request body.
7. Verify responses and structured logs contain no credentials or personal data.

For example, adding POST /v1/orders requires a tenant-scoped identity, a validated
cart, server-owned price calculation, and a reproducible duplicate-request result.
Use a fake inventory service and payment provider to test it without purchases.
