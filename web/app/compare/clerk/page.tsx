import type { Metadata } from "next";
import { ComparisonPage } from "../../../components/comparison-page";

export const metadata: Metadata = {
  title: "Sooauth vs Clerk",
  description:
    "Compare Sooauth and Clerk for SaaS teams. Use standard OIDC, self-host authentication, or choose hosted Sooauth at auth.sooauth.com.",
  alternates: { canonical: "/compare/clerk" },
};

export default function ClerkComparisonPage() {
  return (
    <ComparisonPage
      competitor="Clerk"
      description="Clerk is a polished hosted developer experience. Sooauth is a fit for teams that want hosted auth convenience without giving up a self-hosting option or a standard OIDC boundary."
      bestFor="Teams that value ownership and deployment flexibility"
      points={[
        "Self-host the Community Edition on your own infrastructure",
        "Use any OIDC-compatible framework or SDK",
        "Keep provider-specific social credentials out of your app",
        "Use hosted Sooauth at auth.sooauth.com when operations are not your priority",
      ]}
    />
  );
}
