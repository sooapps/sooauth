import type { Metadata } from "next";
import { authUrl } from "../../../lib/site";
import { C, Code, DashboardPath, DocLink, DocTitle, H2, List, Note, P, Steps, Table } from "../../../components/docs-kit";

export const metadata: Metadata = {
  title: "One-time codes (OTP)",
  description: "Passwordless sign-in with 6-digit codes by email or SMS.",
};

export default function OtpPage() {
  return (
    <article>
      <DocTitle lead={<>Let users sign in with a 6-digit code instead of a password. Turn it on in <DashboardPath>Authentication → Sign-in methods</DashboardPath> by choosing <C>otp</C> (codes only) or <C>both</C> (password or code).</>}>
        One-time codes (OTP)
      </DocTitle>

      <H2 id="how">How it works</H2>
      <Steps>
        <li>The user enters an email or phone number and asks for a code.</li>
        <li>sooauth sends a 6-digit code. It is valid for <strong>5 minutes</strong>.</li>
        <li>The user enters the code. If no account exists yet, one is created (see <DocLink href="#new-users">new users</DocLink>).</li>
        <li>Your app receives an access token, like any other sign-in.</li>
      </Steps>
      <P>The hosted sign-in page and the embed widget do all of this for you. The rest of this page is for custom UIs.</P>

      <H2 id="delivery">Delivery</H2>
      <Table
        head={["Identifier", "How the code is sent"]}
        rows={[
          ["Email", <>Through the project&apos;s email settings (<DashboardPath>Branding → Email</DashboardPath>), or sooauth&apos;s mail server if you haven&apos;t set one up.</>],
          ["Phone", <>sooauth does not send SMS itself. Enable <DashboardPath>“Send phone codes through my webhook”</DashboardPath> and add a webhook for <C>otp.sent</C>; the payload includes the code and your server sends the SMS. See <DocLink href="/docs/webhooks#otp-sent">otp.sent</DocLink>.</>],
          ["Username", "A username has no inbox. The code goes to the account's email (or phone). Unknown usernames get the same response, so usernames cannot be probed."],
        ]}
      />

      <H2 id="limits">Limits</H2>
      <List>
        <li>A code expires after 5 minutes. Sending a new code replaces the old one.</li>
        <li>5 wrong codes discard the code (<C>429 too_many_attempts</C>); request a new one.</li>
        <li>Send is limited to 5 codes per identifier per 10 minutes and 20 per IP per minute (<C>429 rate_limited</C>).</li>
        <li>Codes are random (crypto-grade) and compared in constant time.</li>
      </List>

      <H2 id="api">API</H2>
      <Code title="1. Send a code">{`POST ${authUrl}/auth/otp/send
Content-Type: application/json

{ "client_id": "app_123", "identifier": "ada@example.com" }

200 { "message": "If the account can receive codes, a code was sent.", "expires_in": 300 }`}</Code>
      <Code title="2. Verify it">{`POST ${authUrl}/auth/otp/verify
Content-Type: application/json

{ "client_id": "app_123", "identifier": "ada@example.com", "code": "123456" }

200 {
  "id": "…", "email": "ada@example.com",
  "access_token": "…", "refresh_token": "…", "expires_in": 900, "token_type": "Bearer"
}`}</Code>
      <Table
        head={["Error", "Status", "What to do"]}
        rows={[
          [<C key="1">otp_disabled</C>, "403", "The project's mode is password only."],
          [<C key="2">identifier_not_allowed</C>, "400", "That identifier type is not enabled."],
          [<C key="3">phone_otp_unavailable</C>, "400", "Phone codes need the webhook option."],
          [<C key="4">invalid_code</C>, "400", "Wrong or expired code. Let the user retry."],
          [<C key="5">too_many_attempts</C>, "429", "Ask for a new code."],
          [<C key="6">registration_required</C>, "422", "New user must send the listed sign-up fields (below)."],
          [<C key="7">otp_delivery_failed</C>, "502", "The email could not be sent. Check email settings."],
        ]}
      />

      <H2 id="new-users">New users and sign-up fields</H2>
      <P>
        The first successful code for an unknown email or phone creates the account (marked verified). If the project has required <DocLink href="/docs/registration-fields">sign-up fields</DocLink>, verify returns <C>422</C> with the missing field keys. The code stays valid: show the fields and call verify again with <C>metadata</C>.
      </P>
      <Code>{`422 { "error": "registration_required", "fields": ["company"] }

POST /auth/otp/verify
{ "client_id": "app_123", "identifier": "ada@example.com", "code": "123456",
  "metadata": { "company": "Acme" } }`}</Code>
      <Note>Code sign-in already proves access to the inbox or phone, so <DocLink href="/docs/two-factor">two-step verification</DocLink> is not asked after it.</Note>
    </article>
  );
}
