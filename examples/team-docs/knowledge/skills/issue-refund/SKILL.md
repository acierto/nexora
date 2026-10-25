---
id: skills/issue-refund
name: Issue a refund
description: Steps for adding or changing a refund flow
kind: skill
applies_to:
  projects: [commerce-api]
  paths: ["src/payments/**"]
---
1. Look up the refund policy for the order's market in the payments handbook.
2. Create the refund with an idempotency key and record it before calling the provider.
3. Handle the provider webhook to mark the refund settled or failed.
4. Test partial refunds, duplicate webhooks, and provider timeouts.
