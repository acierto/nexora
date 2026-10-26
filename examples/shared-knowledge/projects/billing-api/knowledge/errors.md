---
id: billing/errors
name: Billing API errors
description: Error handling requirements for billing code
kind: rule
applies_to:
  projects: [billing-api]
  paths: ["src/**/*.go"]
---
Keep billing error responses consistent and avoid exposing sensitive invoice data.

Use invoice_not_found (404), invoice_locked (409), invalid_amount (400), and
internal_error (500) as stable public codes. Include a request_id, never a bank
account, tax identifier, SQL fragment, or raw accounting-provider response.

Finalized invoices cannot be edited; corrections create a credit note linked to
the original invoice. A retry with the same correction key returns the existing
credit note. Test a locked invoice, a missing invoice, and an attempt to read an
invoice owned by another billing tenant. The latter returns invoice_not_found.
