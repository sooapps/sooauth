# Contributing to Sooauth

Thank you for contributing to Sooauth! We appreciate all contributions—from bug fixes and documentation improvements to new identity features.

Here is a pragmatic guide to getting your development environment running and submitting clean pull requests.

---

## Codebase Map

The repository is structured as a Go backend with TypeScript client packages:

```
├── server/          # Go 1.23 backend (OIDC engine, HTTP server, handlers, migrations)
│   ├── cmd/         # Binary entrypoint (cmd/sooauth)
│   ├── internal/    # Internal business logic (auth, crypto, mfa, oidc, social, store)
│   └── migrations/  # SQL schema migrations
├── admin/           # Admin dashboard UI
├── web/             # Marketing website and documentation (Next.js / Cloudflare Workers)
├── sdks/            # Client SDKs
│   └── js/          # Official @sooauth/sdk-js package
├── design-system/   # Shared UI styles and tokens
└── docs/            # Technical specifications and guides
```

---

## Prerequisites

- **Go 1.23+**
- **Docker & Docker Compose** (for PostgreSQL and Redis)
- **Node.js 20+** and **pnpm** (if touching `web/` or `sdks/`)

---

## Local Setup in 3 Minutes

### 1. Clone and Configure
```bash
git clone https://github.com/sooapps/sooauth.git
cd sooauth
cp .env.example .env
```

### 2. Start Supporting Databases
```bash
docker compose up -d postgres redis
```

### 3. Run the Go Server
```bash
cd server
go run ./cmd/sooauth
```

The server will automatically run database migrations on boot.
- Hosted Sign-in: `http://localhost:8080/auth/sign-in`
- Admin Dashboard: `http://localhost:8080/admin/`
- OIDC Discovery: `http://localhost:8080/.well-known/openid-configuration`

---

## Running Tests

Before submitting any code changes, ensure all tests pass:

```bash
# Backend tests
cd server
go test -v ./...

# JavaScript SDK / Frontend tests
pnpm test
```

---

## Development Standards

1. **Go Code Quality:**
   - Format with standard tooling: `go fmt ./...` and `go vet ./...`.
   - Keep packages cohesive within `server/internal/`. Avoid circular dependencies.
   - Any new authentication or cryptographic flow must include unit tests.

2. **No Proprietary Lock-In:**
   - Keep all OAuth/OIDC endpoints compliant with RFC 6749, RFC 7636 (PKCE), and OpenID Connect Core 1.0 specifications.

3. **Commit Messages:**
   - We prefer [Conventional Commits](https://www.conventionalcommits.org/):
     - `feat: add passkey credential nickname edit`
     - `fix: correct token revocation status code`
     - `docs: clarify Next.js callback flow`

---

## Pull Request Process

1. Fork the repository and create your branch from `main`:
   ```bash
   git checkout -b feat/your-feature-name
   ```
2. Make your changes and add tests where appropriate.
3. Verify all tests pass locally (`go test ./...`).
4. Push your branch to your fork and open a Pull Request against `main`.
5. In the PR description, explain:
   - What problem this solves
   - How you verified the change locally
