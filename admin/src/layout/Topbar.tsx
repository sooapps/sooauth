import type { Data } from "../api";
import { useI18n } from "../context";
import { Icon } from "../icons";
import { accountPath, homePath, navigate, projectPath, type Route } from "../router";

// Topbar shows where you are. Inside a project the first crumb switches project
// and keeps the current section.
export function Topbar({ me, route, title, menuOpen, onToggleMenu }: { me: Data | null; route: Route; title: string; menuOpen: boolean; onToggleMenu: () => void }) {
  const { lang, tr, setLang } = useI18n();
  const projects: Data[] = me?.projects || [];
  const project = me?.project;
  return (
    <header className="topbar">
      <button type="button" className="menu-toggle" aria-label={tr("sidebar.menu")} aria-expanded={menuOpen} onClick={onToggleMenu}><Icon name="menu" /></button>
      <div className="crumbs">
        {route.kind === "project" && (
          <>
            <label className="crumb-switch">
              <span className="crumb-mark">{(project?.name || "?").slice(0, 1).toUpperCase()}</span>
              <span className="crumb-name">{project?.name}</span>
              <Icon name="updown" size={14} />
              <select aria-label={tr("sidebar.currentProject")} value={project?.id || ""} onChange={(e) => navigate(projectPath(e.target.value, route.section || "overview"))}>
                {projects.map((p) => <option value={p.id} key={p.id}>{p.name}</option>)}
              </select>
            </label>
            <span className="crumb-sep" aria-hidden="true">/</span>
          </>
        )}
        {route.kind === "account" && (
          <>
            <button type="button" className="crumb-link" onClick={() => navigate(accountPath("billing"))}>{tr("nav.account")}</button>
            <span className="crumb-sep" aria-hidden="true">/</span>
          </>
        )}
        {route.kind === "home"
          ? <span className="crumb-current">{tr("nav.dashboard")}</span>
          : <span className="crumb-current">{title}</span>}
      </div>
      {route.kind !== "home" && <button type="button" className="crumb-home" onClick={() => navigate(homePath)}>{tr("nav.dashboard")}</button>}
      <div className="lang-switch" role="group" aria-label={tr("common.language")}>
        <button type="button" className={lang === "en" ? "active" : ""} onClick={() => setLang("en")}>EN</button>
        <button type="button" className={lang === "tr" ? "active" : ""} onClick={() => setLang("tr")}>TR</button>
      </div>
    </header>
  );
}
