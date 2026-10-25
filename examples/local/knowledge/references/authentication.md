---
id: references/authentication
name: Authentication architecture
description: Boundaries between authentication transport and services
kind: reference
applies_to:
  projects: [commerce-api]
  paths: ["src/auth/**/*.go"]
---
Transport extracts credentials, the authentication service validates them, and
handlers translate typed errors into API responses. Keep token contents out of logs.
