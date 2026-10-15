"use client";

import { useCallback } from "react";
import { buildAiSetupPrompt } from "../lib/ai-setup-prompt";
import { adminUrl, authUrl, oidcDiscoveryUrl } from "../lib/site";
import { buttonClass } from "./ui/button";

export function AiSetupPromptBlock() {
  const prompt = buildAiSetupPrompt({
    issuer: authUrl,
    discoveryUrl: oidcDiscoveryUrl,
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
        AI setup prompt
      </h2>
      <p className="mt-3 max-w-[58ch] text-[15px] leading-[1.6] text-fg-muted">
        Paste this into Cursor, ChatGPT, or Claude. Fill in your stack and callback
        URL — the model implements OIDC against sooauth without adding social keys
        to your app.
      </p>
      <p className="mt-3 max-w-[58ch] border-l-2 border-accent pl-3 text-[14px] leading-[1.6] text-fg">
        Signed in? <a href={`${adminUrl}`} className="underline underline-offset-2">Dashboard → Integration → AI setup prompt</a>{" "}
        gives you this prompt already filled in with your client ID, redirect URLs, sign-in methods and sign-up fields —
        nothing to edit by hand.
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
