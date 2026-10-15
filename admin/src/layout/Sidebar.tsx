import { useState } from "react";
import { post, type Data, type Onboarding } from "../api";
import { useApp, useI18n } from "../context";
import { navGroups } from "../nav";
import { Button, Field, Modal } from "../ui";
import { ProfileMenu, type AppTheme } from "./ProfileMenu";

function NewProjectModal({ onClose, onCreated }: { onClose: () => void; onCreated: (id: string) => void }) {
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
      <Field label={tr("project.name")} value={name} onChange={setName} placeholder="My app" />
    </Modal>
  );
}

export function Sidebar({ me, section, onboarding, theme, onThemeChange, onSwitchProject, open, onNavigate }: {
  me: Data | null;
  section: string;
  onboarding: Onboarding | null;
  theme: AppTheme;
  onThemeChange: (t: AppTheme) => void;
  onSwitchProject: (id: string) => void;
  open: boolean;
  onNavigate: () => void;
}) {
  const { tr } = useI18n();
  const { go } = useApp();
  const [creating, setCreating] = useState(false);
  const projects: Data[] = me?.projects || [];

  return (
    <aside className={`sidebar ${open ? "open" : ""}`}>
      <a className="brand" href="https://sooauth.com">
        <img className="brand-logo brand-logo-light" src={`${import.meta.env.BASE_URL}sooauth-light.png`} alt="sooauth" />
        <img className="brand-logo brand-logo-dark" src={`${import.meta.env.BASE_URL}sooauth-dark.png`} alt="" />
        <span>sooauth</span>
      </a>
      <div className="sidebar-project">
        <label className="project-picker">
          <span>{tr("sidebar.currentProject")}</span>
          <div className="project-select-wrap">
            <select className="project-select" value={me?.project?.id || ""} onChange={(e) => onSwitchProject(e.target.value)} aria-label={tr("sidebar.currentProject")}>
              {projects.map((project) => <option value={project.id} key={project.id}>{project.name}</option>)}
            </select>
          </div>
        </label>
        <button type="button" className="btn-new-project" onClick={() => setCreating(true)}>{tr("sidebar.newProject")}</button>
      </div>
      <nav>
        {navGroups.map((group) => (
          <div className="nav-group" key={group.key || "top"}>
            {group.key && <small>{tr(group.key)}</small>}
            {group.items.map((item) => (
              <button
                className={section === item.id ? "active" : ""}
                key={item.id}
                onClick={() => { go(item.id); onNavigate(); }}
                aria-current={section === item.id ? "page" : undefined}
              >
                <span>{tr(item.labelKey)}</span>
                {item.id === "start" && onboarding && !onboarding.complete && (
                  <em className="nav-badge">{onboarding.done_count}/{onboarding.total}</em>
                )}
              </button>
            ))}
          </div>
        ))}
      </nav>
      <div className="sidebar-links">
        <a href="https://sooauth.com/docs" target="_blank" rel="noopener noreferrer">{tr("sidebar.docs")}</a>
        <a href="/auth/account">{tr("sidebar.security")}</a>
      </div>
      <ProfileMenu me={me} theme={theme} onThemeChange={onThemeChange} />
      {creating && <NewProjectModal onClose={() => setCreating(false)} onCreated={(id) => { setCreating(false); if (id) onSwitchProject(id); else location.reload(); }} />}
    </aside>
  );
}
