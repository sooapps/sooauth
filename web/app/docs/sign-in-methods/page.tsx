import type { Metadata } from "next";
import { authUrl } from "../../../lib/site";
import { C, Code, DashboardPath, DocLink, DocTitle, H2, H3, List, Note, P, Table } from "../../../components/docs-kit";

export const metadata: Metadata = {
  title: "Sign-in methods",
  description: "Choose identifiers (email, username, phone), password or one-time code, and two-step verification.",
};

export default function SignInMethodsPage() {
  return (
    <article>
      <DocTitle lead={<>Every project decides what its users sign in with and how they prove it. Configure it in <DashboardPath>Dashboard → Authentication → Sign-in methods</DashboardPath>. The hosted sign-in page, the embed widget and the API all follow the same settings.</>}>
        Sign-in methods
      </DocTitle>

      <H2 id="identifiers">Identifiers</H2>
      <P>Identifiers are what a user types to say who they are. Enable at least one.</P>
      <Table
        head={["Identifier", "Sign-in", "Notes"]}
        rows={[
          ["Email", "Password or code", "Can receive verification links, sign-in codes and 2FA codes."],
          ["Username", "Password", "A username has no inbox. Code sign-in works only if the account also has an email or phone; the code goes there."],
          ["Phone", "Password or code", <>Codes are sent by your own SMS provider through the <DocLink href="/docs/webhooks#otp-sent">otp.sent webhook</DocLink>.</>],
        ]}
      />
      <P>
        On sign-in the user types one value; sooauth detects whether it is an email (contains <C>@</C>), a phone number (starts with <C>+</C> or is 7+ digits) or a username. An identifier type that is not enabled is rejected with <C>invalid_credentials</C>.
      </P>

      <H2 id="primary-method">Primary method</H2>
      <Table
        head={["Mode", "What users see", "Endpoints that work"]}
        rows={[
          [<C key="p">password</C>, "Identifier + password.", <><C>/auth/sign-in</C>. Code endpoints return <C>403 otp_disabled</C>.</>],
          [<C key="o">otp</C>, "Identifier, then a 6-digit code. No passwords. New users are created on their first code sign-in.", <><C>/auth/otp/send</C>, <C>/auth/otp/verify</C>. Password sign-in returns <C>403 password_login_disabled</C>.</>],
          [<C key="b">both</C>, "Password form with a “Use a one-time code instead” switch. The user picks one.", "All of the above."],
        ]}
      />
      <Note tone="warn">
        <strong>“Password or code” is not two-step verification.</strong> The user chooses either one; one factor is enough. To ask for a password <em>and</em> a code, keep the password mode and turn on <DocLink href="#two-step">two-step verification</DocLink>.
      </Note>

      <H2 id="two-step">Two-step verification (2FA)</H2>
      <P>A second step after a correct password. Pick a policy and the methods users may use.</P>
      <Table
        head={["Policy", "Effect"]}
        rows={[
          [<C key="off">off</C>, "Password only."],
          [<C key="opt">optional</C>, "Users who set up an authenticator app are asked for a code. Others sign in with the password."],
          [<C key="req">required</C>, "Every password sign-in needs a second step. Users without an authenticator get an email code, or set one up on the spot."],
        ]}
      />
      <P>
        Methods: <C>totp</C> (Google Authenticator, 1Password, Authy…) and <C>email_otp</C> (6-digit code by email). Details and API flow: <DocLink href="/docs/two-factor">Two-step verification</DocLink>.
      </P>

      <H2 id="password-and-verification">Passwords and email verification</H2>
      <P>
        In <DashboardPath>Passwords &amp; verification</DashboardPath> set the minimum length and character rules, and whether new users must verify their email (by link, 6-digit code or both) before signing in. Users who sign in with a one-time code are verified automatically.
      </P>

      <H2 id="config-api">Reading the settings from your app</H2>
      <P>Custom UIs read the active settings from the public widget config:</P>
      <Code title="GET /v1/widget/config?client_id=app_…">{`curl "${authUrl}/v1/widget/config?client_id=app_123"

{
  "auth_config": {
    "allowed_identifiers": ["email", "phone"],
    "primary_auth_mode": "both",
    "mfa_policy": "optional",
    "mfa_methods": ["totp", "email_otp"],
    "phone_otp_via_webhook": true
  },
  "password_policy": { "min_length": 8, "hint": "At least 8 characters" },
  "registration_schema": [ … ],
  "providers": ["google"]
}`}</Code>

      <H3>Error codes</H3>
      <Table
        head={["Error", "Status", "Meaning"]}
        rows={[
          [<C key="1">invalid_credentials</C>, "401", "Wrong identifier or password, or the identifier type is disabled."],
          [<C key="2">password_login_disabled</C>, "403", "The project signs in with codes only."],
          [<C key="3">otp_disabled</C>, "403", "The project does not allow code sign-in."],
          [<C key="4">mfa_required</C>, "401", <>Password was correct; finish with <DocLink href="/docs/two-factor#api">/auth/mfa/verify</DocLink>.</>],
          [<C key="5">email_not_verified</C>, "403", "Email verification is required first."],
          [<C key="6">user_disabled</C>, "403", "The project owner disabled this user."],
        ]}
      />
      <List>
        <li>One-time codes: <DocLink href="/docs/otp">One-time codes (OTP)</DocLink></li>
        <li>Extra sign-up fields: <DocLink href="/docs/registration-fields">Sign-up fields</DocLink></li>
      </List>
    </article>
  );
}
