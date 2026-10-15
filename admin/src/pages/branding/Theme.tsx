import { useState } from "react";
import { put, request, type Data } from "../../api";
import { useApp, useI18n } from "../../context";
import { Button, Card, Field, Loading, useLoad } from "../../ui";

export default function Theme() {
  const { tr } = useI18n();
  const { notify } = useApp();
  const [draft, setDraft] = useState<Data | null>(null);
  useLoad(() => request("/theme").then(setDraft));
  if (!draft) return <Loading />;
  const save = async () => {
    try {
      await put("/theme", { brand_name: draft.brand_name, logo_url: draft.logo_url, accent_color: draft.accent_color });
      notify(tr("common.settingsSaved"));
    } catch (e) {
      notify((e as Error).message);
    }
  };
  return (
    <Card title={tr("theme.title")} description={tr("theme.desc")}>
      <div className="settings-layout">
        <div className="form-narrow">
          <Field label={tr("theme.brandName")} value={draft.brand_name || ""} onChange={(v) => setDraft({ ...draft, brand_name: v })} />
          <Field label={tr("theme.logoUrl")} value={draft.logo_url || ""} placeholder="https://myapp.com/logo.png" onChange={(v) => setDraft({ ...draft, logo_url: v })} />
          <label className="field">
            <span>{tr("theme.accent")}</span>
            <div className="color-row">
              <input type="color" value={draft.accent_color || "#FF3B3B"} onChange={(e) => setDraft({ ...draft, accent_color: e.target.value })} />
              <input value={draft.accent_color || "#FF3B3B"} onChange={(e) => setDraft({ ...draft, accent_color: e.target.value })} />
            </div>
          </label>
          <div><Button primary onClick={save}>{tr("common.saveChanges")}</Button></div>
        </div>
        <aside className="preview-card">
          <span className="preview-label">{tr("signIn.previewTitle")}</span>
          <div className="preview-brand">
            {draft.logo_url ? <img src={draft.logo_url} alt="" /> : null}
            <strong>{draft.brand_name || "sooauth"}</strong>
          </div>
          <div className="preview-input">{tr("signIn.id.email")}</div>
          <div className="preview-button" style={{ background: draft.accent_color || undefined }}>{tr("signIn.previewSignIn")}</div>
        </aside>
      </div>
    </Card>
  );
}
