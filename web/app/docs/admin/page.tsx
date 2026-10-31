import type { Metadata } from "next";
import { adminUrl } from "../../../lib/site";
import { C, DocLink, DocTitle, H2, List, Note, P, Table } from "../../../components/docs-kit";

export const metadata: Metadata = {
  title: "Dashboard",
  description: "A map of the sooauth dashboard: where every setting lives and what it does.",
};

export default function AdminDocsPage() {
  return (
    <article>
      <DocTitle lead={<>The <DocLink href={adminUrl}>dashboard</DocLink> manages one project at a time — switch or create projects at the top of the sidebar. Every page has its own link (<C>/dashboard/p/&lt;project&gt;/&lt;page&gt;</C>) and a “Read the docs” link to the matching guide.</>}>
        Dashboard
      </DocTitle>

      <H2 id="navigation">Finding your way around</H2>
      <P>The dashboard home lists your projects and recent activity. Open a project and the sidebar switches to that project&apos;s pages; the arrow at the top goes back. The first item in the breadcrumb switches project and keeps the page you are on. Press <C>⌘K</C> or <C>Ctrl K</C> to jump to any page.</P>

      <H2 id="start">Getting started</H2>
      <P>A checklist that ticks itself as you configure the project, and the setup wizard. The wizard opens automatically for a new project. See <DocLink href="/docs/getting-started">Getting started</DocLink>.</P>

      <H2 id="authentication">Authentication</H2>
      <Table
        head={["Page", "What you set", "Guide"]}
        rows={[
          ["Sign-in methods", "Identifiers (email, username, phone), password / one-time code / both, two-step verification, phone codes via webhook.", <DocLink key="1" href="/docs/sign-in-methods">Sign-in methods</DocLink>],
          ["Social login", "Google and GitHub, platform app or your own OAuth app, Google branding on paid plans.", <DocLink key="2" href="/docs/social-login">Social login</DocLink>],
          ["Sign-up fields", "Extra fields on the sign-up form, including dropdowns from your API.", <DocLink key="3" href="/docs/registration-fields">Sign-up fields</DocLink>],
          ["Passwords & verification", "Password rules and email verification (link, code or both).", <DocLink key="4" href="/docs/sign-in-methods#password-and-verification">Passwords</DocLink>],
        ]}
      />

      <H2 id="integration">Integration</H2>
      <Table
        head={["Page", "What it's for"]}
        rows={[
          ["Keys & URLs", <>Client ID, issuer, discovery URL, tenant ID, endpoints, and the redirect URIs users may return to.</>],
          ["AI setup prompt", "A prompt and snippet for your stack, filled in with this project's real values."],
          ["Live test", <>Sign in to the project as an end user in a popup and see the claims your app receives. Uses <C>prompt=login</C>, so it always shows the sign-in page.</>],
          ["Webhooks", <>Endpoints, events and delivery logs. <DocLink href="/docs/webhooks">Webhooks</DocLink></>],
        ]}
      />

      <H2 id="users">Users</H2>
      <List>
        <li><strong>Users</strong> — search, disable/enable, delete (type the email to confirm), reset 2FA. <DocLink href="/docs/users-and-sessions">Users and sessions</DocLink></li>
        <li><strong>Sessions</strong> — signed-in sessions of this project&apos;s users; revoke any.</li>
        <li><strong>Audit log</strong> — sign-ins and admin actions.</li>
      </List>
      <Note>Your dashboard account is separate from your apps&apos; users. Signing up to your own app with the same email creates a separate user; disabling or deleting it never affects your dashboard access.</Note>

      <H2 id="branding">Branding</H2>
      <P><strong>Theme</strong> sets the brand name, logo and accent color on the hosted page and widget.</P>
      <P>
        <span id="email" className="scroll-mt-24" />
        <strong>Email</strong> sends verification, sign-in and 2FA codes from your own domain through SMTP, Resend, Postmark or Amazon SES. Use <strong>Send test email</strong> before going live. Until it is enabled, emails come from sooauth&apos;s mail server.
      </P>

      <H2 id="project">Project</H2>
      <P><strong>Project settings</strong> holds the name and default language of hosted pages and emails. <strong>Billing</strong> lives under Account on the dashboard home, because your plan covers every project.</P>

      <H2 id="self-hosted">Self-hosted</H2>
      <P>The Docker image serves the same dashboard at <C>/dashboard/</C>. Building from source: run <C>pnpm --filter @sooauth/admin build:embed</C> before <C>go build</C>. See <DocLink href="/docs/self-host">Self-host Sooauth</DocLink>.</P>
    </article>
  );
}
