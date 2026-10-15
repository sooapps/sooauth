import { useState } from "react";
import { put, request, type Data } from "../../api";
import { useApp, useI18n } from "../../context";
import { Button, Card, CopyField, Loading, useLoad } from "../../ui";

// RedirectEditor edits the allowed redirect URIs. Inside the wizard it is
// controlled (value/onChange) and the wizard's Next button saves.
export function RedirectEditor({ initial, onSaved, value, onChange }: { initial: string[]; onSaved?: (uris: string[]) => void; value?: string[]; onChange?: (uris: string[]) => void }) {
  const { tr } = useI18n();
  const { notify, refreshOnboarding } = useApp();
  const [own, setOwn] = useState<string[]>(initial.length ? initial : [""]);
  const controlled = value !== undefined;
  const uris = controlled ? value : own;
  const setUris = (next: string[]) => (controlled ? onChange?.(next) : setOwn(next));
  const save = async () => {
    const clean = uris.map((u) => u.trim()).filter(Boolean);
    try {
      await put("/integration", { redirect_uris: clean });
      notify(tr("integration.saved"));
      refreshOnboarding();
      onSaved?.(clean);
    } catch (e) {
      notify((e as Error).message);
    }
  };
  return (
    <>
      <div className="uri-list">
        {uris.map((uri, index) => (
          <div className="uri-row" key={index}>
            <input value={uri} onChange={(e) => setUris(uris.map((item, i) => (i === index ? e.target.value : item)))} placeholder="https://myapp.com/api/auth/callback/oidc" />
            <Button small onClick={() => setUris(uris.filter((_, i) => i !== index))}>×</Button>
          </div>
        ))}
      </div>
      <div className="actions">
        <Button onClick={() => setUris([...uris, ""])}>{tr("integration.addUri")}</Button>
        {!controlled && <Button primary onClick={save}>{tr("integration.saveUris")}</Button>}
      </div>
    </>
  );
}

export default function KeysUrls() {
  const { tr } = useI18n();
  const [data] = useLoad<Data>(() => request("/integration"));
  if (!data) return <Loading />;
  return (
    <>
      <Card title={tr("integration.connectTitle")} description={tr("integration.connectDesc")}>
        <div className="data-grid">
          <CopyField label={tr("integration.clientId")} value={data.client_id} />
          <CopyField label={tr("integration.issuer")} value={data.issuer} />
          <CopyField label={tr("integration.discovery")} value={data.discovery_url} />
          <CopyField label={tr("integration.tenantId")} value={data.tenant_id} />
        </div>
      </Card>
      <Card title={tr("integration.redirectTitle")} description={tr("integration.redirectDesc")}>
        <RedirectEditor initial={data.redirect_uris || []} />
        <p className="subtle" style={{ marginTop: 14 }}>{tr("integration.testUriNote", { uri: data.test_redirect_uri })}</p>
      </Card>
      <Card title={tr("integration.endpointsTitle")} description={tr("integration.endpointsDesc")}>
        <div className="data-grid">
          <CopyField label="Authorize" value={data.authorize_url} />
          <CopyField label="Token" value={data.token_url} />
          <CopyField label="Userinfo" value={data.userinfo_url} />
          <CopyField label="Widget config" value={`${data.issuer}/v1/widget/config?client_id=${data.client_id}`} />
        </div>
      </Card>
    </>
  );
}
