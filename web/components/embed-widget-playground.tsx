"use client";

import { useState } from "react";
import { Check, Copy, Eye, EyeOff, RotateCcw, Sparkles } from "lucide-react";
import { authUrl } from "../lib/site";

const PRESET_COLORS = [
  { label: "Hot Red", hex: "#FF3B3B" },
  { label: "Soft Beige", hex: "#FFF4E6", textDark: true },
  { label: "Indigo", hex: "#6366F1" },
  { label: "Emerald", hex: "#10B981" },
  { label: "Sunset", hex: "#F97316" },
  { label: "Monochrome", hex: "#18181B" },
];

export function EmbedWidgetPlayground() {
  const [mode, setMode] = useState<"signin" | "signup">("signin");
  const [brandName, setBrandName] = useState("Acme Corp");
  const [accentColor, setAccentColor] = useState("#FF3B3B");
  const [googleEnabled, setGoogleEnabled] = useState(true);
  const [githubEnabled, setGithubEnabled] = useState(true);
  const [rememberMe, setRememberMe] = useState(true);
  const [previewTheme, setPreviewTheme] = useState<"light" | "dark">("light");

  // Interactive password testing state
  const [testPassword, setTestPassword] = useState("");
  const [showPassword, setShowPassword] = useState(false);
  const [testEmail, setTestEmail] = useState("");
  const [copied, setCopied] = useState(false);

  // Password policy controls
  const [minLength, setMinLength] = useState(8);
  const [requireUpper, setRequireUpper] = useState(true);
  const [requireNumber, setRequireNumber] = useState(true);
  const [requireSpecial, setRequireSpecial] = useState(false);

  const resetDefaults = () => {
    setMode("signin");
    setBrandName("Acme Corp");
    setAccentColor("#FF3B3B");
    setGoogleEnabled(true);
    setGithubEnabled(true);
    setRememberMe(true);
    setPreviewTheme("light");
    setTestPassword("");
    setMinLength(8);
    setRequireUpper(true);
    setRequireNumber(true);
    setRequireSpecial(false);
  };

  // Rule verification
  const rules = [
    {
      id: "length",
      label: `At least ${minLength} characters`,
      ok: testPassword.length >= minLength,
    },
    ...(requireUpper
      ? [{ id: "upper", label: "One uppercase letter", ok: /[A-Z]/.test(testPassword) }]
      : []),
    ...(requireNumber
      ? [{ id: "number", label: "One number", ok: /[0-9]/.test(testPassword) }]
      : []),
    ...(requireSpecial
      ? [{ id: "special", label: "One special character", ok: /[^A-Za-z0-9]/.test(testPassword) }]
      : []),
  ];

  const codeSnippet = `<div id="sooauth-widget"></div>
<script
  src="${authUrl}/v1/widget/embed.js"
  data-client-id="app_your_client_id"
  data-mode="${mode}"
  data-target="#sooauth-widget"
  async
></script>`;

  const copySnippet = () => {
    navigator.clipboard.writeText(codeSnippet);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  // Determine text color on top of accent color
  const isAccentLight =
    accentColor.toLowerCase() === "#fff4e6" ||
    accentColor.toLowerCase() === "#ffffff" ||
    accentColor.toLowerCase() === "#e8ff3f";
  const accentFg = isAccentLight ? "#051B23" : "#FFFFFF";

  return (
    <div className="w-full rounded-2xl border border-line bg-surface overflow-hidden shadow-xl font-sans">
      {/* Playground Header Bar */}
      <div className="flex flex-wrap items-center justify-between gap-4 border-b border-line bg-surface-subtle px-6 py-4">
        <div className="flex items-center gap-3">
          <div className="flex h-9 w-9 items-center justify-center rounded-lg bg-accent/10 text-accent">
            <Sparkles size={18} />
          </div>
          <div>
            <h3 className="font-sans text-base font-semibold text-fg tracking-tight text-balance">
              Interactive Embed Widget Playground
            </h3>
            <p className="text-xs text-fg-muted text-pretty">
              Customize social providers, themes, and branding to preview <span className="font-mono text-[11px] text-fg">embed.js</span> in real time.
            </p>
          </div>
        </div>

        <div className="flex items-center gap-2">
          <button
            type="button"
            onClick={resetDefaults}
            className="inline-flex items-center gap-1.5 rounded-lg border border-line bg-field px-3 py-1.5 text-xs font-medium text-fg-muted hover:text-fg hover:border-fg transition active:scale-[0.96]"
          >
            <RotateCcw size={13} />
            <span>Reset</span>
          </button>
          <button
            type="button"
            onClick={copySnippet}
            className="inline-flex items-center gap-1.5 rounded-lg border border-accent bg-accent px-3 py-1.5 text-xs font-semibold text-white hover:bg-accent-hover transition shadow-sm active:scale-[0.96]"
          >
            {copied ? <Check size={13} /> : <Copy size={13} />}
            <span>{copied ? "Copied snippet!" : "Copy snippet"}</span>
          </button>
        </div>
      </div>

      {/* Main Split Body: Controls (Left) & Live Canvas (Right) */}
      <div className="grid grid-cols-1 lg:grid-cols-12 min-h-[580px]">
        {/* Controls Column */}
        <div className="lg:col-span-5 border-b lg:border-b-0 lg:border-r border-line bg-surface p-6 space-y-6 overflow-y-auto max-h-[700px]">
          {/* Mode Selector */}
          <div className="space-y-2">
            <label className="block text-xs font-bold uppercase tracking-wider text-fg-muted">
              Display Mode
            </label>
            <div className="grid grid-cols-2 gap-1 rounded-lg border border-line bg-field p-1">
              <button
                type="button"
                onClick={() => setMode("signin")}
                className={`rounded-md py-2 text-xs font-semibold transition active:scale-[0.96] ${
                  mode === "signin"
                    ? "bg-accent text-white shadow-sm"
                    : "text-fg-muted hover:text-fg"
                }`}
              >
                Sign in (Login)
              </button>
              <button
                type="button"
                onClick={() => setMode("signup")}
                className={`rounded-md py-2 text-xs font-semibold transition active:scale-[0.96] ${
                  mode === "signup"
                    ? "bg-accent text-white shadow-sm"
                    : "text-fg-muted hover:text-fg"
                }`}
              >
                Sign up (Registration)
              </button>
            </div>
          </div>

          {/* Brand Name */}
          <div className="space-y-2">
            <label className="block text-xs font-bold uppercase tracking-wider text-fg-muted">
              Brand Name
            </label>
            <input
              type="text"
              value={brandName}
              onChange={(e) => setBrandName(e.target.value)}
              placeholder="e.g. Acme Corp"
              className="w-full rounded-lg border border-line bg-field px-3 py-2 text-sm text-fg outline-none focus:border-accent focus:ring-2 focus:ring-accent/20 transition"
            />
          </div>

          {/* Social OAuth Providers */}
          <div className="space-y-2">
            <label className="block text-xs font-bold uppercase tracking-wider text-fg-muted">
              Social Providers
            </label>
            <div className="grid grid-cols-2 gap-2">
              <label className="flex items-center gap-2.5 rounded-lg border border-line bg-field p-2.5 text-xs font-medium text-fg cursor-pointer hover:border-fg transition">
                <input
                  type="checkbox"
                  checked={googleEnabled}
                  onChange={(e) => setGoogleEnabled(e.target.checked)}
                  className="h-4 w-4 rounded border-line text-accent accent-[#FF3B3B]"
                />
                <span>Google OAuth</span>
              </label>
              <label className="flex items-center gap-2.5 rounded-lg border border-line bg-field p-2.5 text-xs font-medium text-fg cursor-pointer hover:border-fg transition">
                <input
                  type="checkbox"
                  checked={githubEnabled}
                  onChange={(e) => setGithubEnabled(e.target.checked)}
                  className="h-4 w-4 rounded border-line text-accent accent-[#FF3B3B]"
                />
                <span>GitHub OAuth</span>
              </label>
            </div>
          </div>

          {/* Accent Color Picker */}
          <div className="space-y-2">
            <div className="flex items-center justify-between">
              <label className="text-xs font-bold uppercase tracking-wider text-fg-muted">
                Accent Color
              </label>
              <span className="font-mono text-xs text-fg-muted uppercase">{accentColor}</span>
            </div>
            <div className="flex flex-wrap items-center gap-2">
              {PRESET_COLORS.map((preset) => (
                <button
                  key={preset.hex}
                  type="button"
                  onClick={() => setAccentColor(preset.hex)}
                  title={preset.label}
                  style={{ backgroundColor: preset.hex }}
                  className={`h-7 w-7 rounded-full border border-line transition transform hover:scale-110 active:scale-95 ${
                    accentColor.toLowerCase() === preset.hex.toLowerCase()
                      ? "ring-2 ring-offset-2 ring-fg"
                      : ""
                  }`}
                />
              ))}
              <input
                type="color"
                value={accentColor.startsWith("#") ? accentColor : "#FF3B3B"}
                onChange={(e) => setAccentColor(e.target.value)}
                className="h-7 w-7 rounded-full border border-line cursor-pointer p-0 bg-transparent"
                title="Custom color"
              />
            </div>
          </div>

          {/* Session & Options */}
          <div className="space-y-2">
            <label className="block text-xs font-bold uppercase tracking-wider text-fg-muted">
              Widget Options
            </label>
            <div className="space-y-2">
              <label className="flex items-center gap-2.5 text-xs text-fg cursor-pointer">
                <input
                  type="checkbox"
                  checked={rememberMe}
                  onChange={(e) => setRememberMe(e.target.checked)}
                  className="h-4 w-4 rounded border-line text-accent accent-[#FF3B3B]"
                />
                <span>Show &quot;Remember me&quot; session checkbox</span>
              </label>
            </div>
          </div>

          {/* Password Policy (Sign up mode only) */}
          {mode === "signup" && (
            <div className="space-y-3 rounded-xl border border-line bg-field/60 p-4">
              <label className="block text-xs font-bold uppercase tracking-wider text-fg">
                Password Policy Rules
              </label>
              <div className="space-y-2 text-xs">
                <div className="flex items-center justify-between">
                  <span className="text-fg-muted">Minimum length: {minLength}</span>
                  <input
                    type="range"
                    min="8"
                    max="16"
                    value={minLength}
                    onChange={(e) => setMinLength(Number(e.target.value))}
                    className="w-24 accent-[#FF3B3B]"
                  />
                </div>
                <label className="flex items-center gap-2 text-fg-muted cursor-pointer hover:text-fg">
                  <input
                    type="checkbox"
                    checked={requireUpper}
                    onChange={(e) => setRequireUpper(e.target.checked)}
                    className="accent-[#FF3B3B]"
                  />
                  <span>Require uppercase letter</span>
                </label>
                <label className="flex items-center gap-2 text-fg-muted cursor-pointer hover:text-fg">
                  <input
                    type="checkbox"
                    checked={requireNumber}
                    onChange={(e) => setRequireNumber(e.target.checked)}
                    className="accent-[#FF3B3B]"
                  />
                  <span>Require number</span>
                </label>
                <label className="flex items-center gap-2 text-fg-muted cursor-pointer hover:text-fg">
                  <input
                    type="checkbox"
                    checked={requireSpecial}
                    onChange={(e) => setRequireSpecial(e.target.checked)}
                    className="accent-[#FF3B3B]"
                  />
                  <span>Require special symbol</span>
                </label>
              </div>
            </div>
          )}

          {/* Canvas Background Toggle */}
          <div className="space-y-2 pt-2 border-t border-line">
            <div className="flex items-center justify-between">
              <span className="text-xs font-medium text-fg-muted">Preview Theme:</span>
              <div className="flex items-center gap-1 rounded-md border border-line bg-field p-0.5 text-[11px]">
                <button
                  type="button"
                  onClick={() => setPreviewTheme("light")}
                  className={`px-2 py-0.5 rounded transition ${
                    previewTheme === "light"
                      ? "bg-accent text-white font-semibold"
                      : "text-fg-muted hover:text-fg"
                  }`}
                >
                  ☀️ Light
                </button>
                <button
                  type="button"
                  onClick={() => setPreviewTheme("dark")}
                  className={`px-2 py-0.5 rounded transition ${
                    previewTheme === "dark"
                      ? "bg-accent text-white font-semibold"
                      : "text-fg-muted hover:text-fg"
                  }`}
                >
                  🌙 Dark
                </button>
              </div>
            </div>
          </div>
        </div>

        {/* Live Preview Canvas Column */}
        <div
          className={`lg:col-span-7 flex flex-col items-center justify-center p-6 md:p-10 transition-colors duration-200 ${
            previewTheme === "dark"
              ? "bg-[#0E0F11] text-[#F5F6F6]"
              : "bg-[#FAFAFA] text-[#051B23]"
          }`}
        >
          {/* Simulated Embed Widget Card */}
          <div
            className={`w-full max-w-[380px] rounded-2xl border p-7 transition-all shadow-2xl ${
              previewTheme === "dark"
                ? "bg-[#161719] border-[#2A2B2E] text-[#F5F6F6]"
                : "bg-[#FFFFFF] border-[#E2E4E6] text-[#051B23]"
            }`}
          >
            {/* Widget Header */}
            <div className="text-center mb-6">
              <div className="inline-flex items-center justify-center gap-2 mb-2">
                <span
                  style={{ backgroundColor: accentColor }}
                  className="h-3 w-3 rounded-full inline-block"
                />
                <h4 className="text-lg font-bold tracking-tight">
                  {brandName || "sooauth"}
                </h4>
              </div>
              <p className={`text-xs ${previewTheme === "dark" ? "text-[#9D9E9F]" : "text-[#5D6B70]"}`}>
                {mode === "signin"
                  ? "Sign in to continue to your account"
                  : "Create an account to get started"}
              </p>
            </div>

            {/* Social Buttons */}
            {(googleEnabled || githubEnabled) && (
              <div className="space-y-2.5 mb-5">
                {googleEnabled && (
                  <button
                    type="button"
                    className={`w-full flex items-center justify-center gap-3 rounded-lg border py-2.5 px-4 text-xs font-semibold transition active:scale-[0.96] ${
                      previewTheme === "dark"
                        ? "bg-[#1C1D1F] border-[#353537] text-[#F5F6F6] hover:bg-[#252629]"
                        : "bg-[#FFFFFF] border-[#D0D2D4] text-[#051B23] hover:bg-[#F5F5F5]"
                    }`}
                  >
                    <svg viewBox="0 0 24 24" width="16" height="16">
                      <path fill="#4285F4" d="M23.745 12.27c0-.7-.06-1.4-.19-2.07H12v4.51h6.6c-.29 1.52-1.14 2.82-2.4 3.68v3.05h3.88c2.27-2.09 3.66-5.17 3.66-9.17z"/>
                      <path fill="#34A853" d="M12 24c3.24 0 5.95-1.08 7.93-2.91l-3.88-3.05c-1.08.72-2.45 1.16-4.05 1.16-3.12 0-5.77-2.1-6.72-4.93H1.25v3.15C3.27 21.43 7.35 24 12 24z"/>
                      <path fill="#FBBC05" d="M5.28 14.27c-.25-.72-.38-1.49-.38-2.27s.14-1.55.38-2.27V6.58H1.25C.45 8.18 0 10.03 0 12s.45 3.82 1.25 5.42l4.03-3.15z"/>
                      <path fill="#EA4335" d="M12 4.75c1.77 0 3.35.61 4.6 1.8l3.42-3.42C17.95 1.19 15.24 0 12 0 7.35 0 3.27 2.57 1.25 6.58l4.03 3.15c.95-2.83 3.6-4.98 6.72-4.98z"/>
                    </svg>
                    <span>Continue with Google</span>
                  </button>
                )}

                {githubEnabled && (
                  <button
                    type="button"
                    className={`w-full flex items-center justify-center gap-3 rounded-lg border py-2.5 px-4 text-xs font-semibold transition active:scale-[0.96] ${
                      previewTheme === "dark"
                        ? "bg-[#1C1D1F] border-[#353537] text-[#F5F6F6] hover:bg-[#252629]"
                        : "bg-[#FFFFFF] border-[#D0D2D4] text-[#051B23] hover:bg-[#F5F5F5]"
                    }`}
                  >
                    <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor">
                      <path d="M12 0C5.37 0 0 5.37 0 12c0 5.31 3.435 9.795 8.205 11.385.6.105.825-.255.825-.57 0-.285-.015-1.23-.015-2.235-3.015.555-3.795-.735-4.035-1.41-.135-.345-.72-1.41-1.23-1.695-.42-.225-1.02-.78-.015-.795.945-.015 1.62.87 1.845 1.23 1.08 1.815 2.805 1.305 3.495.99.105-.78.42-1.305.765-1.605-2.67-.3-5.46-1.335-5.46-5.925 0-1.305.465-2.385 1.23-3.225-.12-.3-.54-1.53.12-3.18 0 0 1.005-.315 3.3 1.23.96-.27 1.98-.405 3-.405s2.04.135 3 .405c2.295-1.56 3.3-1.23 3.3-1.23.66 1.65.24 2.88.12 3.18.765.84 1.23 1.905 1.23 3.225 0 4.605-2.805 5.625-5.475 5.925.435.375.81 1.095.81 2.22 0 1.605-.015 2.895-.015 3.3 0 .315.225.69.825.57A12.02 12.02 0 0024 12c0-6.63-5.37-12-12-12z"/>
                    </svg>
                    <span>Continue with GitHub</span>
                  </button>
                )}

                <div className="relative my-4 text-center">
                  <div className={`absolute inset-0 flex items-center ${previewTheme === "dark" ? "border-t border-[#2A2B2E]" : "border-t border-[#E5E7EB]"}`} />
                  <span className={`relative px-3 text-[11px] uppercase tracking-wider font-semibold ${
                    previewTheme === "dark" ? "bg-[#161719] text-[#9D9E9F]" : "bg-[#FFFFFF] text-[#6B7280]"
                  }`}>
                    or with email
                  </span>
                </div>
              </div>
            )}

            {/* Email Form */}
            <form onSubmit={(e) => e.preventDefault()} className="space-y-3.5">
              <div>
                <label className={`block text-xs font-semibold mb-1 ${previewTheme === "dark" ? "text-[#D1D5DB]" : "text-[#374151]"}`}>
                  Email address
                </label>
                <input
                  type="email"
                  value={testEmail}
                  onChange={(e) => setTestEmail(e.target.value)}
                  placeholder="you@example.com"
                  className={`w-full rounded-lg border px-3 py-2 text-xs outline-none transition focus:ring-2 focus:ring-accent/20 ${
                    previewTheme === "dark"
                      ? "bg-[#1C1D1F] border-[#353537] text-[#F5F6F6] focus:border-accent"
                      : "bg-[#FFFFFF] border-[#D0D2D4] text-[#051B23] focus:border-accent"
                  }`}
                />
              </div>

              <div>
                <div className="flex items-center justify-between mb-1">
                  <label className={`block text-xs font-semibold ${previewTheme === "dark" ? "text-[#D1D5DB]" : "text-[#374151]"}`}>
                    Password
                  </label>
                  {mode === "signin" && (
                    <a href="#forgot" onClick={(e) => e.preventDefault()} className="text-[11px] text-accent hover:underline">
                      Forgot password?
                    </a>
                  )}
                </div>
                <div className="relative">
                  <input
                    type={showPassword ? "text" : "password"}
                    value={testPassword}
                    onChange={(e) => setTestPassword(e.target.value)}
                    placeholder="••••••••"
                    className={`w-full rounded-lg border px-3 py-2 pr-9 text-xs outline-none transition focus:ring-2 focus:ring-accent/20 ${
                      previewTheme === "dark"
                        ? "bg-[#1C1D1F] border-[#353537] text-[#F5F6F6] focus:border-accent"
                        : "bg-[#FFFFFF] border-[#D0D2D4] text-[#051B23] focus:border-accent"
                    }`}
                  />
                  <button
                    type="button"
                    onClick={() => setShowPassword(!showPassword)}
                    className="absolute right-2.5 top-1/2 -translate-y-1/2 text-fg-muted hover:text-fg p-0.5"
                  >
                    {showPassword ? <EyeOff size={14} /> : <Eye size={14} />}
                  </button>
                </div>
              </div>

              {/* Password Checklist in Sign Up mode */}
              {mode === "signup" && (
                <ul className="space-y-1 pt-1 pb-1">
                  {rules.map((rule) => (
                    <li
                      key={rule.id}
                      className={`flex items-center gap-2 text-[11px] transition-colors ${
                        !testPassword
                          ? previewTheme === "dark" ? "text-[#9D9E9F]" : "text-[#6B7280]"
                          : rule.ok
                          ? "text-emerald-500 font-medium"
                          : "text-red-500"
                      }`}
                    >
                      <span className="w-3.5 text-center font-bold">
                        {!testPassword ? "○" : rule.ok ? "✓" : "✗"}
                      </span>
                      <span>{rule.label}</span>
                    </li>
                  ))}
                </ul>
              )}

              {/* Remember me option */}
              {rememberMe && mode === "signin" && (
                <label className={`flex items-center gap-2 text-xs cursor-pointer ${previewTheme === "dark" ? "text-[#9D9E9F]" : "text-[#5D6B70]"}`}>
                  <input
                    type="checkbox"
                    defaultChecked
                    className="rounded border-line accent-[#FF3B3B]"
                  />
                  <span>Remember me on this device</span>
                </label>
              )}

              {/* Submit CTA */}
              <button
                type="submit"
                style={{ backgroundColor: accentColor, color: accentFg }}
                className="w-full rounded-lg py-2.5 px-4 text-xs font-bold tracking-tight transition hover:opacity-95 shadow-md active:scale-[0.96]"
              >
                {mode === "signin" ? "Sign in to account" : "Create account"}
              </button>
            </form>

            {/* Widget Mode Switcher */}
            <div className={`mt-5 pt-4 border-t text-center text-xs ${previewTheme === "dark" ? "border-[#2A2B2E] text-[#9D9E9F]" : "border-[#E5E7EB] text-[#6B7280]"}`}>
              <span>
                {mode === "signin" ? "Need an account? " : "Already registered? "}
              </span>
              <button
                type="button"
                onClick={() => setMode(mode === "signin" ? "signup" : "signin")}
                className="font-semibold text-accent hover:underline"
              >
                {mode === "signin" ? "Sign up" : "Sign in"}
              </button>
            </div>

            {/* Brand Footer */}
            <div className="mt-4 text-center">
              <span className={`text-[10px] tracking-wide uppercase font-mono ${previewTheme === "dark" ? "text-[#6E7175]" : "text-[#9CA3AF]"}`}>
                Secured by sooauth
              </span>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
