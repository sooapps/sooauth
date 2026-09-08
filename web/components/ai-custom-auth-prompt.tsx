"use client";

import { useCallback } from "react";
import { buildAiCustomAuthPrompt } from "../lib/ai-setup-prompt";
import { adminUrl, authUrl } from "../lib/site";
import { buttonClass } from "./ui/button";

export function AiCustomAuthPromptBlock() {
  const prompt = buildAiCustomAuthPrompt({
    issuer: authUrl,
    dashboardUrl: adminUrl,
  });

  const copy = useCallback(() => {
    void navigator.clipboard.writeText(prompt);
  }, [prompt]);

  return (
    <section
      id="ai-prompt"
      className="mt-10 border border-line bg-bg-subtle p-6"
    >
      <h2 className="font-sans text-[20px] font-semibold text-fg">
        AI setup prompt (custom UI)
      </h2>
      <p className="mt-3 max-w-[58ch] text-[15px] leading-[1.6] text-fg-muted">
        Paste into Cursor, ChatGPT, or Claude. Fill in your stack and URLs — the
        model wires custom login/sign-up against sooauth (widget config, password
        policy, email link/code verification, social callback). No embed.js.
      </p>
      <pre className="mt-4 max-h-[420px] overflow-auto border border-line bg-bg px-4 py-4 font-mono text-[12px] leading-[1.65] text-fg whitespace-pre-wrap">
        {prompt}
      </pre>
      <button
        type="button"
        onClick={copy}
        className={buttonClass("primary", "sm", "mt-4")}
      >
        Copy prompt
      </button>
    </section>
  );
}
