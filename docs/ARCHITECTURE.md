# Architecture

sooauth is a single Go binary that ships hosted auth pages, OIDC endpoints, and an embedded admin SPA.

## v1 components

| Piece | Role |
| --- | --- |
| `server/` | Auth logic, OIDC, WebAuthn, sessions, JWT/JWKS |
| PostgreSQL | Users, credentials, sessions, refresh tokens, audit log |
| Redis | Rate limits, WebAuthn challenges, ephemeral verification state |
| Webhooks | Signed project events with persisted delivery attempts |
| `admin/` | Operator panel (embedded via `go:embed` in M9) |
| `web/` | Marketing site (separate deploy, Cloudflare) |
| `sdks/js` | Thin OIDC client + middleware + React hook |

## Integration model

Downstream apps (e.g. soobserver) are **OIDC relying parties**. They validate access tokens via JWKS discovery — no vendor-specific protocol.

```
Browser → sooauth hosted pages → session / tokens
App API → validate JWT (JWKS) or introspect
```

## Multi-tenant seam

Tables reserve `tenant_id` / `project_id` for later. v1 is single-tenant.

Projects are already used as the isolation boundary for app users, themes,
providers, redirect URIs, and webhooks. Organization-level roles remain a
roadmap feature.

## soobserver integration (planned)

Replace better-auth with sooauth OIDC client. Self-host: sooauth + soobserver compose stack. SaaS: managed sooauth + managed soobserver.

## Datastores

Two only: Postgres + Redis. No third database in v1.
