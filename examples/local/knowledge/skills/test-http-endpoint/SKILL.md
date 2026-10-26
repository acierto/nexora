---
id: skills/test-http-endpoint
name: Test an HTTP endpoint
description: Test an HTTP endpoint for the fictional commerce API
kind: skill
version: "1.0.0"
applies_to:
  projects: [commerce-api]
  paths: ["src/http/**/*.go"]
---
# Procedure

1. Build a handler with a fake domain service and an in-memory request recorder.
2. Use httptest with a valid identity and assert status, headers, and JSON shape.
3. Table-test malformed JSON, unknown fields, oversized bodies, missing identity,
   wrong tenant, and domain errors. Count service calls for rejected requests.
4. Cancel a request context and verify cancellation reaches the fake service.
5. Check error responses and captured logs contain no credentials or SQL details.

For list endpoints, test an empty page, a full page, and a cursor scoped to a
different tenant. Assert stable ordering when timestamps are equal.
