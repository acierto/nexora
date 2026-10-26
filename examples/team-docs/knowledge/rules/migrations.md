---
id: database/expand-contract
name: Expand and contract schema changes
description: Expand and contract schema changes for the fictional commerce API
kind: rule
version: "1.0.0"
priority: 80
applies_to:
  projects: [commerce-api]
  paths: ["migrations/**/*.sql"]
---
Deploy additive schema changes first, backfill in bounded restartable batches,
and remove obsolete fields only after old readers and writers are retired.
Review table locks and record recovery steps. Application rollback does not undo
schema changes. Test compatibility against a database with existing sample rows.
