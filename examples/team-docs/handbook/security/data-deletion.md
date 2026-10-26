# Customer data deletion

Deletion requests are scoped to a verified tenant and customer identity.

## Erasure and legal holds

Validate the requester, record the case ID, and check for a legal hold before
erasure. Remove customer profile fields and revoke active sessions. Financial
records needed for retention keep transaction amounts and dates while personal
identifiers are minimized. A legal hold blocks automated erasure for held records
and sends the case for review; it does not grant access to another tenant's data.

## Backup restoration checks

Retain a deletion ledger separate from restored application snapshots. After a
backup restore, reapply completed erasures before opening customer access.
Test with synthetic customers, including a held record and a previously erased
profile. Never restore real customer data to a developer laptop for this exercise.
