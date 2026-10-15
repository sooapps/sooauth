import { useState } from "react";
import { put, request, type Data } from "../../api";
import { useApp, useI18n } from "../../context";
import { Button, Card, Check, Field, Loading, useLoad } from "../../ui";

export default function SecurityPolicy() {
  const { tr } = useI18n();
  const { notify } = useApp();
  const [draft, setDraft] = useState<Data | null>(null);
  useLoad(() => request("/project").then(setDraft));
  if (!draft) return <Loading />;
  const set = (patch: Data) => setDraft({ ...draft, ...patch });
  const passwordless = draft.auth_config?.primary_auth_mode === "otp";

  const save = async () => {
    try {
      await put("/project", {
        email_verify_required: Boolean(draft.email_verify_required),
        email_verify_delivery: draft.email_verify_delivery || "link",
        password_min_length: Number(draft.password_min_length || 8),
        password_require_upper: Boolean(draft.password_require_upper),
        password_require_number: Boolean(draft.password_require_number),
        password_require_special: Boolean(draft.password_require_special),
      });
      notify(tr("common.settingsSaved"));
    } catch (e) {
      notify((e as Error).message);
    }
  };

  return (
    <>
      <Card title={tr("security.verifyTitle")} description={tr("security.verifyDesc")}>
        <div className="form-narrow">
          <Check label={tr("project.requireVerify")} hint={tr("security.verifyHint")} checked={Boolean(draft.email_verify_required)} onChange={(on) => set({ email_verify_required: on })} />
          <label className="field">
            <span>{tr("project.delivery")}</span>
            <select value={draft.email_verify_delivery || "link"} onChange={(e) => set({ email_verify_delivery: e.target.value })}>
              <option value="link">{tr("project.deliveryLink")}</option>
              <option value="code">{tr("project.deliveryCode")}</option>
              <option value="both">{tr("project.deliveryBoth")}</option>
            </select>
          </label>
        </div>
      </Card>
      <Card title={tr("security.passwordTitle")} description={passwordless ? tr("security.passwordUnused") : tr("security.passwordDesc")}>
        <div className="settings-layout">
          <div className="form-narrow">
            <Field label={tr("project.minPassword")} type="number" value={draft.password_min_length || 8} onChange={(v) => set({ password_min_length: v })} />
            <Check label={tr("project.requireUpper")} checked={Boolean(draft.password_require_upper)} onChange={(on) => set({ password_require_upper: on })} />
            <Check label={tr("project.requireNumber")} checked={Boolean(draft.password_require_number)} onChange={(on) => set({ password_require_number: on })} />
            <Check label={tr("project.requireSpecial")} checked={Boolean(draft.password_require_special)} onChange={(on) => set({ password_require_special: on })} />
          </div>
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
        </div>
      </Card>
      <div className="actions"><Button primary onClick={save}>{tr("common.saveChanges")}</Button></div>
    </>
  );
}
