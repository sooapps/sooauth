import type { ReactNode } from "react";
import type { Metadata } from "next";
import Link from "next/link";
import { adminUrl, authUrl, signUpUrl } from "../../../lib/site";

export const metadata: Metadata = {
  title: "Getting started",
  description: "From a new account to a working sign-in in your app in about five minutes.",
};

function Step({ n, id, title, children }: { n: number; id?: string; title: string; children: ReactNode }) {
  return (
    <section id={id} className="grid scroll-mt-24 gap-3 border-t border-line py-8 first:border-t-0 first:pt-0">
      <p className="font-mono text-xs text-fg-muted">Step {n}</p>
      <h2 className="font-sans text-[20px] font-semibold text-fg">{title}</h2>
      <div className="grid gap-3 text-[15px] leading-[1.6] text-fg-muted">{children}</div>
    </section>
  );
}

const a = "text-fg underline underline-offset-2";

export default function GettingStartedPage() {
  return (
    <article>
      <h1 className="font-sans text-[32px] font-semibold tracking-tight text-fg">Getting started</h1>
      <p className="mt-3 max-w-[60ch] text-[16px] leading-[1.6] text-fg-muted">
        Create an account, let the setup wizard configure your first project, test the sign-in, then connect
        your app. Auth lives at{" "}
        <a href={authUrl} className="font-mono text-[14px] text-fg hover:underline">{authUrl.replace("https://", "")}</a>.
      </p>

      <div className="mt-10">
        <Step n={1} title="Create your dashboard account">
          <p>
            <a href={signUpUrl} className={a}>Sign up</a> and click the verification link we email you. A first
            project is created for you.
          </p>
          <p>
            This is your <strong>dashboard</strong> account. Users of your apps get their own accounts per project — even
            with the same email they are separate. See{" "}
            <Link href="/docs/users-and-sessions#accounts" className={a}>Users and sessions</Link>.
          </p>
        </Step>

        <Step n={2} title="Follow the setup wizard">
          <p>The <a href={adminUrl} className={a}>dashboard</a> opens a five-step wizard the first time:</p>
          <ol className="grid list-decimal gap-1.5 pl-5">
            <li><strong>Project</strong> — name it and pick what you build (Next.js, SPA, backend, custom UI or embed widget).</li>
            <li><strong>Sign-in methods</strong> — email/username/phone, password or one-time code, optional 2FA. <Link href="/docs/sign-in-methods" className={a}>Details</Link>.</li>
            <li><strong>Callback URL</strong> — where users return after signing in, e.g. <code className="font-mono text-[13px]">http://localhost:3000/api/auth/callback/oidc</code>.</li>
            <li><strong>Connect</strong> — an AI prompt and snippet already filled in with your client ID and URLs.</li>
            <li><strong>Live test</strong> — sign in as a test user and see the token.</li>
          </ol>
          <p>
            Closed it? Everything stays under <strong>Getting started</strong> in the sidebar, with a checklist that ticks
            itself as you go.
          </p>
        </Step>

        <Step n={3} id="test" title="Run a live test">
          <p>
            <strong>Integration → Live test</strong> opens your project&apos;s sign-in page in a popup. Sign up or sign in as an
            end user; the dashboard shows the claims your app will receive (<code className="font-mono text-[13px]">sub</code>,
            email, custom fields…). It always shows the sign-in screen, so you can try different users.
          </p>
        </Step>

        <Step n={4} title="Connect your app">
          <p>Pick the path that matches your stack:</p>
          <ul className="grid list-disc gap-1.5 pl-5">
            <li><Link href="/docs/integrate-oidc" className={a}>OIDC with the hosted page</Link> — the least code; works with any OIDC library.</li>
            <li><Link href="/docs/embed-widget" className={a}>Embed widget</Link> — a drop-in login box on your own page.</li>
            <li><Link href="/docs/custom-auth-ui" className={a}>Custom auth UI</Link> — your own forms over the sooauth API.</li>
          </ul>
          <p>
            Easiest: open <strong>Integration → AI setup prompt</strong> in the dashboard. The prompt includes your real
            client ID, URLs, sign-in methods and sign-up fields — paste it into Claude, Cursor or Copilot.
          </p>
        </Step>

        <Step n={5} title="Go further">
          <ul className="grid list-disc gap-1.5 pl-5">
            <li><Link href="/docs/social-login" className={a}>Social login</Link> with Google or GitHub.</li>
            <li><Link href="/docs/registration-fields" className={a}>Sign-up fields</Link> to collect extra information.</li>
            <li><Link href="/docs/webhooks" className={a}>Webhooks</Link> to sync users into your database.</li>
          </ul>
        </Step>
      </div>
    </article>
  );
}
