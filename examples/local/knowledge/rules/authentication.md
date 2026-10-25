---
id: security/authentication
name: Authentication requirements
description: Protect credentials in commerce authentication changes
kind: rule
version: "1.0.0"
priority: 100
applies_to:
  projects: [commerce-api]
  paths: ["src/auth/**/*.go"]
---
Never log tokens or credentials. Validate expiration and test rejected credentials.
Keep authentication errors consistent without exposing account details.
