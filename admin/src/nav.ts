// Dashboard sections, grouped the way the sidebar shows them. Each section has
// its own URL (/dashboard/p/<project>/<id>) and a matching docs page.
export type Section = { id: string; labelKey: string; introKey: string; docs?: string };
export type NavGroup = { key: string; items: Section[] };

export const navGroups: NavGroup[] = [
  {
    key: "",
    items: [
      { id: "start", labelKey: "nav.start", introKey: "intro.start", docs: "getting-started" },
      { id: "overview", labelKey: "nav.overview", introKey: "intro.overview" },
    ],
  },
  {
    key: "nav.group.auth",
    items: [
      { id: "sign-in", labelKey: "nav.signIn", introKey: "intro.signIn", docs: "sign-in-methods" },
      { id: "social", labelKey: "nav.social", introKey: "intro.social", docs: "social-login" },
      { id: "fields", labelKey: "nav.fields", introKey: "intro.fields", docs: "registration-fields" },
      { id: "security", labelKey: "nav.security", introKey: "intro.security", docs: "sign-in-methods#password-and-verification" },
    ],
  },
  {
    key: "nav.group.integration",
    items: [
      { id: "keys", labelKey: "nav.keys", introKey: "intro.keys", docs: "integrate-oidc" },
      { id: "ai-prompt", labelKey: "nav.aiPrompt", introKey: "intro.aiPrompt", docs: "integrate-oidc#ai-prompt" },
      { id: "test", labelKey: "nav.test", introKey: "intro.test", docs: "getting-started#test" },
      { id: "webhooks", labelKey: "nav.webhooks", introKey: "intro.webhooks", docs: "webhooks" },
    ],
  },
  {
    key: "nav.group.users",
    items: [
      { id: "users", labelKey: "nav.users", introKey: "intro.users", docs: "users-and-sessions" },
      { id: "sessions", labelKey: "nav.sessions", introKey: "intro.sessions", docs: "users-and-sessions#sessions" },
      { id: "audit", labelKey: "nav.audit", introKey: "intro.audit" },
    ],
  },
  {
    key: "nav.group.branding",
    items: [
      { id: "theme", labelKey: "nav.theme", introKey: "intro.theme" },
      { id: "email", labelKey: "nav.email", introKey: "intro.email", docs: "admin#email" },
    ],
  },
  {
    key: "nav.group.project",
    items: [
      { id: "settings", labelKey: "nav.settings", introKey: "intro.settings" },
      { id: "billing", labelKey: "nav.billing", introKey: "intro.billing" },
    ],
  },
];

export const sections: Section[] = navGroups.flatMap((g) => g.items);

export function findSection(id: string): Section | undefined {
  return sections.find((s) => s.id === id);
}

export function groupOf(id: string): string {
  return navGroups.find((g) => g.items.some((s) => s.id === id))?.key || "";
}
