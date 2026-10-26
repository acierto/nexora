---
id: go/conventions
name: Shared Go conventions
description: A reusable baseline across local projects
kind: rule
version: "1.0.0"
priority: 10
---
Keep packages small and test business behavior without requiring network services.

Accept context.Context as the first parameter for request-scoped operations and
propagate cancellation to storage and outbound calls. Wrap errors with operation
context using %w so callers can inspect them with errors.Is or errors.As. Avoid
including tokens, customer payloads, or connection strings in error text.

Inject clocks and external clients when behavior depends on time or network
responses. Prefer table tests for invalid inputs and explicit boundary tests for
empty results, deadlines, and concurrent updates. Close response bodies and check
errors from writes and transaction commits.
