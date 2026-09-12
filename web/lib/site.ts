export const siteName = "Sooauth";
export const siteTitle = "Sooauth — Authentication Without the Homework";
export const siteDescription =
  "Open-source authentication built for vibecoders, AI builders, and indie SaaS. Zero DB migrations, no MAU lock-in, standard OIDC, passkeys, and self-hosting in 10 seconds.";
export const siteUrl =
  process.env.NEXT_PUBLIC_SITE_URL ?? "https://sooauth.com";
export const authUrl =
  process.env.NEXT_PUBLIC_AUTH_URL ?? "https://auth.sooauth.com";

export const signUpUrl = `${authUrl}/auth/sign-up`;
export const signInUrl = `${authUrl}/auth/sign-in`;
export const adminUrl = `${authUrl}/dashboard/`;
export const docsUrl = `${siteUrl}/docs`;
export const docsIntegrateUrl = `${siteUrl}/docs/integrate-oidc`;
export const oidcDiscoveryUrl = `${authUrl}/.well-known/openid-configuration`;

export const githubUrl = "https://github.com/sooapps/sooauth";
export const contactEmail = "info@sooapps.com";

export const socialLinks = [
  { label: "GitHub", href: githubUrl },
];

export const seoLinks = [
  { href: "/docs/getting-started", label: "Getting started" },
  { href: "/docs/integrate-oidc", label: "Connect your app (OIDC)" },
  { href: "/docs/google-oauth", label: "Google login" },
  { href: "/docs/self-host", label: "Self-host" },
  { href: "/blog", label: "Blog" },
  { href: "/compare/better-auth", label: "Sooauth vs Better Auth" },
  { href: "/compare/clerk", label: "Sooauth vs Clerk" },
  { href: "/compare/auth0", label: "Sooauth vs Auth0" },
];

export const testimonials = [
  {
    quote:
      "When building apps with Cursor or Claude, database-heavy auth libraries constantly break with schema migrations. Sooauth gives you 3 env vars and instant OIDC. It just works.",
    name: "Built for",
    role: "Vibecoders & AI Builders",
  },
  {
    quote:
      "We wanted auth we could own without building every screen ourselves. Sooauth gave us standard OIDC and a hosted flow we could ship quickly without vendor lock-in.",
    name: "Built for",
    role: "Indie SaaS teams",
  },
  {
    quote:
      "The important part is the escape hatch: start with the hosted cloud, then self-host with a single docker-compose command when your compliance or data rules require it.",
    name: "Built for",
    role: "Privacy-conscious teams",
  },
];

export const pricingTiers = [
  {
    name: "Community",
    price: "$0",
    cadence: "/ forever",
    description: "100% open-source under AGPL-3.0. Deploy on your own VPS or Docker stack with complete data sovereignty.",
    features: [
      "Single-command Docker Compose setup",
      "Unlimited users, projects & logins",
      "Standard OIDC + PKCE + Passkeys",
      "Google, GitHub, Facebook & X social logins",
      "Embed widgets & hosted auth UI",
      "Isolated Postgres with zero DB migrations",
      "Community support on GitHub Discussions",
    ],
    cta: { label: "Self-host via Docker", href: "/docs/self-host" },
    highlight: false,
  },
  {
    name: "Hosted Cloud",
    price: "$0",
    cadence: "/ public beta",
    description: "Zero infrastructure to manage. High-availability managed cloud hosted at auth.sooauth.com.",
    badge: "100% Free in Beta",
    features: [
      "Instant setup in under 2 minutes",
      "All Pro features unlocked completely free",
      "Custom login branding & accent colors",
      "Custom OAuth app credentials & redirect URIs",
      "Email verification & password reset flows",
      "Permanent grandfathered perks for early adopters",
      "No credit card required to start",
    ],
    cta: { label: "Claim free beta account", href: signUpUrl },
    highlight: true,
  },
  {
    name: "Enterprise",
    price: "Custom",
    cadence: "/ scale",
    description: "For scaling teams requiring dedicated VPC isolation, custom SLAs, compliance guarantees, or hands-on migration.",
    features: [
      "Dedicated multi-region cloud cluster",
      "Custom SLA & direct engineering Slack channel",
      "Enterprise SSO (SAML 2.0 / Okta / Azure AD)",
      "Automated SCIM user provisioning",
      "Custom audit exports & SOC 2 compliance support",
      "White-glove migration from Auth0, Clerk, or Firebase",
    ],
    cta: { label: "Talk to Sooapps", href: `mailto:${contactEmail}?subject=Sooauth%20Enterprise` },
    highlight: false,
  },
];

export const faqs = [
  {
    q: "Why is Sooauth Hosted Cloud currently 100% free?",
    a: "Sooauth is in active public beta. We want developers to test the DX, connect their AI coding tools, and stress-test our OIDC endpoints without billing friction. All Pro capabilities (custom branding, custom OAuth keys, passkeys, unlimited projects) are completely unlocked for free with no credit card required. Early adopters who build on Sooauth now will receive permanent grandfathered perks.",
  },
  {
    q: "What makes Sooauth different from Better Auth?",
    a: "Better Auth manages user and session tables directly inside your application database using ORM adapters (Prisma, Drizzle, etc.). This means every auth feature requires schema migrations, schema upkeep, and tight coupling to TypeScript/Node. Sooauth completely decouples authentication: your application database remains 100% clean, and your app simply validates standard OIDC JWTs. Your AI coding assistants won't get stuck in schema migration loops.",
  },
  {
    q: "What makes Sooauth different from Clerk?",
    a: "Clerk is closed-source, cannot be self-hosted, locks your user data into their proprietary cloud, and scales with aggressive monthly active user (MAU) pricing tiers. Sooauth is AGPL-3.0 open source, gives you an instant one-line self-host option (`curl -sSL ... | bash`), and uses standard OpenID Connect (RFC 6749 / RFC 7636) so you can switch hosting or providers at any time.",
  },
  {
    q: "What makes Sooauth different from Keycloak?",
    a: "Keycloak is an enterprise Java behemoth requiring 1GB+ RAM just to boot, dozens of complex XML/JSON configurations, and significant devops overhead. Sooauth is written in Go, compiles to a lean ~25MB binary, uses minimal memory, boots in milliseconds, and provides modern hosted auth pages out of the box.",
  },
  {
    q: "Can I migrate from Hosted Cloud to self-hosting later?",
    a: "Yes, 100%. Sooauth strictly decouples auth from proprietary storage models. Because both Hosted Cloud and the Community Edition run identical OIDC schemas and Postgres backends, you can export your data and spin up self-hosted Docker containers whenever your compliance or governance policies require.",
  },
  {
    q: "Why is Sooauth ideal for Vibecoders and AI Agents (Cursor, Claude Code, v0)?",
    a: "AI agents excel at standard protocols and struggle with bespoke SDKs and multi-table database migrations. Because Sooauth is pure OIDC with discovery endpoints and hosted auth UI, you can paste one prompt into Cursor or Claude Code, and it writes working auth in seconds using 3 standard environment variables without hallucinations.",
  },
  {
    q: "Does Sooauth touch or modify my app's database?",
    a: "No. Sooauth never touches your application database. Sooauth runs its own isolated storage for credentials, sessions, and audit logs. Your application receives clean, cryptographically signed OIDC tokens.",
  },
  {
    q: "Can I self-host Sooauth on my own VPS?",
    a: "Yes! You can run our automated script `curl -sSL https://raw.githubusercontent.com/sooapps/sooauth/main/install.sh | bash` or use `docker compose up -d`. It sets up PostgreSQL, Redis, and the Sooauth server with auto-generated secure secrets in under 30 seconds.",
  },
  {
    q: "Does Sooauth support passkeys and social logins?",
    a: "Yes. Passkeys (WebAuthn biometric login), Google, and GitHub social sign-in can be toggled in one click from the dashboard. Your application doesn't need to import separate OAuth SDKs or manage provider client secrets.",
  },
];
