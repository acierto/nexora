---
id: api/http-contracts
name: HTTP request and response contracts
description: HTTP request and response contracts for the fictional commerce API
kind: rule
version: "1.0.0"
priority: 80
applies_to:
  projects: [commerce-api]
  paths: ["src/http/**/*.go"]
---
# Request boundaries

Decode a single JSON object with a 1 MiB body limit. Reject unknown fields and
trailing JSON values. Validate before calling a domain service. Use UTC RFC3339
strings for timestamps and integer minor units plus an ISO currency code for money.

# Response contract

Return 201 with a Location header for a newly created resource, 400 for malformed
input, 401 for missing credentials, 403 for insufficient permission, 404 for an
inaccessible resource, and 409 for a state conflict. Map domain errors in the HTTP
layer; never expose SQL errors or provider response bodies.

Errors use `{"error":{"code":"invalid_request","message":"...","request_id":"..."}}`.
Test malformed JSON, a body over the limit, unknown fields, unauthorized access,
and a service error. Assert that rejected requests do not invoke the service.
