# Security

Non-negotiable rules for sooauth. When unsure, choose the conservative standard option.

## Standards

- OAuth 2.0 (RFC 6749) + PKCE for public clients
- OpenID Connect discovery + JWKS
- WebAuthn for passkeys (maintained library, never hand-roll crypto)

## Passwords

- **argon2id** with sane parameters (bcrypt only as documented fallback)
- Never log or return password material
- Uniform responses on reset (no email enumeration)

## Tokens

- Access tokens: ~15 min, signed, rotatable keys via JWKS
- Refresh tokens: rotate every use, stored hashed only
- Refresh reuse → revoke entire token family

## Sessions & cookies

- HttpOnly, Secure, SameSite=Lax/Strict
- CSRF on cookie-authenticated state-changing routes
- Sessions revocable from admin panel

## Verification / reset tokens

- Single-use, short TTL, high entropy, stored hashed

## Rate limiting

- Redis-backed on sign-in, reset, verify, token endpoints
- Per-IP and per-account counters

## Multi-factor authentication

- TOTP secrets are encrypted at rest with `MFA_ENCRYPTION_KEY`.
- Production requires a base64-encoded 32-byte MFA encryption key.
- MFA login challenges are short-lived, hashed, and single-use.
- Backup codes are stored hashed and consumed one time only.
- Passkeys use WebAuthn through the maintained WebAuthn library.

## Production defaults

- Refuse to start with insecure secrets when `SOOAUTH_ENV=production`
- Assume HTTPS termination at the edge
- Strict CORS, security headers on all responses

## Audit

- Append-only audit log: sign-in, failed sign-in, reset, provider change, key rotation, session revoke

## Webhooks

- Webhook payloads are signed with an endpoint-specific HMAC secret.
- Signatures include a timestamp to support replay protection.
- Deliveries are persisted and retried up to three times.
- Endpoint URLs are validated and scoped to the owning project.

## Fail safe

- Ambiguous authz → deny
- Email/Redis outage → never silently grant access

## Known limitations

- Organization-level RBAC, SAML, SCIM, and M2M policy management are roadmap items.
- Bearer access tokens remain valid until their expiry; revoking a browser session immediately revokes the session and refresh token records.
- Operators remain responsible for backups, SMTP security, database access, and infrastructure hardening when self-hosting.
