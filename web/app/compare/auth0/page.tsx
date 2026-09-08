import type { Metadata } from "next";
import { ComparisonPage } from "../../../components/comparison-page";

export const metadata: Metadata = {
  title: "Sooauth vs Auth0",
  description:
    "Compare Sooauth and Auth0 for indie SaaS teams. Standard OIDC, passkeys, hosted auth, and a self-hosting path from Sooapps.",
  alternates: { canonical: "/compare/auth0" },
};

export default function Auth0ComparisonPage() {
  return (
    <ComparisonPage
      competitor="Auth0"
      description="Auth0 is a mature managed identity platform. Sooauth is for teams that want a simpler auth layer with standard OIDC, an open-source Community Edition, and the option to run it themselves."
      bestFor="Indie SaaS and product studios"
      points={[
        "Open-source Community Edition for self-hosting",
        "Standard OIDC integration instead of application lock-in",
        "Hosted email, passkeys, Google, and GitHub flows",
        "Hosted Sooauth at auth.sooauth.com when you want managed operations",
      ]}
    />
  );
}
