# Deploy — Coolify + Cloudflare

Monorepo layout:

| Path | Deploy target |
| --- | --- |
| `Dockerfile` + `docker-compose.coolify.yml` | **Coolify** (VDS) — auth API + login UI + admin |
| `web/` | **Cloudflare Workers** — marketing `sooauth.com` |

Auth never runs on Cloudflare Workers (Postgres + Redis + Go binary required).

---

## 1. GitHub

Push the monorepo. Coolify and Cloudflare both connect via GitHub App.

---

## 2. Coolify (auth.sooauth.com)

### New resource

1. **+ New** → **Docker Compose**
2. Connect GitHub repo (root of monorepo)
3. **Compose file:** `docker-compose.coolify.yml`
4. **Branch:** `main`

### Environment

Copy `.env.coolify.example` into Coolify **Environment Variables**. Minimum:

- `POSTGRES_PASSWORD` — strong random string
- `APP_URL` — `https://auth.sooauth.com`
- `SMTP_URL` + `SMTP_FROM` — real SMTP (verify emails won't send without this)
- `MFA_ENCRYPTION_KEY` — base64-encoded 32-byte key for encrypting TOTP secrets
- `ADMIN_EMAILS` — your email

`AUTO_MIGRATE=true` is already set in the compose file — every deploy runs goose migrations before serving.

### Domain

In Coolify, assign domain to the **`sooauth`** service (not postgres/redis):

- `auth.sooauth.com` → port `8080`

Coolify terminates TLS (Traefik). In Cloudflare DNS:

- `auth` A/CNAME → VDS IP, **Proxied** (orange cloud)
- SSL mode: **Full (strict)**

### Verify

```bash
curl -s https://auth.sooauth.com/healthz
curl -s https://auth.sooauth.com/.well-known/openid-configuration | head
```

Open:

- https://auth.sooauth.com/auth/sign-in
- https://auth.sooauth.com/admin/

---

## 3. Cloudflare (sooauth.com marketing)

Deploy `web/` — Next.js via OpenNext on Workers (same pattern as soobserver).

### Option A — Cloudflare dashboard (GitHub integration)

| Setting | Value |
| --- | --- |
| Root directory | `/` (monorepo root) |
| Build command | `corepack enable && pnpm install && pnpm --filter @sooauth/web cf:build` |
| Deploy command | `pnpm --filter @sooauth/web cf:deploy` |
| Node | 22 |

Environment variables:

| Variable | Example |
| --- | --- |
| `NEXT_PUBLIC_SITE_URL` | `https://sooauth.com` |
| `NEXT_PUBLIC_AUTH_URL` | `https://auth.sooauth.com` |
| `CLOUDFLARE_API_TOKEN` | (CI token if needed) |

### Option B — Local / CI

```bash
pnpm install
pnpm --filter @sooauth/web cf:build
pnpm --filter @sooauth/web cf:deploy
```

Set `NEXT_PUBLIC_*` in `web/.env.production` or CF dashboard before build.

### DNS

- `sooauth.com` / `www` → Cloudflare Workers (automatic after deploy)
- `auth.sooauth.com` → VDS (Coolify) — separate record

Marketing CTAs link to `NEXT_PUBLIC_AUTH_URL` (sign-up / sign-in on Coolify).

---

## 4. What runs where

```
sooauth.com          → Cloudflare Workers (web/)
auth.sooauth.com     → Coolify (Go: /auth/*, /admin/, /oauth/*, OIDC)
  ├ postgres         → Coolify internal volume
  └ redis            → Coolify internal volume
```

Go binary serves **all auth UI** (`/auth/sign-in`, `/admin/`). No separate frontend container for auth.

---

## 5. OAuth clients (soobserver etc.)

After deploy, set redirect URIs in **Admin → Integration** (per project), e.g.:

- soobserver: `https://app.sooobserver.com/api/auth/callback/oidc`

Each sign-up gets its own `client_id` (`app_…`). Dev seed client `soobserver` remains for local testing.

---

## 6. Troubleshooting

| Issue | Fix |
| --- | --- |
| Container crash loop | Check `POSTGRES_PASSWORD` set in Coolify env |
| Migrations fail | Logs: `docker logs` — DB URL uses service name `postgres` |
| Verify email missing | Set `SMTP_URL` |
| Admin empty / 403 | Redeploy latest build — dashboard uses account owner, not ADMIN_EMAILS |
| OIDC redirect mismatch | Admin → Integration tab — add exact callback URL |
| After email verify, app shows wrong error | Add app `/login` to redirect URIs; pass `return_to` on sign-up — not only OIDC callback |
| Code verification | Dashboard → Project → Verification delivery; app calls `POST /auth/verify-code` — see web docs `/docs/custom-auth-ui` |

Manual migrate (one-off):

```bash
docker compose -f docker-compose.coolify.yml exec sooauth sooauth -migrate
```
