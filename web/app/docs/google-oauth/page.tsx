import type { Metadata } from "next";
import { authUrl } from "../../../lib/site";

export const metadata: Metadata = {
  title: "Google login setup",
  description: "Configure Google OAuth for Sooauth sign-in and sign-up.",
};

export default function GoogleOAuthDocsPage() {
  const origin = authUrl;
  const platformRedirect = `${authUrl}/auth/social/google/callback`;

  return (
    <article>
      <h1 className="font-sans text-[32px] font-semibold tracking-tight text-fg">
        Google login setup
      </h1>
      <p className="mt-3 max-w-[52ch] text-[16px] leading-[1.6] text-fg-muted">
        Google buttons appear on sign-in / sign-up after you add credentials. Works
        for both registration and login (same flow).
      </p>

      <section className="mt-10 border border-line p-6">
        <h2 className="font-sans text-[18px] font-semibold text-fg">
          Free vs Pro branding
        </h2>
        <p className="mt-3 text-[15px] leading-[1.6] text-fg-muted">
          <strong className="text-fg">Free</strong> plans use sooauth platform Google OAuth.
          The account picker shows <strong className="text-fg">sooauth.com</strong>.
        </p>
        <p className="mt-3 text-[15px] leading-[1.6] text-fg-muted">
          <strong className="text-fg">Pro+</strong> with a{" "}
          <strong className="text-fg">custom Google OAuth app</strong> can route the callback
          through your app origin so Google shows your domain (e.g. Soobrief). See{" "}
          <a href="#app-callback" className="text-fg hover:underline">app callback</a> below.
        </p>
      </section>

      <section className="mt-10 border border-line p-6">
        <h2 className="font-sans text-[18px] font-semibold text-fg">
          1. Google Cloud Console
        </h2>
        <p className="mt-3 text-[15px] leading-[1.6] text-fg-muted">
          APIs &amp; Services → Credentials → Create OAuth client ID →{" "}
          <strong className="font-medium text-fg">Web application</strong>
        </p>
        <p className="mt-3 text-[15px] leading-[1.6] text-fg-muted">
          OAuth consent screen → set your app name and logo (e.g. Soobrief).
        </p>
      </section>

      <section className="mt-8">
        <h2 className="font-sans text-[20px] font-semibold text-fg">
          2. Authorized JavaScript origins
        </h2>
        <p className="mt-2 text-[15px] text-fg-muted">Add your app origin and sooauth:</p>
        <pre className="mt-3 overflow-x-auto border border-line bg-bg-subtle px-4 py-3 font-mono text-[13px] text-fg">{`https://yourapp.com
${origin}`}</pre>
      </section>

      <section className="mt-8">
        <h2 className="font-sans text-[20px] font-semibold text-fg">
          3. Authorized redirect URIs
        </h2>
        <p className="mt-2 text-[15px] text-fg-muted">
          Dashboard → <strong className="text-fg">Social login → Google</strong> shows the exact
          URI to paste. Platform default:
        </p>
        <pre className="mt-3 overflow-x-auto border border-line bg-bg-subtle px-4 py-3 font-mono text-[13px] text-fg">{platformRedirect}</pre>
        <p className="mt-3 text-[15px] text-fg-muted">
          With Pro + app origin configured, use the app-specific URI from Integration instead
          (e.g. <code className="text-fg">https://api.yourapp.com/api/auth/google/callback</code>).
        </p>
      </section>

      <section className="mt-8">
        <h2 className="font-sans text-[20px] font-semibold text-fg">
          4. Add to sooauth
        </h2>
        <p className="mt-2 text-[15px] leading-[1.6] text-fg-muted">
          Dashboard → <strong className="text-fg">Social login → Google</strong> → Custom OAuth app:
        </p>
        <pre className="mt-4 overflow-x-auto border border-line bg-bg-subtle px-4 py-3 font-mono text-[13px] leading-[1.7] text-fg">{`Client ID + Client secret → Save custom credentials`}</pre>
      </section>

      <section id="app-callback" className="mt-10 border border-line bg-bg-subtle p-6">
        <h2 className="font-sans text-[18px] font-semibold text-fg">
          App callback (Pro — show your domain on Google)
        </h2>
        <p className="mt-3 text-[15px] leading-[1.6] text-fg-muted">
          Google shows the callback domain on the account picker. Route the callback through your app
          so users see your brand instead of sooauth.com.
        </p>
        <ol className="mt-4 list-decimal space-y-3 pl-5 text-[15px] leading-[1.6] text-fg-muted">
          <li>
            Dashboard → <strong className="text-fg">Integration</strong> → set{" "}
            <strong className="text-fg">App origin</strong> (must match a redirect URI host,
            e.g. <code className="text-fg">https://api.soobrief.com</code>).
          </li>
          <li>
            Copy the generated Google redirect URI into Google Cloud Console.
          </li>
          <li>
            Add a proxy route in your app that forwards Google&apos;s query string to sooauth:
          </li>
        </ol>
        <pre className="mt-4 overflow-x-auto border border-line bg-bg-subtle px-4 py-3 font-mono text-[13px] leading-[1.7] text-fg">{`// GET /api/auth/google/callback?code=...&state=...
app.get("/api/auth/google/callback", (req, res) => {
  const target = new URL("${platformRedirect}");
  for (const [key, value] of new URL(req.url, "http://x").searchParams) {
    target.searchParams.set(key, value);
  }
  res.redirect(302, target.toString());
});`}</pre>
        <p className="mt-4 text-[15px] text-fg-muted">
          sooauth still completes the token exchange; your route is only a browser redirect hop.
        </p>
      </section>

      <section className="mt-10 border border-line p-6">
        <h2 className="font-sans text-[18px] font-semibold text-fg">
          GitHub (bonus)
        </h2>
        <p className="mt-3 text-[15px] leading-[1.6] text-fg-muted">
          GitHub → Settings → Developer settings → OAuth App → Authorization callback URL:
        </p>
        <pre className="mt-3 overflow-x-auto border border-line bg-bg-subtle px-4 py-3 font-mono text-[13px] text-fg">{`${authUrl}/auth/social/github/callback`}</pre>
        <p className="mt-3 text-[15px] text-fg-muted">
          Same app-origin pattern works with <code className="text-fg">/api/auth/github/callback</code> on Pro.
        </p>
      </section>
    </article>
  );
}
