---
id: skills/diagnose-inventory
name: Diagnose stock reservation drift
description: Diagnose stock reservation drift for the fictional commerce API
kind: skill
version: "1.0.0"
applies_to:
  projects: [commerce-api]
  paths: ["src/inventory/**"]
---
1. Read the inventory reconciliation runbook before changing counters.
2. Compare active reservation totals to reserved stock within one consistent snapshot.
3. Separate expired reservations from valid holds and consumed reservations.
4. Reproduce drift with synthetic SKUs and a fixed clock in a local test.
5. Propose a version-checked repair and verify available stock stays non-negative.

Do not manually reset reserved to zero: this would make held units sellable again.
