import type { Metadata } from "next";
import { authUrl } from "../../../lib/site";
import { C, Code, DashboardPath, DocLink, DocTitle, H2, List, Note, P, Steps, Table } from "../../../components/docs-kit";

export const metadata: Metadata = {
  title: "Sign-up fields",
  description: "Collect extra information at sign-up with the form builder, including dropdowns loaded from your API.",
};

export default function RegistrationFieldsPage() {
  return (
    <article>
      <DocTitle lead={<>Ask for more than an email and password when users sign up — a company name, a role, a team picked from your own API. Build the form in <DashboardPath>Sign-up fields</DashboardPath>.</>}>
        Sign-up fields
      </DocTitle>

      <H2 id="types">Field types</H2>
      <Table
        head={["Type", "Value sent", "Notes"]}
        rows={[
          [<C key="t">text</C>, "string", "Single line."],
          [<C key="ta">textarea</C>, "string", "Multi-line."],
          [<C key="n">number</C>, "number", ""],
          [<C key="c">checkbox</C>, "true / false", "Use with Required for “I accept the terms”."],
          [<C key="s">select</C>, "option value", <>Options from a fixed list or <DocLink href="#api-options">your API</DocLink>.</>],
        ]}
      />
      <P>Each field has a <strong>key</strong> (e.g. <C>company</C>), a label, optional placeholder and help text, and a <strong>required</strong> flag. Required fields are enforced by the server, not only by the form.</P>

      <H2 id="where">Where the fields appear</H2>
      <List>
        <li><strong>Hosted sign-up page and embed widget:</strong> rendered automatically, nothing to code.</li>
        <li><strong>Custom UI:</strong> read <C>registration_schema</C> from <C>GET /v1/widget/config</C> and render the fields yourself.</li>
        <li><strong>Code-only sign-up:</strong> <C>/auth/otp/verify</C> asks for missing required fields with <C>422 registration_required</C>. See <DocLink href="/docs/otp#new-users">One-time codes</DocLink>.</li>
      </List>

      <H2 id="send">Sending values (custom UI)</H2>
      <Code title="POST /auth/sign-up">{`POST ${authUrl}/auth/sign-up
{
  "client_id": "app_123",
  "email": "ada@example.com",
  "password": "…",
  "metadata": { "company": "Acme", "role": "admin", "terms": true }
}`}</Code>
      <P>A missing required field returns <C>400</C> with <C>field_required:&lt;label&gt;</C>.</P>

      <H2 id="read">Reading values in your app</H2>
      <P>The values are stored on the user and returned as <C>custom_claims</C>:</P>
      <Code title="GET /oauth/userinfo">{`{
  "sub": "f8a9…",
  "email": "ada@example.com",
  "email_verified": true,
  "custom_claims": { "company": "Acme", "role": "admin", "terms": true }
}`}</Code>
      <P>They are also visible in <DashboardPath>Users</DashboardPath> and included as <C>user.Metadata</C> in <DocLink href="/docs/webhooks">webhook</DocLink> payloads.</P>

      <H2 id="api-options">Dropdown options from your API</H2>
      <P>Choose <strong>Loaded from an API</strong> for a select field. sooauth calls your URL from its server, so secret headers never reach the browser.</P>
      <Table
        head={["Setting", "Example", "Meaning"]}
        rows={[
          ["API URL", <C key="u">https://api.example.com/teams</C>, "GET request made by sooauth."],
          ["Headers (JSON)", <C key="h">{`{"Authorization": "Bearer …"}`}</C>, "Sent with the request; stored server-side."],
          ["Items path", <C key="i">data.items</C>, "Where the list is in the response. Empty if the response is a list."],
          ["Label key / Value key", <><C>name</C> / <C>id</C></>, "Which properties become the option label and value."],
          ["Cache", "300", "Seconds to cache the options."],
        ]}
      />
      <Code title="Your API responds">{`{ "data": { "items": [ { "id": "t1", "name": "Red team" }, { "id": "t2", "name": "Blue team" } ] } }`}</Code>
      <Steps>
        <li>Press <strong>Test API</strong> in the dashboard to check the mapping.</li>
        <li>Widgets fetch options from <C>GET /v1/widget/fields/&lt;key&gt;/options?client_id=…</C> (the <C>options_url</C> in the widget config).</li>
      </Steps>
      <Note>Changing a field&apos;s key does not rename values already stored on existing users.</Note>
    </article>
  );
}
