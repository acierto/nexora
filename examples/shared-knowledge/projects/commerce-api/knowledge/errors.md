---
id: commerce/errors
name: Commerce API errors
description: Error handling requirements for commerce code
kind: rule
applies_to:
  projects: [commerce-api]
  paths: ["src/**/*.go"]
---
Translate domain errors into consistent commerce API responses.

Use stable codes: invalid_request (400), unauthenticated (401), forbidden (403),
not_found (404), state_conflict (409), and internal_error (500). Include request_id
for support correlation. Keep the public message independent of provider errors
and SQL diagnostics; retain a sanitized cause in internal structured logs.

An order from another tenant returns not_found. A stale order version returns
state_conflict without emitting a domain event. Test these cases through the HTTP
handler with a fake service rather than matching internal error strings.
