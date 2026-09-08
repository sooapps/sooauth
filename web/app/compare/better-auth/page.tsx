import type { Metadata } from "next";
import { ComparisonPage } from "../../../components/comparison-page";

export const metadata: Metadata = {
  title: "Sooauth vs Better Auth",
  description:
    "Compare Sooauth and Better Auth for AI builders and SaaS teams. Decoupled standard OIDC vs database-coupled auth adapters.",
  alternates: { canonical: "/compare/better-auth" },
};

export default function BetterAuthComparisonPage() {
  return (
    <ComparisonPage
      competitor="Better Auth"
      description="Better Auth embeds user and session tables directly into your application database with ORM adapters. Sooauth keeps identity completely decoupled via standard OIDC, keeping your app database clean and preventing AI agents from breaking migrations."
      bestFor="Vibecoders and teams that want clean database separation without ongoing schema migrations"
      points={[
        "Zero database migrations or schema upkeep in your application database",
        "Pure OpenID Connect (OIDC) standard—compatible with any language, not just TypeScript/Node",
        "Hosted sign-in, passkeys, and security center ready out of the box",
        "Single-command self-hosting with a ~25MB Go binary or hosted cloud at auth.sooauth.com",
      ]}
    />
  );
}
