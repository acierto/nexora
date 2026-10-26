---
id: inventory/reservations
name: Atomic stock reservations
description: Atomic stock reservations for the fictional commerce API
kind: rule
version: "1.0.0"
priority: 70
applies_to:
  projects: [commerce-api]
  paths: ["src/inventory/**/*.go"]
---
# Quantity invariants

available = on_hand - reserved, and neither available nor reserved may be negative.
Reserve with a conditional atomic update or a locked row; a separate availability
read followed by an unconditional write permits overselling. Quantities are
positive integers. A reservation is unique by tenant, checkout, and SKU.

Unpaid reservations expire after 15 minutes. Confirmation consumes the reservation;
cancellation or expiry releases it exactly once. A release racing with confirmation
must leave a single terminal reservation state and balanced stock counters.

# Acceptance cases

With one unit available, two concurrent checkouts produce one reservation and one
out-of-stock error. Repeat release and expiry jobs without changing stock twice.
A failed multi-SKU reservation rolls back every stock update.
