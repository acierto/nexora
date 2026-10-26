---
id: orders/state-transitions
name: Order state transitions
description: Order state transitions for the fictional commerce API
kind: rule
version: "1.0.0"
priority: 70
applies_to:
  projects: [commerce-api]
  paths: ["src/orders/**/*.go"]
---
# Allowed transitions

Orders move pending -> confirmed -> fulfilled. Pending or confirmed orders may
be cancelled; fulfilled orders use the return workflow. Terminal states cannot
be reopened through the update endpoint.

Confirm only after inventory is reserved and payment authorization is recorded.
Use an expected version in each update so concurrent confirmation and cancellation
cannot both succeed. A stale version returns a conflict with no side effects.
Record the state transition and outbox event in the same transaction.

# Acceptance cases

Test each allowed edge, fulfilled -> cancelled rejection, a repeated confirmation,
and two writers using the same version. A duplicate request returns the original
outcome rather than emitting a second OrderConfirmed event.
