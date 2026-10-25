---
id: security/authentication
name: Commerce authentication
description: A shared-source rule restricted to the commerce project
kind: rule
priority: 100
applies_to:
  projects: [commerce-api]
  paths: ["src/auth/**/*.go"]
---
Protect credentials and test invalid and expired authentication tokens.
