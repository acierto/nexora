---
id: skills/add-migration
name: Add a database migration
description: Add a database migration for the fictional commerce API
kind: skill
version: "1.0.0"
applies_to:
  projects: [commerce-api]
  paths: ["migrations/**/*.sql"]
---
# Procedure

1. Identify the last released schema and every application version still running.
2. Write the smallest additive change; state the lock and backfill strategy.
3. Apply it to a disposable database containing rows from the previous schema.
4. Exercise old and new readers and writers against the expanded schema.
5. Run a restartable backfill in batches ordered by primary key; save progress.
6. Verify constraints and counts before scheduling a separate contract migration.

Deliver the SQL, compatibility evidence, and a recovery note. Never run migrations
against a real environment just because this procedure was retrieved.
