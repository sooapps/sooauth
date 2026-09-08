import Link from "next/link";
import { ArrowRight, Bot, Check, Cpu, Database, KeyRound, Mail, Shield, Sparkles, Terminal } from "lucide-react";
import { SiteFooter } from "../components/site-footer";
import { SiteHeader } from "../components/site-header";
import { GridFigure } from "../components/ui/grid-figure";
import { buttonClass } from "../components/ui/button";
import {
  Card,
  Container,
  Eyebrow,
  Section,
  SectionHeading,
} from "../components/ui/layout";
import {
  faqs,
  pricingTiers,
  signInUrl,
  signUpUrl,
  testimonials,
} from "../lib/site";
import { blogPosts } from "../lib/blog";

const faqJsonLd = {
  "@context": "https://schema.org",
  "@type": "FAQPage",
  mainEntity: faqs.map((item) => ({
    "@type": "Question",
    name: item.q,
    acceptedAnswer: { "@type": "Answer", text: item.a },
  })),
};

const steps = [
  {
    title: "Start in Cloud or 10-Second Self-Host",
    body: "Use the free hosted service at auth.sooauth.com, or run 'curl -sSL https://get.sooauth.com | bash' on your own server.",
  },
  {
    title: "Prompt your AI or Grab 3 Env Vars",
    body: "Copy your client ID and issuer URL into Cursor, Claude Code, or your .env file. No database migrations required.",
  },
  {
    title: "Instant OIDC Sign-in",
    body: "Users authenticate via hosted pages, passkeys, Google, or GitHub. Your app simply receives standard, verified OIDC tokens.",
  },
];

const features = [
  {
    icon: Database,
    title: "Zero Database Migrations",
    body: "Unlike Better Auth or NextAuth, Sooauth never touches your app database. No Prisma/Drizzle schema migrations breaking during AI refactors.",
  },
  {
    icon: Shield,
    title: "Pure Standard OIDC",
    body: "Discovery, authorize, token, userinfo, and JWKS endpoints compliant with RFC 6749 and RFC 7636. Works with any language or framework.",
  },
  {
    icon: KeyRound,
    title: "Passkeys and Social Login",
    body: "Biometric WebAuthn passkeys, Google, and GitHub ready on the same sign-in screen. Toggle providers without shipping code.",
  },
  {
    icon: Terminal,
    title: "10-Second Self-Host or Cloud",
    body: "Built in Go with minimal memory usage (~25MB binary). Run anywhere with Docker Compose, or let Sooapps handle operations in the cloud.",
  },
];

const comparisonRows = [
  {
    feature: "Open-source & self-hostable",
    sooauth: "✅ Yes (AGPL-3.0)",
    clerk: "❌ Proprietary cloud only",
    betterAuth: "✅ Yes (MIT)",
    keycloak: "✅ Yes (Apache 2.0)",
  },
  {
    feature: "Modifies your app's database?",
    sooauth: "🛡️ Zero DB migrations",
    clerk: "🛡️ None",
    betterAuth: "⚠️ Requires DB tables & schemas",
    keycloak: "🛡️ None",
  },
  {
    feature: "Protocol compliance",
    sooauth: "✅ Standard OIDC + PKCE",
    clerk: "❌ Proprietary SDK",
    betterAuth: "❌ Custom protocol",
    keycloak: "✅ Standard OIDC & SAML",
  },
  {
    feature: "Memory footprint",
    sooauth: "⚡ Minimal (~25MB Go binary)",
    clerk: "Cloud only",
    betterAuth: "Node.js runtime",
    keycloak: "🐘 Heavy (Java 1GB+ RAM)",
  },
  {
    feature: "MAU pricing tax",
    sooauth: "✅ Free self-host / flat",
    clerk: "❌ $0.02+/MAU spike",
    betterAuth: "✅ Free self-host",
    keycloak: "✅ Free self-host",
  },
  {
    feature: "Prompt-ready for AI agents",
    sooauth: "✨ 1 prompt (standard OIDC)",
    clerk: "⚠️ Vendor SDK methods",
    betterAuth: "⚠️ ORM migration bugs",
    keycloak: "❌ Complex enterprise setup",
  },
];

export default function Page() {
  const featuredArticles = blogPosts.slice(0, 3);

  return (
    <>
      <script
        type="application/ld+json"
        dangerouslySetInnerHTML={{ __html: JSON.stringify(faqJsonLd) }}
      />
      <SiteHeader />

      <main id="main">
        {/* Hero */}
        <Container>
          <div className="grid gap-12 py-[clamp(3.5rem,7vw,6.5rem)] lg:grid-cols-12 lg:items-center lg:gap-8">
            <div className="lg:col-span-7">
              <Eyebrow className="rise rise-1 inline-flex items-center gap-1.5">
                <Sparkles size={14} className="text-fg" aria-hidden />
                Open-source auth for vibecoders & indie SaaS
              </Eyebrow>
              <h1 className="rise rise-2 mt-5 font-sans text-[clamp(2.5rem,5vw,4rem)] font-bold leading-[1.04] tracking-[-0.03em] text-fg">
                Authentication
                <br />
                without the homework.
              </h1>
              <p className="rise rise-3 mt-6 max-w-[54ch] text-[17px] leading-[1.6] text-fg-muted">
                No database migrations in your app. No $0.02/MAU vendor lock-in traps.
                Sooauth gives developers standard OIDC, passkeys, and hosted login pages
                you can self-host in 10 seconds or run on our managed cloud.
              </p>
              <div className="rise rise-4 mt-9 flex flex-wrap gap-3">
                <a href={signUpUrl} className={buttonClass("primary")}>
                  Get started free
                  <ArrowRight size={16} aria-hidden />
                </a>
                <a href="#prompt" className={buttonClass("secondary")}>
                  Vibecoder prompt
                </a>
                <Link
                  href="/docs/getting-started"
                  className={buttonClass("ghost")}
                >
                  Docs
                </Link>
              </div>
            </div>
            <div className="rise rise-4 lg:col-span-5">
              <GridFigure />
            </div>
          </div>
        </Container>

        {/* Vibecoder & AI Prompt Section */}
        <Section id="prompt" labelledBy="prompt-heading">
          <SectionHeading
            id="prompt-heading"
            eyebrow="AI-First & Vibecoder Ready"
            title="Prompt once. Have auth working in seconds."
            lead="Using Cursor, Claude Code, Windsurf, or v0? AI agents understand standard OIDC natively. No hallucinated SDK methods, no database migration collisions."
          />

          <div className="mt-10 grid gap-6 lg:grid-cols-12">
            <div className="border border-line bg-bg-subtle p-6 lg:col-span-7 md:p-8">
              <div className="flex items-center justify-between border-b border-line pb-4">
                <div className="flex items-center gap-2">
                  <Bot size={18} className="text-fg" />
                  <span className="font-mono text-xs font-semibold uppercase tracking-wider text-fg">
                    Copy-Paste Agent Prompt
                  </span>
                </div>
                <span className="font-mono text-[11px] text-fg-muted">
                  For Cursor / Claude Code / v0
                </span>
              </div>
              <pre className="mt-5 overflow-x-auto rounded bg-bg p-4 font-mono text-[13px] leading-[1.7] text-fg select-all border border-line">
{`"Integrate authentication into this app using Sooauth OIDC.
Use process.env.SOOAUTH_ISSUER (default: https://auth.sooauth.com),
process.env.SOOAUTH_CLIENT_ID, and standard PKCE redirect flow.
Direct users to the hosted sign-in page.
Do NOT create user tables or run database migrations in our app database."`}
              </pre>
              <p className="mt-4 text-[13px] text-fg-muted">
                Need machine-readable specs for your agent? See{" "}
                <a href="/llms.txt" className="text-fg underline underline-offset-2">
                  /llms.txt
                </a>{" "}
                and{" "}
                <a href="/llms-full.txt" className="text-fg underline underline-offset-2">
                  /llms-full.txt
                </a>
                .
              </p>
            </div>

            <div className="border border-line bg-bg p-6 lg:col-span-5 md:p-8 flex flex-col justify-between">
              <div>
                <div className="flex items-center gap-2 border-b border-line pb-4">
                  <Terminal size={18} className="text-fg" />
                  <span className="font-mono text-xs font-semibold uppercase tracking-wider text-fg">
                    Self-Host in 10 Seconds
                  </span>
                </div>
                <p className="mt-4 text-[14px] leading-[1.6] text-fg-muted">
                  Prefer running it on your own server? Run our single-command installer to boot Postgres, Redis, and Sooauth with auto-generated secure secrets:
                </p>
                <pre className="mt-4 overflow-x-auto rounded bg-bg-subtle p-3.5 font-mono text-[12px] text-fg border border-line select-all">
curl -sSL https://raw.githubusercontent.com/sooapps/sooauth/main/install.sh | bash
                </pre>
              </div>

              <div className="mt-6 pt-4 border-t border-line flex items-center justify-between text-xs text-fg-muted">
                <span>Docker & Compose required</span>
                <Link href="/docs/self-host" className="text-fg font-medium hover:underline">
                  Full self-host guide →
                </Link>
              </div>
            </div>
          </div>
        </Section>

        {/* How it works */}
        <Section id="how" labelledBy="how-heading">
          <SectionHeading
            id="how-heading"
            eyebrow="How it works"
            title="Three steps. Zero database homework."
          />
          <ol className="mt-12 grid gap-px overflow-hidden border border-line bg-line md:grid-cols-3">
            {steps.map((item, i) => (
              <li key={item.title} className="bg-bg p-6 md:p-8">
                <p className="font-mono text-xs text-fg-muted">
                  {String(i + 1).padStart(2, "0")}
                </p>
                <h3 className="mt-3 font-sans text-[19px] font-semibold text-fg">
                  {item.title}
                </h3>
                <p className="mt-2 text-[15px] leading-[1.6] text-fg-muted">
                  {item.body}
                </p>
              </li>
            ))}
          </ol>
          <div className="mt-10">
            <Link
              href="/docs/integrate-oidc"
              className={buttonClass("secondary")}
            >
              Connect your app
              <ArrowRight size={16} aria-hidden />
            </Link>
          </div>
        </Section>

        {/* Features */}
        <Section id="features" labelledBy="features-heading">
          <SectionHeading
            id="features-heading"
            eyebrow="Core Architecture"
            title="Everything modern auth needs. Nothing it doesn't."
            lead="A clean separation between identity and your application domain. Standard protocols, hosted security, and complete ownership."
          />
          <div className="mt-14 grid gap-px border border-line bg-line md:grid-cols-2">
            {features.map((item) => {
              const Icon = item.icon;
              return (
                <div
                  key={item.title}
                  className="bg-bg p-8 md:p-10 flex flex-col justify-between"
                >
                  <div>
                    <div className="flex h-12 w-12 items-center justify-center border border-line-strong bg-bg-subtle mb-6">
                      <Icon size={24} className="text-fg" aria-hidden />
                    </div>
                    <h3 className="font-sans text-[20px] font-semibold tracking-[-0.01em] text-fg">
                      {item.title}
                    </h3>
                    <p className="mt-3 text-[15px] leading-[1.6] text-fg-muted">
                      {item.body}
                    </p>
                  </div>
                </div>
              );
            })}
          </div>
        </Section>

        {/* Comparison Section with Matrix */}
        <Section id="compare" labelledBy="compare-heading">
          <SectionHeading
            id="compare-heading"
            eyebrow="Honest Comparison"
            title="Sooauth vs. Clerk vs. Better Auth vs. Keycloak"
            lead="Choosing your authentication layer is a deployment and ownership decision. Here is how Sooauth stacks up."
          />

          <div className="mt-10 overflow-x-auto border border-line">
            <table className="w-full text-left border-collapse bg-bg text-[14px]">
              <thead>
                <tr className="border-b border-line bg-bg-subtle font-mono text-xs text-fg uppercase tracking-wider">
                  <th className="p-4 md:p-5">Feature</th>
                  <th className="p-4 md:p-5 text-fg font-bold bg-bg/60 border-l border-r border-line">Sooauth</th>
                  <th className="p-4 md:p-5">Clerk</th>
                  <th className="p-4 md:p-5">Better Auth</th>
                  <th className="p-4 md:p-5">Keycloak</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-line">
                {comparisonRows.map((row) => (
                  <tr key={row.feature} className="hover:bg-bg-subtle/50 transition-colors">
                    <td className="p-4 md:p-5 font-medium text-fg">{row.feature}</td>
                    <td className="p-4 md:p-5 font-semibold text-fg bg-bg-subtle/40 border-l border-r border-line">
                      {row.sooauth}
                    </td>
                    <td className="p-4 md:p-5 text-fg-muted">{row.clerk}</td>
                    <td className="p-4 md:p-5 text-fg-muted">{row.betterAuth}</td>
                    <td className="p-4 md:p-5 text-fg-muted">{row.keycloak}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          <div className="mt-8 grid gap-4 sm:grid-cols-2">
            <Link href="/compare/clerk" className="border border-line bg-bg p-5 hover:bg-bg-subtle transition-colors flex items-center justify-between">
              <div>
                <h4 className="font-medium text-fg">Deep Dive: Sooauth vs Clerk</h4>
                <p className="text-xs text-fg-muted mt-1">Comparing hosted developer experience vs deployment ownership and MAU costs.</p>
              </div>
              <ArrowRight size={16} className="text-fg-muted" />
            </Link>
            <Link href="/blog/sooauth-vs-better-auth-vs-clerk" className="border border-line bg-bg p-5 hover:bg-bg-subtle transition-colors flex items-center justify-between">
              <div>
                <h4 className="font-medium text-fg">Why Vibecoders Are Tired of Complex Auth</h4>
                <p className="text-xs text-fg-muted mt-1">Comparing database-coupled adapters vs lightweight, decoupled OIDC.</p>
              </div>
              <ArrowRight size={16} className="text-fg-muted" />
            </Link>
          </div>
        </Section>

        {/* Blog / Articles Section */}
        <Section id="blog" labelledBy="blog-heading">
          <SectionHeading
            id="blog-heading"
            eyebrow="From the Blog"
            title="Insights for builders & architects."
            lead="Pragmatic engineering notes on authentication, AI-driven development, and self-hosting."
          />
          <div className="mt-12 grid gap-6 md:grid-cols-3">
            {featuredArticles.map((post) => (
              <Card as="article" key={post.slug} className="flex flex-col justify-between p-6">
                <div>
                  <span className="font-mono text-[11px] uppercase tracking-wider text-fg-muted">
                    {post.eyebrow} · {post.readTime}
                  </span>
                  <h3 className="mt-3 font-sans text-[18px] font-semibold text-fg leading-snug">
                    <Link href={`/blog/${post.slug}`} className="hover:underline">
                      {post.title}
                    </Link>
                  </h3>
                  <p className="mt-2 text-[14px] text-fg-muted line-clamp-3 leading-relaxed">
                    {post.description}
                  </p>
                </div>
                <div className="mt-6 pt-4 border-t border-line">
                  <Link
                    href={`/blog/${post.slug}`}
                    className="inline-flex items-center gap-1 text-xs font-medium text-fg hover:underline"
                  >
                    Read article <ArrowRight size={13} aria-hidden />
                  </Link>
                </div>
              </Card>
            ))}
          </div>
        </Section>

        {/* Testimonials */}
        <Section id="testimonials" labelledBy="testimonials-heading">
          <SectionHeading
            id="testimonials-heading"
            eyebrow="Why teams choose Sooauth"
            title="Convenience now. Ownership later."
          />
          <ul className="mt-12 grid gap-6 md:grid-cols-3">
            {testimonials.map((t) => (
              <Card as="li" key={t.quote}>
                <p className="text-[16px] leading-[1.6] text-fg">
                  &ldquo;{t.quote}&rdquo;
                </p>
                <p className="mt-6 font-mono text-xs uppercase tracking-[0.14em] text-fg-muted">
                  {t.name} · {t.role}
                </p>
              </Card>
            ))}
          </ul>
        </Section>

        {/* Pricing */}
        <Section id="pricing" labelledBy="pricing-heading">
          <SectionHeading
            id="pricing-heading"
            eyebrow="Simple pricing"
            title="Start free. Upgrade when you scale."
            lead="Use hosted Sooauth at auth.sooauth.com. If you need full infrastructure control, the Community Edition is 100% open-source and free to self-host."
          />
          <div className="mt-12 grid gap-px border border-line bg-line lg:grid-cols-3">
            {pricingTiers.map((tier) => (
              <article
                key={tier.name}
                className={
                  tier.highlight
                    ? "relative bg-bg-subtle p-8 md:p-10"
                    : "bg-bg p-8 md:p-10"
                }
              >
                {tier.highlight && "badge" in tier && tier.badge ? (
                  <span className="absolute right-8 top-8 border border-line-strong px-2 py-1 font-mono text-[10px] uppercase tracking-[0.14em] text-fg">
                    {tier.badge}
                  </span>
                ) : null}
                <p className="font-mono text-xs uppercase tracking-[0.14em] text-fg-muted">
                  {tier.name}
                </p>
                <p className="mt-4 font-sans text-[40px] font-bold leading-none tracking-[-0.02em] text-fg">
                  {tier.price}
                  {tier.cadence ? (
                    <span className="ml-2 text-[15px] font-normal text-fg-muted">
                      {tier.cadence}
                    </span>
                  ) : null}
                </p>
                <ul className="mt-6 grid gap-2 text-[15px] leading-[1.5] text-fg-muted">
                  {tier.features.map((f) => (
                    <li key={f} className="flex items-center gap-2">
                      <Check size={14} className="text-fg shrink-0" />
                      <span>{f}</span>
                    </li>
                  ))}
                </ul>
                <a
                  href={tier.cta.href}
                  className={buttonClass(
                    tier.highlight ? "primary" : "secondary",
                    "md",
                    "mt-8 w-full",
                  )}
                >
                  {tier.cta.label}
                </a>
              </article>
            ))}
          </div>
          <p className="mt-6 font-mono text-xs text-fg-muted">
            Hosted Sooauth is operated by Sooapps. Community Edition is AGPL-3.0 open source.
          </p>
        </Section>

        {/* Final CTA */}
        <Section id="start" labelledBy="start-heading">
          <div className="grid gap-10 lg:grid-cols-2 lg:items-center">
            <div>
              <h2
                id="start-heading"
                className="font-sans text-[clamp(2rem,4vw,3rem)] font-bold leading-[1.05] tracking-[-0.03em] text-fg"
              >
                Ready to own your sign-in layer?
              </h2>
              <p className="mt-4 max-w-[42ch] text-[16px] leading-[1.6] text-fg-muted">
                Create a project in 60 seconds or run it on your own server.
                Check out the{" "}
                <Link
                  href="/docs/getting-started"
                  className="text-fg underline underline-offset-2"
                >
                  getting started guide
                </Link>
                .
              </p>
              <div className="mt-8 flex flex-wrap gap-3">
                <a href={signUpUrl} className={buttonClass("primary")}>
                  Create free account
                  <ArrowRight size={16} aria-hidden />
                </a>
                <a href={signInUrl} className={buttonClass("secondary")}>
                  Sign in
                </a>
              </div>
            </div>
            <pre className="overflow-x-auto border border-line bg-bg-subtle px-5 py-5 font-mono text-[13px] leading-[1.75] text-fg select-all">{`# .env
SOOAUTH_ISSUER=https://auth.sooauth.com
SOOAUTH_CLIENT_ID=your_client_id
SOOAUTH_REDIRECT_URI=http://localhost:3000/callback`}</pre>
          </div>
        </Section>

        {/* FAQ */}
        <Section id="faq" labelledBy="faq-heading">
          <SectionHeading id="faq-heading" eyebrow="FAQ" title="Frequently asked questions." />
          <dl className="mt-10 grid gap-px border border-line bg-line">
            {faqs.map((item) => (
              <div key={item.q} className="bg-bg p-6 md:p-8">
                <dt className="font-sans text-[17px] font-semibold text-fg">
                  {item.q}
                </dt>
                <dd className="mt-2.5 max-w-[65ch] text-[15px] leading-[1.6] text-fg-muted">
                  {item.a}
                </dd>
              </div>
            ))}
          </dl>
        </Section>
      </main>

      <SiteFooter />
    </>
  );
}
