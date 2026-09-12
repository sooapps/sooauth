# Sooauth Launch Kit 🚀

Complete directory submission guide, ready-to-use copy, founder comments, and platform links for launching Sooauth across 21+ startup & developer platforms.

---

## 1. Quick Copy-Paste Assets

### Product Identity
- **Name:** Sooauth
- **Website:** `https://sooauth.com`
- **Hosted Cloud:** `https://auth.sooauth.com`
- **GitHub Repository:** `https://github.com/sooapps/sooauth`
- **Interactive Playground:** `https://sooauth.com/playground`
- **Documentation:** `https://sooauth.com/docs`
- **License:** AGPL-3.0 (Community Edition) / Free Public Beta (Hosted Cloud)

### Taglines
- **Primary Tagline (Max 60 chars):**  
  `Decoupled, open-source auth without database pollution.`
- **Alternative Tagline:**  
  `Open-source auth & OIDC server with hosted UI & passkeys.`
- **Short Tagline (One sentence):**  
  `Self-hostable, developer-first authentication with zero DB migrations and standard OIDC.`

### Short Description (Max 240 chars)
> Stop polluting your application database with messy user/session tables and breaking migration loops. Sooauth is an open-source, decoupled identity engine supporting OIDC, passkeys, and social logins. Free public beta in the cloud or 1-command Docker self-hosting.

### Long Pitch / Description
> Authentication shouldn't break your database schema or lock your users into proprietary closed clouds.
>
> Modern libraries like Better Auth and NextAuth force user, session, and verification tables directly into your application database using ORM adapters. When AI agents (Cursor, Claude Code) refactor your code, they frequently break schema migrations.
>
> **Sooauth completely decouples identity:**
> - **Zero Database Pollution:** Your application database stays 100% clean. Your app simply validates standard OIDC JWT tokens.
> - **Vibecoder & AI-Ready:** Just paste 3 environment variables (`ISSUER`, `CLIENT_ID`, `REDIRECT_URI`) into your AI prompt. No hallucinations, no broken migration scripts.
> - **Hosted UI & Embed Widget:** Modern split-screen login pages or an embeddable JS widget for your custom app.
> - **Biometric Passkeys & Social Logins:** WebAuthn passkeys, Google, and GitHub login enabled with a single click.
> - **100% Free Public Beta + AGPL-3.0 Open Source:** Start free in the hosted cloud (`auth.sooauth.com`) with all Pro features unlocked, or self-host on your own VPS with `docker compose up -d`.

### Categories & Tags
- `Developer Tools`, `Open Source`, `Authentication`, `Security`, `API`, `Identity & Access Management (IAM)`, `SaaS`, `Self-Hosted`, `Golang`, `Next.js`, `OIDC / OAuth`

---

## 2. Founder / Maker First Comment (Product Hunt & Hacker News)

### Product Hunt Maker Comment
```markdown
Hey Product Hunt! 👋

I'm one of the creators of Sooauth. We built Sooauth after watching developers (and our own AI coding agents) constantly get stuck in database migration hell.

Traditional auth libraries force 5 to 10 extra tables (users, sessions, accounts, verifications) into your application database. Every time you tweak a table or ask an AI agent to refactor a model, migrations break and authentication fails. On the other hand, managed SaaS providers like Clerk and Auth0 lock you into closed ecosystems with aggressive per-user pricing.

Sooauth offers a better way:
1. **Decoupled Architecture:** Your app database remains untouched. Sooauth speaks standard OpenID Connect (RFC 6749 / RFC 7636).
2. **Hosted UI & Embed Widget:** Beautiful split-screen auth pages, passkeys (WebAuthn), and social logins without importing bloated SDKs.
3. **True Ownership:** Hosted Cloud is currently in 100% free public beta (no credit cards, all Pro features unlocked). Prefer self-hosting? The Community Edition is open-source under AGPL-3.0 and runs via Docker in seconds.

Try the live interactive widget preview here: https://sooauth.com/playground
Check out the code on GitHub: https://github.com/sooapps/sooauth

We'd love your feedback, bug reports, and feature requests!
```

### Hacker News Show HN Title & Body
- **Title:** `Show HN: Sooauth – Open-source, decoupled auth without database lock-in`
- **Body:**
```markdown
Hi HN,

We built Sooauth (https://sooauth.com), an open-source identity and OIDC server written in Go with a React/Next.js frontend.

Why another auth solution?
Most modern auth libraries (NextAuth, Better Auth, etc.) inject user and session tables directly into your application database via ORMs. While convenient for simple projects, it tightly couples identity to your runtime schema. When building with AI assistants (Cursor, Claude Code), ORM migrations are frequently a point of failure.

Sooauth completely isolates authentication:
- Your application database has zero auth tables. Your app simply validates signed OIDC tokens.
- Compliant with standard OIDC Discovery, PKCE, JWKS, and UserInfo.
- Built-in biometric passkeys (WebAuthn), Google/GitHub OAuth, and hosted login UI.
- Lightweight Go binary (boots in milliseconds, ~25MB memory footprint).

You can self-host via Docker Compose (`docker compose up -d`) or use our hosted cloud beta at https://auth.sooauth.com (currently 100% free with all Pro features unlocked, no credit card required).

Repository: https://github.com/sooapps/sooauth
Live Widget Playground: https://sooauth.com/playground

We would love to hear your thoughts on decoupled auth architectures, passkey UX, and developer ergonomics.
```

---

## 3. Platform Submission Directory & URLs

### Tier 1: Instant Submission / High DR Backlinks (Submit Today)

| # | Platform | DR | Submission URL | Notes |
|---|---|---|---|---|
| 1 | **AlternativeTo** | 80 | https://alternativeto.net/software/create/ | List as alternative to Clerk, Auth0, Supabase Auth, Keycloak |
| 2 | **DevHunt** | 63 | https://devhunt.org/tool/add | Developer tool specific; high conversion |
| 3 | **SaaSHub** | 80 | https://www.saashub.com/submit | Great SEO & alternatives directory |
| 4 | **Microlaunch** | 64 | https://microlaunch.net/submit | Clean submission form; fast approval |
| 5 | **Uneed** | 75 | https://www.uneed.best/submit-a-tool | Very active indie community |
| 6 | **WhatLaunched Today** | 50+ | http://whatlaunched.today/ | Submit directly |
| 7 | **Peerlist** | 77 | https://peerlist.io/projects | Great tech/developer network |
| 8 | **StartupBase** | 73 | https://startupbase.io/submit | Startup directory |
| 9 | **Tiny Startups** | 71 | https://tinystartups.co/submit | Micro SaaS & indie friendly |
| 10 | **SideProjectors** | 70 | https://www.sideprojectors.com | Developer side project showcase |
| 11 | **Launching Next** | 54 | https://www.launchingnext.com/submit/ | Classic startup listing |
| 12 | **Open Launch** | 68 | https://openlaunch.co | Open directory |
| 13 | **SourceForge** | 92 | https://sourceforge.net/create/ | Open-source software listing |
| 14 | **Micro SaaS Examples** | 48 | https://microsaasexamples.com | Curated list |
| 15 | **Dev Resources** | 40 | https://devresourc.es/submit | Developer curation |
| 16 | **Startup Buffer** | 40 | https://startupbuffer.com/submit | Directory listing |
| 17 | **Tool Folio** | 40 | https://toolfolio.io/submit | Curated tools |
| 18 | **Ctrl Alt CC** | 37 | https://ctrlalt.cc | Startup directory |

### Tier 2: Community Posts (1-2 Days In)

| Platform | DR | Strategy |
|---|---|---|
| **Hacker News (Show HN)** | 91 | Post on a Tuesday or Wednesday morning (between 8am-10am EST). Use the Show HN format provided above. |
| **Indie Hackers** | 81 | Share the journey: "Why I stopped putting auth tables in my app database and built Sooauth in public". |

### Tier 3: Major Launch Events

| Platform | DR | Preparation Checklist |
|---|---|---|
| **Product Hunt** | 91 | Launch on a Tuesday or Wednesday at 00:01 AM PST. Have 5 high-res screenshots (Hot Red design system, embed playground, split-screen login, admin dashboard). Prepare Maker comment. |
| **BetaList** | 76 | Submit at least 2 weeks before public scaling launch for beta queue inclusion. |
