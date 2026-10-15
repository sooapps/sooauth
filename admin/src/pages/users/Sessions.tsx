import { post, request, type Data } from "../../api";
import { useApp, useI18n } from "../../context";
import { Button, Card, Empty, Loading, Table, useLoad } from "../../ui";

export default function Sessions() {
  const { tr } = useI18n();
  const { notify } = useApp();
  const [data, setData] = useLoad<Data>(() => request("/sessions"));
  if (!data) return <Loading />;
  const sessions: Data[] = data.sessions || [];
  return (
    <Card title={tr("sessions.title")} description={tr("sessions.desc")}>
      {sessions.length ? (
        <Table
          headers={[tr("sessions.col.user"), tr("sessions.col.ip"), tr("sessions.col.expires"), ""]}
          rows={sessions.map((s) => [
            s.email || "—",
            s.ip || "—",
            new Date(s.expires_at).toLocaleString(),
            <Button small key="r" onClick={async () => {
              try {
                await post(`/sessions/${s.id}/revoke`);
                notify(tr("sessions.revoked"));
                setData({ ...data, sessions: sessions.filter((item) => item.id !== s.id) });
              } catch (e) {
                notify((e as Error).message);
              }
            }}>{tr("sessions.revoke")}</Button>,
          ])}
        />
      ) : <Empty text={tr("sessions.empty")} />}
    </Card>
  );
}
