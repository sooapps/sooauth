import type { Metadata } from "next";
import { authUrl } from "../../../lib/site";
import { C, Code, DashboardPath, DocLink, DocTitle, H2, List, P, Steps, Table } from "../../../components/docs-kit";

export const metadata: Metadata = {
  title: "Social login",
  description: "Let users sign in with Google or GitHub.",
};

export default function SocialLoginPage() {
  return (
    <article>
      <DocTitle lead={<>Turn on Google or GitHub in <DashboardPath>Social login</DashboardPath>. Social buttons then appear on the hosted page and in the widget. Your app never stores Google or GitHub secrets.</>}>
        Social login
      </DocTitle>

      <H2 id="modes">Shared app or your own app</H2>
      <Table
        head={["Option", "Setup", "What users see"]}
        rows={[
          ["sooauth platform app", "Flip the switch. Nothing else.", "Google's account picker shows sooauth.com."],
          ["Your own OAuth app", "Paste your client ID and secret, and add the callback URL shown in the dashboard to your Google/GitHub app.", "Your app name and logo."],
        ]}
      />
      <P>Step-by-step for Google Cloud Console: <DocLink href="/docs/google-oauth">Google login</DocLink>.</P>

      <H2 id="custom-ui">In a custom UI</H2>
      <Steps>
        <li>Link the button to the start URL with your <C>client_id</C> and the page to return to.</li>
        <li>sooauth sends the user back with <C>?widget_code=…</C>.</li>
        <li>Exchange it for a token.</li>
      </Steps>
      <Code>{`${authUrl}/auth/social/google/start?client_id=app_123&tenant_id=<tenant_id>&return_to=https://myapp.com/login

POST ${authUrl}/v1/widget/exchange
{ "code": "<widget_code>", "client_id": "app_123" }
→ { "access_token": "…", "email": "…", "expires_in": 900 }`}</Code>
      <List>
        <li>The <C>return_to</C> page must be one of the project&apos;s redirect URIs (<DashboardPath>Keys &amp; URLs</DashboardPath>).</li>
        <li>With OIDC (hosted page) you don&apos;t need any of this — the buttons are already on the sign-in page.</li>
        <li>If a user already has a password account in this project with the same email, their first Google/GitHub sign-in is linked to that account.</li>
      </List>
    </article>
  );
}
