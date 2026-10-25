# Architecture overview

commerce-api is a modular monolith written in Go.

## Modules

- `catalog` owns products and prices.
- `orders` owns carts, checkout, and order state.
- `payments` talks to the payment provider and owns captures and refunds.

Modules call each other through Go interfaces in `internal/<module>/api`; they never
read another module's tables.

## Events

State changes publish events to the outbox table. A relay delivers them to the
message bus at least once, so consumers must be idempotent.
