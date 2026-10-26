---
id: security/authentication
name: Authentication requirements
description: Protect credentials in commerce authentication changes
kind: rule
version: "1.0.0"
priority: 100
applies_to:
  projects: [commerce-api]
  paths: ["src/auth/**/*.go"]
---
Never log tokens or credentials. Validate expiration and test rejected credentials.
Keep authentication errors consistent without exposing account details.

Verify the issuer, audience, signature, and expiration before accepting a bearer
token. Use an injected clock for expiry tests. Resolve the customer and tenant
from verified identity, not a caller-supplied customer ID. A missing, malformed,
expired, or revoked token returns the same public 401 response.

Refresh tokens rotate on every use. Reuse of a revoked refresh token invalidates
the whole token family. Store token hashes rather than raw refresh tokens and
test concurrent refresh attempts so only one replacement becomes valid.
