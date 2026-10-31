import { useEffect, useRef, useState } from "react";
import { post, type Data, type Onboarding } from "../api";
import { useApp, useI18n } from "../context";
import { Icon } from "../icons";
import { accountSections, navGroups } from "../nav";
import { accountPath, homePath, navigate, projectPath, type Route } from "../router";
import { Button, Field, Modal } from "../ui";
import { ProfileMenu, type AppTheme } from "./ProfileMenu";

export function NewProjectModal({ onClose, onCreated }: { onClose: () => void; onCreated: (id: string) => void }) {
  const { tr } = useI18n();
  const { notify } = useApp();
  const [name, setName] = useState("");
  const [busy, setBusy] = useState(false);
  const create = async () => {
    if (!name.trim()) return;
    setBusy(true);
    try {
      const res = await post("/projects", { name: name.trim() });
      onCreated(res.project?.id || res.id || "");
    } catch (e) {
      notify((e as Error).message);
    } finally {
      setBusy(false);
    }
  };
  return (
    <Modal title={tr("sidebar.newProjectTitle")} onClose={onClose}
      footer={<><Button onClick={onClose}>{tr("common.cancel")}</Button><Button primary disabled={busy || !name.trim()} onClick={create}>{tr("sidebar.create")}</Button></>}>
      <p className="section-desc">{tr("sidebar.newProjectDesc")}</p>
      <form onSubmit={(e) => { e.preventDefault(); create(); }}>
        <Field label={tr("project.name")} value={name} onChange={setName} placeholder="My app" />
      </form>
    </Modal>
  );
}

function NavItem({ icon, label, active, badge, chevron, external, onClick, href }: {
  icon: string;
  label: string;
  active?: boolean;
  badge?: string;
  chevron?: boolean;
  external?: boolean;
  onClick?: () => void;
  href?: string;
}) {
  const body = (
    <>
      <span className="nav-tile"><Icon name={icon} /></span>
      <span className="nav-label">{label}</span>
      {badge && <em className="nav-badge">{badge}</em>}
      {chevron && <span className="nav-trail"><Icon name="next" size={14} /></span>}
      {external && <span className="nav-trail"><Icon name="external" size={14} /></span>}
    </>
  );
  if (href) {
    return <a className="nav-item" href={href} title={label} target={external ? "_blank" : undefined} rel={external ? "noopener noreferrer" : undefined}>{body}</a>;
  }
  return <button type="button" className={`nav-item ${active ? "active" : ""}`} title={label} aria-current={active ? "page" : undefined} onClick={onClick}>{body}</button>;
}

const levelOrder: Record<string, number> = { home: 0, account: 1, project: 1 };

export function Sidebar({ me, route, onboarding, theme, onThemeChange, open, collapsed, onToggleCollapsed, onNavigate, onOpenSearch }: {
  me: Data | null;
  route: Route;
  onboarding: Onboarding | null;
  theme: AppTheme;
  onThemeChange: (t: AppTheme) => void;
  open: boolean;
  collapsed: boolean;
  onToggleCollapsed: () => void;
  onNavigate: () => void;
  onOpenSearch: () => void;
}) {
  const { tr } = useI18n();
  const { go } = useApp();
  const [creating, setCreating] = useState(false);
  const projects: Data[] = me?.projects || [];
  const level = route.kind === "project" ? "project" : route.kind === "account" ? "account" : "home";

  // Going deeper slides the menu in from the right, going back from the left.
  const previous = useRef(level);
  const direction = levelOrder[level] >= levelOrder[previous.current] ? "forward" : "backward";
  useEffect(() => { previous.current = level; }, [level]);

  const open_ = (path: string) => { navigate(path); onNavigate(); };
  const section = route.kind === "project" || route.kind === "account" ? route.section : "";
  const onSearchMac = /mac/i.test(navigator.platform);

  return (
    <aside className={`sidebar ${open ? "open" : ""}`}>
      <a className="brand" href="https://sooauth.com">
        <img className="brand-logo" src={`${import.meta.env.BASE_URL}sooauth-dark.png`} alt="" />
        <span>sooauth</span>
      </a>

      <button type="button" className="nav-search" onClick={onOpenSearch} title={tr("search.placeholder")}>
        <Icon name="search" size={17} />
        <span>{tr("search.placeholder")}</span>
        <kbd>{onSearchMac ? "⌘K" : "Ctrl K"}</kbd>
      </button>

      <nav key={level} className={`nav-level ${direction}`} aria-label={tr("sidebar.menu")}>
        {level === "home" && (
          <>
            <div className="nav-group">
              <NavItem icon="home" label={tr("nav.dashboard")} active={route.kind === "home"} onClick={() => open_(homePath)} />
            </div>
            <div className="nav-group">
              <small>{tr("nav.projects")}</small>
              {projects.map((p) => (
                <NavItem key={p.id} icon="folder" label={p.name} onClick={() => open_(projectPath(p.id, "overview"))} chevron />
              ))}
              <NavItem icon="plus" label={tr("sidebar.newProjectShort")} onClick={() => setCreating(true)} />
            </div>
            <div className="nav-group">
              <small>{tr("nav.manage")}</small>
              <NavItem icon="account" label={tr("nav.account")} onClick={() => open_(accountPath("billing"))} chevron />
              <NavItem icon="book" label={tr("common.documentation")} href="https://sooauth.com/docs" external />
              <NavItem icon="flask" label={tr("start.helpPlayground")} href="https://sooauth.com/playground" external />
            </div>
          </>
        )}

        {level === "project" && (
          <>
            <button type="button" className="nav-back" title={tr("nav.backToDashboard")} aria-label={tr("nav.backToDashboard")} onClick={() => open_(homePath)}>
              <Icon name="back" size={16} />
              <span>{me?.project?.name || ""}</span>
            </button>
            {navGroups.map((group) => (
              <div className="nav-group divided" key={group[0].id}>
                {group.map((item) => (
                  <NavItem
                    key={item.id}
                    icon={item.icon}
                    label={tr(item.labelKey)}
                    active={section === item.id}
                    badge={item.id === "start" && onboarding && !onboarding.complete ? `${onboarding.done_count}/${onboarding.total}` : undefined}
                    onClick={() => { go(item.id); onNavigate(); }}
                  />
                ))}
              </div>
            ))}
          </>
        )}

        {level === "account" && (
          <>
            <button type="button" className="nav-back" title={tr("nav.backToDashboard")} aria-label={tr("nav.backToDashboard")} onClick={() => open_(homePath)}>
              <Icon name="back" size={16} />
              <span>{tr("nav.account")}</span>
            </button>
            <div className="nav-group divided">
              {accountSections.map((item) => (
                <NavItem key={item.id} icon={item.icon} label={tr(item.labelKey)} active={section === item.id} onClick={() => open_(accountPath(item.id))} />
              ))}
              <NavItem icon="shield" label={tr("common.securityCenter")} href="/auth/account" />
            </div>
          </>
        )}
      </nav>

      <div className="sidebar-foot">
        <ProfileMenu me={me} theme={theme} onThemeChange={onThemeChange} />
        <button type="button" className="nav-collapse" title={tr(collapsed ? "nav.expand" : "nav.collapse")} aria-label={tr(collapsed ? "nav.expand" : "nav.collapse")} aria-pressed={collapsed} onClick={onToggleCollapsed}>
          <Icon name="panel" size={17} />
        </button>
      </div>
      {creating && <NewProjectModal onClose={() => setCreating(false)} onCreated={(id) => { setCreating(false); if (id) location.href = projectPath(id, "overview"); else location.reload(); }} />}
    </aside>
  );
}
