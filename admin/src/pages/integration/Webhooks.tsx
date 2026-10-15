import { useState } from "react";
import { post, request, type Data } from "../../api";
import { useApp, useI18n } from "../../context";
import { Button, Card, Check, CopyField, Empty, Field, Loading, useLoad } from "../../ui";
import { WebhookRow } from "./WebhookRow";

export const webhookEvents = [
  "user.created",
  "user.login",
  "user.login_failed",
  "user.email_verified",
  "user.deleted",
  "password.reset",
  "session.revoked",
  "otp.sent",
];

export default function Webhooks() {
  const { tr } = useI18n();
  const { notify } = useApp();
  const [data, , reload] = useLoad<Data>(() => request("/webhooks"));
  const [url, setURL] = useState("");
  const [events, setEvents] = useState<string[]>(webhookEvents.filter((e) => e !== "otp.sent"));
  const [secret, setSecret] = useState("");
  if (!data) return <Loading />;

  const create = async () => {
    try {
      const result = await post("/webhooks", { url, events });
      setURL("");
      setSecret(result.secret || "");
      notify(tr("webhooks.created"));
      reload();
    } catch (e) {
      notify((e as Error).message);
    }
  };

  return (
    <>
      <Card title={tr("webhooks.addTitle")} description={tr("webhooks.addDesc")}>
        <div className="form-narrow">
          <Field label={tr("webhooks.endpointUrl")} value={url} onChange={setURL} placeholder="https://myapp.com/api/sooauth-webhook" />
          <div className="event-grid">
            {webhookEvents.map((event) => (
              <Check
                key={event}
                label={<code>{event}</code>}
                hint={tr(`webhooks.event.${event}`)}
                checked={events.includes(event)}
                onChange={(on) => setEvents(on ? [...events, event] : events.filter((e) => e !== event))}
              />
            ))}
          </div>
          <div><Button primary onClick={create} disabled={!url.trim()}>{tr("webhooks.create")}</Button></div>
          {secret && (
            <div className="notice">
              <p style={{ margin: "0 0 10px" }}>{tr("webhooks.secretOnce")}</p>
              <CopyField label={tr("webhooks.secret")} value={secret} />
            </div>
          )}
        </div>
      </Card>
      <Card title={tr("webhooks.listTitle", { n: (data.webhooks || []).length })} description={tr("webhooks.listDesc")}>
        {data.webhooks?.length ? (
          <div className="webhook-list">
            {data.webhooks.map((webhook: Data) => <WebhookRow key={webhook.id} webhook={webhook} notify={notify} onDeleted={reload} />)}
          </div>
        ) : <Empty text={tr("webhooks.empty")} />}
      </Card>
    </>
  );
}
