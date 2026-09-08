import { useEffect, useState } from "react";
import "./styles.css";

type Tab = "overview" | "integration" | "providers" | "project" | "users" | "sessions" | "theme" | "billing" | "audit" | "webhooks";
type Data = Record<string, any>;
type AppTheme = "light" | "dark" | "system";

const API = import.meta.env.VITE_API_URL || "";
const tabs: { id: Tab; label: string; group: string }[] = [
  { id: "overview", label: "Overview", group: "Workspace" },
  { id: "integration", label: "Integration", group: "Workspace" },
  { id: "providers", label: "Social login", group: "Workspace" },
  { id: "project", label: "Project", group: "Workspace" },
  { id: "users", label: "Users", group: "Workspace" },
  { id: "sessions", label: "Sessions", group: "Workspace" },
  { id: "theme", label: "Theme", group: "Configuration" },
  { id: "billing", label: "Billing", group: "Configuration" },
  { id: "audit", label: "Audit", group: "Configuration" },
  { id: "webhooks", label: "Webhooks", group: "Configuration" },
];

function csrf() {
  return document.cookie.match(/(?:^|; )sooauth_csrf=([^;]+)/)?.[1] || "";
}

async function request(path: string, options: RequestInit = {}) {
  const method = (options.method || "GET").toString().toUpperCase();
  const headers = new Headers(options.headers);
  if (method !== "GET") {
    headers.set("Content-Type", "application/json");
    headers.set("X-CSRF-Token", decodeURIComponent(csrf()));
  }
  const response = await fetch(`${API}/dashboard/api${path}`, { ...options, headers, credentials: "include" });
  const body = await response.json().catch(() => ({}));
  if (response.status === 401) {
    window.location.href = `${window.location.origin}/auth/sign-in?return_to=/dashboard/`;
    throw new Error("Authentication required");
  }
  if (!response.ok) throw new Error(body.message || body.error || "Request failed");
  return body;
}

function Button({ children, primary, onClick, type = "button", disabled }: { children: React.ReactNode; primary?: boolean; onClick?: () => void; type?: "button" | "submit"; disabled?: boolean }) {
  return <button type={type} disabled={disabled} onClick={onClick} className={`button ${primary ? "primary" : ""}`}>{children}</button>;
}

function Card({ title, description, children, className = "" }: { title?: string; description?: string; children?: React.ReactNode; className?: string }) {
  return <section className={`card ${className}`}>{title && <div className="card-heading"><div><h2>{title}</h2>{description && <p>{description}</p>}</div></div>}{children}</section>;
}

function Field({ label, value, onChange, type = "text", disabled = false }: { label: string; value: string | number; onChange: (value: string) => void; type?: string; disabled?: boolean }) {
  return <label className="field"><span>{label}</span><input disabled={disabled} type={type} value={value} onChange={(e) => onChange(e.target.value)} /></label>;
}

function Integration({ notify }: { notify: (message: string) => void }) {
  const [data, setData] = useState<Data | null>(null);
  const [uris, setUris] = useState<string[]>([]);
  const [origin, setOrigin] = useState("");
  useEffect(() => { request("/integration").then((value) => { setData(value); setUris(value.redirect_uris || [""]); setOrigin(value.social_callback_origin || ""); }).catch((e) => notify(e.message)); }, []);
  if (!data) return <Loading />;
  const copy = (value: string) => navigator.clipboard.writeText(value).then(() => notify("Copied to clipboard."));
  const save = async () => { try { await request("/integration", { method: "PUT", body: JSON.stringify({ redirect_uris: uris.filter(Boolean), social_callback_origin: origin }) }); notify("Integration settings saved."); } catch (e) { notify((e as Error).message); } };
  return <>
    <Card title="Start here" description="Connect your first app in five minutes. Keep this checklist open while you integrate.">
      <div className="checklist"><span>✓ Copy your issuer, client ID, and discovery URL below.</span><span>{uris.length ? "✓" : "○"} Add an exact callback URL under Redirect URIs.</span><span>○ Use the <a href="https://sooauth.com/docs/integrate-oidc#ai-prompt">OIDC setup prompt</a> or your framework&apos;s OIDC library.</span><span>○ Test the hosted sign-in flow before shipping.</span></div>
      <a className="button primary inline" href="/auth/sign-in?return_to=/dashboard/">Test sign-in</a>
    </Card>
    <Card title="Connect your app" description="Copy issuer and client ID into your app's OIDC config. Social keys stay in Sooauth only.">
      <div className="data-grid">{[["Issuer", data.issuer], ["Client ID", data.client_id], ["Discovery URL", data.discovery_url]].map(([label, value]) => <div className="copy-field" key={label as string}><span>{label}</span><div><code>{value}</code><Button onClick={() => copy(value as string)}>Copy</Button></div></div>)}</div>
    </Card>
    <Card title="Google OAuth branding" description="Use a custom Google OAuth app on Pro to show your domain in Google's account picker.">
      {data.can_custom_social_callback ? <><Field label="App origin" value={origin} onChange={setOrigin} /><Button primary onClick={save}>Save Google branding</Button></> : <div className="notice">Free plan: Google sign-in uses sooauth platform OAuth and shows sooauth.com.</div>}
    </Card>
    <Card title="Redirect URIs" description="Add exact callback and login URLs for each environment.">
      <div className="uri-list">{uris.map((uri, index) => <div className="uri-row" key={index}><input value={uri} onChange={(e) => setUris(uris.map((item, i) => i === index ? e.target.value : item))} placeholder="https://yourapp.com/api/auth/callback/oidc" /><Button onClick={() => setUris(uris.filter((_, i) => i !== index))}>×</Button></div>)}</div>
      <div className="actions"><Button onClick={() => setUris([...uris, ""])}>Add URI</Button><Button primary onClick={save}>Save URIs</Button></div>
    </Card>
  </>;
}

function Providers({ notify }: { notify: (message: string) => void }) {
  const [providers, setProviders] = useState<Data[]>([]);
  const [open, setOpen] = useState<string | null>(null);
  const load = () => request("/providers").then((data) => setProviders(data.providers || [])).catch((e) => notify(e.message));
  useEffect(() => { load(); }, []);
  const update = async (provider: string, body: Data) => { try { await request(`/providers/${provider}`, { method: "PUT", body: JSON.stringify(body) }); notify("Provider settings saved."); load(); } catch (e) { notify((e as Error).message); } };
  return <><div className="notice">Enable the providers you want to show on your login page. Expand a provider to configure custom OAuth credentials.</div><div className="provider-list">{providers.map((provider) => <article className={`provider ${open === provider.provider ? "open" : ""}`} key={provider.provider}>
    <button className="provider-head" onClick={() => setOpen(open === provider.provider ? null : provider.provider)}><div><h2>{provider.provider[0].toUpperCase() + provider.provider.slice(1)} <span className={`status ${provider.enabled ? "on" : "off"}`}>{provider.enabled ? "Live" : "Off"}</span></h2><p>{provider.has_custom ? "Custom OAuth app" : provider.platform_ready ? "Using sooauth platform OAuth" : "Add custom credentials to enable"}</p></div><span className="chevron">{open === provider.provider ? "−" : "+"}</span></button>
    {open === provider.provider && <div className="provider-body"><div className="mode-row"><label><input type="radio" checked={!provider.has_custom} onChange={() => update(provider.provider, { enabled: provider.enabled, use_platform: true })} /> Use platform</label><label><input type="radio" checked={provider.has_custom} readOnly /> Custom OAuth app</label></div><Field label="Client ID" value={provider.client_id || ""} onChange={(value) => setProviders(providers.map((item) => item.provider === provider.provider ? { ...item, client_id: value } : item))} /><Field label="Client secret" type="password" value={provider.secret || ""} onChange={(value) => setProviders(providers.map((item) => item.provider === provider.provider ? { ...item, secret: value } : item))} /><Button primary onClick={() => update(provider.provider, { enabled: provider.enabled, custom: true, client_id: provider.client_id || "", client_secret: provider.secret || "" })}>Save custom credentials</Button></div>}
  </article>)}</div></>;
}

function Overview() { const [data, setData] = useState<Data | null>(null); useEffect(() => { Promise.all([request("/users"), request("/sessions"), request("/audit")]).then(([users, sessions, audit]) => setData({ users, sessions, audit })).catch(() => setData({ users: {}, sessions: {}, audit: {} })); }, []); if (!data) return <Loading />; return <><div className="metrics"><Metric value={(data.users.users || []).length} label="App users" /><Metric value={(data.sessions.sessions || []).length} label="Active sessions" /><Metric value={(data.audit.entries || []).length} label="Audit events" /></div><Card title="Recent activity" description="A quick health check for this project.">{data.audit.entries?.length ? <Table headers={["Time", "Action", "IP"]} rows={data.audit.entries.slice(0, 8).map((item: Data) => [item.created_at, item.action, item.ip || "—"])} /> : <Empty text="No events yet." />}</Card></>; }
function Metric({ value, label }: { value: number; label: string }) { return <div className="metric"><strong>{value}</strong><span>{label}</span></div>; }
function Table({ headers, rows }: { headers: string[]; rows: any[][] }) { return <div className="table-wrap"><table><thead><tr>{headers.map((header) => <th key={header}>{header}</th>)}</tr></thead><tbody>{rows.map((row, i) => <tr key={i}>{row.map((cell, j) => <td key={j}>{String(cell)}</td>)}</tr>)}</tbody></table></div>; }
function Empty({ text }: { text: string }) { return <div className="empty">{text}</div>; }
function Loading() { return <div className="loading">Loading workspace…</div>; }
function ThemeControl({ theme, onChange }: { theme: AppTheme; onChange: (theme: AppTheme) => void }) { return <label className="theme-control"><span>Appearance</span><select value={theme} onChange={(event) => onChange(event.target.value as AppTheme)} aria-label="Appearance"><option value="system">System</option><option value="light">Light</option><option value="dark">Dark</option></select></label>; }

function WebhookRow({ webhook, notify }: { webhook: Data; notify: (message: string) => void }) {
  const [open, setOpen] = useState(false);
  const [deliveries, setDeliveries] = useState<Data[]>([]);
  const load = async () => { try { const result = await request(`/webhooks/${webhook.id}/deliveries`); setDeliveries(result.deliveries || []); } catch (e) { notify((e as Error).message); } };
  const test = async () => { try { await request(`/webhooks/${webhook.id}/test`, { method: "POST" }); await load(); notify("Test delivery completed."); setOpen(true); } catch (e) { notify(`Test delivery failed: ${(e as Error).message}`); } };
  return <article className="webhook-row"><div className="webhook-summary"><div><strong>{webhook.url}</strong><p>{(webhook.events || []).join(" · ") || "All events"}</p></div><span className="status on">{webhook.active ? "Active" : "Inactive"}</span></div><div className="webhook-actions"><Button onClick={test}>Send test</Button><Button onClick={async () => { if (!confirm("Delete this webhook endpoint?")) return; try { await request(`/webhooks/${webhook.id}`, { method: "DELETE" }); notify("Webhook deleted. Refresh to update the list."); } catch (e) { notify((e as Error).message); } }}>Delete</Button><Button onClick={() => { setOpen(!open); if (!open) load(); }}>{open ? "Hide deliveries" : "View request log"}</Button></div>{open && <div className="delivery-list">{deliveries.length ? deliveries.map((delivery) => <details className="delivery" key={delivery.id}><summary><span>{delivery.event}</span><span>{delivery.status_code || "Failed"} · {new Date(delivery.created_at).toLocaleString()}</span></summary><div className="delivery-grid"><div><b>Request body</b><pre>{delivery.payload || "{}"}</pre></div><div><b>Response</b><pre>{delivery.response || delivery.error || "No response"}</pre></div></div></details>) : <Empty text="No delivery attempts yet. Send a test to inspect the request and response." />}</div>}</article>;
}

function ResourcePage({ tab, notify }: { tab: Tab; notify: (message: string) => void }) {
  const [data, setData] = useState<Data | null>(null);
  const [webhookURL, setWebhookURL] = useState("");
  const [webhookEvents, setWebhookEvents] = useState(["user.created", "user.login", "user.login_failed", "user.email_verified", "password.reset", "session.revoked"]);
  const endpoint = tab === "users" ? "/users" : tab === "sessions" ? "/sessions" : tab === "audit" ? "/audit" : "/webhooks";
  useEffect(() => { request(endpoint).then(setData).catch((e) => notify(e.message)); }, [endpoint]);
  if (!data) return <Loading />;
  if (tab === "users") return <Card title={`${(data.users || []).length} app users`} description="Manage end users who signed up through your product.">{data.users?.length ? <div className="table-wrap"><table><thead><tr>{["Email", "Method", "Verified", "Status", "Joined", ""].map((h) => <th key={h}>{h}</th>)}</tr></thead><tbody>{data.users.map((u: Data) => <tr key={u.id}><td>{u.email}</td><td>{u.signup_method || "—"}</td><td>{u.email_verified ? "Yes" : "No"}</td><td>{u.disabled ? "Disabled" : "Active"}</td><td>{u.created_at}</td><td><Button onClick={async () => { try { await request(`/users/${u.id}/${u.disabled ? "enable" : "disable"}`, { method: "POST" }); notify("User status updated."); setData({ ...data, users: data.users.map((item: Data) => item.id === u.id ? { ...item, disabled: !item.disabled } : item) }); } catch (e) { notify((e as Error).message); } }}>{u.disabled ? "Enable" : "Disable"}</Button></td></tr>)}</tbody></table></div> : <Empty text="No app users yet." />}</Card>;
  if (tab === "sessions") return <Card title="Active browser sessions" description="Revoke a session to invalidate it immediately.">{data.sessions?.length ? <div className="table-wrap"><table><thead><tr>{["User", "IP", "Expires", ""].map((h) => <th key={h}>{h}</th>)}</tr></thead><tbody>{data.sessions.map((s: Data) => <tr key={s.id}><td>{s.email}</td><td>{s.ip || "—"}</td><td>{s.expires_at}</td><td><Button onClick={async () => { try { await request(`/sessions/${s.id}/revoke`, { method: "POST" }); notify("Session revoked."); setData({ ...data, sessions: data.sessions.filter((item: Data) => item.id !== s.id) }); } catch (e) { notify((e as Error).message); } }}>Revoke</Button></td></tr>)}</tbody></table></div> : <Empty text="No active sessions." />}</Card>;
  if (tab === "audit") return <Card title="Audit log">{data.entries?.length ? <Table headers={["Time", "Action", "IP"]} rows={data.entries.map((e: Data) => [e.created_at, e.action, e.ip || "—"])} /> : <Empty text="No events yet." />}</Card>;
  const createWebhook = async () => { try { const result = await request("/webhooks", { method: "POST", body: JSON.stringify({ url: webhookURL, events: webhookEvents }) }); setWebhookURL(""); setData({ ...data, webhooks: [...(data.webhooks || []), { url: webhookURL, events: webhookEvents, active: true, id: result.id }] }); notify("Webhook endpoint created."); } catch (e) { notify((e as Error).message); } };
  return <><Card title="Add endpoint" description="Signed events keep your application database in sync."><div className="form-narrow"><Field label="Endpoint URL" value={webhookURL} onChange={setWebhookURL} /><div className="event-grid">{["user.created", "user.login", "user.login_failed", "user.email_verified", "password.reset", "session.revoked"].map((event) => <label className="check" key={event}><input type="checkbox" checked={webhookEvents.includes(event)} onChange={() => setWebhookEvents(webhookEvents.includes(event) ? webhookEvents.filter((item) => item !== event) : [...webhookEvents, event])} /> {event}</label>)}</div><Button primary onClick={createWebhook}>Create endpoint</Button></div></Card><Card title={`${(data.webhooks || []).length} webhook endpoints`} description="Inspect the exact request body and response for every test delivery.">{data.webhooks?.length ? <div className="webhook-list">{data.webhooks.map((webhook: Data) => <WebhookRow key={webhook.id} webhook={webhook} notify={notify} />)}</div> : <Empty text="No endpoints configured yet." />}</Card></>;
}

function Settings({ tab, notify }: { tab: "project" | "theme" | "billing"; notify: (message: string) => void }) {
  const [data, setData] = useState<Data | null>(null);
  const [draft, setDraft] = useState<Data>({});
  useEffect(() => { request(tab === "project" ? "/project" : tab === "theme" ? "/theme" : "/billing").then((value) => { setData(value); setDraft(value); }).catch((e) => notify(e.message)); }, [tab]);
  if (!data) return <Loading />;
  if (tab === "billing") return <><Card title="Current plan" description={`${data.plan_label} · ${data.max_projects} project(s) · status ${data.status}`} /> <div className="plans">{(data.plans || []).map((plan: Data) => <Card key={plan.id} title={plan.label} className={plan.id === data.plan ? "current-plan" : ""}><strong className="price">{plan.price_try ? `${plan.price_try.amount_cents / 100} TRY` : "Free"}</strong><ul>{(plan.features || []).map((feature: string) => <li key={feature}>{feature}</li>)}</ul><Button primary={plan.id !== data.plan}>{plan.id === data.plan ? "Current plan" : "Upgrade"}</Button></Card>)}</div></>;
  const project = tab === "project";
  const save = async () => { try { await request(project ? "/project" : "/theme", { method: "PUT", body: JSON.stringify(project ? { name: draft.name, password_min_length: Number(draft.password_min_length || 8), email_verify_required: Boolean(draft.email_verify_required), email_verify_delivery: draft.email_verify_delivery || "link", password_require_upper: Boolean(draft.password_require_upper), password_require_number: Boolean(draft.password_require_number), password_require_special: Boolean(draft.password_require_special) } : { brand_name: draft.brand_name, logo_url: draft.logo_url, accent_color: draft.accent_color }) }); notify("Settings saved."); } catch (e) { notify((e as Error).message); } };
  return <Card title={project ? "Project settings" : "Login page look"} description={project ? "Auth rules for end users signing up through your app." : "Customize the hosted sign-in and sign-up pages."}><div className={project ? "settings-layout" : "form-narrow"}><div className="form-narrow"><Field label={project ? "Project name" : "Brand name"} value={project ? draft.name || "" : draft.brand_name || "sooauth"} onChange={(value) => setDraft({ ...draft, [project ? "name" : "brand_name"]: value })} />{project ? <><label className="check"><input type="checkbox" checked={Boolean(draft.email_verify_required)} onChange={(e) => setDraft({ ...draft, email_verify_required: e.target.checked })} /> Require email verification</label><label className="field"><span>Verification delivery</span><select value={draft.email_verify_delivery || "link"} onChange={(e) => setDraft({ ...draft, email_verify_delivery: e.target.value })}><option value="link">Email link only</option><option value="code">6-digit code only</option><option value="both">Link and code</option></select></label><Field label="Minimum password length" type="number" value={draft.password_min_length || 8} onChange={(value) => setDraft({ ...draft, password_min_length: value })} /><label className="check"><input type="checkbox" checked={Boolean(draft.password_require_upper)} onChange={(e) => setDraft({ ...draft, password_require_upper: e.target.checked })} /> Require uppercase letter</label><label className="check"><input type="checkbox" checked={Boolean(draft.password_require_number)} onChange={(e) => setDraft({ ...draft, password_require_number: e.target.checked })} /> Require number</label><label className="check"><input type="checkbox" checked={Boolean(draft.password_require_special)} onChange={(e) => setDraft({ ...draft, password_require_special: e.target.checked })} /> Require special character</label></> : <Field label="Logo URL" value={draft.logo_url || ""} onChange={(value) => setDraft({ ...draft, logo_url: value })} />}<Button primary onClick={save}>Save changes</Button></div>{project && <aside className="rule-card"><span>Current password rule</span><strong>At least {draft.password_min_length || 8} characters</strong><p>{draft.password_require_upper ? "one uppercase letter, " : ""}{draft.password_require_number ? "one number, " : ""}{draft.password_require_special ? "one special character" : "no additional character rules"}</p><small>Users will see this rule during sign-up.</small></aside>}</div></Card>;
}

export function App() {
  const [tab, setTab] = useState<Tab>((location.hash.slice(1) as Tab) || "overview");
  const [me, setMe] = useState<Data | null>(null);
  const [notice, setNotice] = useState("");
  const [theme, setTheme] = useState<AppTheme>(() => (localStorage.getItem("admin-theme") as AppTheme) || "system");
  const notify = (message: string) => { setNotice(message); window.setTimeout(() => setNotice(""), 3500); };
  useEffect(() => { const apply = () => document.documentElement.dataset.theme = theme === "system" ? (window.matchMedia("(prefers-color-scheme: dark)").matches ? "dark" : "light") : theme; apply(); if (theme !== "system") return; const media = window.matchMedia("(prefers-color-scheme: dark)"); media.addEventListener("change", apply); return () => media.removeEventListener("change", apply); }, [theme]);
  const changeTheme = (next: AppTheme) => { setTheme(next); localStorage.setItem("admin-theme", next); };
  useEffect(() => { request("/me").then(setMe).catch((e) => notify(e.message)); }, []);
  useEffect(() => { history.replaceState(null, "", `/dashboard/#${tab}`); }, [tab]);
  const title = tabs.find((item) => item.id === tab)?.label || "Dashboard";
  const page = tab === "integration" ? <Integration notify={notify} /> : tab === "providers" ? <Providers notify={notify} /> : tab === "overview" ? <Overview /> : ["users", "sessions", "audit", "webhooks"].includes(tab) ? <ResourcePage tab={tab} notify={notify} /> : ["project", "theme", "billing"].includes(tab) ? <Settings tab={tab as "project" | "theme" | "billing"} notify={notify} /> : null;
  const switchProject = async (id: string) => { try { await request(`/projects/${id}/select`, { method: "POST" }); location.reload(); } catch (e) { notify((e as Error).message); } };
  const createProject = async () => { const name = window.prompt("Project name"); if (!name?.trim()) return; try { await request("/projects", { method: "POST", body: JSON.stringify({ name: name.trim() }) }); location.reload(); } catch (e) { notify((e as Error).message); } };
  return <div className="app"><aside className="sidebar"><a className="brand" href="https://sooauth.com"><img className="brand-logo brand-logo-light" src="/sooauth-light.png" alt="sooauth" /><img className="brand-logo brand-logo-dark" src="/sooauth-dark.png" alt="" /><span>sooauth</span></a>{me?.projects?.length ? <div className="sidebar-project"><label className="project-picker"><span>Current project</span><div className="project-select-wrap"><select className="project-select" value={me.project?.id || ""} onChange={(e) => switchProject(e.target.value)} aria-label="Current project">{me.projects.map((project: Data) => <option value={project.id} key={project.id}>{project.name}</option>)}</select></div></label><Button primary onClick={createProject}>+ New project</Button></div> : null}<nav>{["Workspace", "Configuration"].map((group) => <div className="nav-group" key={group}><small>{group}</small>{tabs.filter((item) => item.group === group).map((item) => <button className={tab === item.id ? "active" : ""} key={item.id} onClick={() => setTab(item.id)}>{item.label}</button>)}</div>)}</nav><div className="sidebar-links"><a href="https://sooauth.com/docs">Documentation ↗</a><a href="https://sooauth.com/docs/integrate-oidc#ai-prompt">AI setup prompt ↗</a><a href="/auth/account">Security Center ↗</a></div><div className="profile"><div className="profile-identity"><span className="avatar">{(me?.email || "?").slice(0, 1).toUpperCase()}</span><strong>{me?.email || "Loading account…"}</strong></div><button onClick={() => { fetch(`${API}/auth/sign-out`, { method: "POST", credentials: "include", headers: { "X-CSRF-Token": decodeURIComponent(csrf()) } }).then(() => { location.href = "/auth/sign-in"; }); }}>Sign out</button></div></aside><main><header className="topbar"><div><p className="eyebrow">Workspace</p><h1>{title}{me?.project?.name ? <span> · {me.project.name}</span> : null}</h1></div><ThemeControl theme={theme} onChange={changeTheme} /></header>{notice && <div className="toast">{notice}</div>}<div className="content">{page}</div></main></div>;
}
