import { useState } from "react";
import { put, request, type AuthConfig, type AuthMode, type MFAPolicy } from "../../api";
import { useApp, useI18n } from "../../context";
import { Button, Card, Check, Loading, OptionCards, useLoad } from "../../ui";

export const defaultAuthConfig: AuthConfig = {
  allowed_identifiers: ["email"],
  primary_auth_mode: "password",
  mfa_policy: "off",
  mfa_methods: ["email_otp", "totp"],
  phone_otp_via_webhook: false,
};

export function normalizeAuthConfig(raw: Partial<AuthConfig> | undefined): AuthConfig {
  return { ...defaultAuthConfig, ...(raw || {}) };
}

function toggle(list: string[], item: string, on: boolean) {
  return on ? Array.from(new Set([...list, item])) : list.filter((x) => x !== item);
}

// configWarnings lists combinations that would leave users unable to sign in.
export function configWarnings(cfg: AuthConfig, tr: (k: string) => string): string[] {
  const out: string[] = [];
  const otp = cfg.primary_auth_mode !== "password";
  if (otp && cfg.allowed_identifiers.length === 1 && cfg.allowed_identifiers[0] === "username") {
    out.push(tr("signIn.warnUsernameOtp"));
  }
  if (otp && cfg.allowed_identifiers.includes("phone") && !cfg.phone_otp_via_webhook) {
    out.push(tr("signIn.warnPhoneOtp"));
  }
  if (cfg.mfa_policy === "required" && cfg.mfa_methods.length === 1 && cfg.mfa_methods[0] === "email_otp" && !cfg.allowed_identifiers.includes("email")) {
    out.push(tr("signIn.warnMfaEmail"));
  }
  if (cfg.primary_auth_mode === "otp" && cfg.mfa_policy !== "off") {
    out.push(tr("signIn.noteMfaOtp"));
  }
  return out;
}

// SignInMethodsForm edits identifiers, the primary mode and two-step
// verification. The setup wizard reuses it.
export function SignInMethodsForm({ value, onChange, compact }: { value: AuthConfig; onChange: (next: AuthConfig) => void; compact?: boolean }) {
  const { tr } = useI18n();
  const set = (patch: Partial<AuthConfig>) => onChange({ ...value, ...patch });
  const warnings = configWarnings(value, tr);
  const ids = value.allowed_identifiers;
  return (
    <div className="stack-lg">
      <div>
        <h3 className="section-title">{tr("signIn.identifiersTitle")}</h3>
        <p className="section-desc">{tr("signIn.identifiersDesc")}</p>
        <div className="check-row">
          {(["email", "username", "phone"] as const).map((id) => (
            <Check
              key={id}
              label={tr(`signIn.id.${id}`)}
              hint={tr(`signIn.id.${id}Hint`)}
              checked={ids.includes(id)}
              disabled={ids.length === 1 && ids[0] === id}
              onChange={(on) => set({ allowed_identifiers: toggle(ids, id, on) })}
            />
          ))}
        </div>
      </div>

      <div>
        <h3 className="section-title">{tr("signIn.modeTitle")}</h3>
        <p className="section-desc">{tr("signIn.modeDesc")}</p>
        <OptionCards<AuthMode>
          name="primary_auth_mode"
          value={value.primary_auth_mode}
          onChange={(mode) => set({ primary_auth_mode: mode })}
          options={[
            { value: "password", title: tr("signIn.modePassword"), desc: tr("signIn.modePasswordDesc") },
            { value: "otp", title: tr("signIn.modeOtp"), desc: tr("signIn.modeOtpDesc") },
            { value: "both", title: tr("signIn.modeBoth"), desc: tr("signIn.modeBothDesc") },
          ]}
        />
      </div>

      {!compact && (
        <div>
          <h3 className="section-title">{tr("signIn.mfaTitle")}</h3>
          <p className="section-desc">{tr("signIn.mfaDesc")}</p>
          <OptionCards<MFAPolicy>
            name="mfa_policy"
            value={value.mfa_policy}
            onChange={(policy) => set({ mfa_policy: policy })}
            options={[
              { value: "off", title: tr("signIn.mfaOff"), desc: tr("signIn.mfaOffDesc") },
              { value: "optional", title: tr("signIn.mfaOptional"), desc: tr("signIn.mfaOptionalDesc") },
              { value: "required", title: tr("signIn.mfaRequired"), desc: tr("signIn.mfaRequiredDesc") },
            ]}
          />
          {value.mfa_policy !== "off" && (
            <div className="check-row" style={{ marginTop: 14 }}>
              <Check
                label={tr("signIn.mfaTotp")}
                hint={tr("signIn.mfaTotpHint")}
                checked={value.mfa_methods.includes("totp")}
                disabled={value.mfa_methods.length === 1 && value.mfa_methods[0] === "totp"}
                onChange={(on) => set({ mfa_methods: toggle(value.mfa_methods, "totp", on) })}
              />
              <Check
                label={tr("signIn.mfaEmail")}
                hint={tr("signIn.mfaEmailHint")}
                checked={value.mfa_methods.includes("email_otp")}
                disabled={value.mfa_methods.length === 1 && value.mfa_methods[0] === "email_otp"}
                onChange={(on) => set({ mfa_methods: toggle(value.mfa_methods, "email_otp", on) })}
              />
            </div>
          )}
        </div>
      )}

      {ids.includes("phone") && value.primary_auth_mode !== "password" && (
        <div>
          <h3 className="section-title">{tr("signIn.phoneTitle")}</h3>
          <Check
            label={tr("signIn.phoneWebhook")}
            hint={tr("signIn.phoneWebhookHint")}
            checked={value.phone_otp_via_webhook}
            onChange={(on) => set({ phone_otp_via_webhook: on })}
          />
        </div>
      )}

      {warnings.length > 0 && <div className="notice">{warnings.map((w) => <div key={w}>{w}</div>)}</div>}
    </div>
  );
}

export function SignInPreview({ cfg }: { cfg: AuthConfig }) {
  const { tr } = useI18n();
  const idLabel = cfg.allowed_identifiers.map((id) => tr(`signIn.id.${id}`)).join(" / ");
  const showPassword = cfg.primary_auth_mode !== "otp";
  const showCode = cfg.primary_auth_mode !== "password";
  return (
    <aside className="preview-card" aria-label={tr("signIn.previewTitle")}>
      <span className="preview-label">{tr("signIn.previewTitle")}</span>
      <div className="preview-input">{idLabel}</div>
      {showPassword && <div className="preview-input">{tr("signIn.previewPassword")}</div>}
      <div className="preview-button">{showPassword ? tr("signIn.previewSignIn") : tr("signIn.previewSendCode")}</div>
      {cfg.primary_auth_mode === "both" && <div className="preview-link">{tr("signIn.previewUseCode")}</div>}
      {showCode && cfg.primary_auth_mode === "otp" && <small>{tr("signIn.previewOtpNote")}</small>}
      {cfg.mfa_policy !== "off" && <small>{cfg.mfa_policy === "required" ? tr("signIn.previewMfaRequired") : tr("signIn.previewMfaOptional")}</small>}
    </aside>
  );
}

export default function SignInMethods() {
  const { tr } = useI18n();
  const { notify, refreshOnboarding } = useApp();
  const [cfg, setCfg] = useState<AuthConfig | null>(null);
  const [saving, setSaving] = useState(false);
  useLoad(() => request("/project").then((p) => setCfg(normalizeAuthConfig(p.auth_config))));
  if (!cfg) return <Loading />;

  const save = async () => {
    setSaving(true);
    try {
      await put("/project", { auth_config: cfg });
      notify(tr("common.settingsSaved"));
      refreshOnboarding();
    } catch (e) {
      notify((e as Error).message);
    } finally {
      setSaving(false);
    }
  };

  return (
    <Card>
      <div className="settings-layout">
        <SignInMethodsForm value={cfg} onChange={setCfg} />
        <SignInPreview cfg={cfg} />
      </div>
      <div className="actions sticky-actions">
        <Button primary onClick={save} disabled={saving}>{saving ? tr("common.saving") : tr("common.saveChanges")}</Button>
      </div>
    </Card>
  );
}
