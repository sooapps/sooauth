import { createContext, useContext, useEffect, useMemo, useRef, useState } from "react";
import { getStoredLang, setStoredLang, t, type Lang, type TranslateFn } from "./i18n";
import "./styles.css";

type Tab = "overview" | "integration" | "providers" | "project" | "forms" | "users" | "sessions" | "theme" | "email" | "billing" | "audit" | "webhooks";
type Data = Record<string, any>;
type AppTheme = "light" | "dark" | "system";

const API = import.meta.env.VITE_API_URL || "";

const tabMeta: { id: Tab; labelKey: string; groupKey: string }[] = [
  { id: "overview", labelKey: "nav.overview", groupKey: "nav.group.workspace" },
  { id: "integration", labelKey: "nav.integration", groupKey: "nav.group.workspace" },
  { id: "providers", labelKey: "nav.providers", groupKey: "nav.group.workspace" },
  { id: "project", labelKey: "nav.project", groupKey: "nav.group.workspace" },
  { id: "forms", labelKey: "nav.forms", groupKey: "nav.group.workspace" },
  { id: "users", labelKey: "nav.users", groupKey: "nav.group.workspace" },
  { id: "sessions", labelKey: "nav.sessions", groupKey: "nav.group.workspace" },
  { id: "theme", labelKey: "nav.theme", groupKey: "nav.group.configuration" },
  { id: "email", labelKey: "nav.email", groupKey: "nav.group.configuration" },
  { id: "billing", labelKey: "nav.billing", groupKey: "nav.group.configuration" },
  { id: "audit", labelKey: "nav.audit", groupKey: "nav.group.configuration" },
  { id: "webhooks", labelKey: "nav.webhooks", groupKey: "nav.group.configuration" },
];

const I18nContext = createContext<{ lang: Lang; tr: TranslateFn; setLang: (lang: Lang) => void }>({
  lang: "en",
  tr: (key, vars) => t("en", key, vars),
  setLang: () => undefined,
});

function useI18n() {
  return useContext(I18nContext);
}

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

function Field({ label, value, onChange, type = "text", disabled = false, placeholder }: { label: string; value: string | number; onChange: (value: string) => void; type?: string; disabled?: boolean; placeholder?: string }) {
  return <label className="field"><span>{label}</span><input disabled={disabled} type={type} value={value} placeholder={placeholder} onChange={(e) => onChange(e.target.value)} /></label>;
}

function Integration({ notify }: { notify: (message: string) => void }) {
  const { tr } = useI18n();
  const [data, setData] = useState<Data | null>(null);
  const [uris, setUris] = useState<string[]>([]);
  const [origin, setOrigin] = useState("");
  useEffect(() => { request("/integration").then((value) => { setData(value); setUris(value.redirect_uris || [""]); setOrigin(value.social_callback_origin || ""); }).catch((e) => notify(e.message)); }, []);
  if (!data) return <Loading />;
  const copy = (value: string) => navigator.clipboard.writeText(value).then(() => notify(tr("common.copied")));
  const save = async () => { try { await request("/integration", { method: "PUT", body: JSON.stringify({ redirect_uris: uris.filter(Boolean), social_callback_origin: origin }) }); notify(tr("integration.saved")); } catch (e) { notify((e as Error).message); } };
  return <>
    <Card title={tr("integration.startTitle")} description={tr("integration.startDesc")}>
      <div className="checklist">
        <span>✓ {tr("integration.check1")}</span>
        <span>{uris.length ? "✓" : "○"} {tr("integration.check2")}</span>
        <span>○ {tr("integration.check3Before")}<a href="https://sooauth.com/docs/integrate-oidc#ai-prompt">{tr("integration.check3Link")}</a>{tr("integration.check3After")}</span>
        <span>○ {tr("integration.check4")}</span>
      </div>
      <a className="button primary inline" href="/auth/sign-in?return_to=/dashboard/">{tr("integration.testSignIn")}</a>
    </Card>
    <Card title={tr("integration.connectTitle")} description={tr("integration.connectDesc")}>
      <div className="data-grid">{[
        [tr("integration.issuer"), data.issuer],
        [tr("integration.clientId"), data.client_id],
        [tr("integration.discovery"), data.discovery_url],
      ].map(([label, value]) => <div className="copy-field" key={label as string}><span>{label}</span><div><code>{value}</code><Button onClick={() => copy(value as string)}>{tr("common.copy")}</Button></div></div>)}</div>
    </Card>
    <Card title={tr("integration.googleTitle")} description={tr("integration.googleDesc")}>
      {data.can_custom_social_callback ? <><Field label={tr("integration.appOrigin")} value={origin} onChange={setOrigin} /><Button primary onClick={save}>{tr("integration.saveGoogle")}</Button></> : <div className="notice">{tr("integration.freeNotice")}</div>}
    </Card>
    <Card title={tr("integration.redirectTitle")} description={tr("integration.redirectDesc")}>
      <div className="uri-list">{uris.map((uri, index) => <div className="uri-row" key={index}><input value={uri} onChange={(e) => setUris(uris.map((item, i) => i === index ? e.target.value : item))} placeholder="https://yourapp.com/api/auth/callback/oidc" /><Button onClick={() => setUris(uris.filter((_, i) => i !== index))}>×</Button></div>)}</div>
      <div className="actions"><Button onClick={() => setUris([...uris, ""])}>{tr("integration.addUri")}</Button><Button primary onClick={save}>{tr("integration.saveUris")}</Button></div>
    </Card>
  </>;
}

function ProviderCard({
  provider,
  isOpen,
  onToggleOpen,
  onUpdate,
  notify,
}: {
  provider: Data;
  isOpen: boolean;
  onToggleOpen: () => void;
  onUpdate: (provider: string, body: Data) => Promise<void>;
  notify: (message: string) => void;
}) {
  const { tr } = useI18n();
  const [selectedMode, setSelectedMode] = useState<"platform" | "custom">(
    provider.has_custom || !provider.platform_ready ? "custom" : "platform"
  );
  const [clientId, setClientId] = useState(provider.client_id || "");
  const [clientSecret, setClientSecret] = useState("");
  const [saving, setSaving] = useState(false);
  const [toggling, setToggling] = useState(false);

  useEffect(() => {
    if (provider.has_custom) {
      setSelectedMode("custom");
    } else if (provider.platform_ready) {
      setSelectedMode("platform");
    } else {
      setSelectedMode("custom");
    }
    setClientId(provider.client_id || "");
  }, [provider.has_custom, provider.platform_ready, provider.client_id]);

  const label = provider.provider ? provider.provider[0].toUpperCase() + provider.provider.slice(1) : "";
  const canToggle = Boolean(provider.has_custom || provider.platform_ready);

  const handleToggle = async (e: React.ChangeEvent<HTMLInputElement>) => {
    e.stopPropagation();
    const nextEnabled = e.target.checked;
    if (nextEnabled && !canToggle) {
      notify(tr("providers.needCustomFirst", { name: label }));
      if (!isOpen) onToggleOpen();
      return;
    }
    try {
      setToggling(true);
      await onUpdate(provider.provider, { enabled: nextEnabled });
    } finally {
      setToggling(false);
    }
  };

  const handleSaveCustom = async () => {
    if (!clientId.trim()) {
      notify(tr("providers.clientIdRequired", { name: label }));
      return;
    }
    if (!clientSecret && !provider.has_client_secret) {
      notify(tr("providers.clientSecretRequired", { name: label }));
      return;
    }
    try {
      setSaving(true);
      await onUpdate(provider.provider, {
        enabled: true,
        custom: true,
        client_id: clientId.trim(),
        client_secret: clientSecret,
      });
      setClientSecret("");
      notify(tr("providers.customSaved", { name: label }));
    } finally {
      setSaving(false);
    }
  };

  const handleRevertPlatform = async () => {
    try {
      setSaving(true);
      await onUpdate(provider.provider, {
        enabled: provider.enabled,
        use_platform: true,
      });
      setSelectedMode("platform");
      notify(tr("providers.switchedPlatform", { name: label }));
    } finally {
      setSaving(false);
    }
  };

  const copy = (val: string) => {
    navigator.clipboard.writeText(val).then(() => notify(tr("common.copied")));
  };

  const modeDesc = provider.has_custom
    ? tr("providers.modeCustom", { hint: provider.client_id_hint ? ` · ${provider.client_id_hint}` : "" })
    : provider.platform_ready
    ? tr("providers.modePlatform")
    : tr("providers.modeUnavailable");

  return (
    <article className={`provider ${isOpen ? "open" : ""}`}>
      <div
        className="provider-head"
        onClick={onToggleOpen}
        role="button"
        tabIndex={0}
        onKeyDown={(e) => {
          if (e.key === "Enter" || e.key === " ") {
            e.preventDefault();
            onToggleOpen();
          }
        }}
      >
        <div className="provider-info">
          <h2>
            {label}
            <span className={`status ${provider.enabled ? "on" : "off"}`}>
              {provider.enabled ? tr("providers.live") : tr("providers.off")}
            </span>
          </h2>
          <p>{modeDesc}</p>
        </div>
        <div className="provider-controls" onClick={(e) => e.stopPropagation()}>
          <label
            className={`switch ${!canToggle && !provider.enabled ? "disabled" : ""}`}
            title={
              canToggle
                ? provider.enabled
                  ? tr("providers.toggleOff")
                  : tr("providers.toggleOn")
                : tr("providers.toggleNeedCreds")
            }
          >
            <input
              type="checkbox"
              checked={Boolean(provider.enabled)}
              disabled={toggling}
              onChange={handleToggle}
              aria-label={tr("providers.toggleAria", { name: label })}
            />
            <span className="slider"></span>
          </label>
          <button
            type="button"
            className="chevron-btn"
            onClick={onToggleOpen}
            aria-label={isOpen ? tr("providers.collapse", { name: label }) : tr("providers.expand", { name: label })}
          >
            <span className="chevron">{isOpen ? "−" : "+"}</span>
          </button>
        </div>
      </div>

      {isOpen && (
        <div className="provider-body">
          <div className="mode-row">
            <label
              className={`mode-label ${selectedMode === "platform" ? "active" : ""} ${
                !provider.platform_ready ? "disabled" : ""
              }`}
            >
              <input
                type="radio"
                name={`provider-mode-${provider.provider}`}
                value="platform"
                checked={selectedMode === "platform"}
                disabled={!provider.platform_ready}
                onChange={() => setSelectedMode("platform")}
              />
              <span>{tr("providers.usePlatform")}</span>
              {!provider.platform_ready && <span className="mode-pill">{tr("providers.unavailable")}</span>}
            </label>
            <label className={`mode-label ${selectedMode === "custom" ? "active" : ""}`}>
              <input
                type="radio"
                name={`provider-mode-${provider.provider}`}
                value="custom"
                checked={selectedMode === "custom"}
                onChange={() => setSelectedMode("custom")}
              />
              <span>{tr("providers.customApp")}</span>
            </label>
          </div>

          {selectedMode === "platform" ? (
            <div className="platform-info-box">
              <p>
                {tr("providers.platformInfo")}
                {provider.provider === "google" && tr("providers.platformGoogleFree")}
              </p>
              {provider.has_custom && (
                <div style={{ marginTop: 14 }}>
                  <Button disabled={saving} onClick={handleRevertPlatform}>
                    {tr("providers.switchPlatform")}
                  </Button>
                </div>
              )}
            </div>
          ) : (
            <div className="custom-fields-panel">
              {provider.callback_url && (
                <div className="callback-box">
                  <span className="callback-label">
                    {tr("providers.callbackLabel", { name: label })}
                  </span>
                  <div className="copy-field">
                    <code>{provider.callback_url}</code>
                    <Button onClick={() => copy(provider.callback_url)}>{tr("common.copy")}</Button>
                  </div>
                </div>
              )}

              <Field
                label={tr("providers.clientId")}
                value={clientId}
                onChange={setClientId}
                placeholder={tr("providers.clientIdPlaceholder", { name: label })}
              />

              <div className="field-group">
                <Field
                  label={provider.has_client_secret ? tr("providers.clientSecretSaved") : tr("providers.clientSecret")}
                  type="password"
                  value={clientSecret}
                  onChange={setClientSecret}
                  placeholder={
                    provider.has_client_secret
                      ? tr("providers.secretKeepPlaceholder")
                      : tr("common.required")
                  }
                />
                {provider.has_client_secret && (
                  <p className="field-note">
                    {tr("providers.secretNote")}
                  </p>
                )}
              </div>

              <div className="actions" style={{ marginTop: 16, display: "flex", gap: 10, alignItems: "center" }}>
                <Button primary disabled={saving} onClick={handleSaveCustom}>
                  {saving ? tr("common.saving") : tr("providers.saveCustom")}
                </Button>
                {provider.has_custom && provider.platform_ready && (
                  <Button disabled={saving} onClick={handleRevertPlatform}>
                    {tr("providers.usePlatformInstead")}
                  </Button>
                )}
              </div>
            </div>
          )}
        </div>
      )}
    </article>
  );
}

function Providers({ notify }: { notify: (message: string) => void }) {
  const { tr } = useI18n();
  const [providers, setProviders] = useState<Data[]>([]);
  const [open, setOpen] = useState<string | null>(null);

  const load = () =>
    request("/providers")
      .then((data) => {
        const list = data.providers || [];
        setProviders(list);
        if (!open && list.length > 0) {
          setOpen(list[0].provider);
        }
      })
      .catch((e) => notify(e.message));

  useEffect(() => {
    load();
  }, []);

  const update = async (provider: string, body: Data) => {
    try {
      await request(`/providers/${provider}`, {
        method: "PUT",
        body: JSON.stringify(body),
      });
      load();
    } catch (e) {
      notify((e as Error).message);
      throw e;
    }
  };

  return (
    <>
      <div className="notice">
        {tr("providers.notice")}
      </div>
      <div className="provider-list">
        {providers.map((provider) => (
          <ProviderCard
            key={provider.provider}
            provider={provider}
            isOpen={open === provider.provider}
            onToggleOpen={() => setOpen(open === provider.provider ? null : provider.provider)}
            onUpdate={update}
            notify={notify}
          />
        ))}
      </div>
    </>
  );
}

function Overview() {
  const { tr } = useI18n();
  const [data, setData] = useState<Data | null>(null);
  useEffect(() => {
    Promise.all([request("/users"), request("/sessions"), request("/audit")])
      .then(([users, sessions, audit]) => setData({ users, sessions, audit }))
      .catch(() => setData({ users: {}, sessions: {}, audit: {} }));
  }, []);
  if (!data) return <Loading />;
  return (
    <>
      <div className="metrics">
        <Metric value={(data.users.users || []).length} label={tr("overview.users")} />
        <Metric value={(data.sessions.sessions || []).length} label={tr("overview.sessions")} />
        <Metric value={(data.audit.entries || []).length} label={tr("overview.audit")} />
      </div>
      <Card title={tr("overview.recent")} description={tr("overview.recentDesc")}>
        {data.audit.entries?.length ? (
          <Table
            headers={[tr("overview.col.time"), tr("overview.col.action"), tr("overview.col.ip")]}
            rows={data.audit.entries.slice(0, 8).map((item: Data) => [item.created_at, item.action, item.ip || tr("common.emDash")])}
          />
        ) : (
          <Empty text={tr("overview.empty")} />
        )}
      </Card>
    </>
  );
}

function Metric({ value, label }: { value: number; label: string }) {
  return <div className="metric"><strong>{value}</strong><span>{label}</span></div>;
}

function Table({ headers, rows }: { headers: string[]; rows: any[][] }) {
  return (
    <div className="table-wrap">
      <table>
        <thead><tr>{headers.map((header) => <th key={header}>{header}</th>)}</tr></thead>
        <tbody>{rows.map((row, i) => <tr key={i}>{row.map((cell, j) => <td key={j}>{String(cell)}</td>)}</tr>)}</tbody>
      </table>
    </div>
  );
}

function Empty({ text }: { text: string }) {
  return <div className="empty">{text}</div>;
}

function Loading() {
  const { tr } = useI18n();
  return <div className="loading">{tr("common.loading")}</div>;
}

function UserProfileMenu({ me, theme, onThemeChange }: { me: Data | null; theme: AppTheme; onThemeChange: (theme: AppTheme) => void }) {
  const { tr } = useI18n();
  const [open, setOpen] = useState(false);
  const menuRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const handleOutsideClick = (e: MouseEvent) => {
      if (menuRef.current && !menuRef.current.contains(e.target as Node)) {
        setOpen(false);
      }
    };
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Escape") setOpen(false);
    };
    if (open) {
      document.addEventListener("mousedown", handleOutsideClick);
      document.addEventListener("keydown", handleKeyDown);
    }
    return () => {
      document.removeEventListener("mousedown", handleOutsideClick);
      document.removeEventListener("keydown", handleKeyDown);
    };
  }, [open]);

  const handleSignOut = () => {
    fetch(`${API}/auth/sign-out`, {
      method: "POST",
      credentials: "include",
      headers: { "X-CSRF-Token": decodeURIComponent(csrf()) }
    }).then(() => {
      location.href = "/auth/sign-in";
    });
  };

  return (
    <div className="user-profile-menu-wrap" ref={menuRef}>
      <button
        type="button"
        className={`profile-card-trigger ${open ? "open" : ""}`}
        onClick={() => setOpen(!open)}
        aria-expanded={open}
        aria-haspopup="true"
        aria-label={tr("profile.ariaLabel")}
      >
        <div className="profile-identity">
          <span className="avatar">{(me?.email || "?").slice(0, 1).toUpperCase()}</span>
          <div className="profile-info">
            <strong title={me?.email}>{me?.email || tr("common.loadingAccount")}</strong>
            <span className="profile-badge">{tr("profile.platformAdmin")}</span>
          </div>
        </div>
        <svg className="profile-chevron" viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
          <polyline points="6 9 12 15 18 9"></polyline>
        </svg>
      </button>

      {open && (
        <div className="profile-popover" role="menu">
          <div className="popover-header">
            <strong>{me?.email}</strong>
            <small>{tr("profile.accountSettings")}</small>
          </div>

          <div className="popover-divider" />

          <div className="popover-section">
            <span className="popover-label">{tr("profile.appearance")}</span>
            <div className="theme-toggle-group">
              {(["light", "dark", "system"] as AppTheme[]).map((item) => (
                <button
                  key={item}
                  type="button"
                  className={`theme-pill ${theme === item ? "active" : ""}`}
                  onClick={() => { onThemeChange(item); setOpen(false); }}
                >
                  {item === "light" && tr("profile.themeLight")}
                  {item === "dark" && tr("profile.themeDark")}
                  {item === "system" && tr("profile.themeSystem")}
                </button>
              ))}
            </div>
          </div>

          <div className="popover-divider" />

          <div className="popover-links">
            <a href="/auth/account" className="popover-item">
              <span>{tr("common.securityCenter")}</span>
              <span>↗</span>
            </a>
            <a href="https://sooauth.com/docs" target="_blank" rel="noopener noreferrer" className="popover-item">
              <span>{tr("common.documentation")}</span>
              <span>↗</span>
            </a>
          </div>

          <div className="popover-divider" />

          <button type="button" className="popover-item danger" onClick={handleSignOut}>
            <span>{tr("profile.signOut")}</span>
            <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"></path>
              <polyline points="16 17 21 12 16 7"></polyline>
              <line x1="21" y1="12" x2="9" y2="12"></line>
            </svg>
          </button>
        </div>
      )}
    </div>
  );
}

function WebhookRow({ webhook, notify }: { webhook: Data; notify: (message: string) => void }) {
  const { tr } = useI18n();
  const [open, setOpen] = useState(false);
  const [deliveries, setDeliveries] = useState<Data[]>([]);
  const load = async () => {
    try {
      const result = await request(`/webhooks/${webhook.id}/deliveries`);
      setDeliveries(result.deliveries || []);
    } catch (e) {
      notify((e as Error).message);
    }
  };
  const test = async () => {
    try {
      await request(`/webhooks/${webhook.id}/test`, { method: "POST" });
      await load();
      notify(tr("webhooks.testOk"));
      setOpen(true);
    } catch (e) {
      notify(tr("webhooks.testFail", { message: (e as Error).message }));
    }
  };
  return (
    <article className="webhook-row">
      <div className="webhook-summary">
        <div>
          <strong>{webhook.url}</strong>
          <p>{(webhook.events || []).join(" · ") || tr("webhooks.allEvents")}</p>
        </div>
        <span className="status on">{webhook.active ? tr("common.active") : tr("common.inactive")}</span>
      </div>
      <div className="webhook-actions">
        <Button onClick={test}>{tr("webhooks.sendTest")}</Button>
        <Button onClick={async () => {
          if (!confirm(tr("webhooks.deleteConfirm"))) return;
          try {
            await request(`/webhooks/${webhook.id}`, { method: "DELETE" });
            notify(tr("webhooks.deleted"));
          } catch (e) {
            notify((e as Error).message);
          }
        }}>{tr("common.delete")}</Button>
        <Button onClick={() => { setOpen(!open); if (!open) load(); }}>
          {open ? tr("webhooks.hideDeliveries") : tr("webhooks.viewLog")}
        </Button>
      </div>
      {open && (
        <div className="delivery-list">
          {deliveries.length ? deliveries.map((delivery) => (
            <details className="delivery" key={delivery.id}>
              <summary>
                <span>{delivery.event}</span>
                <span>{delivery.status_code || tr("webhooks.failed")} · {new Date(delivery.created_at).toLocaleString()}</span>
              </summary>
              <div className="delivery-grid">
                <div><b>{tr("webhooks.requestBody")}</b><pre>{delivery.payload || "{}"}</pre></div>
                <div><b>{tr("webhooks.response")}</b><pre>{delivery.response || delivery.error || tr("webhooks.noResponse")}</pre></div>
              </div>
            </details>
          )) : <Empty text={tr("webhooks.noDeliveries")} />}
        </div>
      )}
    </article>
  );
}

function ResourcePage({ tab, notify }: { tab: Tab; notify: (message: string) => void }) {
  const { tr } = useI18n();
  const [data, setData] = useState<Data | null>(null);
  const [webhookURL, setWebhookURL] = useState("");
  const [webhookEvents, setWebhookEvents] = useState(["user.created", "user.login", "user.login_failed", "user.email_verified", "password.reset", "session.revoked"]);
  const endpoint = tab === "users" ? "/users" : tab === "sessions" ? "/sessions" : tab === "audit" ? "/audit" : "/webhooks";
  useEffect(() => { request(endpoint).then(setData).catch((e) => notify(e.message)); }, [endpoint]);
  if (!data) return <Loading />;
  if (tab === "users") {
    return (
      <Card title={tr("users.title", { n: (data.users || []).length })} description={tr("users.desc")}>
        {data.users?.length ? (
          <div className="table-wrap">
            <table>
              <thead>
                <tr>
                  {[tr("users.col.email"), tr("users.col.method"), tr("users.col.verified"), tr("users.col.status"), tr("users.col.joined"), ""].map((h, i) => <th key={h || `empty-${i}`}>{h}</th>)}
                </tr>
              </thead>
              <tbody>
                {data.users.map((u: Data) => (
                  <tr key={u.id}>
                    <td>
                      <div>
                        <strong>{u.email || u.username || u.phone || tr("common.emDash")}</strong>
                      </div>
                      {(u.username || u.phone) && (u.email ? (
                        <div style={{ fontSize: "11px", color: "var(--muted)", marginTop: "2px" }}>
                          {u.username && <span style={{ marginRight: "6px" }}>@{u.username}</span>}
                          {u.phone && <span>{u.phone}</span>}
                        </div>
                      ) : null)}
                    </td>
                    <td>{u.signup_method || tr("common.emDash")}</td>
                    <td>{u.email_verified || u.phone_verified_at ? tr("common.yes") : tr("common.no")}</td>
                    <td>{u.disabled ? tr("common.disabled") : tr("common.active")}</td>
                    <td>{u.created_at}</td>
                    <td>
                      <Button onClick={async () => {
                        try {
                          await request(`/users/${u.id}/${u.disabled ? "enable" : "disable"}`, { method: "POST" });
                          notify(tr("users.statusUpdated"));
                          setData({ ...data, users: data.users.map((item: Data) => item.id === u.id ? { ...item, disabled: !item.disabled } : item) });
                        } catch (e) {
                          notify((e as Error).message);
                        }
                      }}>{u.disabled ? tr("common.enable") : tr("common.disable")}</Button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        ) : <Empty text={tr("users.empty")} />}
      </Card>
    );
  }
  if (tab === "sessions") {
    return (
      <Card title={tr("sessions.title")} description={tr("sessions.desc")}>
        {data.sessions?.length ? (
          <div className="table-wrap">
            <table>
              <thead>
                <tr>{[tr("sessions.col.user"), tr("sessions.col.ip"), tr("sessions.col.expires"), ""].map((h, i) => <th key={h || `empty-${i}`}>{h}</th>)}</tr>
              </thead>
              <tbody>
                {data.sessions.map((s: Data) => (
                  <tr key={s.id}>
                    <td>{s.email}</td>
                    <td>{s.ip || tr("common.emDash")}</td>
                    <td>{s.expires_at}</td>
                    <td>
                      <Button onClick={async () => {
                        try {
                          await request(`/sessions/${s.id}/revoke`, { method: "POST" });
                          notify(tr("sessions.revoked"));
                          setData({ ...data, sessions: data.sessions.filter((item: Data) => item.id !== s.id) });
                        } catch (e) {
                          notify((e as Error).message);
                        }
                      }}>{tr("sessions.revoke")}</Button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        ) : <Empty text={tr("sessions.empty")} />}
      </Card>
    );
  }
  if (tab === "audit") {
    return (
      <Card title={tr("audit.title")}>
        {data.entries?.length ? (
          <Table
            headers={[tr("overview.col.time"), tr("overview.col.action"), tr("overview.col.ip")]}
            rows={data.entries.map((e: Data) => [e.created_at, e.action, e.ip || tr("common.emDash")])}
          />
        ) : <Empty text={tr("audit.empty")} />}
      </Card>
    );
  }
  const createWebhook = async () => {
    try {
      const result = await request("/webhooks", { method: "POST", body: JSON.stringify({ url: webhookURL, events: webhookEvents }) });
      setWebhookURL("");
      setData({ ...data, webhooks: [...(data.webhooks || []), { url: webhookURL, events: webhookEvents, active: true, id: result.id }] });
      notify(tr("webhooks.created"));
    } catch (e) {
      notify((e as Error).message);
    }
  };
  return (
    <>
      <Card title={tr("webhooks.addTitle")} description={tr("webhooks.addDesc")}>
        <div className="form-narrow">
          <Field label={tr("webhooks.endpointUrl")} value={webhookURL} onChange={setWebhookURL} />
          <div className="event-grid">
            {["user.created", "user.login", "user.login_failed", "user.email_verified", "password.reset", "session.revoked"].map((event) => (
              <label className="check" key={event}>
                <input
                  type="checkbox"
                  checked={webhookEvents.includes(event)}
                  onChange={() => setWebhookEvents(webhookEvents.includes(event) ? webhookEvents.filter((item) => item !== event) : [...webhookEvents, event])}
                /> {event}
              </label>
            ))}
          </div>
          <Button primary onClick={createWebhook}>{tr("webhooks.create")}</Button>
        </div>
      </Card>
      <Card title={tr("webhooks.listTitle", { n: (data.webhooks || []).length })} description={tr("webhooks.listDesc")}>
        {data.webhooks?.length ? (
          <div className="webhook-list">
            {data.webhooks.map((webhook: Data) => <WebhookRow key={webhook.id} webhook={webhook} notify={notify} />)}
          </div>
        ) : <Empty text={tr("webhooks.empty")} />}
      </Card>
    </>
  );
}

function Settings({ tab, notify }: { tab: "project" | "theme" | "billing"; notify: (message: string) => void }) {
  const { tr } = useI18n();
  const [data, setData] = useState<Data | null>(null);
  const [draft, setDraft] = useState<Data>({});
  useEffect(() => {
    request(tab === "project" ? "/project" : tab === "theme" ? "/theme" : "/billing")
      .then((value) => { setData(value); setDraft(value); })
      .catch((e) => notify(e.message));
  }, [tab]);
  if (!data) return <Loading />;
  if (tab === "billing") {
    return (
      <>
        <Card title={tr("billing.currentTitle")} description={tr("billing.currentDesc", { plan: data.plan_label, projects: data.max_projects, status: data.status })} />
        {data.public_beta && (
          <div className="notice" style={{ margin: "16px 0", padding: "12px 16px", borderRadius: "8px", background: "var(--accent-soft)", borderLeft: "3px solid var(--accent)" }}>
            🚀 <strong>{tr("billing.betaNotice")}</strong>{tr("billing.betaBody")}
          </div>
        )}
        <div className="plans">
          {(data.plans || []).map((plan: Data) => (
            <Card key={plan.id} title={plan.label} className={plan.id === data.plan ? "current-plan" : ""}>
              <strong className="price">
                {data.public_beta
                  ? (plan.id === "free" ? tr("billing.forever") : tr("billing.publicBeta"))
                  : plan.price_try ? `${plan.price_try.amount_cents / 100} TRY` : tr("billing.free")}
              </strong>
              <ul>{(plan.features || []).map((feature: string) => <li key={feature}>{feature}</li>)}</ul>
              <Button primary={plan.id !== data.plan}>
                {plan.id === data.plan ? tr("billing.currentPlan") : data.public_beta ? tr("billing.includedBeta") : tr("billing.upgrade")}
              </Button>
            </Card>
          ))}
        </div>
      </>
    );
  }
  const project = tab === "project";
  const save = async () => {
    try {
      await request(project ? "/project" : "/theme", {
        method: "PUT",
        body: JSON.stringify(project ? {
          name: draft.name,
          password_min_length: Number(draft.password_min_length || 8),
          email_verify_required: Boolean(draft.email_verify_required),
          email_verify_delivery: draft.email_verify_delivery || "link",
          password_require_upper: Boolean(draft.password_require_upper),
          password_require_number: Boolean(draft.password_require_number),
          password_require_special: Boolean(draft.password_require_special),
          default_locale: draft.default_locale || "en",
        } : {
          brand_name: draft.brand_name,
          logo_url: draft.logo_url,
          accent_color: draft.accent_color,
        }),
      });
      notify(tr("common.settingsSaved"));
    } catch (e) {
      notify((e as Error).message);
    }
  };
  return (
    <Card
      title={project ? tr("project.title") : tr("theme.title")}
      description={project ? tr("project.desc") : tr("theme.desc")}
    >
      <div className={project ? "settings-layout" : "form-narrow"}>
        <div className="form-narrow">
          <Field
            label={project ? tr("project.name") : tr("theme.brandName")}
            value={project ? draft.name || "" : draft.brand_name || "sooauth"}
            onChange={(value) => setDraft({ ...draft, [project ? "name" : "brand_name"]: value })}
          />
          {project ? (
            <>
              <label className="check">
                <input type="checkbox" checked={Boolean(draft.email_verify_required)} onChange={(e) => setDraft({ ...draft, email_verify_required: e.target.checked })} />
                {tr("project.requireVerify")}
              </label>
              <label className="field">
                <span>{tr("project.delivery")}</span>
                <select value={draft.email_verify_delivery || "link"} onChange={(e) => setDraft({ ...draft, email_verify_delivery: e.target.value })}>
                  <option value="link">{tr("project.deliveryLink")}</option>
                  <option value="code">{tr("project.deliveryCode")}</option>
                  <option value="both">{tr("project.deliveryBoth")}</option>
                </select>
              </label>
              <label className="field">
                <span>{tr("project.defaultLocale")}</span>
                <select value={draft.default_locale || "en"} onChange={(e) => setDraft({ ...draft, default_locale: e.target.value })}>
                  <option value="en">{tr("project.localeEn")}</option>
                  <option value="tr">{tr("project.localeTr")}</option>
                </select>
              </label>
              <Field label={tr("project.minPassword")} type="number" value={draft.password_min_length || 8} onChange={(value) => setDraft({ ...draft, password_min_length: value })} />
              <label className="check">
                <input type="checkbox" checked={Boolean(draft.password_require_upper)} onChange={(e) => setDraft({ ...draft, password_require_upper: e.target.checked })} />
                {tr("project.requireUpper")}
              </label>
              <label className="check">
                <input type="checkbox" checked={Boolean(draft.password_require_number)} onChange={(e) => setDraft({ ...draft, password_require_number: e.target.checked })} />
                {tr("project.requireNumber")}
              </label>
              <label className="check">
                <input type="checkbox" checked={Boolean(draft.password_require_special)} onChange={(e) => setDraft({ ...draft, password_require_special: e.target.checked })} />
                {tr("project.requireSpecial")}
              </label>
            </>
          ) : (
            <Field label={tr("theme.logoUrl")} value={draft.logo_url || ""} onChange={(value) => setDraft({ ...draft, logo_url: value })} />
          )}
          <Button primary onClick={save}>{tr("common.saveChanges")}</Button>
        </div>
        {project && (
          <aside className="rule-card">
            <span>{tr("project.ruleTitle")}</span>
            <strong>{tr("project.ruleAtLeast", { n: draft.password_min_length || 8 })}</strong>
            <p>
              {draft.password_require_upper ? tr("project.ruleUpper") : ""}
              {draft.password_require_number ? tr("project.ruleNumber") : ""}
              {draft.password_require_special ? tr("project.ruleSpecial") : tr("project.ruleNone")}
            </p>
            <small>{tr("project.ruleUsersSee")}</small>
          </aside>
        )}
      </div>
    </Card>
  );
}

type SelectOption = { label: string; value: string };
type OptionsSource = {
  type: "static" | "dynamic_api";
  options?: SelectOption[];
  url?: string;
  method?: string;
  headers?: Record<string, string>;
  items_path?: string;
  label_key?: string;
  value_key?: string;
  cache_ttl_seconds?: number;
};
type RegistrationField = {
  id: string;
  label: string;
  type: "text" | "textarea" | "number" | "select" | "checkbox";
  required: boolean;
  placeholder?: string;
  description?: string;
  options_source?: OptionsSource;
};

function FormsBuilder({ notify }: { notify: (message: string) => void }) {
  const { tr } = useI18n();
  const [data, setData] = useState<Data | null>(null);
  const [allowedEmail, setAllowedEmail] = useState(true);
  const [allowedUsername, setAllowedUsername] = useState(false);
  const [allowedPhone, setAllowedPhone] = useState(false);
  const [primaryAuthMode, setPrimaryAuthMode] = useState<"password" | "otp" | "both">("password");
  const [fields, setFields] = useState<RegistrationField[]>([]);
  const [saving, setSaving] = useState(false);
  const [testResult, setTestResult] = useState<Record<string, { loading?: boolean; count?: number; error?: string }>>({});

  useEffect(() => {
    request("/project")
      .then((val) => {
        setData(val);
        const authCfg = val.auth_config || {};
        const allowed = authCfg.allowed_identifiers || ["email"];
        setAllowedEmail(allowed.includes("email"));
        setAllowedUsername(allowed.includes("username"));
        setAllowedPhone(allowed.includes("phone"));
        setPrimaryAuthMode(authCfg.primary_auth_mode || "password");
        setFields(val.registration_schema || []);
      })
      .catch((e) => notify(e.message));
  }, []);

  if (!data) return <Loading />;

  const save = async () => {
    const allowed = [];
    if (allowedEmail) allowed.push("email");
    if (allowedUsername) allowed.push("username");
    if (allowedPhone) allowed.push("phone");
    if (allowed.length === 0) {
      notify("Please select at least one sign-in identifier.");
      return;
    }

    try {
      setSaving(true);
      await request("/project", {
        method: "PUT",
        body: JSON.stringify({
          auth_config: {
            allowed_identifiers: allowed,
            primary_auth_mode: primaryAuthMode,
          },
          registration_schema: fields,
        }),
      });
      notify(tr("common.settingsSaved"));
    } catch (e) {
      notify((e as Error).message);
    } finally {
      setSaving(false);
    }
  };

  const addField = () => {
    const newId = "field_" + Date.now().toString(36);
    setFields([
      ...fields,
      {
        id: newId,
        label: "New Field",
        type: "text",
        required: false,
        placeholder: "",
        description: "",
      },
    ]);
  };

  const removeField = (index: number) => {
    setFields(fields.filter((_, i) => i !== index));
  };

  const updateField = (index: number, patch: Partial<RegistrationField>) => {
    setFields(fields.map((f, i) => (i === index ? { ...f, ...patch } : f)));
  };

  const testApi = async (index: number, src: OptionsSource) => {
    const field = fields[index];
    setTestResult((prev) => ({ ...prev, [field.id]: { loading: true } }));
    try {
      const res = await request("/registration-schema/test-api", {
        method: "POST",
        body: JSON.stringify(src),
      });
      setTestResult((prev) => ({ ...prev, [field.id]: { count: res.count } }));
    } catch (e) {
      setTestResult((prev) => ({ ...prev, [field.id]: { error: (e as Error).message } }));
    }
  };

  return (
    <>
      <Card title={tr("forms.identifiersTitle")} description={tr("forms.identifiersDesc")}>
        <div style={{ display: "grid", gap: "16px", maxWidth: "600px" }}>
          <div style={{ display: "flex", gap: "24px", flexWrap: "wrap" }}>
            <label className="check" style={{ display: "flex", alignItems: "center", gap: "8px" }}>
              <input
                type="checkbox"
                checked={allowedEmail}
                onChange={(e) => setAllowedEmail(e.target.checked)}
              />
              <span>{tr("forms.idEmail")}</span>
            </label>
            <label className="check" style={{ display: "flex", alignItems: "center", gap: "8px" }}>
              <input
                type="checkbox"
                checked={allowedUsername}
                onChange={(e) => setAllowedUsername(e.target.checked)}
              />
              <span>{tr("forms.idUsername")}</span>
            </label>
            <label className="check" style={{ display: "flex", alignItems: "center", gap: "8px" }}>
              <input
                type="checkbox"
                checked={allowedPhone}
                onChange={(e) => setAllowedPhone(e.target.checked)}
              />
              <span>{tr("forms.idPhone")}</span>
            </label>
          </div>

          <div style={{ marginTop: "12px" }}>
            <span style={{ display: "block", fontSize: "14px", fontWeight: 600, marginBottom: "8px" }}>
              {tr("forms.authModeTitle")}
            </span>
            <p style={{ fontSize: "13px", color: "var(--text-muted)", margin: "0 0 12px" }}>
              {tr("forms.authModeDesc")}
            </p>
            <div style={{ display: "grid", gap: "10px" }}>
              <label style={{ display: "flex", alignItems: "center", gap: "8px", cursor: "pointer" }}>
                <input
                  type="radio"
                  name="auth_mode"
                  value="password"
                  checked={primaryAuthMode === "password"}
                  onChange={() => setPrimaryAuthMode("password")}
                />
                <span>{tr("forms.modePassword")}</span>
              </label>
              <label style={{ display: "flex", alignItems: "center", gap: "8px", cursor: "pointer" }}>
                <input
                  type="radio"
                  name="auth_mode"
                  value="otp"
                  checked={primaryAuthMode === "otp"}
                  onChange={() => setPrimaryAuthMode("otp")}
                />
                <span>{tr("forms.modeOtp")}</span>
              </label>
              <label style={{ display: "flex", alignItems: "center", gap: "8px", cursor: "pointer" }}>
                <input
                  type="radio"
                  name="auth_mode"
                  value="both"
                  checked={primaryAuthMode === "both"}
                  onChange={() => setPrimaryAuthMode("both")}
                />
                <span>{tr("forms.modeBoth")}</span>
              </label>
            </div>
          </div>
        </div>
      </Card>

      <Card title={tr("forms.schemaTitle")} description={tr("forms.schemaDesc")}>
        <div style={{ display: "grid", gap: "20px" }}>
          {fields.length === 0 ? (
            <div className="notice" style={{ padding: "16px", borderRadius: "8px" }}>
              {tr("forms.noFields")}
            </div>
          ) : (
            fields.map((field, idx) => {
              const src = field.options_source || { type: "static", options: [] };
              const tRes = testResult[field.id];

              return (
                <div
                  key={field.id}
                  style={{
                    border: "1px solid var(--border)",
                    borderRadius: "10px",
                    padding: "16px",
                    background: "var(--card-bg, rgba(255,255,255,0.02))",
                    display: "grid",
                    gap: "14px",
                  }}
                >
                  <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center" }}>
                    <div style={{ display: "flex", alignItems: "center", gap: "10px" }}>
                      <strong style={{ fontSize: "15px" }}>{field.label || field.id}</strong>
                      <span
                        style={{
                          fontSize: "11px",
                          padding: "2px 8px",
                          borderRadius: "4px",
                          background: "var(--accent-soft, #eee)",
                          color: "var(--accent, #333)",
                          fontWeight: 600,
                        }}
                      >
                        {field.type.toUpperCase()}
                      </span>
                      {field.required && (
                        <span
                          style={{
                            fontSize: "11px",
                            padding: "2px 8px",
                            borderRadius: "4px",
                            background: "rgba(255,59,59,0.1)",
                            color: "#FF3B3B",
                            fontWeight: 600,
                          }}
                        >
                          {tr("common.required")}
                        </span>
                      )}
                    </div>
                    <Button onClick={() => removeField(idx)}>{tr("forms.removeField")}</Button>
                  </div>

                  <div style={{ display: "grid", gridTemplateColumns: "repeat(auto-fit, minmax(200px, 1fr))", gap: "12px" }}>
                    <Field
                      label={tr("forms.fieldLabel")}
                      value={field.label}
                      onChange={(v) => updateField(idx, { label: v })}
                    />
                    <Field
                      label={tr("forms.fieldId")}
                      value={field.id}
                      onChange={(v) => updateField(idx, { id: v.toLowerCase().replace(/[^a-z0-9_]/g, "_") })}
                    />
                    <label className="field">
                      <span>{tr("forms.fieldType")}</span>
                      <select
                        value={field.type}
                        onChange={(e) => {
                          const nextType = e.target.value as RegistrationField["type"];
                          const patch: Partial<RegistrationField> = { type: nextType };
                          if (nextType === "select" && !field.options_source) {
                            patch.options_source = { type: "static", options: [] };
                          }
                          updateField(idx, patch);
                        }}
                      >
                        <option value="text">{tr("forms.typeText")}</option>
                        <option value="textarea">{tr("forms.typeTextarea")}</option>
                        <option value="number">{tr("forms.typeNumber")}</option>
                        <option value="select">{tr("forms.typeSelect")}</option>
                        <option value="checkbox">{tr("forms.typeCheckbox")}</option>
                      </select>
                    </label>
                  </div>

                  <div style={{ display: "grid", gridTemplateColumns: "repeat(auto-fit, minmax(200px, 1fr))", gap: "12px" }}>
                    <Field
                      label={tr("forms.fieldPlaceholder")}
                      value={field.placeholder || ""}
                      onChange={(v) => updateField(idx, { placeholder: v })}
                    />
                    <Field
                      label={tr("forms.fieldDesc")}
                      value={field.description || ""}
                      onChange={(v) => updateField(idx, { description: v })}
                    />
                    <label className="check" style={{ display: "flex", alignItems: "center", gap: "8px", marginTop: "24px" }}>
                      <input
                        type="checkbox"
                        checked={field.required}
                        onChange={(e) => updateField(idx, { required: e.target.checked })}
                      />
                      <span>{tr("forms.fieldRequired")}</span>
                    </label>
                  </div>

                  {field.type === "select" && (
                    <div
                      style={{
                        marginTop: "8px",
                        padding: "12px",
                        borderRadius: "8px",
                        border: "1px dashed var(--border)",
                        background: "rgba(0,0,0,0.02)",
                      }}
                    >
                      <label className="field" style={{ marginBottom: "12px" }}>
                        <span>{tr("forms.optionsSource")}</span>
                        <select
                          value={src.type || "static"}
                          onChange={(e) => {
                            const nextSrcType = e.target.value as "static" | "dynamic_api";
                            updateField(idx, {
                              options_source: {
                                ...src,
                                type: nextSrcType,
                              },
                            });
                          }}
                        >
                          <option value="static">{tr("forms.sourceStatic")}</option>
                          <option value="dynamic_api">{tr("forms.sourceDynamic")}</option>
                        </select>
                      </label>

                      {src.type === "dynamic_api" ? (
                        <div style={{ display: "grid", gap: "10px" }}>
                          <Field
                            label={tr("forms.apiUrl")}
                            value={src.url || ""}
                            placeholder="https://api.riotgames.com/lol/ranked/v4/entries/by-summoner/..."
                            onChange={(v) =>
                              updateField(idx, {
                                options_source: { ...src, url: v },
                              })
                            }
                          />
                          <Field
                            label={tr("forms.apiHeaders") + " (JSON)"}
                            value={src.headers ? JSON.stringify(src.headers) : ""}
                            placeholder='{"X-Riot-Token": "RGAPI-..."}'
                            onChange={(v) => {
                              try {
                                const parsed = v.trim() ? JSON.parse(v) : {};
                                updateField(idx, {
                                  options_source: { ...src, headers: parsed },
                                });
                              } catch {
                                /* ignore invalid json while typing */
                              }
                            }}
                          />
                          <div style={{ display: "grid", gridTemplateColumns: "repeat(auto-fit, minmax(140px, 1fr))", gap: "10px" }}>
                            <Field
                              label={tr("forms.itemsPath")}
                              value={src.items_path || ""}
                              placeholder="data.ranks"
                              onChange={(v) =>
                                updateField(idx, {
                                  options_source: { ...src, items_path: v },
                                })
                              }
                            />
                            <Field
                              label={tr("forms.labelKey")}
                              value={src.label_key || ""}
                              placeholder="name or tier"
                              onChange={(v) =>
                                updateField(idx, {
                                  options_source: { ...src, label_key: v },
                                })
                              }
                            />
                            <Field
                              label={tr("forms.valueKey")}
                              value={src.value_key || ""}
                              placeholder="id or tier"
                              onChange={(v) =>
                                updateField(idx, {
                                  options_source: { ...src, value_key: v },
                                })
                              }
                            />
                            <Field
                              label={tr("forms.cacheTtl")}
                              type="number"
                              value={src.cache_ttl_seconds || 300}
                              onChange={(v) =>
                                updateField(idx, {
                                  options_source: { ...src, cache_ttl_seconds: Number(v) || 300 },
                                })
                              }
                            />
                          </div>
                          <div style={{ display: "flex", alignItems: "center", gap: "12px", marginTop: "6px" }}>
                            <Button
                              onClick={() => testApi(idx, src)}
                              disabled={tRes?.loading || !src.url}
                            >
                              {tRes?.loading ? tr("forms.testingApi") : tr("forms.testApi")}
                            </Button>
                            {tRes?.count !== undefined && (
                              <span style={{ color: "#2E7D32", fontSize: "13px", fontWeight: 600 }}>
                                ✓ {tr("forms.testSuccess", { n: tRes.count })}
                              </span>
                            )}
                            {tRes?.error && (
                              <span style={{ color: "#D32F2F", fontSize: "13px" }}>
                                ✗ {tr("forms.testFailed", { err: tRes.error })}
                              </span>
                            )}
                          </div>
                        </div>
                      ) : (
                        <div style={{ display: "grid", gap: "8px" }}>
                          <span style={{ fontSize: "13px", fontWeight: 600 }}>{tr("forms.staticOptions")}</span>
                          {(src.options || []).map((opt, optIdx) => (
                            <div key={optIdx} style={{ display: "flex", gap: "8px", alignItems: "center" }}>
                              <input
                                placeholder={tr("forms.optLabel")}
                                value={opt.label}
                                onChange={(e) => {
                                  const nextOpts = [...(src.options || [])];
                                  nextOpts[optIdx] = { ...opt, label: e.target.value };
                                  updateField(idx, { options_source: { ...src, options: nextOpts } });
                                }}
                                style={{ flex: 1, padding: "8px", borderRadius: "6px", border: "1px solid var(--border)" }}
                              />
                              <input
                                placeholder={tr("forms.optValue")}
                                value={opt.value}
                                onChange={(e) => {
                                  const nextOpts = [...(src.options || [])];
                                  nextOpts[optIdx] = { ...opt, value: e.target.value };
                                  updateField(idx, { options_source: { ...src, options: nextOpts } });
                                }}
                                style={{ flex: 1, padding: "8px", borderRadius: "6px", border: "1px solid var(--border)" }}
                              />
                              <Button
                                onClick={() => {
                                  const nextOpts = (src.options || []).filter((_, i) => i !== optIdx);
                                  updateField(idx, { options_source: { ...src, options: nextOpts } });
                                }}
                              >
                                ×
                              </Button>
                            </div>
                          ))}
                          <div>
                            <Button
                              onClick={() => {
                                const nextOpts = [...(src.options || []), { label: "", value: "" }];
                                updateField(idx, { options_source: { ...src, options: nextOpts } });
                              }}
                            >
                              {tr("forms.addOption")}
                            </Button>
                          </div>
                        </div>
                      )}
                    </div>
                  )}
                </div>
              );
            })
          )}

          <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginTop: "10px" }}>
            <Button onClick={addField}>{tr("forms.addField")}</Button>
            <Button primary onClick={save} disabled={saving}>
              {saving ? tr("common.saving") : tr("common.saveChanges")}
            </Button>
          </div>
        </div>
      </Card>
    </>
  );
}

function EmailSettings({ notify }: { notify: (message: string) => void }) {
  const { tr } = useI18n();
  const [data, setData] = useState<Data | null>(null);
  const [draft, setDraft] = useState<Data>({});
  const [testing, setTesting] = useState(false);
  const [testTo, setTestTo] = useState("");

  const load = () => {
    request("/email")
      .then((val) => {
        setData(val);
        setDraft(val);
      })
      .catch((e) => notify(e.message));
  };

  useEffect(() => {
    load();
  }, []);

  if (!data) return <Loading />;

  const provider = draft.provider || "smtp";

  const save = async () => {
    try {
      await request("/email", {
        method: "PUT",
        body: JSON.stringify({
          enabled: Boolean(draft.enabled),
          provider: draft.provider || "smtp",
          from_name: draft.from_name || "",
          from_email: draft.from_email || "",
          reply_to: draft.reply_to || "",
          smtp_host: draft.smtp_host || "",
          smtp_port: Number(draft.smtp_port || 587),
          smtp_user: draft.smtp_user || "",
          smtp_password: draft.smtp_password || undefined,
          smtp_tls_mode: draft.smtp_tls_mode || "starttls",
          api_key: draft.api_key || undefined,
          aws_access_key_id: draft.aws_access_key_id || "",
          aws_secret_key: draft.aws_secret_key || undefined,
          aws_region: draft.aws_region || "us-east-1",
        }),
      });
      notify(tr("email.saved"));
      load();
    } catch (e) {
      notify((e as Error).message);
    }
  };

  const sendTest = async () => {
    setTesting(true);
    try {
      const res = await request("/email/test", {
        method: "POST",
        body: JSON.stringify({
          to: testTo.trim() || undefined,
          provider: draft.provider || "smtp",
          from_name: draft.from_name || "",
          from_email: draft.from_email || "",
          reply_to: draft.reply_to || "",
          smtp_host: draft.smtp_host || "",
          smtp_port: Number(draft.smtp_port || 587),
          smtp_user: draft.smtp_user || "",
          smtp_password: draft.smtp_password || undefined,
          smtp_tls_mode: draft.smtp_tls_mode || "starttls",
          api_key: draft.api_key || undefined,
          aws_access_key_id: draft.aws_access_key_id || "",
          aws_secret_key: draft.aws_secret_key || undefined,
          aws_region: draft.aws_region || "us-east-1",
        }),
      });
      notify(res.message || tr("email.testSent"));
    } catch (e) {
      notify(tr("email.testFailed", { message: (e as Error).message }));
    } finally {
      setTesting(false);
    }
  };

  return (
    <>
      <Card title={tr("email.title")} description={tr("email.desc")}>
        <div className="form-narrow">
          <label className="check">
            <input
              type="checkbox"
              checked={Boolean(draft.enabled)}
              onChange={(e) => setDraft({ ...draft, enabled: e.target.checked })}
            />
            <strong>{tr("email.enable")}</strong>
          </label>

          {!draft.enabled && (
            <div className="notice">
              {tr("email.disabledNotice", { from: data.platform_from || tr("email.platformDefault") })}
            </div>
          )}

          <label className="field">
            <span>{tr("email.provider")}</span>
            <select
              value={provider}
              onChange={(e) => setDraft({ ...draft, provider: e.target.value })}
            >
              <option value="smtp">{tr("email.providerSmtp")}</option>
              <option value="resend">{tr("email.providerResend")}</option>
              <option value="postmark">{tr("email.providerPostmark")}</option>
              <option value="ses">{tr("email.providerSes")}</option>
            </select>
          </label>

          <Field
            label={tr("email.fromName")}
            value={draft.from_name || ""}
            onChange={(v) => setDraft({ ...draft, from_name: v })}
            placeholder="e.g. Acme Security"
          />
          <Field
            label={tr("email.fromEmail")}
            type="email"
            value={draft.from_email || ""}
            onChange={(v) => setDraft({ ...draft, from_email: v })}
            placeholder="e.g. auth@acme.com"
          />
          <Field
            label={tr("email.replyTo")}
            type="email"
            value={draft.reply_to || ""}
            onChange={(v) => setDraft({ ...draft, reply_to: v })}
            placeholder="e.g. support@acme.com"
          />

          {provider === "smtp" && (
            <>
              <Field
                label={tr("email.smtpHost")}
                value={draft.smtp_host || ""}
                onChange={(v) => setDraft({ ...draft, smtp_host: v })}
                placeholder="smtp.example.com"
              />
              <Field
                label={tr("email.smtpPort")}
                type="number"
                value={draft.smtp_port || 587}
                onChange={(v) => setDraft({ ...draft, smtp_port: v })}
              />
              <Field
                label={tr("email.smtpUser")}
                value={draft.smtp_user || ""}
                onChange={(v) => setDraft({ ...draft, smtp_user: v })}
                placeholder="username or API user"
              />
              <Field
                label={data.smtp_has_password ? tr("email.smtpPasswordSaved") : tr("email.smtpPassword")}
                type="password"
                value={draft.smtp_password || ""}
                onChange={(v) => setDraft({ ...draft, smtp_password: v })}
                placeholder={data.smtp_has_password ? "••••••••••••" : tr("email.enterPassword")}
              />
              <label className="field">
                <span>{tr("email.tlsMode")}</span>
                <select
                  value={draft.smtp_tls_mode || "starttls"}
                  onChange={(e) => setDraft({ ...draft, smtp_tls_mode: e.target.value })}
                >
                  <option value="starttls">{tr("email.tlsStarttls")}</option>
                  <option value="implicit">{tr("email.tlsImplicit")}</option>
                  <option value="none">{tr("email.tlsNone")}</option>
                </select>
              </label>
            </>
          )}

          {provider === "resend" && (
            <Field
              label={data.has_api_key ? tr("email.resendKeySaved") : tr("email.resendKey")}
              type="password"
              value={draft.api_key || ""}
              onChange={(v) => setDraft({ ...draft, api_key: v })}
              placeholder={data.has_api_key ? "re_••••••••••••" : "re_123456789"}
            />
          )}

          {provider === "postmark" && (
            <Field
              label={data.has_api_key ? tr("email.postmarkTokenSaved") : tr("email.postmarkToken")}
              type="password"
              value={draft.api_key || ""}
              onChange={(v) => setDraft({ ...draft, api_key: v })}
              placeholder={data.has_api_key ? "••••••••-••••-••••" : "Your Postmark server token"}
            />
          )}

          {provider === "ses" && (
            <>
              <Field
                label={tr("email.awsKey")}
                value={draft.aws_access_key_id || ""}
                onChange={(v) => setDraft({ ...draft, aws_access_key_id: v })}
                placeholder="AKIAIOSFODNN7EXAMPLE"
              />
              <Field
                label={data.aws_has_secret_key ? tr("email.awsSecretSaved") : tr("email.awsSecret")}
                type="password"
                value={draft.aws_secret_key || ""}
                onChange={(v) => setDraft({ ...draft, aws_secret_key: v })}
                placeholder={data.aws_has_secret_key ? "••••••••••••" : "Secret access key"}
              />
              <Field
                label={tr("email.awsRegion")}
                value={draft.aws_region || "us-east-1"}
                onChange={(v) => setDraft({ ...draft, aws_region: v })}
                placeholder="us-east-1"
              />
            </>
          )}

          <div className="actions">
            <Button primary onClick={save}>{tr("email.save")}</Button>
          </div>
        </div>
      </Card>

      <Card title={tr("email.testTitle")} description={tr("email.testDesc")}>
        <div className="form-narrow">
          <Field
            label={tr("email.testRecipient")}
            type="email"
            value={testTo}
            onChange={setTestTo}
            placeholder={tr("email.testPlaceholder")}
          />
          <Button onClick={sendTest} disabled={testing}>
            {testing ? tr("email.sending") : tr("email.sendTest")}
          </Button>
        </div>
      </Card>
    </>
  );
}

export function App() {
  const [tab, setTab] = useState<Tab>((location.hash.slice(1) as Tab) || "overview");
  const [me, setMe] = useState<Data | null>(null);
  const [notice, setNotice] = useState("");
  const [theme, setTheme] = useState<AppTheme>(() => (localStorage.getItem("admin-theme") as AppTheme) || "system");
  const [lang, setLang] = useState<Lang>(() => getStoredLang());
  const notify = (message: string) => { setNotice(message); window.setTimeout(() => setNotice(""), 3500); };

  const changeLang = (next: Lang) => {
    setStoredLang(next);
    setLang(next);
    document.documentElement.lang = next;
  };

  const tr = useMemo<TranslateFn>(() => (key, vars) => t(lang, key, vars), [lang]);

  useEffect(() => {
    document.documentElement.lang = lang;
  }, [lang]);

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

  const changeTheme = (next: AppTheme) => { setTheme(next); localStorage.setItem("admin-theme", next); };
  useEffect(() => { request("/me").then(setMe).catch((e) => notify(e.message)); }, []);
  useEffect(() => { history.replaceState(null, "", `/dashboard/#${tab}`); }, [tab]);

  const title = tr(tabMeta.find((item) => item.id === tab)?.labelKey || "nav.overview");
  const page = tab === "integration" ? <Integration notify={notify} />
    : tab === "providers" ? <Providers notify={notify} />
    : tab === "overview" ? <Overview />
    : tab === "email" ? <EmailSettings notify={notify} />
    : tab === "forms" ? <FormsBuilder notify={notify} />
    : ["users", "sessions", "audit", "webhooks"].includes(tab) ? <ResourcePage tab={tab} notify={notify} />
    : ["project", "theme", "billing"].includes(tab) ? <Settings tab={tab as "project" | "theme" | "billing"} notify={notify} />
    : null;

  const switchProject = async (id: string) => {
    try {
      document.cookie = `sooauth_project=${encodeURIComponent(id)}; path=/; max-age=31536000; SameSite=Lax`;
      await request(`/projects/${id}/select`, { method: "POST" });
      location.reload();
    } catch (e) {
      notify((e as Error).message);
    }
  };

  const createProject = async () => {
    const name = window.prompt(tr("project.promptName"));
    if (!name?.trim()) return;
    try {
      await request("/projects", { method: "POST", body: JSON.stringify({ name: name.trim() }) });
      location.reload();
    } catch (e) {
      notify((e as Error).message);
    }
  };

  const navGroups = ["nav.group.workspace", "nav.group.configuration"] as const;

  return (
    <I18nContext.Provider value={{ lang, tr, setLang: changeLang }}>
      <div className="app">
        <aside className="sidebar">
          <a className="brand" href="https://sooauth.com">
            <img className="brand-logo brand-logo-light" src="/sooauth-light.png" alt="sooauth" />
            <img className="brand-logo brand-logo-dark" src="/sooauth-dark.png" alt="" />
            <span>sooauth</span>
          </a>
          {me?.projects?.length ? (
            <div className="sidebar-project">
              <label className="project-picker">
                <span>{tr("sidebar.currentProject")}</span>
                <div className="project-select-wrap">
                  <select className="project-select" value={me.project?.id || ""} onChange={(e) => switchProject(e.target.value)} aria-label={tr("sidebar.currentProject")}>
                    {me.projects.map((project: Data) => (
                      <option value={project.id} key={project.id}>{project.name}</option>
                    ))}
                  </select>
                </div>
              </label>
              <button type="button" className="btn-new-project" onClick={createProject}>
                {tr("sidebar.newProject")}
              </button>
            </div>
          ) : null}
          <nav>
            {navGroups.map((groupKey) => (
              <div className="nav-group" key={groupKey}>
                <small>{tr(groupKey)}</small>
                {tabMeta.filter((item) => item.groupKey === groupKey).map((item) => (
                  <button className={tab === item.id ? "active" : ""} key={item.id} onClick={() => setTab(item.id)}>
                    {tr(item.labelKey)}
                  </button>
                ))}
              </div>
            ))}
          </nav>
          <div className="sidebar-links">
            <a href="https://sooauth.com/docs" target="_blank" rel="noopener noreferrer">{tr("sidebar.docs")}</a>
            <a href="https://sooauth.com/docs/integrate-oidc#ai-prompt" target="_blank" rel="noopener noreferrer">{tr("sidebar.aiPrompt")}</a>
            <a href="/auth/account">{tr("sidebar.security")}</a>
          </div>
          <UserProfileMenu me={me} theme={theme} onThemeChange={changeTheme} />
        </aside>
        <main>
          <header className="topbar">
            <div>
              <p className="eyebrow">{tr("nav.group.workspace")}</p>
              <h1>{title}{me?.project?.name ? <span> · {me.project.name}</span> : null}</h1>
            </div>
            <div className="lang-switch" role="group" aria-label={tr("common.language")}>
              <button type="button" className={lang === "en" ? "active" : ""} onClick={() => changeLang("en")}>EN</button>
              <button type="button" className={lang === "tr" ? "active" : ""} onClick={() => changeLang("tr")}>TR</button>
            </div>
          </header>
          {notice && <div className="toast">{notice}</div>}
          <div className="content">{page}</div>
        </main>
      </div>
    </I18nContext.Provider>
  );
}
