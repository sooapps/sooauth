import { useState } from "react";
import { request, type Data } from "../../api";
import { useI18n } from "../../context";
import { Button, ConfirmDialog, Empty } from "../../ui";

export function WebhookRow({ webhook, notify, onDeleted }: { webhook: Data; notify: (message: string) => void; onDeleted: () => void }) {
  const { tr } = useI18n();
  const [open, setOpen] = useState(false);
  const [confirming, setConfirming] = useState(false);
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
        <Button danger onClick={() => setConfirming(true)}>{tr("common.delete")}</Button>
        <Button onClick={() => { setOpen(!open); if (!open) load(); }}>
          {open ? tr("webhooks.hideDeliveries") : tr("webhooks.viewLog")}
        </Button>
      </div>
      {confirming && (
        <ConfirmDialog
          title={tr("webhooks.deleteTitle")}
          body={tr("webhooks.deleteConfirm")}
          confirmLabel={tr("common.delete")}
          onClose={() => setConfirming(false)}
          onConfirm={async () => {
            try {
              await request(`/webhooks/${webhook.id}`, { method: "DELETE" });
              notify(tr("webhooks.deleted"));
              onDeleted();
            } catch (e) {
              notify((e as Error).message);
            }
          }}
        />
      )}
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

