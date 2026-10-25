# Token rotation

Access tokens expire after 15 minutes and refresh tokens after 30 days.

## Rotating refresh tokens

Every refresh issues a new refresh token and revokes the previous one. If a revoked
refresh token is used again, revoke the whole token family and force a new login:
this usually means the token leaked.

## Service credentials

Service-to-service credentials rotate every 90 days through the secrets manager.
Never commit credentials or print them in logs or error messages.
