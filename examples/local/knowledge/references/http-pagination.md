---
id: references/http-pagination
name: Cursor pagination contract
description: Cursor pagination contract for the fictional commerce API
kind: reference
version: "1.0.0"
applies_to:
  projects: [commerce-api]
  paths: ["src/http/**/*.go"]
---
# List contract

GET /v1/orders accepts limit (default 25, range 1..100) and an opaque cursor.
Sort by created_at descending, then order_id descending to break ties. Return
items and next_cursor; next_cursor is null when there are no more rows.

The cursor encodes the last ordering tuple and a filter fingerprint. Validate its
shape and version. Reject a cursor for different filters; always obtain the tenant
from authentication rather than trusting a tenant encoded in a cursor.

Fetch limit+1 rows to detect another page. Do not use offset pagination for this
changing dataset. Test identical timestamps, an empty result, malformed cursors,
and a resource inserted between page requests.
