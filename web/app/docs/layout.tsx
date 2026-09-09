import type { ReactNode } from "react";
import Link from "next/link";
import { DocsNav } from "../../components/docs-nav";
import { buttonClass } from "../../components/ui/button";
import { Eyebrow } from "../../components/ui/layout";
import { ThemeToggle } from "../../components/ui/theme-toggle";
import { BrandLogo } from "../../components/brand-logo";
import { adminUrl, signInUrl } from "../../lib/site";

export default function DocsLayout({ children }: { children: ReactNode }) {
  return (
    <div className="min-h-dvh bg-bg">
      <header className="sticky top-0 z-100 border-b border-line bg-bg/90 backdrop-blur-md">
        <div className="container-swiss flex h-16 items-center justify-between gap-4">
          <Link
            href="/"
            className="text-fg"
          >
            <BrandLogo />
          </Link>
          <nav aria-label="Docs header" className="flex items-center gap-2">
            <Link href={adminUrl} className={buttonClass("ghost", "sm")}>
              Dashboard
            </Link>
            <ThemeToggle />
            <a href={signInUrl} className={buttonClass("primary", "sm")}>
              Sign in
            </a>
          </nav>
        </div>
      </header>
      <div className="container-swiss grid gap-10 py-14 md:grid-cols-[220px_minmax(0,1fr)]">
        <aside className="md:sticky md:top-24 md:self-start">
          <Eyebrow className="mb-3">Documentation</Eyebrow>
          <DocsNav />
          <p className="mt-8 border-t border-line pt-6">
            <Link
              href={adminUrl}
              className="font-sans text-sm font-medium text-fg underline underline-offset-2"
            >
              Open dashboard →
            </Link>
          </p>
        </aside>
        <main id="main" className="min-w-0">
          {children}
        </main>
      </div>
    </div>
  );
}
