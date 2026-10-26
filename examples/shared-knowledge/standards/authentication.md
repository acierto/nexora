---
id: security/authentication
name: Commerce authentication
description: A shared-source rule restricted to the commerce project
kind: rule
priority: 100
applies_to:
  projects: [commerce-api]
  paths: ["src/auth/**/*.go"]
---
Protect credentials and test invalid and expired authentication tokens.

Check issuer, audience, signature, and expiry before establishing identity. The
commerce tenant comes from verified claims and must scope every resource lookup.
Return a uniform 401 for invalid credentials, and keep a foreign tenant's resource
indistinguishable from a missing resource. Never accept a request field as proof
of tenant membership. Test two tenants with colliding external resource IDs.
