# Webhook delivery operations

The inbox separates accepting a verified event from applying its business effects.

## Signature mismatch diagnosis

When webhook signature mismatch rates rise, compare the configured key version,
timestamp skew, and whether middleware changed the raw body. Verify exact bytes
before JSON decoding. During key rotation accept old and new keys only for the
configured overlap. Do not include signatures or secrets in diagnostics.

## Duplicate inbox delivery

A repeated provider event ID receives a successful acknowledgement after the
original durable inbox write. Reprocess a failed inbox row with its existing ID;
do not manufacture a new event. A crash after persistence but before acknowledgement
can cause redelivery, which must produce one business effect.
