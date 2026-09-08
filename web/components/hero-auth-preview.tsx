"use client";

import { useState } from "react";
import { Check, Copy, Fingerprint, KeyRound, Lock, ShieldCheck, Terminal } from "lucide-react";

type TabId = "ui" | "jwt" | "env";

export function HeroAuthPreview() {
  const [activeTab, setActiveTab] = useState<TabId>("ui");
  const [copied, setCopied] = useState(false);

  const envSnippet = `SOOAUTH_ISSUER="https://auth.sooauth.com"
SOOAUTH_CLIENT_ID="app_live_8392"
SOOAUTH_REDIRECT_URI="http://localhost:3000/callback"`;

  const copyEnv = () => {
    navigator.clipboard.writeText(envSnippet);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <div className="relative w-full rounded-lg border border-line bg-bg-subtle shadow-2xl overflow-hidden font-sans">
      {/* Window Titlebar */}
      <div className="flex items-center justify-between border-b border-line bg-bg/80 px-4 py-3 backdrop-blur">
        <div className="flex items-center gap-2">
          <span className="h-2.5 w-2.5 rounded-full bg-red-500/80" />
          <span className="h-2.5 w-2.5 rounded-full bg-yellow-500/80" />
          <span className="h-2.5 w-2.5 rounded-full bg-green-500/80" />
          <span className="ml-2 font-mono text-[11px] text-fg-muted">
            auth.sooauth.com
          </span>
        </div>
        <div className="flex items-center gap-1.5 font-mono text-[10px] text-lime-400">
          <span className="inline-block h-1.5 w-1.5 rounded-full bg-lime-400 animate-pulse" />
          <span>OIDC / LIVE</span>
        </div>
      </div>

      {/* Mode Tabs */}
      <div className="flex border-b border-line bg-bg font-mono text-xs">
        <button
          type="button"
          onClick={() => setActiveTab("ui")}
          className={`flex items-center gap-1.5 px-4 py-2.5 border-r border-line transition-colors ${
            activeTab === "ui"
              ? "bg-bg-subtle text-fg font-semibold border-b-2 border-b-fg"
              : "text-fg-muted hover:text-fg"
          }`}
        >
          <Lock size={12} />
          <span>Hosted UI</span>
        </button>
        <button
          type="button"
          onClick={() => setActiveTab("jwt")}
          className={`flex items-center gap-1.5 px-4 py-2.5 border-r border-line transition-colors ${
            activeTab === "jwt"
              ? "bg-bg-subtle text-fg font-semibold border-b-2 border-b-fg"
              : "text-fg-muted hover:text-fg"
          }`}
        >
          <ShieldCheck size={12} />
          <span>Verified JWT</span>
        </button>
        <button
          type="button"
          onClick={() => setActiveTab("env")}
          className={`flex items-center gap-1.5 px-4 py-2.5 transition-colors ${
            activeTab === "env"
              ? "bg-bg-subtle text-fg font-semibold border-b-2 border-b-fg"
              : "text-fg-muted hover:text-fg"
          }`}
        >
          <Terminal size={12} />
          <span>3 Env Vars</span>
        </button>
      </div>

      {/* Tab Contents */}
      <div className="p-6 md:p-7 min-h-[330px] flex flex-col justify-center">
        {activeTab === "ui" && (
          <div className="mx-auto w-full max-w-sm space-y-4">
            <div className="text-center">
              <div className="inline-flex h-9 w-9 items-center justify-center rounded-md border border-line bg-bg mb-2">
                <KeyRound size={18} className="text-fg" />
              </div>
              <h4 className="text-sm font-semibold text-fg">Sign in to your app</h4>
              <p className="text-xs text-fg-muted">Hosted by Sooauth · RFC 7636 PKCE</p>
            </div>

            {/* Passkey Button */}
            <button
              type="button"
              className="flex w-full items-center justify-center gap-2 rounded-md bg-fg px-4 py-2.5 text-xs font-semibold text-bg transition hover:opacity-90"
            >
              <Fingerprint size={16} />
              <span>Continue with Passkey</span>
            </button>

            {/* Social Logins */}
            <div className="grid grid-cols-2 gap-2">
              <button
                type="button"
                className="flex items-center justify-center gap-1.5 rounded-md border border-line bg-bg px-3 py-2 text-xs font-medium text-fg hover:bg-bg-subtle"
              >
                <span>GitHub</span>
              </button>
              <button
                type="button"
                className="flex items-center justify-center gap-1.5 rounded-md border border-line bg-bg px-3 py-2 text-xs font-medium text-fg hover:bg-bg-subtle"
              >
                <span>Google</span>
              </button>
            </div>

            {/* Magic Link */}
            <div className="relative flex items-center justify-center">
              <div className="absolute inset-0 flex items-center"><div className="w-full border-t border-line" /></div>
              <span className="relative bg-bg-subtle px-2 font-mono text-[10px] uppercase text-fg-muted">or magic link</span>
            </div>

            <div className="flex gap-2">
              <input
                type="email"
                disabled
                placeholder="developer@indiesaas.com"
                className="flex-1 rounded-md border border-line bg-bg px-3 py-1.5 text-xs text-fg-muted"
              />
              <button
                type="button"
                disabled
                className="rounded-md border border-line bg-bg px-3 py-1.5 text-xs font-medium text-fg-muted"
              >
                Send
              </button>
            </div>

            <p className="text-center font-mono text-[10px] text-fg-muted">
              Zero database writes in your application.
            </p>
          </div>
        )}

        {activeTab === "jwt" && (
          <div className="space-y-3 font-mono text-xs">
            <div className="flex items-center justify-between border-b border-line/60 pb-2 text-[11px] text-fg-muted">
              <span className="flex items-center gap-1.5 text-lime-400">
                <Check size={13} />
                <span>RS256 Signature Validated</span>
              </span>
              <span>claims: 6</span>
            </div>
            <pre className="overflow-x-auto rounded bg-bg p-4 text-[12px] leading-[1.6] text-fg border border-line">
{`{
  "iss": "https://auth.sooauth.com",
  "sub": "usr_01jk98az34m",
  "aud": "app_live_8392",
  "email": "developer@indiesaas.com",
  "email_verified": true,
  "auth_method": "webauthn_passkey",
  "db_migrations": 0
}`}
            </pre>
            <p className="text-[11px] text-fg-muted">
              Standard OIDC ID token. Any JWT validator or framework middleware decodes it instantly.
            </p>
          </div>
        )}

        {activeTab === "env" && (
          <div className="space-y-3 font-mono text-xs">
            <div className="flex items-center justify-between border-b border-line/60 pb-2 text-[11px] text-fg-muted">
              <span>.env.local</span>
              <button
                type="button"
                onClick={copyEnv}
                className="flex items-center gap-1 text-fg hover:text-fg-muted"
              >
                {copied ? <Check size={12} className="text-lime-400" /> : <Copy size={12} />}
                <span>{copied ? "Copied" : "Copy"}</span>
              </button>
            </div>
            <pre className="overflow-x-auto rounded bg-bg p-4 text-[12px] leading-[1.8] text-fg border border-line select-all">
{envSnippet}
            </pre>
            <div className="rounded border border-line bg-bg p-3 text-[11px] text-fg-muted">
              💡 No Prisma schema. No Drizzle migrations. Your database remains 100% clean for your own product data.
            </div>
          </div>
        )}
      </div>

      {/* Footer Status Bar */}
      <div className="flex items-center justify-between border-t border-line bg-bg/90 px-4 py-2 font-mono text-[11px] text-fg-muted">
        <span>RFC 6749 · RFC 7636</span>
        <span className="text-lime-400">0 database migrations</span>
      </div>
    </div>
  );
}
