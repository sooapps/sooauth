import { useApp, useI18n } from "../../context";
import { Button, Card, Loading } from "../../ui";

const stepTargets: Record<string, string> = {
  project: "settings",
  sign_in: "sign-in",
  redirect: "keys",
  integrate: "ai-prompt",
  test: "test",
};

// GettingStarted is the persistent setup checklist. Steps tick themselves from
// real project state (redirect URIs, users, sign-ins) where possible.
export default function GettingStarted() {
  const { tr } = useI18n();
  const { onboarding, go, openWizard } = useApp();
  if (!onboarding) return <Loading />;
  const pct = Math.round((onboarding.done_count / onboarding.total) * 100);
  return (
    <>
      <Card
        title={onboarding.complete ? tr("start.doneTitle") : tr("start.title")}
        description={onboarding.complete ? tr("start.doneDesc") : tr("start.desc")}
        actions={<Button primary={!onboarding.complete} onClick={openWizard}>{tr("start.openWizard")}</Button>}
      >
        <div className="progress-row">
          <div className="progress"><span style={{ width: `${pct}%` }} /></div>
          <span>{tr("start.progress", { done: onboarding.done_count, total: onboarding.total })}</span>
        </div>
        <ol className="checklist-steps">
          {onboarding.steps.map((step, i) => (
            <li key={step.id} className={step.done ? "done" : ""}>
              <span className="step-mark">{step.done ? "✓" : i + 1}</span>
              <div>
                <strong>{tr(`start.step.${step.id}`)}</strong>
                <p>{tr(`start.step.${step.id}Desc`)}</p>
              </div>
              <Button small primary={!step.done} onClick={() => go(stepTargets[step.id])}>{step.done ? tr("start.review") : tr("start.go")}</Button>
            </li>
          ))}
        </ol>
      </Card>
      <Card title={tr("start.helpTitle")}>
        <div className="link-grid">
          <a href="https://sooauth.com/docs/getting-started" target="_blank" rel="noopener noreferrer"><strong>{tr("start.helpDocs")}</strong><span>{tr("start.helpDocsDesc")}</span></a>
          <button type="button" onClick={() => go("ai-prompt")}><strong>{tr("start.helpPrompt")}</strong><span>{tr("start.helpPromptDesc")}</span></button>
          <a href="https://sooauth.com/playground" target="_blank" rel="noopener noreferrer"><strong>{tr("start.helpPlayground")}</strong><span>{tr("start.helpPlaygroundDesc")}</span></a>
        </div>
      </Card>
    </>
  );
}
