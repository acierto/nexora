# Customer support incident triage

Start with a tenant-scoped order ID and a request ID rather than personal data.

## Safe diagnostic export

A support diagnostic export includes state changes, request IDs, operation IDs,
and timestamps. Redact email addresses, street addresses, access credentials,
and raw provider payloads. Give the download a short expiry and audit who created
it and why. Validate that every row belongs to the requesting tenant.

## Charged without confirmation

When a customer reports a charge but no confirmed order, inspect the recorded
payment operation and reservation state. An authorization hold is different from
a settled capture. Reconcile the existing operation before considering a refund;
never issue a second charge to see whether the first one succeeded.
