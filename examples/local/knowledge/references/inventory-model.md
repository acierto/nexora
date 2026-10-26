---
id: references/inventory-model
name: Inventory reservation model
description: Inventory reservation model for the fictional commerce API
kind: reference
version: "1.0.0"
applies_to:
  projects: [commerce-api]
  paths: ["src/inventory/**/*.go"]
---
# Records and clock

Stock is keyed by (tenant_id, sku); reservations by (tenant_id, checkout_id, sku).
Reservations hold quantity, expires_at, and state: active, consumed, or released.
Use an injected UTC clock to test expiry at the exact deadline without sleeping.

At expires_at <= now an active reservation is eligible for release. A worker must
claim the state change atomically. Consumed and released records remain as an
audit trail; deleting them would remove the deduplication key.

Example: on_hand=8 and reserved=3 gives available=5. Reserving 2 yields reserved=5.
Releasing that reservation restores reserved=3. Consuming it reduces both on_hand
and reserved by 2, preserving availability for unrelated checkouts.
