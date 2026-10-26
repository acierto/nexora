# Schema evolution

Application releases and database schema versions advance independently.

## Expand contract backfill

Add a nullable field, ship compatible readers and writers, then backfill existing
rows in bounded batches ordered by primary key. Save a checkpoint so a restarted
job resumes safely. Validate row counts and constraints before enforcing NOT NULL.
Remove the old field in a later release after all old code has stopped using it.

## Blocking migration recovery

If a migration waits on a table lock, stop the release and inspect active
transactions. Do not kill sessions or drop indexes without an approved recovery
plan. A rolled-back application image still runs against the expanded schema;
destructive schema rollback is not the default recovery step.
