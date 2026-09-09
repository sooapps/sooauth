# Contributing to Sooauth

Thank you for your interest in contributing to Sooauth! 🎉

We welcome contributions of all kinds—whether you are fixing a small typo, adding translations, improving documentation, reporting a bug, or implementing a major identity feature.

If this is your first time contributing to open source, don't worry! This guide provides a welcoming, step-by-step roadmap to get you up and running smoothly.

---

## Quick Navigation

- [First-Time Contributor Roadmap](#first-time-contributor-roadmap)
- [How to Contribute](#how-to-contribute)
- [Discussion vs. Issue vs. Pull Request](#discussion-vs-issue-vs-pull-request)
- [Step-by-Step GitHub Lifecycle (Fork to PR)](#step-by-step-github-lifecycle-fork-to-pr)
- [Codebase Map](#codebase-map)
- [Prerequisites & Local Setup](#prerequisites--local-setup)
- [Running Tests](#running-tests)
- [Development & Commit Standards](#development--commit-standards)
- [Need Help?](#need-help)

---

## First-Time Contributor Roadmap

Here is how a contribution flows from an idea to merged code:

```text
┌─────────────────────────┐     ┌─────────────────────────┐     ┌─────────────────────────┐
│  1. Find or Propose     │ ──> │   2. Discuss & Align    │ ──> │   3. Fork & Branch      │
│  - Browse issues        │     │  - Comment on issue     │     │  - Fork to your account │
│  - Or start Discussion  │     │  - Get it assigned      │     │  - git checkout -b feat/│
└─────────────────────────┘     └─────────────────────────┘     └─────────────────────────┘
                                                                             │
                                                                             ▼
┌─────────────────────────┐     ┌─────────────────────────┐     ┌─────────────────────────┐
│     6. Merged! 🎉       │ <── │   5. Open Pull Request  │ <── │   4. Develop & Test     │
│  - Celebrated by team   │     │  - Link issue (Fixes #) │     │  - Run Go/JS tests      │
│  - Listed in changelog  │     │  - Address feedback     │     │  - Follow standards     │
└─────────────────────────┘     └─────────────────────────┘     └─────────────────────────┘
```

---

## How to Contribute

### 1. Beginner-Friendly Issues
If you are looking for somewhere to start, look at our curated issue tags:
- [`good first issue`](https://github.com/sooapps/sooauth/issues?q=is%3Aissue+is%3Aopen+label%3A%22good+first+issue%22): Self-contained tasks designed specifically for beginners.
- [`help wanted`](https://github.com/sooapps/sooauth/issues?q=is%3Aissue+is%3Aopen+label%3A%22help+wanted%22): Well-defined tasks where we actively welcome community collaboration.

### 2. Non-Code Contributions (Equally Valuable!)
You do **not** have to write Go or TypeScript to make a meaningful impact:
- **Documentation & Guides:** Clarify explanations, fix typos, or contribute tutorials for frameworks (e.g. Next.js, SvelteKit, FastAPI, Vue).
- **Internationalization (i18n):** Help us localize hosted authentication pages and transactional emails into your native language.
- **Bug Reports & Testing:** Test self-hosting setups, report edge cases, or provide clean reproduction steps.
- **UI & Accessibility:** Report contrast or styling bugs and suggest usability improvements.

---

## Discussion vs. Issue vs. Pull Request

Not sure where to share your idea? Use this simple rule of thumb:

| Action | Where to Go | Examples |
| :--- | :--- | :--- |
| **Ask a question or get help** | [GitHub Discussions (Q&A)](https://github.com/sooapps/sooauth/discussions/categories/q-a) | *"How do I configure custom claims in Go?"*, *"Stuck on local Docker setup"* |
| **Brainstorm an idea or new feature** | [GitHub Discussions (Ideas & RFCs)](https://github.com/sooapps/sooauth/discussions/categories/ideas-rfcs) | *"Proposal for a Python FastAPI SDK"*, *"Idea for Passkey auto-fill UX"* |
| **Report a confirmed bug** | [GitHub Issues (Bug Report)](https://github.com/sooapps/sooauth/issues/new?template=bug_report.yml) | *"Favicon does not load in Safari"*, *"Revocation endpoint returns 500 on missing ID"* |
| **Submit a code or doc fix** | [Pull Request](https://github.com/sooapps/sooauth/pulls) | A branch fixing a bug or adding a feature with tests. |

> [!TIP]
> If you're planning a non-trivial feature or refactor, **always discuss it first in Discussions or an Issue** before writing code. This ensures alignment with the project architecture and avoids wasted effort.

---

## Step-by-Step GitHub Lifecycle (Fork to PR)

### 1. Fork and Clone
Click the **Fork** button at the top right of the [sooapps/sooauth repository](https://github.com/sooapps/sooauth).

Then clone your fork locally:
```bash
git clone https://github.com/<your-username>/sooauth.git
cd sooauth
```

Add the upstream repository so you can sync upstream changes anytime:
```bash
git remote add upstream https://github.com/sooapps/sooauth.git
```

### 2. Create a Topic Branch
Make sure your local `main` is up to date, then branch off:
```bash
git checkout main
git pull upstream main
git checkout -b feat/your-feature-name
```

**Branch Naming Conventions:**
- `feat/feature-name` for new capabilities or enhancements.
- `fix/bug-description` for bug fixes.
- `docs/doc-update` for documentation changes.
- `refactor/scope` for restructuring code without behavioral changes.

### 3. Make Your Changes and Test Locally
Follow our [Development Standards](#development-and-commit-standards) and verify your changes pass all tests:
```bash
# Go server tests
cd server
go test -v ./...

# Frontend / SDK tests
pnpm test
```

### 4. Commit Your Changes
We follow [Conventional Commits](https://www.conventionalcommits.org/):
```bash
git add .
git commit -m "feat(auth): add Turkish localization to hosted sign-in page"
```

Common prefixes:
- `feat:` A new user-facing feature or API method.
- `fix:` A bug fix.
- `docs:` Documentation improvements or typo fixes.
- `test:` Adding or updating tests.
- `refactor:` Code refactoring without changing functionality.

### 5. Push and Open a Pull Request
Push your branch to your personal fork:
```bash
git push -u origin feat/your-feature-name
```

Then go to the [sooauth repository](https://github.com/sooapps/sooauth) on GitHub and click **Compare & pull request**.

In your PR description:
- Provide a concise summary of what changed and why.
- Link the related issue using GitHub keywords so it closes automatically upon merge (e.g. `Fixes #42` or `Closes #18`).
- Mention how you tested the changes.
- If your PR is not ready for review yet, click the dropdown and choose **Create Draft Pull Request**.

---

## Codebase Map

The repository is structured as a Go backend with TypeScript client packages:

```text
├── server/          # Go 1.23 backend (OIDC engine, HTTP server, handlers, migrations)
│   ├── cmd/         # Binary entrypoint (cmd/sooauth)
│   ├── internal/    # Internal business logic (auth, crypto, mfa, oidc, social, store)
│   └── migrations/  # SQL schema migrations
├── admin/           # Admin dashboard UI (Vite + React)
├── web/             # Marketing website and documentation (Next.js / Cloudflare Workers)
├── sdks/            # Client SDKs
│   └── js/          # Official @sooauth/sdk-js package
├── design-system/   # Shared UI styles and tokens
└── docs/            # Technical specifications and architectural guides
```

---

## Prerequisites & Local Setup

### Prerequisites
- **Go 1.23+**
- **Docker & Docker Compose** (for PostgreSQL and Redis)
- **Node.js 20+** and **pnpm** (if modifying `web/`, `admin/`, or `sdks/`)

### Setup in 3 Minutes

1. **Configure Environment:**
   ```bash
   cp .env.example .env
   ```

2. **Start Supporting Databases:**
   ```bash
   docker compose up -d postgres redis
   ```

3. **Run the Go Server:**
   ```bash
   cd server
   go run ./cmd/sooauth
   ```

The server automatically applies database migrations on startup.
- Hosted Sign-in: `http://localhost:8080/auth/sign-in`
- Admin Dashboard: `http://localhost:8080/admin/`
- OIDC Discovery: `http://localhost:8080/.well-known/openid-configuration`

---

## Running Tests

Before opening a pull request, ensure all test suites pass:

```bash
# Backend unit & integration tests
cd server
go test -v ./...

# JavaScript SDK / Frontend typecheck & build
pnpm --filter @sooauth/sdk-js test
pnpm build
```

---

## Development & Commit Standards

1. **Go Code Quality:**
   - Format with standard tooling: `go fmt ./...` and `go vet ./...`.
   - Keep packages cohesive within `server/internal/`. Avoid circular dependencies.
   - Any new authentication, cryptographic flow, or API endpoint must include unit tests.

2. **Open Standards Compliance:**
   - Keep all OAuth/OIDC endpoints compliant with RFC 6749, RFC 7636 (PKCE), and OpenID Connect Core 1.0 specifications. Avoid proprietary lock-in.

3. **Frontend & SDK Quality:**
   - Run `pnpm typecheck` and `pnpm lint` before pushing changes to frontend or SDK packages.
   - Ensure components support both light and dark themes with accessible contrast ratios (WCAG AA).

---

## Need Help?

Stuck or have a question? You don't have to figure it out alone:
- Say hello or ask for guidance in [🤝 Looking to contribute? Start here!](https://github.com/sooapps/sooauth/discussions).
- Ask technical setup questions in [Discussions Q&A](https://github.com/sooapps/sooauth/discussions/categories/q-a).
- Report security issues privately per our [Security Policy](docs/SECURITY.md).

Thank you for helping make Sooauth the best open-source authentication platform! 🚀
