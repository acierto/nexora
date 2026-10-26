# Inventory reconciliation

This runbook diagnoses differences between stock counters and reservation records.

## Reservation drift investigation

For a single tenant and SKU, compare reserved stock with the sum of active
reservation quantities in a consistent database snapshot. Exclude consumed and
released rows. Inspect expiry worker progress before concluding the counter is wrong.
Capture record versions and the observation time; never reset reserved to zero.

## Expired stock holds

Unpaid stock holds expire after 15 minutes. The expiry worker atomically changes
active to released and decrements reserved. If confirmation wins the same state
check, expiry must do nothing. Repeating the job must not free units twice.
Test at the exact expires_at boundary with an injected clock.
