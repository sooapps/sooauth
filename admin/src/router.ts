import { useEffect, useState } from "react";

// A tiny history-based router for the dashboard. URLs look like
// /dashboard/p/<projectId>/<section>, so every page has a shareable link and
// back/forward work.

export type Route =
  | { kind: "project"; projectId: string; section: string }
  | { kind: "home" }
  | { kind: "account"; section: string }
  | { kind: "test-callback" }
  | { kind: "root"; legacyHash: string };

const BASE = "/dashboard";

export function parseRoute(pathname = location.pathname, hash = location.hash): Route {
  const rest = pathname.startsWith(BASE) ? pathname.slice(BASE.length) : pathname;
  const parts = rest.split("/").filter(Boolean);
  if (parts[0] === "test-callback") return { kind: "test-callback" };
  if (parts[0] === "home") return { kind: "home" };
  if (parts[0] === "account") return { kind: "account", section: parts[1] || "billing" };
  if (parts[0] === "p" && parts[1]) {
    return { kind: "project", projectId: parts[1], section: parts[2] || "" };
  }
  return { kind: "root", legacyHash: hash.replace(/^#/, "") };
}

export function projectPath(projectId: string, section: string) {
  return `${BASE}/p/${projectId}/${section}`;
}

export const homePath = `${BASE}/home`;

export function accountPath(section: string) {
  return `${BASE}/account/${section}`;
}

const listeners = new Set<() => void>();

export function navigate(path: string, opts: { replace?: boolean } = {}) {
  if (path === location.pathname + location.search) return;
  if (opts.replace) history.replaceState(null, "", path);
  else history.pushState(null, "", path);
  window.scrollTo(0, 0);
  listeners.forEach((fn) => fn());
}

export function useRoute(): Route {
  const [route, setRoute] = useState<Route>(() => parseRoute());
  useEffect(() => {
    const update = () => setRoute(parseRoute());
    listeners.add(update);
    window.addEventListener("popstate", update);
    return () => {
      listeners.delete(update);
      window.removeEventListener("popstate", update);
    };
  }, []);
  return route;
}

// Old dashboard links used #tab hashes; map them to the new sections.
export const legacySections: Record<string, string> = {
  overview: "overview",
  integration: "keys",
  providers: "social",
  project: "settings",
  forms: "fields",
  users: "users",
  sessions: "sessions",
  theme: "theme",
  email: "email",
  audit: "audit",
  webhooks: "webhooks",
};
