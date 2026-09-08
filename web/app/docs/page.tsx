import type { Metadata } from "next";
import Link from "next/link";
import { ArrowRight } from "lucide-react";
import { adminUrl, authUrl, docsIntegrateUrl, signUpUrl } from "../../lib/site";

export const metadata: Metadata = {
  title: "Documentation",
  description: "Get started with Sooauth — accounts, OIDC, dashboard, self-hosting, and app integration.",
};

const cards = [
  {
    href: "/docs/getting-started",
    title: "Getting started",
    body: "Create an account, verify email, open the dashboard.",
  },
  {
    href: "/docs/integrate-oidc",
    title: "Connect your app (OIDC)",
    body: "OIDC discovery, client ID, redirect URI — any framework or SPA.",
  },
  {
    href: "/docs/custom-auth-ui",
    title: "Custom auth UI",
    body: "Your own login/sign-up — widget config, password policy, email link/code verify, AI prompt.",
  },
  {
    href: "/docs/custom-auth-ui#ai-prompt",
    title: "AI prompt (custom UI)",
    body: "Copy-paste for Cursor / ChatGPT to wire custom auth against sooauth APIs.",
  },
  {
    href: "/docs/embed-widget",
    title: "Embed widget",
    body: "Drop-in embed.js — sign-in/sign-up script, password rules, sooauth:login event.",
  },
  {
    href: "/docs/integrate-oidc#ai-prompt",
    title: "AI setup prompt",
    body: "Copy-paste prompt for Cursor / ChatGPT to wire OIDC into your codebase.",
  },
  {
    href: "/docs/admin",
    title: "Dashboard",
    body: "Integration keys, project auth rules, users, sessions, branding.",
  },
  {
    href: "/docs/google-oauth",
    title: "Google login",
    body: "Custom OAuth app, Free vs Pro branding, app callback proxy for your domain.",
  },
  {
    href: "/docs/self-host",
    title: "Self-host Sooauth",
    body: "Run the open-source Community Edition on your own server with Docker Compose.",
  },
  {
    href: "/compare/auth0",
    title: "Sooauth vs Auth0",
    body: "Understand the trade-offs between managed identity and an auth layer you can own.",
  },
  {
    href: "/compare/clerk",
    title: "Sooauth vs Clerk",
    body: "Compare hosted developer experience, OIDC portability, and self-hosting options.",
  },
];

export default function DocsHubPage() {
  return (
    <div>
      <h1 className="font-sans text-[32px] font-semibold tracking-tight text-fg">
        Documentation
      </h1>
      <p className="mt-3 max-w-[52ch] text-[16px] leading-[1.6] text-fg-muted">
        Sooauth is open-source authentication from Sooapps. End users sign up at{" "}
        <span className="font-mono text-[14px] text-fg">{authUrl.replace("https://", "")}</span>.
        Your apps use standard OIDC — no custom SDK required (one is available if you want it). Run it yourself or use the hosted service at auth.sooauth.com.
      </p>

      <section className="mt-10 border border-line bg-bg-subtle p-6">
        <h2 className="font-sans text-[18px] font-semibold text-fg">
          New here? Start in 5 minutes
        </h2>
        <ol className="mt-4 grid gap-2 text-[15px] leading-[1.55] text-fg-muted">
          <li>1. <a href={signUpUrl} className="text-fg underline-offset-2 hover:underline">Create an account</a></li>
          <li>2. Verify your email from the inbox link</li>
          <li>3. Open <a href={adminUrl} className="text-fg underline-offset-2 hover:underline">dashboard</a> and copy Integration values</li>
          <li>4. <a href={`${docsIntegrateUrl}#ai-prompt`} className="text-fg underline-offset-2 hover:underline">Use the AI prompt</a> or follow <Link href="/docs/integrate-oidc" className="text-fg underline-offset-2 hover:underline">Connect your app</Link></li>
        </ol>
      </section>

      <ul className="mt-10 grid gap-4 sm:grid-cols-2">
        {cards.map((card) => (
          <li key={card.href}>
            <Link
              href={card.href}
              className="group flex h-full flex-col border border-line p-5 transition-colors hover:border-line-strong hover:bg-bg-subtle"
            >
              <h2 className="font-sans text-[17px] font-semibold text-fg">
                {card.title}
              </h2>
              <p className="mt-2 flex-1 text-[14px] leading-[1.55] text-fg-muted">
                {card.body}
              </p>
              <span className="mt-4 inline-flex items-center gap-1 font-sans text-sm text-fg group-hover:text-fg">
                Read
                <ArrowRight size={14} aria-hidden />
              </span>
            </Link>
          </li>
        ))}
      </ul>
    </div>
  );
}
