"use client";

import { useEffect, useState, type ReactNode } from "react";
import {
  buildPrompt,
  callbackUrl,
  defaultAppUrl,
  isBrowserOnly,
  missingValues,
  stackLabels,
  type ColorScheme,
  type Design,
  type EmailVerify,
  type PromptOptions,
  type SignIn,
  type Stack,
  type TwoStep,
} from "../lib/ai-setup-prompt";
import { adminUrl, authUrl } from "../lib/site";
import { cn } from "../lib/cn";
import { buttonClass } from "./ui/button";

const accents = ["#FF3B3B", "#2563EB", "#16A34A", "#7C3AED", "#EA580C", "#111827"];

function Choice<T extends string>({ label, value, options, onChange }: {
  label: string;
  value: T;
  options: { value: T; label: string; hint?: string }[];
  onChange: (value: T) => void;
}) {
  return (
    <fieldset className="grid gap-2">
      <legend className="text-[13px] font-semibold text-fg">{label}</legend>
      <div className="mt-2 flex flex-wrap gap-2">
        {options.map((o) => (
          <label
            key={o.value}
            className={cn(
              "cursor-pointer border px-3 py-1.5 text-[13px] transition-colors has-[:focus-visible]:outline has-[:focus-visible]:outline-2 has-[:focus-visible]:outline-accent",
              value === o.value ? "border-accent bg-accent/10 font-medium text-fg" : "border-line text-fg-muted hover:text-fg",
            )}
            title={o.hint}
          >
            <input type="radio" className="sr-only" checked={value === o.value} onChange={() => onChange(o.value)} />
            {o.label}
          </label>
        ))}
      </div>
    </fieldset>
  );
}

function Text({ label, value, onChange, placeholder, missing, hint }: {
  label: string;
  value: string;
  onChange: (value: string) => void;
  placeholder?: string;
  missing?: boolean;
  hint?: ReactNode;
}) {
  return (
    <label className="grid gap-1.5 text-[13px] font-semibold text-fg">
      {label}
      <input
        value={value}
        placeholder={placeholder}
        onChange={(e) => onChange(e.target.value)}
        className={cn("border bg-bg px-3 py-2 font-mono text-[13px] font-normal text-fg outline-none focus:border-accent", missing ? "border-accent" : "border-line")}
      />
      {hint && <span className="font-sans text-[12px] font-normal text-fg-muted">{hint}</span>}
    </label>
  );
}

// AiPromptBuilder turns a few choices into a prompt for a coding assistant. The
// prompt updates as you choose, and a value left empty is called out both here
// and inside the prompt.
export function AiPromptBuilder({ initialDesign = "hosted" }: { initialDesign?: Design }) {
  const [o, setO] = useState<PromptOptions>({
    issuer: authUrl,
    stack: "react",
    otherStack: "",
    design: initialDesign,
    matchExisting: true,
    accent: accents[0],
    colorScheme: "system",
    signIn: "password",
    twoStep: "off",
    emailVerify: "link",
    social: [],
    extraFields: false,
    clientId: "",
    appUrl: defaultAppUrl.react,
    language: "",
  });
  const [copied, setCopied] = useState(false);
  const set = (patch: Partial<PromptOptions>) => setO((prev) => ({ ...prev, ...patch }));
  useEffect(() => { setCopied(false); }, [o]);

  const prompt = buildPrompt(o);
  const missing = missingValues(o);
  const copy = async () => {
    await navigator.clipboard.writeText(prompt);
    setCopied(true);
  };
  const setStack = (stack: Stack) => set({ stack, appUrl: o.appUrl === defaultAppUrl[o.stack] ? defaultAppUrl[stack] : o.appUrl });
  const toggleSocial = (name: string) => set({ social: o.social.includes(name) ? o.social.filter((s) => s !== name) : [...o.social, name] });

  return (
    <section id="ai-prompt" className="mt-10 scroll-mt-24 border border-line bg-bg-subtle p-6">
      <h2 className="font-sans text-[20px] font-semibold text-fg">AI setup prompt</h2>
      <p className="mt-3 max-w-[62ch] text-[15px] leading-[1.6] text-fg-muted">
        Choose how you are building and copy the result into Claude, Cursor or Copilot. The prompt changes as you choose.
      </p>
      <p className="mt-3 max-w-[62ch] border-l-2 border-accent pl-3 text-[14px] leading-[1.6] text-fg">
        Signed in? <a href={adminUrl} className="underline underline-offset-2">Your project → AI setup prompt</a>{" "}
        already knows your client ID, URLs and sign-in settings.
      </p>

      <div className="mt-6 grid gap-5">
        <Choice label="What do you build with?" value={o.stack} onChange={setStack}
          options={(Object.keys(stackLabels) as Stack[]).map((s) => ({ value: s, label: stackLabels[s] }))} />
        {o.stack === "other" && <Text label="Your stack" value={o.otherStack} onChange={(v) => set({ otherStack: v })} placeholder="Flutter, Rails, Astro…" missing={!o.otherStack.trim()} />}

        <Choice<Design> label="Which sign-in design?" value={o.design} onChange={(design) => set({ design })} options={[
          { value: "hosted", label: "Ready-made page", hint: "Users go to a sooauth page and come back. Least code." },
          { value: "widget", label: "Ready-made widget", hint: "A login box embedded in your own page." },
          { value: "own", label: "My own design", hint: "Your own forms, built on the sooauth API." },
        ]} />

        {o.design === "own" && (
          <>
            <Choice label="Starting point" value={o.matchExisting ? "match" : "new"} onChange={(v) => set({ matchExisting: v === "match" })} options={[
              { value: "match", label: "Match my existing design" },
              { value: "new", label: "Design it from scratch" },
            ]} />
            <Choice<ColorScheme> label="Color scheme" value={o.colorScheme} onChange={(colorScheme) => set({ colorScheme })} options={[
              { value: "system", label: "Follow system" },
              { value: "light", label: "Light" },
              { value: "dark", label: "Dark" },
            ]} />
          </>
        )}

        <fieldset>
          <legend className="text-[13px] font-semibold text-fg">Accent color</legend>
          <div className="mt-2 flex flex-wrap items-center gap-2">
            {accents.map((c) => (
              <button key={c} type="button" aria-label={`Use ${c}`} aria-pressed={o.accent === c} onClick={() => set({ accent: c })}
                className={cn("h-8 w-8 border-2", o.accent === c ? "border-fg" : "border-transparent")} style={{ background: c }} />
            ))}
            <input aria-label="Custom accent color" type="color" value={o.accent} onChange={(e) => set({ accent: e.target.value.toUpperCase() })} className="h-8 w-10 cursor-pointer border border-line bg-bg" />
            <span className="font-mono text-[13px] text-fg-muted">{o.accent}</span>
          </div>
          {o.design !== "own" && <p className="mt-2 text-[12px] text-fg-muted">Ready-made designs take this from dashboard → Theme, so the prompt lists it as a step for you.</p>}
        </fieldset>

        <Choice<SignIn> label="How do users sign in?" value={o.signIn} onChange={(signIn) => set({ signIn })} options={[
          { value: "password", label: "Password" },
          { value: "otp", label: "One-time code" },
          { value: "both", label: "Password or code" },
        ]} />
        <Choice<TwoStep> label="Two-step verification" value={o.twoStep} onChange={(twoStep) => set({ twoStep })} options={[
          { value: "off", label: "Off" },
          { value: "optional", label: "Optional" },
          { value: "required", label: "Required" },
        ]} />
        <Choice<EmailVerify> label="Email verification" value={o.emailVerify} onChange={(emailVerify) => set({ emailVerify })} options={[
          { value: "link", label: "Link" },
          { value: "code", label: "6-digit code" },
          { value: "both", label: "Link and code" },
          { value: "off", label: "Off" },
        ]} />

        <fieldset>
          <legend className="text-[13px] font-semibold text-fg">Extras</legend>
          <div className="mt-2 flex flex-wrap gap-x-5 gap-y-2 text-[13px] text-fg-muted">
            {["Google", "GitHub"].map((name) => (
              <label key={name} className="flex cursor-pointer items-center gap-2">
                <input type="checkbox" checked={o.social.includes(name)} onChange={() => toggleSocial(name)} />{name} login
              </label>
            ))}
            <label className="flex cursor-pointer items-center gap-2">
              <input type="checkbox" checked={o.extraFields} onChange={() => set({ extraFields: !o.extraFields })} />Extra sign-up fields
            </label>
          </div>
        </fieldset>

        <div className="grid gap-4 sm:grid-cols-2">
          <Text label="Client ID" value={o.clientId} onChange={(v) => set({ clientId: v })} placeholder="app_…" missing={!o.clientId.trim()} hint="Dashboard → Keys & URLs" />
          <Text label="Your app's URL" value={o.appUrl} onChange={(v) => set({ appUrl: v })} hint={<>Add <span className="font-mono">{callbackUrl(o)}</span> as a redirect URI.</>} />
          <Text label="Auth server" value={o.issuer} onChange={(v) => set({ issuer: v.trim().replace(/\/+$/, "") })} hint="Change this if you self-host." />
          <Text label="UI language (optional)" value={o.language} onChange={(v) => set({ language: v })} placeholder="Turkish" />
        </div>
      </div>

      {missing.length > 0 && (
        <p role="status" className="mt-6 border-l-2 border-accent bg-accent/5 px-3 py-2 text-[14px] text-fg">
          Still empty: {missing.join(", ")}. You can copy now; the prompt tells your assistant to ask you for it.
        </p>
      )}
      {o.design !== "hosted" && !isBrowserOnly(o.stack) && o.stack !== "other" && o.stack !== "nextjs" && (
        <p className="mt-3 text-[13px] text-fg-muted">Your server keeps the session, so tokens stay out of the browser.</p>
      )}

      <pre className="mt-4 max-h-[460px] overflow-auto whitespace-pre-wrap border border-line bg-bg px-4 py-4 font-mono text-[12px] leading-[1.65] text-fg">{prompt}</pre>
      <button type="button" onClick={copy} className={buttonClass("primary", "sm", "mt-4")}>{copied ? "Copied" : "Copy prompt"}</button>
    </section>
  );
}
