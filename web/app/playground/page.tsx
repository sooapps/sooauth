import type { Metadata } from "next";
import Link from "next/link";
import { SiteFooter } from "../../components/site-footer";
import { SiteHeader } from "../../components/site-header";
import { Container, Eyebrow } from "../../components/ui/layout";
import { EmbedWidgetPlayground } from "../../components/embed-widget-playground";

export const metadata: Metadata = {
  title: "Embed Widget Playground",
  description:
    "Design, customize, and test the Sooauth embed.js drop-in auth widget with live preview and code export.",
  alternates: { canonical: "/playground" },
};

export default function PlaygroundPage() {
  return (
    <>
      <SiteHeader />
      <main id="main" className="pb-24">
        <Container>
          <div className="max-w-[720px] pt-12 pb-8">
            <Eyebrow>Interactive Preview</Eyebrow>
            <h1 className="mt-4 font-sans text-[clamp(2.2rem,4.5vw,3.5rem)] font-bold leading-[1.05] tracking-[-0.03em] text-fg text-balance">
              Test your embed widget before shipping.
            </h1>
            <p className="mt-4 text-[17px] leading-[1.6] text-fg-muted text-pretty">
              Customize social providers, brand identity, accent colors, and password policy rules. See the live widget update instantly and copy ready-to-paste integration snippets.
            </p>
          </div>

          <div className="mt-4">
            <EmbedWidgetPlayground />
          </div>

          <div className="mt-12 rounded-xl border border-line bg-surface p-6 flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
            <div>
              <h3 className="font-semibold text-fg text-sm">Need complete control over HTML &amp; CSS?</h3>
              <p className="text-xs text-fg-muted mt-1">
                Build your own bespoke auth forms with the Sooauth REST &amp; OIDC API or AI setup prompts.
              </p>
            </div>
            <div className="flex items-center gap-3">
              <Link
                href="/docs/embed-widget"
                className="inline-flex items-center rounded-md border border-line bg-field px-3.5 py-2 text-xs font-semibold text-fg hover:border-fg transition active:scale-[0.96]"
              >
                Embed Docs
              </Link>
              <Link
                href="/docs/custom-auth-ui"
                className="inline-flex items-center rounded-md bg-fg px-3.5 py-2 text-xs font-semibold text-bg hover:opacity-90 transition active:scale-[0.96]"
              >
                Custom Auth UI
              </Link>
            </div>
          </div>
        </Container>
      </main>
      <SiteFooter />
    </>
  );
}
