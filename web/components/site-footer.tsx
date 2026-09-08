import Link from "next/link";
import { ArrowRight, Github } from "lucide-react";
import { companyAddress, companyLegalName, legalLinks, sooappsUrl } from "../lib/legal";
import {
  contactEmail,
  seoLinks,
  signInUrl,
  signUpUrl,
  socialLinks,
} from "../lib/site";
import { buttonClass } from "./ui/button";
import { Container } from "./ui/layout";
import { BrandLogo } from "./brand-logo";

const footLink =
  "inline-flex cursor-pointer items-center font-sans text-sm text-fg-muted transition-colors duration-200 hover:text-fg";

export function SiteFooter() {
  const year = new Date().getFullYear();

  return (
    <footer className="border-t border-line">
      <Container className="py-16">
        <div className="grid gap-10 md:grid-cols-[1.4fr_repeat(3,minmax(0,1fr))]">
          <div>
            <BrandLogo />
            <p className="mt-3 max-w-[32ch] text-[15px] leading-[1.55] text-fg-muted">
              Open-source authentication you can own. Self-host Sooauth or let
              Sooapps run it for you.
            </p>
            <address className="mt-5 max-w-[38ch] not-italic text-[12px] leading-[1.6] text-fg-muted">
              <strong className="font-medium text-fg">{companyLegalName}</strong>
              <br />
              {companyAddress}
              <br />
              <a href="mailto:info@sooapps.com" className="underline underline-offset-2 hover:text-fg">
                info@sooapps.com
              </a>
            </address>
            <a
              href={signUpUrl}
              className={buttonClass("primary", "md", "mt-6")}
            >
              Get started
              <ArrowRight size={16} aria-hidden />
            </a>
          </div>

          <div>
            <p className="font-mono text-xs uppercase tracking-[0.14em] text-fg-muted">
              Product
            </p>
            <ul className="mt-4 grid gap-2">
              <li>
                <a href="#how" className={footLink}>
                  How it works
                </a>
              </li>
              <li>
                <a href="#pricing" className={footLink}>
                  Pricing
                </a>
              </li>
              <li>
                <a href="#faq" className={footLink}>
                  FAQ
                </a>
              </li>
              <li>
                <Link href="/blog" className={footLink}>
                  Blog
                </Link>
              </li>
            </ul>
          </div>

          <div>
            <p className="font-mono text-xs uppercase tracking-[0.14em] text-fg-muted">
              Docs
            </p>
            <ul className="mt-4 grid gap-2">
              {seoLinks.map((item) => (
                <li key={item.href}>
                  <Link href={item.href} className={footLink}>
                    {item.label}
                  </Link>
                </li>
              ))}
            </ul>
          </div>

          <div>
            <p className="font-mono text-xs uppercase tracking-[0.14em] text-fg-muted">
              Account
            </p>
            <ul className="mt-4 grid gap-2">
              <li>
                <a href={signInUrl} className={footLink}>
                  Sign in
                </a>
              </li>
              <li>
                <a href={signUpUrl} className={footLink}>
                  Create account
                </a>
              </li>
              <li>
                <a href={`mailto:${contactEmail}`} className={footLink}>
                  Contact
                </a>
              </li>
            </ul>
          </div>
        </div>

        <div className="mt-12 flex flex-wrap items-center justify-between gap-4 border-t border-line pt-6">
          <p className="font-mono text-xs text-fg-muted">© {year} Sooapps · Sooauth</p>

          <div className="flex items-center gap-4">
            {socialLinks.map((s) => (
              <a
                key={s.href}
                href={s.href}
                target="_blank"
                rel="noopener noreferrer"
                aria-label={s.label}
                className="text-fg-muted transition-colors hover:text-fg"
              >
                <Github size={16} aria-hidden />
              </a>
            ))}
            <a
              href={socialLinks[0]?.href}
              target="_blank"
              rel="noopener noreferrer"
              className="font-mono text-xs text-fg-muted hover:text-fg"
            >
              Source ↗
            </a>
            <a
              href={sooappsUrl}
              target="_blank"
              rel="noopener noreferrer"
              className="font-mono text-xs text-fg-muted hover:text-fg"
            >
              A Sooapps product ↗
            </a>
          </div>
        </div>

        <ul className="mt-4 flex flex-wrap gap-x-4 gap-y-2">
          {legalLinks.map((item) => (
            <li key={item.href}>
              <a
                href={item.href}
                className="font-sans text-xs text-fg-muted transition-colors hover:text-fg"
              >
                {item.label}
              </a>
            </li>
          ))}
        </ul>
      </Container>
    </footer>
  );
}
