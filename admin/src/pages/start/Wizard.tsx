import { useEffect, useState } from "react";
import { put, request, type AuthConfig } from "../../api";
import { useApp, useI18n } from "../../context";
import { defaultCallback, type AppType } from "../../lib/prompts";
import { Button, Field, Loading, Modal } from "../../ui";
import { normalizeAuthConfig, SignInMethodsForm } from "../auth/SignInMethods";
import { AppTypePicker, PromptPanel } from "../integration/AiPrompt";
import { RedirectEditor } from "../integration/KeysUrls";
import { LiveTest } from "../integration/LiveTest";

const steps = ["project", "sign_in", "redirect", "integrate", "test"] as const;

// Wizard walks a new owner through setup: what they build, how users sign in,
// where to send them back, how to integrate, and a live test.
export default function Wizard({ onClose }: { onClose: () => void }) {
  const { tr } = useI18n();
  const { notify, onboarding, refreshOnboarding, refreshMe, go } = useApp();
  const [step, setStep] = useState(0);
  const [name, setName] = useState("");
  const [appType, setAppType] = useState<AppType>((onboarding?.app_type as AppType) || "nextjs");
  const [auth, setAuth] = useState<AuthConfig | null>(null);
  const [uris, setUris] = useState<string[] | null>(null);
  const [busy, setBusy] = useState(false);

  useEffect(() => {
    Promise.all([request("/project"), request("/integration")])
      .then(([project, integration]) => {
        setName(project.name || "");
        setAuth(normalizeAuthConfig(project.auth_config));
        const current: string[] = integration.redirect_uris || [];
        const onlyDefault = current.length === 1 && current[0] === defaultCallback.nextjs;
        setUris(current.length && !onlyDefault ? current : []);
      })
      .catch((e) => notify((e as Error).message));
  }, [notify]);

  const mark = (id: string, extra: Record<string, unknown> = {}) => put("/onboarding", { complete: id, ...extra });

  const close = async () => {
    await put("/onboarding", { dismissed: true }).catch(() => undefined);
    await refreshOnboarding();
    onClose();
  };

  const next = async () => {
    setBusy(true);
    try {
      const id = steps[step];
      if (id === "project") {
        await put("/project", { name });
        await mark("project", { app_type: appType });
        await refreshMe();
      } else if (id === "sign_in" && auth) {
        await put("/project", { auth_config: auth });
        await mark("sign_in");
      } else if (id === "redirect" && uris) {
        // The suggestion is shown in the field, so an untouched field means it.
        const clean = (uris.length ? uris : [defaultCallback[appType]]).map((u) => u.trim()).filter(Boolean);
        if (!clean.length) {
          notify(tr("wizard.redirect.required"));
          return;
        }
        await put("/integration", { redirect_uris: clean });
        await mark("redirect");
      } else if (id === "integrate") {
        await mark("integrate");
      }
      await refreshOnboarding();
      if (step === steps.length - 1) {
        await close();
        go("start");
        return;
      }
      setStep(step + 1);
    } catch (e) {
      notify((e as Error).message);
    } finally {
      setBusy(false);
    }
  };

  if (!auth || uris === null) {
    return <Modal title={tr("wizard.title")} onClose={close} wide><Loading /></Modal>;
  }

  const id = steps[step];
  const suggested = defaultCallback[appType];
  return (
    <Modal
      wide
      title={tr("wizard.title")}
      onClose={close}
      footer={
        <>
          <span className="wizard-count">{tr("wizard.stepOf", { n: step + 1, total: steps.length })}</span>
          <div className="actions">
            <Button onClick={close}>{tr("wizard.later")}</Button>
            {step > 0 && <Button onClick={() => setStep(step - 1)}>{tr("wizard.back")}</Button>}
            <Button primary disabled={busy} onClick={next}>{step === steps.length - 1 ? tr("wizard.finish") : tr("wizard.next")}</Button>
          </div>
        </>
      }
    >
      <div className="wizard-steps">
        {steps.map((s, i) => <span key={s} className={i === step ? "active" : i < step ? "done" : ""}>{tr(`start.step.${s}`)}</span>)}
      </div>

      <h3 className="wizard-heading">{tr(`wizard.${id}.title`)}</h3>
      <p className="section-desc">{tr(`wizard.${id}.desc`)}</p>

      {id === "project" && (
        <div className="stack-lg">
          <div className="form-narrow"><Field label={tr("project.name")} value={name} onChange={setName} /></div>
          <AppTypePicker value={appType} onChange={setAppType} />
        </div>
      )}

      {id === "sign_in" && <SignInMethodsForm value={auth} onChange={setAuth} />}

      {id === "redirect" && (
        <div className="stack">
          <div className="notice">{tr("wizard.redirect.suggest", { uri: suggested })}</div>
          <RedirectEditor initial={[]} value={uris.length ? uris : [suggested]} onChange={setUris} />
        </div>
      )}

      {id === "integrate" && <PromptPanel type={appType} />}

      {id === "test" && <LiveTest compact />}
    </Modal>
  );
}
