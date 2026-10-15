import { useEffect, useState } from "react";
import { request, type Data } from "../../api";
import { useApp, useI18n } from "../../context";
import { Button, Card, Field, Loading } from "../../ui";

export default function EmailSettings() {
  const { notify } = useApp();
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

