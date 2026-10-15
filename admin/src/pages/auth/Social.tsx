import { useEffect, useState } from "react";
import { put, request, type Data } from "../../api";
import { useApp, useI18n } from "../../context";
import { Button, Card, Field, Loading } from "../../ui";

export function ProviderCard({
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

function ProviderList() {
  const { notify } = useApp();
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


// GoogleBranding lets paid plans show their own domain in Google's account
// picker by routing the Google callback through the app's origin.
function GoogleBranding() {
  const { tr } = useI18n();
  const { notify } = useApp();
  const [data, setData] = useState<Data | null>(null);
  const [origin, setOrigin] = useState("");
  useEffect(() => { request("/integration").then((v) => { setData(v); setOrigin(v.social_callback_origin || ""); }).catch((e) => notify(e.message)); }, []);
  if (!data) return null;
  const save = async () => {
    try {
      await put("/integration", { redirect_uris: data.redirect_uris, social_callback_origin: origin });
      notify(tr("integration.saved"));
    } catch (e) {
      notify((e as Error).message);
    }
  };
  return (
    <Card title={tr("integration.googleTitle")} description={tr("integration.googleDesc")}>
      {data.can_custom_social_callback ? (
        <div className="form-narrow">
          <Field label={tr("integration.appOrigin")} value={origin} onChange={setOrigin} placeholder="https://myapp.com" />
          <div><Button primary onClick={save}>{tr("integration.saveGoogle")}</Button></div>
        </div>
      ) : <div className="notice">{tr("integration.freeNotice")}</div>}
    </Card>
  );
}

export default function Social() {
  return (
    <>
      <ProviderList />
      <GoogleBranding />
    </>
  );
}
