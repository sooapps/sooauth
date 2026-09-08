import type { Metadata } from "next";
import { AiSetupPromptBlock } from "../../../components/ai-setup-prompt";
import { adminUrl, authUrl, oidcDiscoveryUrl } from "../../../lib/site";

export const metadata: Metadata = {
  title: "Connect your app",
  description: "OIDC integration for Sooauth — discovery, client, redirect URI, and PKCE.",
};

export default function IntegrateOidcPage() {
  const issuer = authUrl;

  return (
    <article>
      <h1 className="font-sans text-[32px] font-semibold tracking-tight text-fg">
        Connect your app (OIDC)
      </h1>
      <p className="mt-3 max-w-[52ch] text-[16px] leading-[1.6] text-fg-muted">
        Sooauth speaks standard OpenID Connect. Your app redirects users to
        authorize, exchanges a code for tokens, and reads userinfo — same as
        Auth0 or Google. Works with any stack: Next.js, Go, mobile, SPA.
      </p>

      <AiSetupPromptBlock />

      <section className="mt-10 border border-line p-6">
        <h2 className="font-sans text-[18px] font-semibold text-fg">
          Discovery
        </h2>
        <p className="mt-3 text-[15px] leading-[1.6] text-fg-muted">
          Fetch metadata once at startup:
        </p>
        <pre className="mt-4 overflow-x-auto border border-line bg-bg-subtle px-4 py-3 font-mono text-[13px] leading-[1.7] text-fg">{oidcDiscoveryUrl}</pre>
      </section>

      <section className="mt-10">
        <h2 className="font-sans text-[22px] font-semibold text-fg">
          OAuth client
        </h2>
        <p className="mt-3 text-[15px] leading-[1.6] text-fg-muted">
          In{" "}
          <a href={adminUrl} className="text-fg hover:underline">
            dashboard → Integration
          </a>
          , each project has:
        </p>
        <ul className="mt-4 grid gap-2 text-[15px] text-fg-muted">
          <li>· <strong className="font-medium text-fg">client_id</strong> — public identifier (copy from dashboard)</li>
          <li>· <strong className="font-medium text-fg">redirect_uris</strong> — exact callback URLs (must match byte-for-byte). For custom login pages, include <span className="font-mono text-[13px]">/login</span> and <span className="font-mono text-[13px]">/register</span> in addition to your OIDC callback.</li>
          <li>· PKCE required for public clients (SPAs, mobile, serverless)</li>
        </ul>
      </section>

      <section className="mt-10">
        <h2 className="font-sans text-[22px] font-semibold text-fg">
          Environment variables (typical)
        </h2>
        <p className="mt-3 text-[15px] leading-[1.6] text-fg-muted">
          Names vary by framework — point them at your dashboard values:
        </p>
        <pre className="mt-4 overflow-x-auto border border-line bg-bg-subtle px-4 py-3 font-mono text-[13px] leading-[1.7] text-fg">{`# Example — rename to match your app
AUTH_OIDC_ISSUER=${issuer}
AUTH_OIDC_CLIENT_ID=app_your_client_id_from_dashboard
AUTH_OIDC_REDIRECT_URI=https://yourapp.com/api/auth/callback/oidc`}</pre>
        <p className="mt-4 text-[15px] leading-[1.6] text-fg-muted">
          Add the same redirect URI in dashboard before testing. For custom auth UI (email/social),
          also register your app&apos;s <span className="font-mono text-[13px]">/login</span> and{" "}
          <span className="font-mono text-[13px]">/register</span> URLs — used for email verification
          return and social <span className="font-mono text-[13px]">return_to</span>, not only the OIDC
          callback. Social login (Google, GitHub, etc.) is configured in dashboard → Social login
          only — never in your app&apos;s env.
        </p>
      </section>

      <section className="mt-10">
        <h2 className="font-sans text-[22px] font-semibold text-fg">
          Authorization flow
        </h2>
        <ol className="mt-4 grid gap-3 text-[15px] leading-[1.6] text-fg-muted">
          <li>1. App sends user to <span className="font-mono text-[13px]">/oauth/authorize</span> with <span className="font-mono text-[13px]">client_id</span>, <span className="font-mono text-[13px]">redirect_uri</span>, <span className="font-mono text-[13px]">code_challenge</span> (PKCE)</li>
          <li>2. User signs in on hosted pages if needed</li>
          <li>3. Browser returns to your callback with <span className="font-mono text-[13px]">code</span></li>
          <li>4. Backend POSTs to <span className="font-mono text-[13px]">/oauth/token</span> with code + verifier</li>
          <li>5. Use access token on <span className="font-mono text-[13px]">/oauth/userinfo</span> or validate JWT via JWKS</li>
        </ol>
      </section>

      <section className="mt-10 border border-line bg-bg-subtle p-6">
        <h2 className="font-sans text-[18px] font-semibold text-fg">
          JS SDK (optional)
        </h2>
        <p className="mt-3 text-[15px] leading-[1.6] text-fg-muted">
          Monorepo package <span className="font-mono text-[13px]">@sooauth/sdk-js</span>{" "}
          wraps redirect + callback + session middleware for Next.js.
        </p>
        <pre className="mt-4 overflow-x-auto border border-line bg-bg-subtle px-4 py-3 font-mono text-[13px] leading-[1.7] text-fg">{`import { createSooauthClient } from "@sooauth/sdk-js";

const client = createSooauthClient({
  issuer: "${issuer}",
  clientId: "your-client-id",
  redirectUri: "https://yourapp.com/callback",
});`}</pre>
      </section>
    </article>
  );
}
