"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { cn } from "../lib/cn";

const links = [
  { href: "/docs", label: "Overview", exact: true },
  { href: "/docs/getting-started", label: "Getting started" },
  { href: "/docs/integrate-oidc", label: "Connect your app (OIDC)" },
  { href: "/docs/custom-auth-ui", label: "Custom auth UI" },
  { href: "/docs/embed-widget", label: "Embed widget" },
  { href: "/playground", label: "Widget playground ↗" },
  { href: "/docs/google-oauth", label: "Google login" },
  { href: "/docs/admin", label: "Dashboard" },
  { href: "/docs/self-host", label: "Self-host Sooauth" },
];

export function DocsNav() {
  const pathname = usePathname();
  return (
    <nav aria-label="Documentation" className="grid gap-1">
      {links.map((item) => {
        const current = item.exact
          ? pathname === item.href
          : pathname === item.href || pathname.startsWith(`${item.href}/`);
        return (
          <Link
            key={item.href}
            href={item.href}
            aria-current={current ? "page" : undefined}
            className={cn(
              "border-l-2 px-3 py-2 font-sans text-sm transition-colors duration-200",
              current
                ? "border-accent font-medium text-fg"
                : "border-transparent text-fg-muted hover:border-line hover:text-fg",
            )}
          >
            {item.label}
          </Link>
        );
      })}
    </nav>
  );
}
