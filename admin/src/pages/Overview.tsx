import { request, type Data } from "../api";
import { useApp, useI18n } from "../context";
import { Button, Card, Empty, Loading, Metric, Table, useLoad } from "../ui";

export default function Overview() {
  const { tr } = useI18n();
  const { onboarding, go } = useApp();
  const [data] = useLoad<Data>(() =>
    Promise.all([request("/users"), request("/sessions"), request("/audit")])
      .then(([users, sessions, audit]) => ({ users, sessions, audit }))
      .catch(() => ({ users: {}, sessions: {}, audit: {} })),
  );
  if (!data) return <Loading />;
  return (
    <>
      {onboarding && !onboarding.complete && (
        <Card title={tr("overview.setupTitle", { done: onboarding.done_count, total: onboarding.total })} description={tr("overview.setupDesc")}
          actions={<Button primary onClick={() => go("start")}>{tr("overview.setupCta")}</Button>}>
          <div className="progress"><span style={{ width: `${(onboarding.done_count / onboarding.total) * 100}%` }} /></div>
        </Card>
      )}
      <div className="metrics">
        <Metric value={(data.users.users || []).length} label={tr("overview.users")} />
        <Metric value={(data.sessions.sessions || []).length} label={tr("overview.sessions")} />
        <Metric value={(data.audit.entries || []).length} label={tr("overview.audit")} />
      </div>
      <Card title={tr("overview.recent")} description={tr("overview.recentDesc")}>
        {data.audit.entries?.length ? (
          <Table
            headers={[tr("overview.col.time"), tr("overview.col.action"), tr("overview.col.ip")]}
            rows={data.audit.entries.slice(0, 8).map((item: Data) => [new Date(item.created_at).toLocaleString(), item.action, item.ip || "—"])}
          />
        ) : <Empty text={tr("overview.empty")} />}
      </Card>
    </>
  );
}
