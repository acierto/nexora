# Outbox relay operations

Order transitions and their event records commit in the same transaction.

## Relay backlog triage

An increasing outbox backlog means committed events are waiting for publication.
Compare the oldest undelivered timestamp, relay progress, and publisher errors.
Pause new relay workers if they amplify a downstream failure; preserve pending
rows. Resume with bounded retries after the destination is healthy.

## Publication crash recovery

A crash after publication but before marking an outbox row delivered may publish
the same event again. Keep its event_id unchanged. Consumers deduplicate that ID;
changing it converts a safe retry into a second business operation. Verify recovery
with a fake publisher that fails after accepting a message.
