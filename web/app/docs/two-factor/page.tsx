import type { Metadata } from "next";
import { authUrl } from "../../../lib/site";
import { C, Code, DashboardPath, DocLink, DocTitle, H2, H3, List, Note, P, Steps, Table } from "../../../components/docs-kit";

export const metadata: Metadata = {
  title: "Two-step verification (2FA)",
  description: "Ask for an authenticator app code or an emailed code after the password.",
};

export default function TwoFactorPage() {
  return (
    <article>
      <DocTitle lead={<>Two-step verification asks for a second code after a correct password. Configure it in <DashboardPath>Sign-in methods → Two-step verification</DashboardPath>.</>}>
        Two-step verification (2FA)
      </DocTitle>

      <H2 id="policy">Policy and methods</H2>
      <Table
        head={["Setting", "Values"]}
        rows={[
          [<C key="p">mfa_policy</C>, <><C>off</C> · <C>optional</C> (only users with an authenticator) · <C>required</C> (everyone)</>],
          [<C key="m">mfa_methods</C>, <><C>totp</C> authenticator app · <C>email_otp</C> code by email</>],
        ]}
      />
      <P>What a user is asked for after the password:</P>
      <Table
        head={["Situation", "Second step (method)"]}
        rows={[
          ["User has an authenticator app", <><C>totp</C> — the app code or a backup code. Can switch to an email code if <C>email_otp</C> is enabled.</>],
          ["Required, no authenticator, email enabled", <><C>email_otp</C> — a 6-digit code is emailed.</>],
          ["Required, authenticator only, not set up yet", <><C>totp_setup</C> — the user scans a QR code and confirms a code; backup codes are returned.</>],
          ["Optional, no authenticator", "No second step."],
        ]}
      />
      <Note>2FA applies to password sign-ins. <DocLink href="/docs/otp">Code sign-in</DocLink> and social sign-in are not asked again.</Note>

      <H2 id="api">API flow</H2>
      <Steps>
        <li>Sign in with the password as usual.</li>
        <li>If a second step is needed you get <C>401 mfa_required</C> with a <C>challenge</C> (valid 5 minutes) and the <C>method</C>.</li>
        <li>Collect the code and call <C>/auth/mfa/verify</C>. Success returns the tokens.</li>
      </Steps>
      <Code title="1. Sign in">{`POST ${authUrl}/auth/sign-in
{ "client_id": "app_123", "identifier": "ada@example.com", "password": "…" }

401 { "error": "mfa_required", "challenge": "c_…", "method": "totp", "methods": ["totp", "email_otp"] }`}</Code>
      <Code title="2. Verify">{`POST ${authUrl}/auth/mfa/verify
{ "challenge": "c_…", "code": "123456", "method": "totp" }

200 { "access_token": "…", "refresh_token": "…", "expires_in": 900 }`}</Code>
      <H3>Switching to an email code / resending</H3>
      <Code>{`POST ${authUrl}/auth/mfa/resend
{ "challenge": "c_…" }    → then verify with "method": "email_otp"`}</Code>
      <H3>First-time setup (method totp_setup)</H3>
      <Code>{`POST ${authUrl}/auth/mfa/setup
{ "challenge": "c_…" }

200 { "secret": "JBSW…", "otpauth_url": "otpauth://totp/…", "qr_code": "data:image/png;base64,…" }

POST ${authUrl}/auth/mfa/verify
{ "challenge": "c_…", "code": "123456", "method": "totp_setup" }

200 { "access_token": "…", "backup_codes": ["A1B2C3D4E5", …] }`}</Code>
      <P>Show the backup codes once and ask the user to store them. Each works one time instead of an app code.</P>
      <Table
        head={["Error", "Status", "Meaning"]}
        rows={[
          [<C key="1">invalid_mfa_code</C>, "401", "Wrong code, or the challenge expired/was used."],
          [<C key="2">too_many_attempts</C>, "429", "Too many wrong email codes; sign in again."],
          [<C key="3">rate_limited</C>, "429", "Too many tries for this challenge."],
        ]}
      />

      <H2 id="self-service">Let users manage their authenticator</H2>
      <P>With the user&apos;s access token (<C>Authorization: Bearer …</C>):</P>
      <Code>{`GET  ${authUrl}/v1/me/mfa                 → { "policy", "methods", "totp_enabled" }
POST ${authUrl}/v1/me/mfa/totp/start      → { "secret", "otpauth_url", "qr_code" }
POST ${authUrl}/v1/me/mfa/totp/confirm    { "code" } → { "enabled": true, "backup_codes": [...] }
POST ${authUrl}/v1/me/mfa/totp/disable    { "code" } → { "enabled": false }`}</Code>
      <List>
        <li>Lost phone? The project owner can press <DashboardPath>Reset 2FA</DashboardPath>. If 2FA is required the user sets it up again on the next sign-in.</li>
        <li>The hosted sign-in page and the embed widget already show all of these screens.</li>
      </List>
    </article>
  );
}
