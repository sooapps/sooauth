import type { ReactNode } from "react";
import type { Metadata } from "next";
import Link from "next/link";
import { adminUrl, authUrl, docsIntegrateUrl, signInUrl, signUpUrl } from "../../../lib/site";

export const metadata: Metadata = {
  title: "Getting started",
  description: "Create a Sooauth account and open the dashboard.",
};

function Step({
  n,
  title,
  children,
}: {
  n: number;
  title: string;
  children: ReactNode;
}) {
  return (
    <section className="grid gap-3 border-t border-line py-8 first:border-t-0 first:pt-0">
      <p className="font-mono text-xs text-fg-muted">Step {n}</p>
      <h2 className="font-sans text-[20px] font-semibold text-fg">{title}</h2>
      <div className="text-[15px] leading-[1.6] text-fg-muted">{children}</div>
    </section>
  );
}

export default function GettingStartedPage() {
  return (
    <article>
      <h1 className="font-sans text-[32px] font-semibold tracking-tight text-fg">
        Getting started
      </h1>
      <p className="mt-3 max-w-[52ch] text-[16px] leading-[1.6] text-fg-muted">
        You need a Sooauth account before connecting apps. Auth lives at{" "}
        <a href={authUrl} className="font-mono text-[14px] text-fg hover:underline">
          {authUrl.replace("https://", "")}
        </a>{" "}
        — separate from this marketing site.
      </p>

      <div className="mt-10">
        <Step n={1} title="Create an account">
          <p>
            Go to{" "}
            <a href={signUpUrl} className="text-fg underline-offset-2 hover:underline">
              sign up
            </a>
            . Use a real email — verification is required in production.
          </p>
        </Step>

        <Step n={2} title="Verify email">
          <p>
            Check your inbox for the verification link (platform dashboard accounts always use a
            link). Click it — you should see a success message and a link to sign in.
          </p>
          <p className="mt-3 text-fg-muted">
            App end-users can use link, 6-digit code, or both — configured per project in dashboard → Project. See{" "}
            <Link href="/docs/custom-auth-ui#verify-code" className="text-fg underline-offset-2 hover:underline">
              verify code
            </Link>
            .
          </p>
          <p className="mt-3 text-fg-muted">
            No email? Confirm SMTP is configured on the auth server (operators only).
          </p>
        </Step>

        <Step n={3} title="Sign in">
          <p>
            <a href={signInUrl} className="text-fg underline-offset-2 hover:underline">
              Sign in
            </a>{" "}
            with your email and password. Passkeys and social login appear when
            enabled in the dashboard.
          </p>
        </Step>

        <Step n={4} title="Open dashboard">
          <p>
            After sign-in, open{" "}
            <a href={adminUrl} className="text-fg underline-offset-2 hover:underline">
              dashboard
            </a>
            . Copy issuer and client ID from Integration, toggle social login,
            and manage app users.
          </p>
        </Step>

        <Step n={5} title="Connect your app">
          <p>
            Follow{" "}
            <Link href="/docs/integrate-oidc" className="text-fg underline-offset-2 hover:underline">
              Connect your app (OIDC)
            </Link>
            {" "}or use the{" "}
            <a href={`${docsIntegrateUrl}#ai-prompt`} className="text-fg underline-offset-2 hover:underline">
              AI setup prompt
            </a>
            {" "}with Cursor or ChatGPT. Any OIDC-compatible stack works.
          </p>
        </Step>
      </div>
    </article>
  );
}
