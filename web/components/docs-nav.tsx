"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { cn } from "../lib/cn";

// Grouped like the dashboard sidebar so each settings page has a matching guide.
const groups: { title?: string; links: { href: string; label: string; exact?: boolean }[] }[] = [
  {
    links: [
      { href: "/docs", label: "Overview", exact: true },
      { href: "/docs/getting-started", label: "Getting started" },
    ],
  },
  {
    title: "Authentication",
    links: [
      { href: "/docs/sign-in-methods", label: "Sign-in methods" },
      { href: "/docs/otp", label: "One-time codes (OTP)" },
      { href: "/docs/two-factor", label: "Two-step verification" },
      { href: "/docs/social-login", label: "Social login" },
      { href: "/docs/google-oauth", label: "Google OAuth app" },
      { href: "/docs/registration-fields", label: "Sign-up fields" },
    ],
  },
  {
    title: "Connect your app",
    links: [
      { href: "/docs/integrate-oidc", label: "OIDC (hosted page)" },
      { href: "/docs/custom-auth-ui", label: "Custom auth UI" },
      { href: "/docs/embed-widget", label: "Embed widget" },
      { href: "/playground", label: "Widget playground ↗" },
      { href: "/docs/webhooks", label: "Webhooks" },
    ],
  },
  {
    title: "Manage",
    links: [
      { href: "/docs/users-and-sessions", label: "Users and sessions" },
      { href: "/docs/admin", label: "Dashboard" },
      { href: "/docs/self-host", label: "Self-host Sooauth" },
    ],
  },
];

export function DocsNav() {
  const pathname = usePathname();
  return (
    <nav aria-label="Documentation" className="grid gap-5">
      {groups.map((group, gi) => (
        <div key={gi} className="grid gap-1">
          {group.title && (
            <p className="px-3 pb-1 font-mono text-[11px] uppercase tracking-wider text-fg-muted">{group.title}</p>
          )}
          {group.links.map((item) => {
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
        </div>
      ))}
    </nav>
  );
}
