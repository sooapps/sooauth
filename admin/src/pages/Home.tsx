import { useState } from "react";
import { request, type Data } from "../api";
import { useApp, useI18n } from "../context";
import { Icon } from "../icons";
import { NewProjectModal } from "../layout/Sidebar";
import { navigate, projectPath } from "../router";
import { Empty, Loading, useLoad } from "../ui";

// Home lists every project in the account next to what happened last.
export default function Home() {
  const { tr } = useI18n();
  const { me } = useApp();
  const [creating, setCreating] = useState(false);
  const [data] = useLoad<Data>(() => request("/home"), [me?.projects?.length]);
  if (!data) return <Loading />;
  const projects: Data[] = data.projects || [];
  const recent: Data[] = data.recent || [];
  return (
    <div className="home-grid">
      <section>
        <div className="home-heading">
          <h1>{tr("nav.projects")}</h1>
          <button type="button" className="icon-button" aria-label={tr("sidebar.newProjectTitle")} title={tr("sidebar.newProjectTitle")} onClick={() => setCreating(true)}><Icon name="plus" size={16} /></button>
        </div>
        <div className="project-list">
          <div className="project-row head">
            <span>{tr("home.project")}</span>
            <span>{tr("nav.users")}</span>
            <span>{tr("integration.clientId")}</span>
          </div>
          {projects.map((p) => (
            <button type="button" className="project-row" key={p.id} onClick={() => navigate(projectPath(p.id, "overview"))}>
              <span className="project-name"><span className="nav-tile"><Icon name="folder" /></span><strong>{p.name}</strong></span>
              <span>{p.users}</span>
              <code>{p.client_id}</code>
            </button>
          ))}
        </div>
      </section>
      <section>
        <div className="home-heading"><h2>{tr("overview.recent")}</h2></div>
        {recent.length ? (
          <ol className="activity-list">
            {recent.map((item, i) => (
              <li key={i}>
                <button type="button" onClick={() => navigate(projectPath(item.project_id, "audit"))}>
                  <strong>{item.action}</strong>
                  <span>{item.project}</span>
                  <time dateTime={item.created_at}>{new Date(item.created_at).toLocaleString()}</time>
                </button>
              </li>
            ))}
          </ol>
        ) : <Empty text={tr("overview.empty")} />}
      </section>
      {creating && <NewProjectModal onClose={() => setCreating(false)} onCreated={(id) => { setCreating(false); location.href = projectPath(id, "overview"); }} />}
    </div>
  );
}
