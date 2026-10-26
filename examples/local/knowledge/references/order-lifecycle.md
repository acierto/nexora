---
id: references/order-lifecycle
name: Order lifecycle and ownership
description: Order lifecycle and ownership for the fictional commerce API
kind: reference
version: "1.0.0"
applies_to:
  projects: [commerce-api]
  paths: ["src/orders/**/*.go"]
---
# Domain record

An order contains tenant_id, order_id, customer_id, version, currency, line items,
and state. Prices are snapshots in integer minor units. Later catalog changes do
not rewrite an accepted order. The orders module owns this record.

| State | Entry condition | Permitted next states |
| --- | --- | --- |
| pending | validated checkout persisted | confirmed, cancelled |
| confirmed | stock reserved and authorization recorded | fulfilled, cancelled |
| fulfilled | shipment accepted | none; use a return record |
| cancelled | cancellation won the version check | none |

Payment authorization and capture are separate. A provider timeout is an unknown
outcome, not proof of failure. Reconciliation looks up the original operation key.
Shipping consumes a confirmed order; it cannot authorize payment itself.
