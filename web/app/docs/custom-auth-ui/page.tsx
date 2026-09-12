import type { Metadata } from "next";
import Link from "next/link";
import { AiCustomAuthPromptBlock } from "../../../components/ai-custom-auth-prompt";
import { adminUrl, authUrl } from "../../../lib/site";

export const metadata: Metadata = {
  title: "Custom auth UI",
  description:
    "Build your own login and sign-up with sooauth APIs — widget config, password policy, email and social auth.",
};

export default function CustomAuthUiPage() {
  const issuer = authUrl;

  return (
    <article>
      <h1 className="font-sans text-[32px] font-semibold tracking-tight text-fg">
        Custom auth UI
      </h1>
      <p className="mt-3 max-w-[58ch] text-[16px] leading-[1.6] text-fg-muted">
        Use your own login and sign-up pages — React, Vue, mobile, anything. Call
        sooauth HTTP APIs directly. Dashboard project settings (password rules,
        social providers, branding) flow through{" "}
        <span className="font-mono text-[14px] text-fg">/v1/widget/config</span>.
        No embed.js required.
      </p>

      <p className="mt-4 text-[15px] leading-[1.6] text-fg-muted">
        Prefer a drop-in script? See{" "}
        <Link href="/docs/embed-widget" className="text-fg underline-offset-2 hover:underline">
          Embed widget
        </Link>
        . Prefer full OIDC redirect? See{" "}
        <Link href="/docs/integrate-oidc" className="text-fg underline-offset-2 hover:underline">
          Connect your app (OIDC)
        </Link>
        .
      </p>

      <AiCustomAuthPromptBlock />

      <section className="mt-10">
        <h2 className="font-sans text-[22px] font-semibold text-fg">
          Integration overview
        </h2>
        <ol className="mt-4 grid gap-3 text-[15px] leading-[1.6] text-fg-muted">
          <li>1. Copy <span className="font-mono text-[13px]">client_id</span> and redirect URIs from{" "}
            <a href={adminUrl} className="text-fg underline-offset-2 hover:underline">dashboard → Integration</a></li>
          <li>2. On sign-in / sign-up page load → fetch widget config</li>
          <li>3. Render social buttons from <span className="font-mono text-[13px]">providers</span>; build password checklist from <span className="font-mono text-[13px]">password_policy</span></li>
          <li>4. POST email auth to <span className="font-mono text-[13px]">/auth/sign-in</span> or <span className="font-mono text-[13px]">/auth/sign-up</span> with <span className="font-mono text-[13px]">client_id</span></li>
          <li>5. Social: redirect to <span className="font-mono text-[13px]">/auth/social/…/start</span>, handle <span className="font-mono text-[13px]">widget_code</span> on return</li>
          <li>6. Store <span className="font-mono text-[13px]">access_token</span> in your session; protect app routes</li>
        </ol>
      </section>

      <section className="mt-10 border border-line p-6">
        <h2 className="font-sans text-[18px] font-semibold text-fg">
          1. Widget config
        </h2>
        <p className="mt-3 text-[15px] leading-[1.6] text-fg-muted">
          Single source of truth for UI — call once per page (or via your backend proxy):
        </p>
        <pre className="mt-4 overflow-x-auto border border-line bg-bg-subtle px-4 py-3 font-mono text-[13px] leading-[1.7] text-fg">{`GET ${issuer}/v1/widget/config?client_id=app_your_client_id

{
  "client_id": "app_…",
  "tenant_id": "uuid",
  "issuer": "${issuer}",
  "providers": ["google", "github"],
  "brand_name": "My App",
  "accent_color": "#FF3B3B",
  "email_verify_required": true,
  "email_verify_delivery": "link",
  "password_policy": {
    "min_length": 12,
    "require_uppercase": true,
    "require_number": true,
    "require_special": false,
    "hint": "At least 12 characters, one uppercase letter, one number"
  }
}`}</pre>
        <p className="mt-4 text-[15px] leading-[1.6] text-fg-muted">
          CORS is enabled. In production, proxy through your backend with{" "}
          <span className="font-mono text-[13px]">cache: no-store</span> so policy
          changes apply immediately.
        </p>
      </section>

      <section className="mt-10 border border-line p-6">
        <h2 className="font-sans text-[18px] font-semibold text-fg">
          2. Password policy UI (sign-up)
        </h2>
        <p className="mt-3 text-[15px] leading-[1.6] text-fg-muted">
          Rules are set in dashboard → <strong className="font-medium text-fg">Project settings</strong>.
          Build a live checklist from <span className="font-mono text-[13px]">password_policy</span>:
        </p>
        <ul className="mt-4 grid gap-2 text-[15px] text-fg-muted">
          <li>· Min length → <span className="font-mono text-[13px]">min_length</span></li>
          <li>· Uppercase → <span className="font-mono text-[13px]">require_uppercase</span></li>
          <li>· Number → <span className="font-mono text-[13px]">require_number</span></li>
          <li>· Special char → <span className="font-mono text-[13px]">require_special</span></li>
        </ul>
        <p className="mt-4 text-[15px] leading-[1.6] text-fg-muted">
          Validate client-side before submit; server always re-validates. On failure:
        </p>
        <pre className="mt-4 overflow-x-auto border border-line bg-bg-subtle px-4 py-3 font-mono text-[13px] leading-[1.7] text-fg">{`HTTP 400
{ "error": "weak_password", "message": "At least 12 characters, one uppercase letter, one number" }`}</pre>
      </section>

      <section className="mt-10 border border-line p-6">
        <h2 className="font-sans text-[18px] font-semibold text-fg">
          3. Email sign-up and sign-in
        </h2>
        <pre className="mt-4 overflow-x-auto border border-line bg-bg-subtle px-4 py-3 font-mono text-[13px] leading-[1.7] text-fg">{`POST ${issuer}/auth/sign-up
Content-Type: application/json

{ "email": "user@example.com", "password": "…", "client_id": "app_your_client_id", "return_to": "https://yourapp.com/login" }

POST ${issuer}/auth/sign-in
Content-Type: application/json

{ "email": "user@example.com", "password": "…", "client_id": "app_your_client_id" }

// Success (app user — not dashboard)
{ "access_token": "…", "email": "…", "expires_in": 3600 }`}</pre>
        <p className="mt-4 text-[15px] leading-[1.6] text-fg-muted">
          With <span className="font-mono text-[13px]">client_id</span>, users are app
          end-users linked to your project — not sooauth dashboard accounts. Optional{" "}
          <span className="font-mono text-[13px]">return_to</span> sets where email
          verification sends the user after they confirm (must match a redirect URI host;
          otherwise the first <span className="font-mono text-[13px]">/login</span> URI is used).
          Verification emails use your project name as the brand. Delivery mode is set in
          dashboard → Project: <span className="font-mono text-[13px]">link</span>,{" "}
          <span className="font-mono text-[13px]">code</span>, or{" "}
          <span className="font-mono text-[13px]">both</span> (see{" "}
          <span className="font-mono text-[13px]">email_verify_delivery</span> in widget config).
        </p>
      </section>

      <section id="verify-link" className="mt-10 border border-line p-6 scroll-mt-24">
        <h2 className="font-sans text-[18px] font-semibold text-fg">
          3a. Verify with email link
        </h2>
        <p className="mt-3 text-[15px] leading-[1.6] text-fg-muted">
          Default when <span className="font-mono text-[13px]">email_verify_delivery</span> is{" "}
          <span className="font-mono text-[13px]">link</span> or{" "}
          <span className="font-mono text-[13px]">both</span>. The mail contains a button to{" "}
          <span className="font-mono text-[13px]">{issuer}/auth/verify?token=…&amp;client_id=…&amp;return_to=…</span>.
          After verification, the user is redirected to your{" "}
          <span className="font-mono text-[13px]">return_to</span> URL with{" "}
          <span className="font-mono text-[13px]">?email_verified=1</span>. Show a success message on
          your login page — do not treat this query param as an OAuth error.
        </p>
      </section>

      <section id="verify-code" className="mt-10 border border-line p-6 scroll-mt-24">
        <h2 className="font-sans text-[18px] font-semibold text-fg">
          3b. Verify with 6-digit code
        </h2>
        <pre className="mt-4 overflow-x-auto border border-line bg-bg-subtle px-4 py-3 font-mono text-[13px] leading-[1.7] text-fg">{`POST ${issuer}/auth/verify-code
Content-Type: application/json

{ "email": "user@example.com", "code": "123456", "client_id": "app_your_client_id" }

// Success — verifies email and returns an app session
{ "message": "Email verified. You are signed in.", "access_token": "…", "email": "…", "expires_in": 3600 }`}</pre>
        <p className="mt-4 text-[15px] leading-[1.6] text-fg-muted">
          Show a code input on your check-email screen when{" "}
          <span className="font-mono text-[13px]">email_verify_delivery</span> is{" "}
          <span className="font-mono text-[13px]">code</span> or{" "}
          <span className="font-mono text-[13px]">both</span>. Resend with{" "}
          <span className="font-mono text-[13px]">POST /auth/resend-verification</span> and{" "}
          <span className="font-mono text-[13px]">client_id</span>.
        </p>
      </section>

      <section className="mt-10 border border-line p-6">
        <h2 className="font-sans text-[18px] font-semibold text-fg">
          4. Social login (Google, GitHub)
        </h2>
        <p className="mt-3 text-[15px] leading-[1.6] text-fg-muted">
          Use <span className="font-mono text-[13px]">tenant_id</span> and{" "}
          <span className="font-mono text-[13px]">client_id</span> from widget config.
          Social keys stay in sooauth dashboard only.
        </p>
        <pre className="mt-4 overflow-x-auto border border-line bg-bg-subtle px-4 py-3 font-mono text-[13px] leading-[1.7] text-fg">{`<!-- Your button links to: -->
${issuer}/auth/social/google/start
  ?tenant_id=TENANT_UUID
  &client_id=app_your_client_id
  &return_to=https://yourapp.com/api/auth/callback/social

<!-- User returns with ?widget_code=… on return_to -->
POST ${issuer}/v1/widget/exchange
{ "code": "widget_code_value", "client_id": "app_your_client_id" }`}</pre>
        <p className="mt-4 text-[15px] leading-[1.6] text-fg-muted">
          Register <span className="font-mono text-[13px]">return_to</span> host in
          dashboard redirect URIs. Custom Google OAuth (your brand on consent screen)
          is configured in dashboard → Social login.
        </p>
      </section>

      <section className="mt-10 border border-line p-6">
        <h2 className="font-sans text-[18px] font-semibold text-fg">
          5. Backend proxy (recommended)
        </h2>
        <p className="mt-3 text-[15px] leading-[1.6] text-fg-muted">
          Pattern used by production apps — thin API routes in your app:
        </p>
        <ul className="mt-4 grid gap-2 text-[15px] text-fg-muted">
          <li>· <span className="font-mono text-[13px]">GET /api/auth/widget/config</span> → proxies widget config</li>
          <li>· <span className="font-mono text-[13px]">POST /api/auth/sign-in/email</span> → proxies sign-in</li>
          <li>· <span className="font-mono text-[13px]">POST /api/auth/sign-up/email</span> → proxies sign-up (pass <span className="font-mono text-[13px]">return_to</span>)</li>
          <li>· <span className="font-mono text-[13px]">POST /api/auth/verify-code</span> → proxies code verification + session</li>
          <li>· <span className="font-mono text-[13px]">POST /api/auth/resend-verification</span> → resend mail (with <span className="font-mono text-[13px]">client_id</span>)</li>
          <li>· <span className="font-mono text-[13px]">GET /api/auth/social/[provider]/start</span> → redirect helper</li>
          <li>· <span className="font-mono text-[13px]">GET /api/auth/callback/social</span> → exchanges widget_code, sets session cookie</li>
        </ul>
        <p className="mt-4 text-[15px] leading-[1.6] text-fg-muted">
          Env: <span className="font-mono text-[13px]">SOOAUTH_ISSUER</span>,{" "}
          <span className="font-mono text-[13px]">SOOAUTH_CLIENT_ID</span>
        </p>
      </section>

      <section className="mt-10">
        <h2 className="font-sans text-[22px] font-semibold text-fg">
          Choose an approach
        </h2>
        <div className="mt-4 overflow-x-auto">
          <table className="w-full border-collapse text-left text-[14px] text-fg-muted">
            <thead>
              <tr className="border-b border-line text-fg-muted">
                <th className="py-2 pr-4 font-medium">Approach</th>
                <th className="py-2 pr-4 font-medium">Best for</th>
                <th className="py-2 font-medium">Password UI</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-line">
              <tr>
                <td className="py-3 pr-4 text-fg">Custom UI (this doc)</td>
                <td className="py-3 pr-4">Full design control, mobile, existing design system</td>
                <td className="py-3">You build checklist from config API</td>
              </tr>
              <tr>
                <td className="py-3 pr-4">
                  <Link href="/docs/embed-widget" className="text-fg hover:underline">embed.js</Link>
                </td>
                <td className="py-3 pr-4">Static sites, quick prototype</td>
                <td className="py-3">Automatic with <span className="font-mono text-[13px]">data-mode=signup</span></td>
              </tr>
              <tr>
                <td className="py-3 pr-4">
                  <Link href="/docs/integrate-oidc" className="text-fg hover:underline">OIDC redirect</Link>
                </td>
                <td className="py-3 pr-4">Server apps, hosted sooauth login page</td>
                <td className="py-3">Hosted pages / your OIDC callback only</td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>
    </article>
  );
}
