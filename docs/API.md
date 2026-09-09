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
| GET | `/auth/account/password` | Check whether current account has a password |
| PUT | `/auth/account/password` | Change existing account password |
| POST | `/auth/account/password/set` | Set password for social-only login users |
| POST | `/auth/forgot-password` | Start a password reset (link or 6-digit code) |
| POST | `/auth/reset-password` | Complete a password reset with token or code |
| POST | `/auth/resend-verification` | Resend email verification |
| POST | `/auth/verify-code` | Verify a six-digit app code |

Mutating browser requests that use cookies require the `X-CSRF-Token` header.
Public app endpoints use `client_id` to select a project. Never put provider
secrets in an application; social provider credentials belong in the dashboard.

### Sign In & Remember Me

`POST /auth/sign-in`

```json
{
  "email": "user@example.com",
  "password": "yourPassword123!",
  "client_id": "app_your_client_id",
  "remember_me": true
}
```

* `email` *(required)*: User's email address.
* `password` *(required)*: User's password.
* `client_id` *(optional)*: Scopes authentication to a specific tenant application.
* `remember_me` *(optional, default: `true`)*:
  * When `true`: Sets persistent session and CSRF cookies (`Max-Age: 30 days`), and 30-day session and refresh tokens.
  * When `false`: Sets transient browser session cookies (omitting `Max-Age` / `Expires`, expiring on browser close), and 24-hour session and refresh tokens.


## Password Reset (Forgot Password)

Password reset supports both tenant-scoped users and platform administrators.

### Request Password Reset

`POST /auth/forgot-password`

```json
{
  "email": "user@example.com",
  "client_id": "app_your_client_id",
  "return_to": "https://myapp.com/dashboard",
  "delivery": "link"
}
```

* `email` *(required)*: User's registered email address.
* `client_id` *(optional)*: Scopes lookup and reset links to your tenant project.
* `return_to` *(optional)*: Redirect URL after resetting password on hosted UI.
* `delivery` *(optional)*: `"link"` (default) sends a 1-hour secure link button; `"code"` sends a 15-minute 6-digit numeric OTP code. If omitted, uses the tenant's default delivery preference.

Response:

```json
{
  "message": "If an account exists for that email, we sent reset instructions.",
  "delivery": "link"
}
```

### Complete Password Reset

`POST /auth/reset-password`

**Option A: Reset via Token Link:**
```json
{
  "token": "raw-token-from-email-link",
  "password": "newSecurePassword123!"
}
```

**Option B: Reset Password with 6-Digit OTP Code**

`POST /auth/reset-password`

```json
{
  "email": "user@example.com",
  "code": "583920",
  "password": "brand-new-secure-password"
}
```

On success, existing sessions and refresh tokens for that user are revoked, and `{"message": "Password updated. Sign in with your new password."}` is returned.

## In-App Password Management (Change / Set Password)

Authenticated users (both tenant application users via Bearer token and platform administrators via session cookies) can manage their credentials from within their app or security settings.

### Check Password Status

`GET /auth/account/password`

Header: `Authorization: Bearer <access_token>`

Response:
```json
{
  "has_password": true
}
```
If `has_password` is `false`, the user signed up via a social provider (Google, GitHub) and does not have a local password yet.

### Set Password (Social Login Users)

`POST /auth/account/password/set`

Header: `Authorization: Bearer <access_token>`

```json
{
  "new_password": "new-strong-password-123"
}
```
* Fails with `400` (`password_already_set`) if the user already has a password hash established.
* Revokes other active sessions and logs a `password_set` audit event.

### Change Password

`PUT /auth/account/password`

Header: `Authorization: Bearer <access_token>`

```json
{
  "current_password": "current-password-123",
  "new_password": "new-strong-password-456"
}
```
* Validates `current_password` and checks `new_password` against policy (minimum 8 characters).
* Revokes other active sessions and logs a `password_changed` audit event.

## Multi-Factor Authentication (MFA)

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

## Tenant Email & Transactional Provider Settings (Admin API)

Tenant administrators can configure custom transactional email providers for outbound verification emails and password reset links/codes, overriding the platform default SMTP.

| Method | Endpoint | Purpose |
| --- | --- | --- |
| GET | `/dashboard/api/email` | Get tenant email configuration (sensitive credentials masked) |
| PUT | `/dashboard/api/email` | Update email provider configuration and credentials |
| POST | `/dashboard/api/email/test` | Send a diagnostic test email to verify delivery |

Requires tenant admin authentication (session cookie + `X-CSRF-Token` header).

### Get Email Settings

`GET /dashboard/api/email`

Response:
```json
{
  "enabled": true,
  "provider": "smtp",
  "from_name": "Acme Auth",
  "from_email": "auth@acme.com",
  "reply_to": "support@acme.com",
  "smtp_host": "smtp.mailgun.org",
  "smtp_port": 587,
  "smtp_user": "postmaster@acme.com",
  "smtp_password_set": true,
  "smtp_tls_mode": "starttls",
  "resend_api_key_set": false,
  "postmark_token_set": false,
  "ses_region": "",
  "ses_access_key_id": "",
  "ses_secret_access_key_set": false
}
```

### Update Email Settings

`PUT /dashboard/api/email`

* Supported providers: `"smtp"`, `"resend"`, `"postmark"`, `"ses"`.
* Passwords and secret keys are encrypted at rest with AES-256-GCM.
* If a password/token field is sent as empty string or omitted when already set, the existing stored encrypted secret is retained.

Example (Custom SMTP):
```json
{
  "enabled": true,
  "provider": "smtp",
  "from_name": "Acme Auth",
  "from_email": "auth@acme.com",
  "reply_to": "support@acme.com",
  "smtp_host": "smtp.mailgun.org",
  "smtp_port": 587,
  "smtp_user": "postmaster@acme.com",
  "smtp_password": "super-secret-smtp-password",
  "smtp_tls_mode": "starttls"
}
```

Example (Resend):
```json
{
  "enabled": true,
  "provider": "resend",
  "from_name": "Acme Auth",
  "from_email": "auth@acme.com",
  "resend_api_key": "re_123456789"
}
```

Example (AWS SES):
```json
{
  "enabled": true,
  "provider": "ses",
  "from_name": "Acme Auth",
  "from_email": "auth@acme.com",
  "ses_region": "us-east-1",
  "ses_access_key_id": "AKIAIOSFODNN7EXAMPLE",
  "ses_secret_access_key": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
}
```

### Send Test Email

`POST /dashboard/api/email/test`

Tests deliverability using the provided configuration before saving, or tests the currently saved settings.

Request:
```json
{
  "recipient": "admin@example.com",
  "settings": {
    "enabled": true,
    "provider": "resend",
    "from_name": "Acme Auth",
    "from_email": "auth@acme.com",
    "resend_api_key": "re_123456789"
  }
}
```

Response:
```json
{
  "success": true,
  "message": "Test email sent successfully to admin@example.com"
}
```

## Errors and limits

Errors use a stable machine-readable `error` field and may include a human
readable `message`. Common statuses are `400` invalid input, `401` invalid or
missing credentials, `403` verification/MFA/CSRF required, `404` not found,
`409` conflict, and `429` rate limited.

Access tokens are short-lived and refresh tokens rotate on use. API consumers
should cache discovery and JWKS responses according to their HTTP cache headers.

