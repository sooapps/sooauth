# API Reference

Sooauth exposes standard OpenID Connect endpoints plus a small set of account
and widget APIs. The issuer is the public `APP_URL` of the deployment.

## OpenID Connect

| Method | Endpoint | Purpose |
| --- | --- | --- |
| GET | `/.well-known/openid-configuration` | Discovery metadata |
| GET | `/.well-known/jwks.json` | Public signing keys |
| GET | `/oauth/authorize` | Authorization code flow with PKCE |
| POST | `/oauth/token` | Exchange authorization code or refresh token |
| GET | `/oauth/userinfo` | Read the authenticated user |

Use an OIDC library whenever possible. Do not implement token validation or
PKCE by hand.

## Hosted account API

| Method | Endpoint | Purpose |
| --- | --- | --- |
| POST | `/auth/sign-up` | Create a platform or app user |
| POST | `/auth/sign-in` | Sign in with email and password |
| POST | `/auth/sign-out` | Revoke the current cookie session |
| GET | `/auth/me` | Read the current account |
| POST | `/auth/forgot-password` | Start a password reset |
| POST | `/auth/reset-password` | Complete a password reset |
| POST | `/auth/resend-verification` | Resend email verification |
| POST | `/auth/verify-code` | Verify a six-digit app code |

Mutating browser requests that use cookies require the `X-CSRF-Token` header.
Public app endpoints use `client_id` to select a project. Never put provider
secrets in an application; social provider credentials belong in the dashboard.

## Passkeys and MFA

| Method | Endpoint | Purpose |
| --- | --- | --- |
| POST | `/auth/passkey/register/begin` | Start passkey registration |
| POST | `/auth/passkey/register/finish` | Finish passkey registration |
| POST | `/auth/passkey/sign-in/begin` | Start passkey sign-in |
| POST | `/auth/passkey/sign-in/finish` | Finish passkey sign-in |
| POST | `/auth/mfa/start` | Create TOTP setup data |
| POST | `/auth/mfa/confirm` | Enable TOTP after code verification |
| POST | `/auth/mfa/verify` | Complete an MFA login challenge |
| POST | `/auth/mfa/disable` | Disable MFA with a valid code |

When MFA is enabled, password sign-in returns `403` with:

```json
{
  "error": "mfa_required",
  "challenge": "one-time-challenge-token"
}
```

The challenge is short-lived and can only be used once. The client submits the
challenge and an authenticator code or backup code to `/auth/mfa/verify`.

## Widget API

| Method | Endpoint | Purpose |
| --- | --- | --- |
| GET | `/v1/widget/config?client_id=...` | Read project auth configuration |
| POST | `/v1/widget/exchange` | Exchange a social callback code |
| GET | `/v1/widget/embed.js` | Download the drop-in widget |

## Errors and limits

Errors use a stable machine-readable `error` field and may include a human
readable `message`. Common statuses are `400` invalid input, `401` invalid or
missing credentials, `403` verification/MFA/CSRF required, `404` not found,
`409` conflict, and `429` rate limited.

Access tokens are short-lived and refresh tokens rotate on use. API consumers
should cache discovery and JWKS responses according to their HTTP cache headers.
