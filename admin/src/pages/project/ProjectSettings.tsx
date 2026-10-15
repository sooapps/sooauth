import { useState } from "react";
import { put, request, type Data } from "../../api";
import { useApp, useI18n } from "../../context";
import { Button, Card, CopyField, Field, Loading, useLoad } from "../../ui";

export default function ProjectSettings() {
  const { tr } = useI18n();
  const { notify, projectId } = useApp();
  const [draft, setDraft] = useState<Data | null>(null);
  useLoad(() => request("/project").then(setDraft));
  if (!draft) return <Loading />;
  const save = async () => {
    try {
      await put("/project", { name: draft.name, default_locale: draft.default_locale || "en" });
      notify(tr("common.settingsSaved"));
    } catch (e) {
      notify((e as Error).message);
    }
  };
  return (
    <Card title={tr("settings.generalTitle")} description={tr("settings.generalDesc")}>
      <div className="form-narrow">
        <Field label={tr("project.name")} value={draft.name || ""} onChange={(v) => setDraft({ ...draft, name: v })} />
        <label className="field">
          <span>{tr("project.defaultLocale")}</span>
          <select value={draft.default_locale || "en"} onChange={(e) => setDraft({ ...draft, default_locale: e.target.value })}>
            <option value="en">{tr("project.localeEn")}</option>
            <option value="tr">{tr("project.localeTr")}</option>
          </select>
          <small className="field-hint">{tr("settings.localeHint")}</small>
        </label>
        <CopyField label={tr("settings.projectId")} value={projectId} />
        <div><Button primary onClick={save}>{tr("common.saveChanges")}</Button></div>
      </div>
    </Card>
  );
}
