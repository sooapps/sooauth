import { useCallback, useEffect, useMemo, useState, type ReactElement } from "react";
import { post, request, type Data, type Onboarding } from "./api";
import { AppContext, I18nContext } from "./context";
import { getStoredLang, setStoredLang, t, type Lang, type TranslateFn } from "./i18n";
import { Sidebar } from "./layout/Sidebar";
import type { AppTheme } from "./layout/ProfileMenu";
import { findSection, groupOf } from "./nav";
import { legacySections, navigate, projectPath, useRoute } from "./router";
import { Loading, PageIntro } from "./ui";
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
      navigate(projectPath(activeId, legacy || "overview"), { replace: true });
      return;
    }
    if (route.kind !== "project") return;
    if (route.projectId !== activeId) {
      if (!projects.some((p) => p.id === route.projectId)) {
        navigate(projectPath(activeId, route.section || "overview"), { replace: true });
        return;
      }
      setSwitching(true);
      post(`/projects/${route.projectId}/select`)
        .then(() => request("/me"))
        .then(setMe)
        .catch((e) => notify((e as Error).message))
        .finally(() => setSwitching(false));
      return;
    }
    if (!findSection(route.section)) {
      navigate(projectPath(activeId, "overview"), { replace: true });
    }
  }, [me, route, activeId, isCallback, notify]);

  useEffect(() => {
    if (!activeId) return;
    request("/onboarding")
      .then((state: Onboarding) => {
        setOnboarding(state);
        // Open the wizard once, for a fresh project; afterwards the checklist
        // under Getting started is the way back in.
        if (!state.dismissed && state.done_count === 0) setWizardOpen(true);
      })
      .catch(() => undefined);
  }, [activeId]);

  const go = useCallback((section: string) => {
    if (activeId) navigate(projectPath(activeId, section));
  }, [activeId]);

  if (isCallback) {
    return <I18nContext.Provider value={{ lang, tr, setLang: changeLang }}><TestCallback /></I18nContext.Provider>;
  }

  const section = route.kind === "project" ? route.section : "";
  const meta = findSection(section);
  const ready = me && route.kind === "project" && route.projectId === activeId && meta && !switching;
  const Page = meta ? pages[meta.id] : undefined;

  return (
    <I18nContext.Provider value={{ lang, tr, setLang: changeLang }}>
      <AppContext.Provider value={{ me, projectId: activeId, notify, go, onboarding, refreshOnboarding, openWizard: () => setWizardOpen(true) }}>
        <div className="app">
          <Sidebar
            me={me}
            section={section}
            onboarding={onboarding}
            theme={theme}
            onThemeChange={changeTheme}
            onSwitchProject={(id) => navigate(projectPath(id, section || "overview"))}
            open={menuOpen}
            onNavigate={() => setMenuOpen(false)}
          />
          <main>
            <header className="topbar">
              <button type="button" className="menu-toggle" aria-label={tr("sidebar.menu")} aria-expanded={menuOpen} onClick={() => setMenuOpen(!menuOpen)}>☰</button>
              <div>
                <p className="eyebrow">{meta && groupOf(meta.id) ? tr(groupOf(meta.id)) : me?.project?.name || ""}</p>
                <h1>{meta ? tr(meta.labelKey) : ""}{me?.project?.name && meta && groupOf(meta.id) ? <span> · {me.project.name}</span> : null}</h1>
              </div>
              <div className="lang-switch" role="group" aria-label={tr("common.language")}>
                <button type="button" className={lang === "en" ? "active" : ""} onClick={() => changeLang("en")}>EN</button>
                <button type="button" className={lang === "tr" ? "active" : ""} onClick={() => changeLang("tr")}>TR</button>
              </div>
            </header>
            {notice && <div className="toast" role="status">{notice}</div>}
            <div className="content">
              {ready && meta && Page ? (
                <div key={`${activeId}:${meta.id}`}>
                  <PageIntro text={tr(meta.introKey)} docs={meta.docs} />
                  <Page />
                </div>
              ) : <Loading />}
            </div>
          </main>
          {wizardOpen && ready && <Wizard onClose={() => setWizardOpen(false)} />}
        </div>
      </AppContext.Provider>
    </I18nContext.Provider>
  );
}
