import Link from "next/link";
import { adminUrl, signInUrl } from "../lib/site";
import { buttonClass } from "./ui/button";
import { Container } from "./ui/layout";
import { ThemeToggle } from "./ui/theme-toggle";
import { BrandLogo } from "./brand-logo";

export function SiteHeader() {
  return (
    <header className="sticky top-0 z-100 border-b border-line bg-bg/90 backdrop-blur-md">
      <Container className="flex h-16 items-center justify-between gap-6">
        <Link
          href="/"
          className="text-fg"
        >
          <BrandLogo />
        </Link>
        <nav
          aria-label="Marketing"
          className="flex flex-1 items-center justify-end gap-1 sm:gap-2"
        >
          <a
            href="#how"
            className={buttonClass("ghost", "sm", "hidden sm:inline-flex")}
          >
            How it works
          </a>
          <a
            href="#pricing"
            className={buttonClass("ghost", "sm", "hidden sm:inline-flex")}
          >
            Pricing
          </a>
          <Link href="/docs" className={buttonClass("ghost", "sm")}>
            Docs
          </Link>
          <Link href="/playground" className={buttonClass("ghost", "sm")}>
            Playground
          </Link>
          <Link
            href="/blog"
            className={buttonClass("ghost", "sm", "hidden sm:inline-flex")}
          >
            Blog
          </Link>
          <a
            href={adminUrl}
            className={buttonClass("ghost", "sm", "hidden sm:inline-flex")}
          >
            Dashboard
          </a>
          <ThemeToggle />
          <a href={signInUrl} className={buttonClass("primary", "sm")}>
            Sign in
          </a>
        </nav>
      </Container>
    </header>
  );
}
