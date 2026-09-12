import type { Metadata } from "next";
import Link from "next/link";
import { authUrl } from "../../../lib/site";
import { EmbedWidgetPlayground } from "../../../components/embed-widget-playground";

export const metadata: Metadata = {
  title: "Embed widget",
  description:
    "Drop-in sign-in and sign-up with embed.js — social login, email/password, password policy, and events.",
};

export default function EmbedWidgetPage() {
  const issuer = authUrl;

  return (
    <article>
      <h1 className="font-sans text-[32px] font-semibold tracking-tight text-fg">
        Embed widget
      </h1>
      <p className="mt-3 max-w-[58ch] text-[16px] leading-[1.6] text-fg-muted">
        The fastest way to add sooauth login to any site — no React required. Load{" "}
        <span className="font-mono text-[14px] text-fg">embed.js</span> from your
        auth server; it renders Google/GitHub buttons and email forms, reads your
        dashboard theme, and fires a browser event when sign-in succeeds.
      </p>
      <p className="mt-4 text-[15px] leading-[1.6] text-fg-muted">
        Building your own UI instead? See{" "}
        <Link href="/docs/custom-auth-ui" className="text-fg underline-offset-2 hover:underline">
          Custom auth UI
        </Link>{" "}
        (API reference +{" "}
        <Link href="/docs/custom-auth-ui#ai-prompt" className="text-fg underline-offset-2 hover:underline">
          AI prompt
        </Link>
        ).
      </p>

      <section className="mt-10">
        <h2 className="font-sans text-[22px] font-semibold text-fg">
          Quick start
        </h2>
        <p className="mt-3 text-[15px] leading-[1.6] text-fg-muted">
          Copy from dashboard → Integration, or paste this (replace{" "}
          <span className="font-mono text-[13px]">app_…</span> with your client ID):
        </p>
        <pre className="mt-4 overflow-x-auto border border-line bg-bg-subtle px-4 py-3 font-mono text-[13px] leading-[1.7] text-fg">{`<div id="sooauth-widget"></div>
<script
  src="${issuer}/v1/widget/embed.js"
  data-client-id="app_your_client_id"
  data-mode="signin"
  data-target="#sooauth-widget"
  async
></script>`}</pre>
      </section>

      <section className="mt-12">
        <div className="flex items-center justify-between mb-4">
          <div>
            <h2 className="font-sans text-[22px] font-semibold text-fg">
              Interactive playground
            </h2>
            <p className="mt-1 text-[15px] leading-[1.6] text-fg-muted">
              Configure options below to see real-time updates and preview your drop-in login widget.
            </p>
          </div>
          <Link
            href="/playground"
            className="hidden sm:inline-flex items-center text-xs font-semibold text-accent hover:underline"
          >
            Open full page ↗
          </Link>
        </div>
        <EmbedWidgetPlayground />
      </section>

      <section className="mt-12">
        <h2 className="font-sans text-[22px] font-semibold text-fg">
          Script attributes
        </h2>
        <div className="mt-4 overflow-x-auto">
          <table className="w-full border-collapse text-left text-[14px] text-fg-muted">
            <thead>
              <tr className="border-b border-line text-fg-muted">
                <th className="py-2 pr-4 font-medium">Attribute</th>
                <th className="py-2 pr-4 font-medium">Required</th>
                <th className="py-2 font-medium">Description</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-line">
              <tr>
                <td className="py-3 pr-4 font-mono text-[13px] text-fg">data-client-id</td>
                <td className="py-3 pr-4">Yes</td>
                <td className="py-3">Public OAuth client ID from dashboard → Integration</td>
              </tr>
              <tr>
                <td className="py-3 pr-4 font-mono text-[13px] text-fg">data-mode</td>
                <td className="py-3 pr-4">No</td>
                <td className="py-3">
                  <span className="font-mono text-[13px]">signin</span> (default) or{" "}
                  <span className="font-mono text-[13px]">signup</span>. Sign-up mode shows
                  live password rules from your project settings.
                </td>
              </tr>
              <tr>
                <td className="py-3 pr-4 font-mono text-[13px] text-fg">data-target</td>
                <td className="py-3 pr-4">No</td>
                <td className="py-3">CSS selector where the widget mounts (default{" "}
                  <span className="font-mono text-[13px]">#sooauth-widget</span>)</td>
              </tr>
              <tr>
                <td className="py-3 pr-4 font-mono text-[13px] text-fg">data-issuer</td>
                <td className="py-3 pr-4">No</td>
                <td className="py-3">Auth base URL if script is loaded from another origin (default: script origin)</td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>

      <section className="mt-10 border border-line bg-bg-subtle p-6">
        <h2 className="font-sans text-[18px] font-semibold text-fg">
          Password requirements (sign-up)
        </h2>
        <p className="mt-3 text-[15px] leading-[1.6] text-fg-muted">
          Rules come from dashboard → <strong className="font-medium text-fg">Project settings</strong>{" "}
          (min length, uppercase, number, special). The widget reads them from{" "}
          <span className="font-mono text-[13px]">GET /v1/widget/config</span>.
        </p>
        <ul className="mt-4 grid gap-2 text-[15px] text-fg-muted">
          <li>
            · Use <span className="font-mono text-[13px]">data-mode=&quot;signup&quot;</span> on
            your registration page — embed.js shows a live ✓/✗ checklist and disables submit until
            the password passes.
          </li>
          <li>
            · <span className="font-mono text-[13px]">data-mode=&quot;signin&quot;</span> does{" "}
            <em>not</em> show rules (sign-in only validates on the server).
          </li>
          <li>
            · Custom UI (React, mobile, etc.)? See{" "}
            <Link href="/docs/custom-auth-ui" className="text-fg underline-offset-2 hover:underline">
              Custom auth UI
            </Link>{" "}
            — fetch config and build your own checklist.
          </li>
        </ul>
        <pre className="mt-4 overflow-x-auto border border-line bg-bg-subtle px-4 py-3 font-mono text-[13px] leading-[1.7] text-fg">{`<!-- Sign-up page with live password rules -->
<div id="sooauth-widget"></div>
<script
  src="${issuer}/v1/widget/embed.js"
  data-client-id="app_your_client_id"
  data-mode="signup"
  async
></script>`}</pre>
      </section>

      <section className="mt-10">
        <h2 className="font-sans text-[18px] font-semibold text-fg">
          Email verification
        </h2>
        <p className="mt-3 text-[15px] leading-[1.6] text-fg-muted">
          embed.js uses the <strong className="font-medium text-fg">link</strong> flow only
          (user clicks the button in the email). For 6-digit codes or link+code, set{" "}
          <span className="font-mono text-[13px]">email_verify_delivery</span> in dashboard → Project
          and build a check-email screen with{" "}
          <Link href="/docs/custom-auth-ui#verify-code" className="text-fg underline-offset-2 hover:underline">
            Custom auth UI
          </Link>
          .
        </p>
      </section>

      <section className="mt-10">
        <h2 className="font-sans text-[22px] font-semibold text-fg">
          After sign-in (JavaScript event)
        </h2>
        <p className="mt-3 text-[15px] leading-[1.6] text-fg-muted">
          Email sign-in and Google/GitHub (via widget redirect) both dispatch:
        </p>
        <pre className="mt-4 overflow-x-auto border border-line bg-bg-subtle px-4 py-3 font-mono text-[13px] leading-[1.7] text-fg">{`document.addEventListener("sooauth:login", (e) => {
  const { access_token, email, expires_in } = e.detail;
  // Send access_token to your backend, set session cookie, redirect, etc.
});`}</pre>
        <p className="mt-4 text-[15px] leading-[1.6] text-fg-muted">
          Google/GitHub return to your page with{" "}
          <span className="font-mono text-[13px]">?widget_code=…</span> in the URL; embed.js
          exchanges it automatically and then fires the same event.
        </p>
      </section>

      <section className="mt-10">
        <h2 className="font-sans text-[22px] font-semibold text-fg">
          Password reset (Forgot password)
        </h2>
        <p className="mt-3 text-[15px] leading-[1.6] text-fg-muted">
          embed.js includes a built-in &quot;Forgot password?&quot; toggle on sign-in. Users can request
          password reset instructions, or enter their 6-digit OTP code directly inside the widget to choose a new password without leaving your page.
        </p>
        <p className="mt-3 text-[15px] leading-[1.6] text-fg-muted">
          When a user successfully updates their password, embed.js fires:
        </p>
        <pre className="mt-4 overflow-x-auto border border-line bg-bg-subtle px-4 py-3 font-mono text-[13px] leading-[1.7] text-fg">{`document.addEventListener("sooauth:password_reset", (e) => {
  console.log("Password reset completed for:", e.detail.email);
});`}</pre>
      </section>

      <section className="mt-10">
        <h2 className="font-sans text-[22px] font-semibold text-fg">
          Custom UI instead?
        </h2>
        <p className="mt-3 text-[15px] leading-[1.6] text-fg-muted">
          embed.js is optional. For your own components, use the same APIs documented under{" "}
          <Link href="/docs/custom-auth-ui" className="text-fg underline-offset-2 hover:underline">
            Custom auth UI
          </Link>{" "}
          — widget config, sign-in/sign-up POST, email verify (link/code), social callback, and an{" "}
          <Link href="/docs/custom-auth-ui#ai-prompt" className="text-fg underline-offset-2 hover:underline">
            AI setup prompt
          </Link>
          .
        </p>
      </section>

      <section className="mt-10">
        <h2 className="font-sans text-[22px] font-semibold text-fg">
          embed.js vs OIDC redirect
        </h2>
        <ul className="mt-4 grid gap-3 text-[15px] leading-[1.6] text-fg-muted">
          <li>
            · <strong className="font-medium text-fg">embed.js</strong> — drop-in UI, returns
            an access token via event (good for SPAs that manage their own session)
          </li>
          <li>
            · <Link href="/docs/integrate-oidc" className="text-fg underline-offset-2 hover:underline">OIDC redirect</Link> — full authorize → callback flow with authorization code (recommended for server-side apps)
          </li>
        </ul>
        <p className="mt-4 text-[15px] leading-[1.6] text-fg-muted">
          You can use both in the same project — e.g. OIDC for the main app and embed.js on a
          marketing landing page.
        </p>
      </section>
    </article>
  );
}
