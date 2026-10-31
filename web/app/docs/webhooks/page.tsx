import type { Metadata } from "next";
import { C, Code, DashboardPath, DocTitle, H2, List, P, Table } from "../../../components/docs-kit";

export const metadata: Metadata = {
  title: "Webhooks",
  description: "Receive signed HTTP requests when users sign up, sign in or are deleted.",
};

export default function WebhooksPage() {
  return (
    <article>
      <DocTitle lead={<>Add an endpoint in <DashboardPath>Webhooks</DashboardPath> and pick events. sooauth POSTs JSON to your URL and signs every request.</>}>Webhooks</DocTitle>

      <H2 id="events">Events</H2>
      <Table
        head={["Event", "When"]}
        rows={[
          [<C key="1">user.created</C>, "A user signed up (password, code or social)."],
          [<C key="2">user.login</C>, "A user signed in."],
          [<C key="3">user.login_failed</C>, "Wrong password."],
          [<C key="4">user.email_verified</C>, "Email verified."],
          [<C key="5">user.deleted</C>, "A user was deleted from the dashboard."],
          [<C key="6">password.reset</C>, "Password reset completed."],
          [<C key="7">session.revoked</C>, "A user signed out or a session was revoked."],
          [<C key="8" >otp.sent</C>, <span key="8d" id="otp-sent">A one-time code was sent. For phone numbers the payload contains <C>code</C> so your server can send the SMS.</span>],
        ]}
      />

      <H2 id="payload">Payload</H2>
      <Code>{`POST https://myapp.com/api/sooauth-webhook
X-Sooauth-Event: user.created
X-Sooauth-Delivery: 5c0e…
X-Sooauth-Signature: t=1760550000,v1=8f3a…

{
  "id": "e1b2…",
  "type": "user.created",
  "created_at": "2026-10-15T12:00:00Z",
  "tenant_id": "ec43…",
  "user": { "ID": "f8a9…", "Email": "ada@example.com", "Username": null, "Phone": null, "Metadata": { "company": "Acme" } },
  "data": { "method": "email", "ip": "203.0.113.4" }
}`}</Code>
      <Code title="otp.sent for a phone number">{`{ "type": "otp.sent", "data": { "identifier": "+905551234567", "channel": "phone", "purpose": "sign_in", "code": "123456", "expires_in": 300 } }`}</Code>

      <H2 id="verify">Verifying the signature</H2>
      <P>The secret is shown once when you create the endpoint. Compute HMAC-SHA256 over <C>{"<t>.<raw body>"}</C> and compare with <C>v1</C>. Reject requests older than 5 minutes.</P>
      <Code title="Node.js">{`import crypto from "node:crypto";

export function verifySooauth(rawBody, header, secret) {
  const parts = Object.fromEntries(header.split(",").map((p) => p.split("=")));
  const expected = crypto.createHmac("sha256", secret).update(\`\${parts.t}.\${rawBody}\`).digest("hex");
  const fresh = Math.abs(Date.now() / 1000 - Number(parts.t)) < 300;
  return fresh && crypto.timingSafeEqual(Buffer.from(expected), Buffer.from(parts.v1));
}`}</Code>
      <List>
        <li>Respond with any 2xx quickly. A failed delivery is retried up to 3 times within about a second, then recorded as failed — design your handler to be idempotent using <C>X-Sooauth-Delivery</C>.</li>
        <li>Every delivery and response is visible under the endpoint&apos;s <strong>View log</strong>.</li>
        <li><strong>Send test</strong> posts a <C>webhook.test</C> event.</li>
      </List>
    </article>
  );
}
