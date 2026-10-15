import { request, type Data } from "../../api";
import { useI18n } from "../../context";
import { Card, Empty, Loading, Table, useLoad } from "../../ui";

export default function Audit() {
  const { tr } = useI18n();
  const [data] = useLoad<Data>(() => request("/audit"));
  if (!data) return <Loading />;
  return (
    <Card title={tr("audit.title")}>
      {data.entries?.length ? (
        <Table
          headers={[tr("overview.col.time"), tr("overview.col.action"), tr("overview.col.ip")]}
          rows={data.entries.map((e: Data) => [new Date(e.created_at).toLocaleString(), e.action, e.ip || "—"])}
        />
      ) : <Empty text={tr("audit.empty")} />}
    </Card>
  );
}
