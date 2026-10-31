import type { Metadata } from "next";
import { C, DashboardPath, DocLink, DocTitle, H2, List, Note, P, Table } from "../../../components/docs-kit";

export const metadata: Metadata = {
  title: "Users and sessions",
  description: "Disable or delete users, revoke sessions, and how dashboard accounts stay separate from app users.",
};

export default function UsersAndSessionsPage() {
  return (
    <article>
      <DocTitle lead={<>Manage the people who signed up to your project in <DashboardPath>Users</DashboardPath>.</>}>Users and sessions</DocTitle>

      <H2 id="accounts">Dashboard account vs. app users</H2>
      <P>
        Your sooauth dashboard account and the users of your projects are completely separate, even when they use the same email address. Signing up to your own app with your dashboard email creates a new, independent user in that project. Disabling or deleting that user never signs you out of the dashboard, and your dashboard account never appears in a project&apos;s user list.
      </P>
      <Note>To test your app, use the <DocLink href="/docs/getting-started#test">live test</DocLink> in the dashboard. It always shows the project&apos;s sign-in page, even if you are signed in to the dashboard.</Note>

      <H2 id="disable-delete">Disable or delete</H2>
      <Table
        head={["Action", "Effect", "Undo"]}
        rows={[
          ["Disable", "Sign-in is blocked. All sessions and refresh tokens are revoked immediately; existing access tokens stop working.", "Enable"],
          ["Delete", <>The user, passwords, sessions, refresh tokens, linked social accounts, 2FA and pending codes are removed. A <C key="d">user.deleted</C> webhook is sent.</>, "Not possible"],
          ["Reset 2FA", "Removes the user's authenticator app. They set it up again if 2FA is required.", "—"],
        ]}
      />
      <P>Delete asks you to type the user&apos;s email to confirm.</P>

      <H2 id="sessions">Sessions</H2>
      <P>
        <DashboardPath>Sessions</DashboardPath> lists signed-in sessions of this project&apos;s users. Revoking one signs that device out at its next request. Your own dashboard sessions are not listed here; manage them in the Security Center.
      </P>
      <List>
        <li>Refresh tokens rotate on every use; reusing an old one revokes the whole family.</li>
      </List>
    </article>
  );
}
