---
id: security/tenant-isolation
name: Tenant ownership checks
description: Tenant ownership checks for the fictional commerce API
kind: rule
version: "1.0.0"
priority: 100
applies_to:
  projects: [commerce-api]
  paths: ["src/orders/**/*.go", "src/inventory/**/*.go", "src/http/**/*.go"]
---
# Resolve the tenant from trusted identity

Derive tenant_id from verified authentication context. A request header, URL,
or JSON field cannot change the caller's tenant. Every read and mutation must
constrain both tenant_id and resource_id, including nested resources and joins.

Return the same 404 response for another tenant's resource and a missing resource.
Apply the ownership check inside the transaction that changes state. Include the
tenant in cache keys and uniqueness constraints for customer-supplied identifiers.

# Acceptance cases

Seed tenants A and B with the same external order reference. A can read and update
only its own order. Submitting B's order ID returns 404 and changes neither row.
A line-item ID from B must also be rejected when attached to A's order.
