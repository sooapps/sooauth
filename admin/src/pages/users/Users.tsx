import { useMemo, useState } from "react";
import { post, request, type Data } from "../../api";
import { useApp, useI18n } from "../../context";
import { Button, Card, ConfirmDialog, Empty, Loading, useLoad } from "../../ui";

function displayName(u: Data) {
  return u.email || u.username || u.phone || "—";
}

export default function Users() {
  const { tr } = useI18n();
  const { notify, go } = useApp();
  const [data, setData, reload] = useLoad<Data>(() => request("/users"));
  const [query, setQuery] = useState("");
  const [deleting, setDeleting] = useState<Data | null>(null);
  const [resetting, setResetting] = useState<Data | null>(null);
  const users: Data[] = data?.users || [];
  const filtered = useMemo(() => {
    const q = query.trim().toLowerCase();
    if (!q) return users;
    return users.filter((u) => [u.email, u.username, u.phone].some((v) => String(v || "").toLowerCase().includes(q)));
  }, [users, query]);
  if (!data) return <Loading />;

  const toggleDisabled = async (u: Data) => {
    try {
      await post(`/users/${u.id}/${u.disabled ? "enable" : "disable"}`);
      notify(u.disabled ? tr("users.enabled") : tr("users.disabledMsg"));
      setData({ ...data, users: users.map((item) => (item.id === u.id ? { ...item, disabled: !item.disabled } : item)) });
    } catch (e) {
      notify((e as Error).message);
    }
  };

  return (
    <Card
      title={tr("users.title", { n: users.length })}
      description={tr("users.desc")}
      actions={users.length > 0 ? <input className="search" type="search" value={query} placeholder={tr("users.search")} onChange={(e) => setQuery(e.target.value)} /> : undefined}
    >
      {users.length === 0 ? (
        <Empty text={tr("users.empty")}>
          <Button primary onClick={() => go("test")}>{tr("users.emptyCta")}</Button>
        </Empty>
      ) : (
        <div className="table-wrap">
          <table>
            <thead>
              <tr>{[tr("users.col.user"), tr("users.col.method"), tr("users.col.verified"), tr("users.col.mfa"), tr("users.col.status"), tr("users.col.joined"), ""].map((h, i) => <th key={h || `h${i}`}>{h}</th>)}</tr>
            </thead>
            <tbody>
              {filtered.map((u) => (
                <tr key={u.id}>
                  <td>
                    <strong>{displayName(u)}</strong>
                    {u.email && (u.username || u.phone) && (
                      <div className="subtle">{u.username && <span>@{u.username} </span>}{u.phone && <span>{u.phone}</span>}</div>
                    )}
                  </td>
                  <td>{u.signup_method || "—"}</td>
                  <td>{u.email_verified || u.phone_verified ? tr("common.yes") : tr("common.no")}</td>
                  <td>{u.mfa_enabled ? <span className="status on">{tr("users.mfaOn")}</span> : <span className="status off">{tr("users.mfaOff")}</span>}</td>
                  <td>{u.disabled ? <span className="status off">{tr("common.disabled")}</span> : <span className="status on">{tr("common.active")}</span>}</td>
                  <td>{new Date(u.created_at).toLocaleDateString()}</td>
                  <td>
                    <div className="row-actions">
                      <Button small onClick={() => toggleDisabled(u)}>{u.disabled ? tr("common.enable") : tr("common.disable")}</Button>
                      {u.mfa_enabled && <Button small onClick={() => setResetting(u)}>{tr("users.resetMfa")}</Button>}
                      <Button small danger onClick={() => setDeleting(u)}>{tr("common.delete")}</Button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
      {deleting && (
        <ConfirmDialog
          title={tr("users.deleteTitle")}
          body={<><p>{tr("users.deleteBody", { user: displayName(deleting) })}</p><p className="subtle">{tr("users.deleteVsDisable")}</p></>}
          confirmLabel={tr("users.deleteConfirm")}
          confirmText={displayName(deleting)}
          onClose={() => setDeleting(null)}
          onConfirm={async () => {
            try {
              await request(`/users/${deleting.id}`, { method: "DELETE" });
              notify(tr("users.deleted"));
              reload();
            } catch (e) {
              notify((e as Error).message);
            }
          }}
        />
      )}
      {resetting && (
        <ConfirmDialog
          title={tr("users.resetMfaTitle")}
          body={tr("users.resetMfaBody", { user: displayName(resetting) })}
          confirmLabel={tr("users.resetMfa")}
          onClose={() => setResetting(null)}
          onConfirm={async () => {
            try {
              await post(`/users/${resetting.id}/mfa/reset`);
              notify(tr("users.mfaResetDone"));
              reload();
            } catch (e) {
              notify((e as Error).message);
            }
          }}
        />
      )}
    </Card>
  );
}
