# Checkout recovery

Checkout coordinates order creation, stock reservation, and payment authorization.

## Ambiguous authorization

An authorization timeout leaves the outcome unknown. Mark the operation pending
reconciliation and look up the original idempotency key with the provider. Do not
create a fresh authorization key: the first request may already have succeeded.
A confirmed authorization lets checkout resume; a confirmed decline releases stock.

## Conflicting checkout retries

The same checkout key and request fingerprint returns the existing order. The same
key with a different cart returns 409. Store the fingerprint and order reference
before making an external request. Test a timeout followed by a retry and a
concurrent pair of requests with the same key.
