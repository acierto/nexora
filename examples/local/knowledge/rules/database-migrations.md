---
id: database/expand-contract
name: Compatible database migrations
description: Compatible database migrations for the fictional commerce API
kind: rule
version: "1.0.0"
priority: 80
applies_to:
  projects: [commerce-api]
  paths: ["migrations/**/*.sql"]
---
# Expand before contracting

Ship additive schema changes before code depends on them. Add nullable columns,
backfill in bounded restartable batches, then enforce constraints in a later
release. A renamed column requires a period in which both application versions
can operate; do not drop the old column during the first deploy.

Give every migration a unique increasing version and a clear purpose. Review lock
behavior and table size before adding an index or validating a constraint. Do not
put network calls, live credentials, or destructive sample-data cleanup in a migration.

# Review evidence

Test migration from the previous schema with existing rows and verify the previous
application can still read and write. Document recovery steps for a failed backfill.
Application rollback does not imply automatic schema rollback.
