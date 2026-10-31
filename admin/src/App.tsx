import { useCallback, useEffect, useMemo, useState, type ReactElement } from "react";
import { post, request, type Data, type Onboarding } from "./api";
import { AppContext, I18nContext } from "./context";
import { getStoredLang, setStoredLang, t, type Lang, type TranslateFn } from "./i18n";
import { CommandPalette } from "./layout/CommandPalette";
import { Sidebar } from "./layout/Sidebar";
import { Topbar } from "./layout/Topbar";
import type { AppTheme } from "./layout/ProfileMenu";
import { findAccountSection, findSection } from "./nav";
import { accountPath, homePath, legacySections, navigate, projectPath, useRoute } from "./router";
import { Loading, PageIntro } from "./ui";
import Home from "./pages/Home";
import Overview from "./pages/Overview";
import GettingStarted from "./pages/start/GettingStarted";
import Wizard from "./pages/start/Wizard";
import SignInMethods from "./pages/auth/SignInMethods";
import Social from "./pages/auth/Social";
import RegistrationFields from "./pages/auth/RegistrationFields";
import SecurityPolicy from "./pages/auth/SecurityPolicy";
import KeysUrls from "./pages/integration/KeysUrls";
import AiPrompt from "./pages/integration/AiPrompt";
import LiveTestPage from "./pages/integration/LiveTest";
import TestCallback from "./pages/integration/TestCallback";
import Webhooks from "./pages/integration/Webhooks";
import Users from "./pages/users/Users";
import Sessions from "./pages/users/Sessions";
import Audit from "./pages/users/Audit";
import Theme from "./pages/branding/Theme";
import EmailDelivery from "./pages/branding/EmailDelivery";
import ProjectSettings from "./pages/project/ProjectSettings";
import Billing from "./pages/project/Billing";
import "./styles.css";
import "./shell.css";

const pages: Record<string, () => ReactElement> = {
  start: GettingStarted,
  overview: Overview,
  "sign-in": SignInMethods,
  social: Social,
  fields: RegistrationFields,
  security: SecurityPolicy,
  keys: KeysUrls,
  "ai-prompt": AiPrompt,
  test: LiveTestPage,
  webhooks: Webhooks,
  users: Users,
  sessions: Sessions,
  audit: Audit,
  theme: Theme,
  email: EmailDelivery,
  settings: ProjectSettings,
  billing: Billing,
};

function useCollapsed() {
  const [collapsed, setCollapsed] = useState(() => {
    try {
      return localStorage.getItem("admin-sidebar") === "rail";
    } catch {
      return false;
    }
  });
  const toggle = () => {
    setCollapsed(!collapsed);
    try { localStorage.setItem("admin-sidebar", collapsed ? "full" : "rail"); } catch { /* ignore */ }
  };
  return [collapsed, toggle] as const;
}

function useTheme() {
  const [theme, setTheme] = useState<AppTheme>(() => {
    try {
      return (localStorage.getItem("admin-theme") as AppTheme) || "system";
    } catch {
      return "system";
    }
  });
  useEffect(() => {
    const apply = () => {
      document.documentElement.dataset.theme = theme === "system"
        ? (window.matchMedia("(prefers-color-scheme: dark)").matches ? "dark" : "light")
        : theme;
    };
    apply();
    if (theme !== "system") return;
    const media = window.matchMedia("(prefers-color-scheme: dark)");
    media.addEventListener("change", apply);
    return () => media.removeEventListener("change", apply);
  }, [theme]);
  const change = (next: AppTheme) => {
    setTheme(next);
    try { localStorage.setItem("admin-theme", next); } catch { /* ignore */ }
  };
  return [theme, change] as const;
}

export function App() {
  const route = useRoute();
  const [lang, setLang] = useState<Lang>(() => getStoredLang());
  const tr = useMemo<TranslateFn>(() => (key, vars) => t(lang, key, vars), [lang]);
  const [theme, changeTheme] = useTheme();
  const [me, setMe] = useState<Data | null>(null);
  const [notice, setNotice] = useState("");
  const [onboarding, setOnboarding] = useState<Onboarding | null>(null);
  const [wizardOpen, setWizardOpen] = useState(false);
  const [menuOpen, setMenuOpen] = useState(false);
  const [switching, setSwitching] = useState(false);
  const [searchOpen, setSearchOpen] = useState(false);
  const [collapsed, toggleCollapsed] = useCollapsed();

  const notify = useCallback((message: string) => {
    setNotice(message);
    window.setTimeout(() => setNotice(""), 3500);
  }, []);

  const changeLang = (next: Lang) => {
    setStoredLang(next);
    setLang(next);
  };
  useEffect(() => { document.documentElement.lang = lang; }, [lang]);

  const isCallback = route.kind === "test-callback";
  useEffect(() => {
    if (isCallback) return;
    request("/me").then(setMe).catch((e) => notify((e as Error).message));
  }, [isCallback, notify]);

  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === "k") {
        e.preventDefault();
        setSearchOpen((v) => !v);
      }
    };
    document.addEventListener("keydown", onKey);
    return () => document.removeEventListener("keydown", onKey);
  }, []);

  const refreshMe = useCallback(async () => {
    setMe(await request("/me"));
  }, []);

  const activeId: string = me?.project?.id || "";
  const refreshOnboarding = useCallback(async () => {
    try {
      setOnboarding(await request("/onboarding"));
    } catch {
      /* the checklist is optional */
    }
  }, []);

  // Keep the URL and the active project in sync. The URL wins: opening a link
  // to another project switches to it.
  useEffect(() => {
    if (!me || isCallback) return;
    const projects: Data[] = me.projects || [];
    if (route.kind === "root") {
      const legacy = legacySections[route.legacyHash];
      if (route.legacyHash === "billing") navigate(accountPath("billing"), { replace: true });
      else if (legacy || projects.length === 1) navigate(projectPath(activeId, legacy || "overview"), { replace: true });
      else navigate(homePath, { replace: true });
      return;
    }
    if (route.kind !== "project") return;
    if (route.section === "billing") {
      navigate(accountPath("billing"), { replace: true });
      return;
    }
    if (route.projectId !== activeId) {
      if (!projects.some((p) => p.id === route.projectId)) {
        navigate(projectPath(activeId, route.section || "overview"), { replace: true });
        return;
      }
      setSwitching(true);
      const target = route.projectId;
      const section = route.section || "overview";
      // Whatever happens, leave the URL on a project we can show: a mismatch
      // here would re-run this effect forever.
      post(`/projects/${target}/select`)
        .then(() => request("/me"))
        .then((fresh: Data) => {
          setMe(fresh);
          if (fresh.project?.id !== target) {
            notify(tr("project.switchFailed"));
            navigate(projectPath(fresh.project?.id, section), { replace: true });
          }
        })
        .catch((e) => {
          notify((e as Error).message);
          navigate(projectPath(activeId, section), { replace: true });
        })
        .finally(() => setSwitching(false));
      return;
    }
    if (!findSection(route.section)) {
      navigate(projectPath(activeId, "overview"), { replace: true });
    }
  }, [me, route, activeId, isCallback, notify, tr]);

  const inProject = route.kind === "project";
  useEffect(() => {
    if (!activeId || !inProject) return;
    request("/onboarding")
      .then((state: Onboarding) => {
        setOnboarding(state);
        // Open the wizard once, for a fresh project; afterwards the checklist
        // under Getting started is the way back in.
        if (!state.dismissed && state.done_count === 0) setWizardOpen(true);
      })
      .catch(() => undefined);
  }, [activeId, inProject]);

  const go = useCallback((section: string) => {
    if (activeId) navigate(projectPath(activeId, section));
  }, [activeId]);

  if (isCallback) {
    return <I18nContext.Provider value={{ lang, tr, setLang: changeLang }}><TestCallback /></I18nContext.Provider>;
  }

  const meta = route.kind === "project" ? findSection(route.section) : route.kind === "account" ? findAccountSection(route.section) : undefined;
  const ready = me && !switching && (route.kind === "home" || (meta && (route.kind === "account" || (route.kind === "project" && route.projectId === activeId))));
  const Page = meta ? pages[meta.id] : undefined;
  const title = meta ? tr(meta.labelKey) : "";

  return (
    <I18nContext.Provider value={{ lang, tr, setLang: changeLang }}>
      <AppContext.Provider value={{ me, projectId: activeId, notify, go, onboarding, refreshOnboarding, refreshMe, openWizard: () => setWizardOpen(true) }}>
        <div className={`app ${collapsed ? "rail" : ""}`}>
          <Sidebar
            me={me}
            route={route}
            onboarding={onboarding}
            theme={theme}
            onThemeChange={changeTheme}
            open={menuOpen}
            collapsed={collapsed}
            onToggleCollapsed={toggleCollapsed}
            onNavigate={() => setMenuOpen(false)}
            onOpenSearch={() => { setMenuOpen(false); setSearchOpen(true); }}
          />
          {menuOpen && <button type="button" className="nav-scrim" aria-label={tr("common.close")} onClick={() => setMenuOpen(false)} />}
          <main>
            <Topbar me={me} route={route} title={title} menuOpen={menuOpen} onToggleMenu={() => setMenuOpen(!menuOpen)} />
            {notice && <div className="toast" role="status">{notice}</div>}
            <div className="content">
              {!ready ? <Loading /> : route.kind === "home" ? <Home /> : meta && Page ? (
                <div key={`${activeId}:${meta.id}`}>
                  <h1 className="page-title">{title}</h1>
                  <PageIntro text={tr(meta.introKey)} docs={meta.docs} />
                  <Page />
                </div>
              ) : <Loading />}
            </div>
          </main>
          {wizardOpen && ready && inProject && <Wizard onClose={() => setWizardOpen(false)} />}
          {searchOpen && <CommandPalette me={me} onClose={() => setSearchOpen(false)} />}
        </div>
      </AppContext.Provider>
    </I18nContext.Provider>
  );
}
