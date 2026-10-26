---
id: decisions/money-minor-units
name: Represent money in integer minor units
description: Represent money in integer minor units for the fictional commerce API
kind: decision
version: "1.0.0"
applies_to:
  projects: [commerce-api]
  paths: ["src/orders/**/*.go", "src/payments/**/*.go"]
---
# Decision

Represent monetary amounts as integer minor units with an explicit currency.
For this sample, EUR and USD use two fractional digits and JPY uses zero. Reject
unsupported currencies rather than guessing precision. Never add different currencies.

# Why

Binary floating point introduces rounding differences between checkout, provider
requests, and refunds. Integer arithmetic makes the stored and transmitted amount
comparable. Check for overflow before multiplying quantity by unit price.

# Consequences

EUR 12.34 is stored as amount_minor=1234 and currency=EUR. JPY 500 is stored as
amount_minor=500. Define discount rounding explicitly: allocate any remainder
by stable line-item ID so totals are reproducible across retries.
