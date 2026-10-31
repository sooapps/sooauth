// Dashboard navigation. The sidebar has three levels: the account (home), the
// inside of one project, and account settings. Moving between them replaces the
// menu. Every section has its own URL and, where one exists, a docs page.
export type Section = { id: string; labelKey: string; introKey: string; icon: string; docs?: string };

// Project sections, split into groups the sidebar separates with a divider.
export const navGroups: Section[][] = [
  [
    { id: "start", labelKey: "nav.start", introKey: "intro.start", icon: "rocket", docs: "getting-started" },
    { id: "overview", labelKey: "nav.overview", introKey: "intro.overview", icon: "gauge" },
  ],
  [
    { id: "sign-in", labelKey: "nav.signIn", introKey: "intro.signIn", icon: "key", docs: "sign-in-methods" },
    { id: "social", labelKey: "nav.social", introKey: "intro.social", icon: "users", docs: "social-login" },
    { id: "fields", labelKey: "nav.fields", introKey: "intro.fields", icon: "form", docs: "registration-fields" },
    { id: "security", labelKey: "nav.security", introKey: "intro.security", icon: "shield", docs: "sign-in-methods#password-and-verification" },
  ],
  [
    { id: "keys", labelKey: "nav.keys", introKey: "intro.keys", icon: "plug", docs: "integrate-oidc" },
    { id: "ai-prompt", labelKey: "nav.aiPrompt", introKey: "intro.aiPrompt", icon: "sparkle", docs: "integrate-oidc#ai-prompt" },
    { id: "test", labelKey: "nav.test", introKey: "intro.test", icon: "play", docs: "getting-started#test" },
    { id: "webhooks", labelKey: "nav.webhooks", introKey: "intro.webhooks", icon: "hook", docs: "webhooks" },
  ],
  [
    { id: "users", labelKey: "nav.users", introKey: "intro.users", icon: "users", docs: "users-and-sessions" },
    { id: "sessions", labelKey: "nav.sessions", introKey: "intro.sessions", icon: "devices", docs: "users-and-sessions#sessions" },
    { id: "audit", labelKey: "nav.audit", introKey: "intro.audit", icon: "history" },
  ],
  [
    { id: "theme", labelKey: "nav.theme", introKey: "intro.theme", icon: "palette" },
    { id: "email", labelKey: "nav.email", introKey: "intro.email", icon: "mail", docs: "admin#email" },
    { id: "settings", labelKey: "nav.settings", introKey: "intro.settings", icon: "gear" },
  ],
];

export const sections: Section[] = navGroups.flat();

export const accountSections: Section[] = [
  { id: "billing", labelKey: "nav.billing", introKey: "intro.billing", icon: "card" },
];

export function findSection(id: string): Section | undefined {
  return sections.find((s) => s.id === id);
}

export function findAccountSection(id: string): Section | undefined {
  return accountSections.find((s) => s.id === id);
}
