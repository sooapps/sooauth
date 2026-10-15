import { useState } from "react";
import { put } from "../../api";
import { useApp, useI18n } from "../../context";
import { appTypes, buildPrompt, snippetFor, type AppType } from "../../lib/prompts";
import { useProjectInfo } from "../../lib/projectInfo";
import { Button, Card, Loading } from "../../ui";

export function AppTypePicker({ value, onChange }: { value: AppType; onChange: (t: AppType) => void }) {
  const { tr } = useI18n();
  return (
    <div className="option-cards cols-5" role="radiogroup">
      {appTypes.map((type) => (
        <label key={type} className={`option-card ${value === type ? "active" : ""}`}>
          <input type="radio" name="app_type" checked={value === type} onChange={() => onChange(type)} />
          <span className="option-title">{tr(`appType.${type}`)}</span>
          <span className="option-desc">{tr(`appType.${type}Desc`)}</span>
        </label>
      ))}
    </div>
  );
}

// PromptPanel shows the filled-in prompt and quick-start snippet for a stack.
export function PromptPanel({ type, onCopied }: { type: AppType; onCopied?: () => void }) {
  const { tr } = useI18n();
  const { notify } = useApp();
  const { info, error } = useProjectInfo();
  if (error) return <div className="notice">{error}</div>;
  if (!info) return <Loading />;
  const prompt = buildPrompt(info, type);
  const copy = async (text: string) => {
    await navigator.clipboard.writeText(text);
    notify(tr("common.copied"));
    onCopied?.();
  };
  return (
    <div className="stack">
      <div className="prompt-box">
        <div className="prompt-head">
          <span>{tr("aiPrompt.promptLabel")}</span>
          <Button small primary onClick={() => copy(prompt)}>{tr("aiPrompt.copyPrompt")}</Button>
        </div>
        <pre>{prompt}</pre>
      </div>
      <div className="prompt-box">
        <div className="prompt-head">
          <span>{tr("aiPrompt.snippetLabel")}</span>
          <Button small onClick={() => copy(snippetFor(info, type))}>{tr("common.copy")}</Button>
        </div>
        <pre>{snippetFor(info, type)}</pre>
      </div>
    </div>
  );
}

export default function AiPrompt() {
  const { tr } = useI18n();
  const { onboarding, refreshOnboarding } = useApp();
  const [type, setType] = useState<AppType>((onboarding?.app_type as AppType) || "nextjs");
  const markIntegrated = () => put("/onboarding", { complete: "integrate", app_type: type }).then(refreshOnboarding).catch(() => undefined);
  return (
    <>
      <Card title={tr("aiPrompt.stackTitle")} description={tr("aiPrompt.stackDesc")}>
        <AppTypePicker value={type} onChange={setType} />
      </Card>
      <Card title={tr("aiPrompt.title")} description={tr("aiPrompt.desc")}>
        <PromptPanel type={type} onCopied={markIntegrated} />
      </Card>
    </>
  );
}
