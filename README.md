# Sooauth

**Authentication without the homework.** No database migrations, no custom schema upkeep, and no $0.02/MAU vendor traps. Standard OIDC, passkeys, social login, and hosted auth pages built for vibecoders, AI-assisted developers, and indie hackers.

Use the hosted cloud at [auth.sooauth.com](https://auth.sooauth.com) or self-host your own instance in seconds.

[![License: AGPL v3](https://img.shields.io/badge/License-AGPL_v3-blue.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/sooapps/sooauth)](https://goreportcard.com/report/github.com/sooapps/sooauth)
[![AI Ready](https://img.shields.io/badge/AI_Ready-llms.txt-success.svg)](https://auth.sooauth.com/llms.txt)

---

## Why Sooauth?

Most auth tools force a bad tradeoff: either pay steep monthly taxes for hosted convenience, or spend hours maintaining database migrations, schema adapters, and runtime dependencies.

```
┌────────────────────────────────────────────────────────────────────────┐
│                              THE AUTH LANDSCAPE                        │
│                                                                        │
│  Complex         │ [Keycloak]    - Heavy Java stack, 1GB+ RAM, 100 configs │
│  (Overengineered)│ [Better Auth] - Manages DB schemas, migrations, adapters │
│                  │ [Clerk]       - Great DX, but closed-source & MAU fees │
│  ────────────────┼──────────────────────────────────────────────────── │
│  Lean & Fast     │ ✨ [SOOAUTH]   - Single Go binary, standard OIDC,      │
│  (AI-First)      │   hosted pages ready, 3-line env configuration.     │
└────────────────────────────────────────────────────────────────────────┘
```

| Feature | Sooauth | Clerk | Better Auth | Keycloak |
| :--- | :---: | :---: | :---: | :---: |
| **Open Source & Self-Hostable** | ✅ Yes (AGPL-3.0) | ❌ No | ✅ Yes | ✅ Yes |
| **Your App's DB Touched?** | 🛡️ **Zero DB migrations** | 🛡️ None | ⚠️ Modifies app DB | 🛡️ None |
| **Standard OIDC Protocol** | ✅ Pure OIDC + PKCE | ❌ Proprietary | ❌ Custom | ✅ Standard |
| **Footprint / Memory** | ⚡ Minimal (Go ~25MB) | Cloud only | Node.js runtime | 🐘 Heavy (Java 1GB+) |
| **No MAU Price Surprises** | ✅ Free self-host / flat | ❌ \$0.02+/MAU | ✅ Free self-host | ✅ Free self-host |
| **Vibecoder / AI-Prompt Ready** | ✅ 1-prompt setup | ⚠️ Custom SDKs | ⚠️ DB errors | ❌ Complex setup |

---

## 🤖 Vibecoder & AI Prompt ("Prompt-Ready Auth")

If you are building your app with **Cursor, Claude Code, Windsurf, or v0**, paste this single prompt into your AI agent:

> *"Integrate user authentication using Sooauth OIDC. Use standard OIDC with `SOOAUTH_ISSUER` (default: `https://auth.sooauth.com`), `SOOAUTH_CLIENT_ID`, and `SOOAUTH_REDIRECT_URI`. Direct users to the hosted sign-in page and exchange the authorization code with PKCE. Do NOT create user tables or run database migrations in our app database."*

*(You can also drop our pre-built [`.cursorrules`](.cursorrules) or [`.cursor/rules/sooauth.mdc`](.cursor/rules/sooauth.mdc) directly into your project!)*

---

## Quick Start

### 1. Cloud (Instant)
Sign up for free at [auth.sooauth.com/auth/sign-up](https://auth.sooauth.com/auth/sign-up), create a project, and grab your `Client ID`.

### 2. Self-Host (10 Seconds)

Run our automated installer on any Linux/macOS server with Docker:

```bash
curl -sSL https://raw.githubusercontent.com/sooapps/sooauth/main/install.sh | bash
```

Or run via Docker Compose manually:

```bash
git clone https://github.com/sooapps/sooauth.git
cd sooauth
cp .env.example .env
docker compose up -d
```

Your services are live:
- **Sign-in / Sign-up:** `http://localhost:8080/auth/sign-in`
- **Admin Panel:** `http://localhost:8080/admin/`
- **OIDC Discovery:** `http://localhost:8080/.well-known/openid-configuration`

---

## 30-Second Client Integration

### Next.js / React (JavaScript SDK)

```bash
npm install @sooauth/sdk-js
```

Add your environment variables:
```env
NEXT_PUBLIC_SOOAUTH_ISSUER=https://auth.sooauth.com
NEXT_PUBLIC_SOOAUTH_CLIENT_ID=your_client_id
NEXT_PUBLIC_SOOAUTH_REDIRECT_URI=http://localhost:3000/callback
```

Initiate login:
```typescript
import { createSooauthClient } from "@sooauth/sdk-js";

export const auth = createSooauthClient({
  issuer: process.env.NEXT_PUBLIC_SOOAUTH_ISSUER!,
  clientId: process.env.NEXT_PUBLIC_SOOAUTH_CLIENT_ID!,
  redirectUri: process.env.NEXT_PUBLIC_SOOAUTH_REDIRECT_URI!,
});

// Trigger hosted login with passkeys, Google, GitHub, or password
await auth.signInWithRedirect();
```

Handle callback (`/callback` page):
```typescript
const session = await auth.handleCallback();
console.log("Authenticated user:", session.user);
```

### Backend Route Protection (Any Language)
Because Sooauth emits standard OIDC JWTs, verify Bearer tokens with any standard JWKS library (e.g. `jose` in Node.js, `PyJWT` in Python, or Go's `oidc`):

```typescript
import { createRemoteJWKSet, jwtVerify } from "jose";

const JWKS = createRemoteJWKSet(new URL("https://auth.sooauth.com/.well-known/jwks.json"));

export async function verifySession(token: string) {
  const { payload } = await jwtVerify(token, JWKS, {
    issuer: "https://auth.sooauth.com",
  });
  return payload; // { sub, email, ... }
}
```

---

## Features

- **Standard OIDC:** Discovery (`/.well-known/openid-configuration`), Authorize, Token, UserInfo, and JWKS endpoints.
- **Passkeys (WebAuthn):** Frictionless passwordless biometric authentication.
- **Social Login:** Google and GitHub OAuth out of the box with zero provider SDKs required in client apps.
- **Security Center:** Self-service user dashboard for active sessions, revocations, passkey management, and password changes.
- **Multi-Factor Auth (MFA):** Encrypted TOTP authenticator app support with emergency recovery backup codes.
- **Embedded Admin Dashboard:** Manage users, monitor active sessions, configure providers, and inspect audit logs directly at `/admin/`.
- **Signed Webhooks:** Real-time event notifications with HMAC-SHA256 signatures for user creations and logins.

---

## Local Development

```bash
# 1. Start Postgres & Redis
docker compose up -d postgres redis

# 2. Run Go server
cd server
go run ./cmd/sooauth

# 3. Run all tests
go test ./...
```

---

## Documentation

- [API Reference](docs/API.md)
- [Architecture](docs/ARCHITECTURE.md)
- [Self-Host Deployment Guide](docs/DEPLOY.md)
- [Security Model](docs/SECURITY.md)
- [Contributing Guide](CONTRIBUTING.md)
- [Licensing](docs/LICENSING.md)
- [AI Integration Specification (llms.txt)](https://sooauth.com/llms.txt)

---

## License

Licensed under the [GNU Affero General Public License v3.0 (AGPL-3.0-or-later)](LICENSE).
Built with care by [Sooapps](https://sooapps.com).
