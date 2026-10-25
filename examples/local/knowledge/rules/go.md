---
id: go/conventions
name: Go conventions
description: Shared baseline for Go development
kind: rule
version: "1.0.0"
priority: 10
tags: [go]
---
Keep packages focused. Return errors with enough context to diagnose failures.
Test behavior at package boundaries without external services when possible.
