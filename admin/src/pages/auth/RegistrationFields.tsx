import { useState } from "react";
import { post, put, request, type OptionsSource, type RegistrationField } from "../../api";
import { useApp, useI18n } from "../../context";
import { Button, Card, Check, Field, Loading, useLoad } from "../../ui";

type TestState = Record<string, { loading?: boolean; count?: number; error?: string }>;

function FieldEditor({ field, onChange, onRemove, onMove, test, onTest }: {
  field: RegistrationField;
  onChange: (patch: Partial<RegistrationField>) => void;
  onRemove: () => void;
  onMove: (dir: -1 | 1) => void;
  test?: TestState[string];
  onTest: (src: OptionsSource) => void;
}) {
  const { tr } = useI18n();
  const src: OptionsSource = field.options_source || { type: "static", options: [] };
  const setSrc = (patch: Partial<OptionsSource>) => onChange({ options_source: { ...src, ...patch } });
  const [headersText, setHeadersText] = useState(src.headers ? JSON.stringify(src.headers) : "");
  const [headersError, setHeadersError] = useState(false);

  return (
    <div className="field-editor">
      <div className="field-editor-head">
        <div className="field-editor-title">
          <strong>{field.label || field.id}</strong>
          <span className="pill">{tr(`fields.type.${field.type}`)}</span>
          {field.required && <span className="pill accent">{tr("common.required")}</span>}
        </div>
        <div className="actions">
          <Button small onClick={() => onMove(-1)}>↑</Button>
          <Button small onClick={() => onMove(1)}>↓</Button>
          <Button small danger onClick={onRemove}>{tr("fields.remove")}</Button>
        </div>
      </div>

      <div className="grid-3">
        <Field label={tr("fields.label")} value={field.label} onChange={(v) => onChange({ label: v })} />
        <Field label={tr("fields.id")} hint={tr("fields.idHint")} value={field.id} onChange={(v) => onChange({ id: v.toLowerCase().replace(/[^a-z0-9_]/g, "_") })} />
        <label className="field">
          <span>{tr("fields.typeLabel")}</span>
          <select
            value={field.type}
            onChange={(e) => {
              const type = e.target.value as RegistrationField["type"];
              onChange(type === "select" && !field.options_source ? { type, options_source: { type: "static", options: [] } } : { type });
            }}
          >
            {(["text", "textarea", "number", "select", "checkbox"] as const).map((type) => (
              <option key={type} value={type}>{tr(`fields.type.${type}`)}</option>
            ))}
          </select>
        </label>
      </div>

      <div className="grid-3">
        <Field label={tr("fields.placeholder")} value={field.placeholder || ""} onChange={(v) => onChange({ placeholder: v })} />
        <Field label={tr("fields.help")} value={field.description || ""} onChange={(v) => onChange({ description: v })} />
        <div className="field-check">
          <Check label={tr("fields.required")} checked={field.required} onChange={(on) => onChange({ required: on })} />
        </div>
      </div>

      {field.type === "select" && (
        <div className="subpanel">
          <label className="field">
            <span>{tr("fields.optionsSource")}</span>
            <select value={src.type} onChange={(e) => setSrc({ type: e.target.value as OptionsSource["type"] })}>
              <option value="static">{tr("fields.sourceStatic")}</option>
              <option value="dynamic_api">{tr("fields.sourceApi")}</option>
            </select>
          </label>

          {src.type === "dynamic_api" ? (
            <div className="stack">
              <p className="section-desc">{tr("fields.apiExplain")}</p>
              <Field label={tr("fields.apiUrl")} value={src.url || ""} placeholder="https://api.example.com/teams" onChange={(v) => setSrc({ url: v })} />
              <label className="field">
                <span>{tr("fields.apiHeaders")}</span>
                <input
                  value={headersText}
                  placeholder='{"Authorization": "Bearer …"}'
                  onChange={(e) => {
                    setHeadersText(e.target.value);
                    try {
                      setSrc({ headers: e.target.value.trim() ? JSON.parse(e.target.value) : {} });
                      setHeadersError(false);
                    } catch {
                      setHeadersError(true);
                    }
                  }}
                />
                <small className={`field-hint ${headersError ? "error" : ""}`}>{headersError ? tr("fields.headersInvalid") : tr("fields.headersHint")}</small>
              </label>
              <div className="grid-4">
                <Field label={tr("fields.itemsPath")} hint={tr("fields.itemsPathHint")} value={src.items_path || ""} placeholder="data.items" onChange={(v) => setSrc({ items_path: v })} />
                <Field label={tr("fields.labelKey")} value={src.label_key || ""} placeholder="name" onChange={(v) => setSrc({ label_key: v })} />
                <Field label={tr("fields.valueKey")} value={src.value_key || ""} placeholder="id" onChange={(v) => setSrc({ value_key: v })} />
                <Field label={tr("fields.cacheTtl")} type="number" value={src.cache_ttl_seconds || 300} onChange={(v) => setSrc({ cache_ttl_seconds: Number(v) || 300 })} />
              </div>
              <div className="actions">
                <Button onClick={() => onTest(src)} disabled={test?.loading || !src.url}>{test?.loading ? tr("fields.testing") : tr("fields.testApi")}</Button>
                {test?.count !== undefined && <span className="ok-text">✓ {tr("fields.testOk", { n: test.count })}</span>}
                {test?.error && <span className="error-text">✗ {test.error}</span>}
              </div>
            </div>
          ) : (
            <div className="stack">
              {(src.options || []).map((opt, i) => (
                <div className="uri-row" key={i}>
                  <input placeholder={tr("fields.optLabel")} value={opt.label} onChange={(e) => setSrc({ options: (src.options || []).map((o, j) => (j === i ? { ...o, label: e.target.value } : o)) })} />
                  <input placeholder={tr("fields.optValue")} value={opt.value} onChange={(e) => setSrc({ options: (src.options || []).map((o, j) => (j === i ? { ...o, value: e.target.value } : o)) })} />
                  <Button small onClick={() => setSrc({ options: (src.options || []).filter((_, j) => j !== i) })}>×</Button>
                </div>
              ))}
              <div><Button small onClick={() => setSrc({ options: [...(src.options || []), { label: "", value: "" }] })}>{tr("fields.addOption")}</Button></div>
            </div>
          )}
        </div>
      )}
    </div>
  );
}

export default function RegistrationFields() {
  const { tr } = useI18n();
  const { notify } = useApp();
  const [fields, setFields] = useState<RegistrationField[] | null>(null);
  const [saving, setSaving] = useState(false);
  const [tests, setTests] = useState<TestState>({});
  useLoad(() => request("/project").then((p) => setFields(p.registration_schema || [])));
  if (!fields) return <Loading />;

  const update = (index: number, patch: Partial<RegistrationField>) => setFields(fields.map((f, i) => (i === index ? { ...f, ...patch } : f)));
  const move = (index: number, dir: -1 | 1) => {
    const next = [...fields];
    const target = index + dir;
    if (target < 0 || target >= next.length) return;
    [next[index], next[target]] = [next[target], next[index]];
    setFields(next);
  };
  const add = () => setFields([...fields, { id: `field_${fields.length + 1}`, label: tr("fields.newField"), type: "text", required: false, placeholder: "", description: "" }]);
  const test = async (field: RegistrationField, src: OptionsSource) => {
    setTests((t) => ({ ...t, [field.id]: { loading: true } }));
    try {
      const res = await post("/registration-schema/test-api", src);
      setTests((t) => ({ ...t, [field.id]: { count: res.count } }));
    } catch (e) {
      setTests((t) => ({ ...t, [field.id]: { error: (e as Error).message } }));
    }
  };
  const save = async () => {
    const ids = fields.map((f) => f.id);
    if (ids.some((id) => !id) || new Set(ids).size !== ids.length) {
      notify(tr("fields.duplicateIds"));
      return;
    }
    setSaving(true);
    try {
      await put("/project", { registration_schema: fields });
      notify(tr("common.settingsSaved"));
    } catch (e) {
      notify((e as Error).message);
    } finally {
      setSaving(false);
    }
  };

  return (
    <>
      <Card title={tr("fields.title")} description={tr("fields.desc")}>
        <div className="stack-lg">
          {fields.length === 0 ? (
            <div className="empty">{tr("fields.empty")}</div>
          ) : (
            fields.map((field, i) => (
              <FieldEditor
                key={i}
                field={field}
                onChange={(patch) => update(i, patch)}
                onRemove={() => setFields(fields.filter((_, j) => j !== i))}
                onMove={(dir) => move(i, dir)}
                test={tests[field.id]}
                onTest={(src) => test(field, src)}
              />
            ))
          )}
          <div className="actions split">
            <Button onClick={add}>{tr("fields.add")}</Button>
            <Button primary onClick={save} disabled={saving}>{saving ? tr("common.saving") : tr("common.saveChanges")}</Button>
          </div>
        </div>
      </Card>
      <Card title={tr("fields.howTitle")}>
        <ol className="how-list">
          <li>{tr("fields.how1")}</li>
          <li>{tr("fields.how2")}</li>
          <li>{tr("fields.how3")}</li>
        </ol>
        <pre className="code-block">{`// sign-up body\n{ "email": "ada@example.com", "password": "…", "client_id": "app_…",\n  "metadata": { ${fields.length ? fields.slice(0, 2).map((f) => `"${f.id}": …`).join(", ") : `"company": "Acme"`} } }\n\n// GET /oauth/userinfo → custom_claims\n{ "sub": "…", "email": "ada@example.com", "custom_claims": { … } }`}</pre>
      </Card>
    </>
  );
}
