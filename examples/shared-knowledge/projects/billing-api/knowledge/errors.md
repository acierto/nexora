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
