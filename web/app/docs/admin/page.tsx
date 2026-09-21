import type { Metadata } from "next";
import { adminUrl, authUrl } from "../../../lib/site";

export const metadata: Metadata = {
  title: "Dashboard",
  description: "Manage integration, social login, users, and branding in the Sooauth dashboard.",
};

export default function AdminDocsPage() {
  return (
    <article>
      <h1 className="font-sans text-[32px] font-semibold tracking-tight text-fg">
        Dashboard
      </h1>
      <p className="mt-3 max-w-[52ch] text-[16px] leading-[1.6] text-fg-muted">
        Built into the auth server at{" "}
        <a href={adminUrl} className="font-mono text-[14px] text-fg hover:underline">
          {authUrl.replace("https://", "")}/dashboard/
        </a>
        . Sign up or sign in — your account automatically gets a project with OIDC credentials.
        No env vars to configure per customer.
      </p>

      <section className="mt-10 grid gap-8">
        <div>
          <h2 className="font-sans text-[20px] font-semibold text-fg">Integration</h2>
          <p className="mt-2 text-[15px] leading-[1.6] text-fg-muted">
            Issuer, client ID, discovery URL, and redirect URIs for your app. Created on sign-up.
          </p>
        </div>
        <div>
          <h2 className="font-sans text-[20px] font-semibold text-fg">Project</h2>
          <p className="mt-2 text-[15px] leading-[1.6] text-fg-muted">
            Project name, password policy, and email verification for app users (password sign-up).
          </p>
          <ul className="mt-3 grid gap-2 text-[15px] leading-[1.6] text-fg-muted">
            <li>
              · <strong className="font-medium text-fg">Require email verification</strong> — block sign-in until verified
            </li>
            <li>
              · <strong className="font-medium text-fg">Verification delivery</strong> —{" "}
              <span className="font-mono text-[13px]">link</span> (default),{" "}
              <span className="font-mono text-[13px]">code</span> (6-digit), or{" "}
              <span className="font-mono text-[13px]">both</span>. Exposed to your app as{" "}
              <span className="font-mono text-[13px]">email_verify_delivery</span> in widget config.
            </li>
          </ul>
          <p className="mt-3 text-[15px] leading-[1.6] text-fg-muted">
            Code mode uses{" "}
            <span className="font-mono text-[13px]">POST /auth/verify-code</span> from your check-email UI. See{" "}
            <a href="/docs/custom-auth-ui#verify-code" className="text-fg underline-offset-2 hover:underline">
              Custom auth UI → verify code
            </a>
            .
          </p>
        </div>
        <div>
          <h2 className="font-sans text-[20px] font-semibold text-fg">Registration &amp; Fields</h2>
          <p className="mt-2 text-[15px] leading-[1.6] text-fg-muted">
            Configure how end users sign up and sign in, including allowed identifiers, passwordless mode, and custom registration fields.
          </p>
          <ul className="mt-3 grid gap-2 text-[15px] leading-[1.6] text-fg-muted">
            <li>
              · <strong className="font-medium text-fg">Allowed Identifiers</strong> — Choose any combination of <span className="font-mono text-[13px]">Email</span>, <span className="font-mono text-[13px]">Username</span>, and <span className="font-mono text-[13px]">Phone</span>. You can allow username-only registration for gaming communities or phone-only for mobile apps.
            </li>
            <li>
              · <strong className="font-medium text-fg">Primary Authentication Mode</strong> — Switch between <span className="font-mono text-[13px]">Password</span> (with customizable complexity rules) and <span className="font-mono text-[13px]">Passwordless OTP</span> (6-digit one-time code sent via email or SMS).
            </li>
            <li>
              · <strong className="font-medium text-fg">Custom Form Builder</strong> — Define additional fields collected during registration: <span className="font-mono text-[13px]">text</span>, <span className="font-mono text-[13px]">textarea</span>, <span className="font-mono text-[13px]">number</span>, <span className="font-mono text-[13px]">select</span>, and <span className="font-mono text-[13px]">checkbox</span>. Each field supports custom labels, placeholders, descriptions, and required constraints.
            </li>
            <li>
              · <strong className="font-medium text-fg">Dynamic External API Options</strong> — Connect dropdowns to any external REST API (e.g. Riot Games League of Legends ranks, country lists, department directories). Configure URL, request headers (e.g. API tokens), JSONPath extraction, and cache TTL.
            </li>
            <li>
              · <strong className="font-medium text-fg">Interactive Test API</strong> — Test external API endpoints directly inside the dashboard with live latency metrics and a visual preview of extracted options before saving.
            </li>
            <li>
              · <strong className="font-medium text-fg">Security Guarantee</strong> — External API secrets and tokens are securely stored on your server and proxied via <span className="font-mono text-[13px]">/v1/widget/fields/&#123;field_id&#125;/options</span>. Sensitive headers are never leaked to client browsers.
            </li>
            <li>
              · <strong className="font-medium text-fg">Metadata &amp; OIDC Claims</strong> — All custom field inputs are saved in <span className="font-mono text-[13px]">users.metadata</span> and exposed in <span className="font-mono text-[13px]">/oauth/userinfo</span> under <span className="font-mono text-[13px]">custom_claims</span>, along with standard claims (<span className="font-mono text-[13px]">preferred_username</span>, <span className="font-mono text-[13px]">phone_number</span>, <span className="font-mono text-[13px]">phone_number_verified</span>).
            </li>
          </ul>
        </div>
        <div>
          <h2 className="font-sans text-[20px] font-semibold text-fg">App users</h2>
          <p className="mt-2 text-[15px] leading-[1.6] text-fg-muted">
             End users who sign in through your app — not Sooauth dashboard accounts. Each project has
            its own user pool: the same email can register in two apps with different passwords.
            OIDC <span className="font-mono text-[13px]">sub</span> is per project user, not shared
            across projects. Remove a user to unlink them from this project.
          </p>
        </div>
        <div>
          <h2 className="font-sans text-[20px] font-semibold text-fg">Sessions</h2>
          <p className="mt-2 text-[15px] leading-[1.6] text-fg-muted">
            Active sessions per user. Revoke suspicious sessions instantly.
          </p>
        </div>
        <div>
          <h2 className="font-sans text-[20px] font-semibold text-fg">OAuth clients</h2>
          <p className="mt-2 text-[15px] leading-[1.6] text-fg-muted">
            Each project gets a public OIDC client. Set{" "}
            <span className="font-mono text-[13px]">redirect_uris</span> in the Integration tab —
            mismatch causes &quot;redirect_uri invalid&quot; errors.
          </p>
        </div>
        <div>
          <h2 className="font-sans text-[20px] font-semibold text-fg">Integration</h2>
          <p className="mt-2 text-[15px] leading-[1.6] text-fg-muted">
            Issuer, client ID, redirect URIs, embed widget snippet.{" "}
            <strong className="text-fg">Google sign-in branding</strong> (Pro + custom Google OAuth):
            set your app origin so Google shows your domain — Free plans always use sooauth platform OAuth.
          </p>
        </div>
        <div>
          <h2 className="font-sans text-[20px] font-semibold text-fg">Providers</h2>
          <p className="mt-2 text-[15px] leading-[1.6] text-fg-muted">
            Google and GitHub social login. Step-by-step:{" "}
            <a href="/docs/google-oauth" className="text-fg underline-offset-2 hover:underline">
              Google login setup
            </a>
            .
          </p>
        </div>
        <div>
          <h2 className="font-sans text-[20px] font-semibold text-fg">Theme</h2>
          <p className="mt-2 text-[15px] leading-[1.6] text-fg-muted">
            Brand name, logo URL, accent color on hosted sign-in pages.
          </p>
        </div>
        <div>
          <h2 className="font-sans text-[20px] font-semibold text-fg">Email</h2>
          <p className="mt-2 text-[15px] leading-[1.6] text-fg-muted">
            Configure custom outbound transactional email credentials for email verifications and
            password reset links/codes. Choose between:
          </p>
          <ul className="mt-3 grid gap-2 text-[15px] leading-[1.6] text-fg-muted">
            <li>
              · <strong className="font-medium text-fg">Custom SMTP</strong> — Host, port (587, 465, 25), username, password, and TLS mode (STARTTLS, Implicit TLS, None).
            </li>
            <li>
              · <strong className="font-medium text-fg">Resend</strong> — Direct transactional sending using your Resend API key.
            </li>
            <li>
              · <strong className="font-medium text-fg">Postmark</strong> — Transactional sending with your Postmark Server API Token.
            </li>
            <li>
              · <strong className="font-medium text-fg">Amazon SES</strong> — Transactional sending via AWS SES v2 with region, Access Key ID, and Secret Access Key.
            </li>
          </ul>
          <p className="mt-3 text-[15px] leading-[1.6] text-fg-muted">
            All API keys and SMTP passwords are encrypted at rest with AES-256-GCM. Use the built-in{" "}
            <strong className="font-medium text-fg">Send test email</strong> tool in the dashboard to verify deliverability before saving. If custom email is disabled or unconfigured, Sooauth seamlessly falls back to platform SMTP.
          </p>
        </div>
        <div>
          <h2 className="font-sans text-[20px] font-semibold text-fg">Audit log</h2>
          <p className="mt-2 text-[15px] leading-[1.6] text-fg-muted">
            Sign-in, sign-up, token events for security review.
          </p>
        </div>
      </section>
    </article>
  );
}
