export type BlogSection = {
  heading: string;
  paragraphs: string[];
  bullets?: string[];
};

export type BlogPost = {
  slug: string;
  title: string;
  description: string;
  date: string;
  readTime: string;
  eyebrow: string;
  sections: BlogSection[];
};

export const blogPosts: BlogPost[] = [
  {
    slug: "why-we-built-sooauth",
    title: "Why we built an open-source authentication layer for SaaS teams",
    description:
      "Sooauth gives small teams a practical path between building auth from scratch and locking their product into one provider.",
    date: "2026-09-07",
    readTime: "5 min read",
    eyebrow: "Product",
    sections: [
      {
        heading: "Authentication should not decide where your product can run",
        paragraphs: [
          "Most teams start with a hosted auth provider because it is the right engineering decision. Sign-up, password reset, email verification, OAuth, sessions, and security edge cases are not the place to learn by trial and error.",
          "The problem appears later. Data residency changes, deployment requirements grow, pricing becomes harder to predict, or a product team wants one identity layer across several applications. Moving away from a provider can be harder than adopting one.",
        ],
      },
      {
        heading: "A practical escape hatch",
        paragraphs: [
          "Sooauth is built around a simple promise: use a hosted flow when speed matters, and keep a self-hosting path when ownership matters. The application integration stays standard through OIDC, so the identity layer does not leak provider-specific code into every product.",
        ],
        bullets: [
          "Standard OIDC discovery, authorization, tokens, userinfo, and JWKS",
          "Hosted email, passkeys, and social login flows",
          "A dashboard for projects, providers, sessions, and branding",
          "A Community Edition for teams that want to run the stack themselves",
        ],
      },
      {
        heading: "Built for small teams first",
        paragraphs: [
          "Sooauth is not trying to replace every enterprise identity suite on day one. It is for indie SaaS teams, product studios, and privacy-conscious companies that want a clean auth foundation without a multi-quarter platform project.",
          "The hosted Sooauth service at auth.sooauth.com is operated by Sooapps for teams that want to start without running the infrastructure themselves. Self-hosting remains available when ownership or data residency becomes the priority.",
        ],
      },
    ],
  },
  {
    slug: "self-hosted-auth-vs-managed-auth",
    title: "Self-hosted auth vs managed auth: which should your SaaS use?",
    description:
      "A practical guide to choosing between running authentication yourself and using a managed auth service.",
    date: "2026-09-07",
    readTime: "6 min read",
    eyebrow: "Guide",
    sections: [
      {
        heading: "Start with the risk you are trying to reduce",
        paragraphs: [
          "Managed authentication reduces operational work. Self-hosted authentication reduces dependency on an external service. Neither option is universally better; the right choice depends on which risk your team can handle more comfortably.",
        ],
        bullets: [
          "Choose managed auth when speed, uptime, and low operations are the priority.",
          "Choose self-hosted auth when data control, deployment ownership, or predictable infrastructure matters most.",
          "Choose a hybrid path when you want to start managed and keep a migration option.",
        ],
      },
      {
        heading: "The hybrid path is often the practical one",
        paragraphs: [
          "A good auth architecture keeps the application integration portable. OIDC gives your app a stable contract while the deployment can change behind it. This lets a small team launch with hosted authentication and revisit operations after product-market fit, instead of paying the complexity cost on day one.",
        ],
      },
      {
        heading: "Questions to ask before choosing",
        paragraphs: [
          "Ask where user data must live, who owns incident response, how much authentication downtime your product can tolerate, and whether the provider supports a real export or migration path. Pricing is important, but switching cost is often the larger long-term cost.",
        ],
      },
    ],
  },
  {
    slug: "oidc-for-indie-saas",
    title: "Why OIDC is the best default for an indie SaaS auth integration",
    description:
      "OIDC gives your application a standard identity contract and keeps provider-specific authentication details out of your codebase.",
    date: "2026-09-07",
    readTime: "5 min read",
    eyebrow: "Engineering",
    sections: [
      {
        heading: "Your app should consume identity, not implement it",
        paragraphs: [
          "An application usually needs to know who a user is, whether their email is verified, and what access they have. It should not need to own password hashing, social OAuth edge cases, WebAuthn ceremonies, token signing, or session recovery.",
          "OIDC creates a standard boundary between those responsibilities. The auth provider handles identity. Your app validates tokens and applies its own authorization rules.",
        ],
      },
      {
        heading: "The integration stays portable",
        paragraphs: [
          "With discovery and JWKS, most frameworks can integrate using an existing OIDC library. That means fewer custom dependencies and a much clearer migration story if your deployment or provider changes later.",
        ],
        bullets: [
          "Discovery describes the provider automatically.",
          "PKCE protects public clients during authorization.",
          "JWKS lets your API verify signed tokens without a shared secret.",
          "Userinfo provides a standard identity endpoint.",
        ],
      },
      {
        heading: "A small contract with a large payoff",
        paragraphs: [
          "For a small SaaS team, standard protocol support is a force multiplier. You can use the same mental model across web apps, internal tools, and future products instead of creating a new auth integration every time.",
        ],
      },
    ],
  },
  {
    slug: "passkeys-without-the-magic",
    title: "Passkeys without the magic: what your SaaS still needs to own",
    description:
      "Passkeys remove passwords from the sign-in flow, but the surrounding account and recovery design still matters.",
    date: "2026-09-08",
    readTime: "5 min read",
    eyebrow: "Security",
    sections: [
      {
        heading: "A passkey is a credential, not an account strategy",
        paragraphs: [
          "WebAuthn handles a strong authentication ceremony in the browser. Your product still needs clear account linking, recovery, session revocation, and device management around it.",
        ],
        bullets: [
          "Give users a useful device label.",
          "Show when a credential was added.",
          "Keep a recovery method that does not weaken the entire account.",
        ],
      },
      {
        heading: "Make the security model visible",
        paragraphs: [
          "A compact security center should let people review sessions, remove passkeys, and change their password without making them search through account settings. Good security UX is mostly good information architecture.",
        ],
      },
    ],
  },
  {
    slug: "oauth-provider-setup-checklist",
    title: "The OAuth provider setup checklist we use before shipping",
    description:
      "A practical checklist for callback URLs, secrets, environments, and the details that break social login in production.",
    date: "2026-09-08",
    readTime: "4 min read",
    eyebrow: "Engineering",
    sections: [
      {
        heading: "Start with exact URLs",
        paragraphs: [
          "OAuth providers compare redirect URLs strictly. Treat every environment as a separate configuration and copy the exact callback URI instead of constructing one by hand.",
        ],
        bullets: [
          "Register development, preview, and production URLs separately.",
          "Keep provider secrets on the auth service, never in the browser.",
          "Test both successful and cancelled consent flows.",
        ],
      },
      {
        heading: "Separate platform OAuth from customer branding",
        paragraphs: [
          "A platform OAuth app is a useful default for getting started. Teams that need their own domain in the provider account picker should configure a custom app and verify its branding requirements independently.",
        ],
      },
    ],
  },
  {
    slug: "sooauth-vs-better-auth-vs-clerk",
    title: "Why vibecoders are tired of complex auth: Sooauth vs Better Auth vs Clerk",
    description:
      "Comparing auth strategies for AI builders: database-coupled adapters vs proprietary SaaS lock-in vs lightweight, decoupled OIDC.",
    date: "2026-09-08",
    readTime: "6 min read",
    eyebrow: "Comparison",
    sections: [
      {
        heading: "The hidden tax of database-coupled authentication",
        paragraphs: [
          "When modern developers build applications with AI tools like Cursor or Claude Code, they want to move fast. Libraries like Better Auth or NextAuth/Auth.js promise complete control by putting user, account, session, and verification tables directly into your application database.",
          "In theory, that sounds great. In practice, it creates constant friction: every new auth feature requires schema migrations, adapter updates, and careful ORM configuration. When an AI coding assistant tries to refactor your database models, auth tables often get mangled, leading to broken migrations and halted deployments.",
        ],
      },
      {
        heading: "The SaaS lock-in and MAU price trap",
        paragraphs: [
          "The alternative has traditionally been hosted providers like Clerk or Auth0. They deliver great developer experience, but at a steep price: closed-source code, zero self-hosting capability, and pricing models that charge aggressive per-MAU rates as soon as your product starts gaining traction.",
          "Worse, their proprietary SDKs tightly couple your frontend and backend to their specific vendor APIs, making migration an expensive multi-week nightmare.",
        ],
      },
      {
        heading: "The Sooauth approach: decoupled, open, standard",
        paragraphs: [
          "Sooauth takes a different path: complete separation of concerns via standard OpenID Connect (OIDC). Your application database remains 100% focused on your business domain. Your auth service runs independently—either hosted at auth.sooauth.com or self-hosted via Docker.",
        ],
        bullets: [
          "Zero database migrations: No user or session tables cluttering your app database.",
          "Standard OIDC with PKCE: Works with any language, framework, or AI coding assistant.",
          "Freedom to self-host: AGPL-3.0 open source with a single-command install script.",
          "Built-in hosted UI: Branded login, passkeys, TOTP MFA, and social providers ready out of the box.",
        ],
      },
    ],
  },
  {
    slug: "ai-first-auth-cursor-claude-code",
    title: "Prompt-Ready Auth: building full-stack apps with AI without database migrations",
    description:
      "How decoupling your auth layer with standard OIDC makes Cursor, Windsurf, and Claude Code 10x more reliable.",
    date: "2026-09-08",
    readTime: "5 min read",
    eyebrow: "AI & Vibecoding",
    sections: [
      {
        heading: "Why AI coding assistants struggle with traditional auth",
        paragraphs: [
          "Large language models excel at well-defined protocols and clear API contracts. Where they consistently fail is multi-step state synchronization: keeping a Prisma/Drizzle schema in sync with third-party auth plugins, managing cookie encryption subtleties, and handling edge cases in custom session stores.",
          "When you ask an AI assistant to 'add authentication' using a database-coupled auth library, it often generates hundreds of lines of boilerplate, adds 5 new tables to your schema, and introduces subtle migration errors that stall development.",
        ],
      },
      {
        heading: "The power of 3 environment variables",
        paragraphs: [
          "With Sooauth, your prompt to Cursor or Claude Code becomes remarkably simple: 'Integrate authentication using Sooauth standard OIDC.'",
          "Because OIDC is an established Internet standard (RFC 6749 & RFC 7636), every modern AI model already understands the exact flow. The AI only needs to wire up 3 environment variables (ISSUER, CLIENT_ID, REDIRECT_URI) and handle standard Bearer tokens.",
        ],
        bullets: [
          "Zero database collisions between your AI-generated models and auth state.",
          "Consistent JWT claims that any API route or middleware can verify using standard JWKS.",
          "Pre-built hosted login screens, eliminating the need to prompt AI for login UI, reset password flows, and passkey registration.",
        ],
      },
    ],
  },
  {
    slug: "self-hosting-auth-in-30-seconds",
    title: "Self-hosting authentication in 30 seconds with Go and Docker",
    description:
      "Why a single Go binary beats heavy Java identity providers for modern indie infrastructure.",
    date: "2026-09-08",
    readTime: "4 min read",
    eyebrow: "DevOps",
    sections: [
      {
        heading: "The problem with traditional self-hosted identity",
        paragraphs: [
          "For years, if you wanted to self-host an open-source OIDC provider, your main option was Keycloak. While Keycloak is powerful for Fortune 500 enterprises, running a Java-based monolith that requires 1GB+ of RAM, extensive XML/JSON tuning, and minutes to boot is overkill for indie hackers and nimble startups.",
        ],
      },
      {
        heading: "Go: compiled, lean, and instant",
        paragraphs: [
          "Sooauth is built from the ground up in Go. It compiles to a single ~25MB binary that consumes less than 50MB of RAM under normal loads, connects to PostgreSQL and Redis, and boots in milliseconds.",
          "With our one-line installer (`curl -sSL https://raw.githubusercontent.com/sooapps/sooauth/main/install.sh | bash`), any developer can deploy a complete, production-ready auth system with automated secret generation and healthchecks on a $5/month VPS in under half a minute.",
        ],
      },
    ],
  },
];

export function getBlogPost(slug: string) {
  return blogPosts.find((post) => post.slug === slug);
}
