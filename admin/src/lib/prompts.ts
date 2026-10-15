// Builds copy-paste prompts for AI coding assistants, filled in with the
// project's real settings so nothing has to be edited by hand.

import type { AuthConfig, RegistrationField } from "../api";

export type AppType = "nextjs" | "spa" | "backend" | "custom" | "widget";

export const appTypes: AppType[] = ["nextjs", "spa", "backend", "custom", "widget"];

export type ProjectInfo = {
  projectName: string;
  issuer: string;
  clientId: string;
  tenantId: string;
  redirectUris: string[];
  dashboardUrl: string;
  providers: string[];
  auth: AuthConfig;
  fields: RegistrationField[];
  passwordHint: string;
  emailVerifyRequired: boolean;
  emailVerifyDelivery: string;
};

const stackNames: Record<AppType, string> = {
  nextjs: "Next.js (App Router)",
  spa: "React / Vue single-page app",
  backend: "Server-rendered backend (Node, Go, Python, PHP …)",
  custom: "My own login and sign-up pages (custom UI over the sooauth API)",
  widget: "Drop-in sooauth embed widget",
};

export const defaultCallback: Record<AppType, string> = {
  nextjs: "http://localhost:3000/api/auth/callback/oidc",
  spa: "http://localhost:5173/callback",
  backend: "http://localhost:8000/auth/callback",
  custom: "http://localhost:3000/login",
  widget: "http://localhost:3000/login",
};

function callbackFor(info: ProjectInfo, type: AppType) {
  return info.redirectUris.find((u) => u && u !== defaultCallback.nextjs) || info.redirectUris[0] || defaultCallback[type];
}

function describeAuth(info: ProjectInfo) {
  const a = info.auth;
  const ids = a.allowed_identifiers.join(", ");
  const mode =
    a.primary_auth_mode === "otp"
      ? "one-time code only (no passwords)"
      : a.primary_auth_mode === "both"
        ? "password OR a one-time code (the user chooses on the sign-in screen)"
        : "password";
  const lines = [
    `- Users sign in with: ${ids}`,
    `- Primary method: ${mode}`,
    `- Two-step verification: ${a.mfa_policy}${a.mfa_policy !== "off" ? ` (methods: ${a.mfa_methods.join(", ")})` : ""}`,
    `- Social login enabled: ${info.providers.length ? info.providers.join(", ") : "none"}`,
  ];
  if (a.primary_auth_mode !== "otp") lines.push(`- Password policy: ${info.passwordHint || "minimum 8 characters"}`);
  lines.push(`- Email verification: ${info.emailVerifyRequired ? `required (${info.emailVerifyDelivery})` : "not required"}`);
  return lines.join("\n");
}

function describeFields(info: ProjectInfo) {
  if (!info.fields.length) return "- No extra sign-up fields.";
  return info.fields
    .map((f) => {
      const opts = f.type === "select" ? (f.options_source?.type === "dynamic_api" ? " (options loaded from the widget config `options_url`)" : ` (options: ${(f.options_source?.options || []).map((o) => o.value).join(", ")})`) : "";
      return `- \`${f.id}\` — ${f.label}, ${f.type}${f.required ? ", required" : ""}${opts}`;
    })
    .join("\n");
}

function header(info: ProjectInfo, type: AppType) {
  return `Integrate sooauth authentication into my application "${info.projectName}".

## Project values (already filled in — use exactly these)
- Issuer / API base: ${info.issuer}
- OIDC discovery: ${info.issuer}/.well-known/openid-configuration
- client_id: ${info.clientId}
- tenant_id: ${info.tenantId}
- Redirect / callback URL: ${callbackFor(info, type)}
- Registered redirect URIs: ${info.redirectUris.join(", ") || "(none yet — add one in the dashboard)"}
- Dashboard: ${info.dashboardUrl}

## How sign-in is configured
${describeAuth(info)}

## Extra sign-up fields (sent as \`metadata\`, returned as \`custom_claims\` from /oauth/userinfo)
${describeFields(info)}

## My stack
- ${stackNames[type]}`;
}

function oidcPrompt(info: ProjectInfo, type: AppType) {
  return `${header(info, type)}

## Rules
1. Use the OIDC authorization code flow with PKCE (public client — never put a client secret in the browser).
2. Send users to the sooauth hosted sign-in page. It already handles ${info.auth.primary_auth_mode === "password" ? "passwords" : "one-time codes"}${info.auth.mfa_policy !== "off" ? ", two-step verification" : ""}${info.providers.length ? ` and ${info.providers.join("/")} sign-in` : ""}; do not build a login form.
3. Authorize: ${info.issuer}/oauth/authorize?client_id=${info.clientId}&response_type=code&scope=openid%20email%20profile&redirect_uri=<callback>&code_challenge=<S256>&code_challenge_method=S256&state=<random>
4. Exchange the code at ${info.issuer}/oauth/token (form-encoded: grant_type=authorization_code, code, redirect_uri, client_id, code_verifier). Validate the id_token against ${info.issuer}/.well-known/jwks.json or call ${info.issuer}/oauth/userinfo with the access token.
5. Refresh with grant_type=refresh_token at the same token endpoint.
6. Do NOT add Google/GitHub OAuth keys to my app — social login is configured in the sooauth dashboard only.

## Deliverables
- Environment variables (SOOAUTH_ISSUER=${info.issuer}, SOOAUTH_CLIENT_ID=${info.clientId}, SOOAUTH_REDIRECT_URI=${callbackFor(info, type)})
- Sign-in redirect and callback route
- Session storage (httpOnly cookie) and sign-out
- Middleware that protects private routes

Implement step by step and explain each file you change.`;
}

function customPrompt(info: ProjectInfo) {
  const a = info.auth;
  const steps: string[] = [
    `1. On page load: GET ${info.issuer}/v1/widget/config?client_id=${info.clientId} → providers, password_policy, auth_config, registration_schema (render fields from it), email_verify_*.`,
  ];
  if (a.primary_auth_mode !== "otp") {
    steps.push(`2. Sign-up: POST ${info.issuer}/auth/sign-up { ${a.allowed_identifiers.map((i) => (i === "email" ? "email" : i)).join(" | ")}, password, client_id: "${info.clientId}", metadata: { …extra fields } }.`);
    steps.push(`3. Sign-in: POST ${info.issuer}/auth/sign-in { identifier, password, client_id: "${info.clientId}" } → { access_token, refresh_token, expires_in }.`);
  }
  if (a.primary_auth_mode !== "password") {
    steps.push(`${steps.length + 1}. One-time code: POST ${info.issuer}/auth/otp/send { identifier, client_id } then POST ${info.issuer}/auth/otp/verify { identifier, code, client_id, metadata? }. A 422 { error: "registration_required", fields } means a new user must also send the listed metadata fields. 429 "too_many_attempts" means request a new code.`);
  }
  if (a.primary_auth_mode === "both") steps.push(`${steps.length + 1}. Show a "use a code instead" switch on the sign-in form so users can pick password or code.`);
  if (a.mfa_policy !== "off") {
    steps.push(`${steps.length + 1}. Two-step verification: sign-in may return 401 { error: "mfa_required", challenge, method } where method is "totp", "email_otp" or "totp_setup". Ask for the 6-digit code and POST ${info.issuer}/auth/mfa/verify { challenge, code, method }. For "totp_setup" first POST ${info.issuer}/auth/mfa/setup { challenge } and show the returned qr_code / secret. Resend email codes with POST ${info.issuer}/auth/mfa/resend { challenge }.`);
  }
  if (info.emailVerifyRequired) steps.push(`${steps.length + 1}. Email verification is required (${info.emailVerifyDelivery}): after sign-up show a check-email screen; codes go to POST ${info.issuer}/auth/verify-code { email, code, client_id }.`);
  if (info.providers.length) steps.push(`${steps.length + 1}. Social buttons (${info.providers.join(", ")}): link to ${info.issuer}/auth/social/<provider>/start?tenant_id=${info.tenantId}&client_id=${info.clientId}&return_to=<my login URL>. On return, exchange ?widget_code= with POST ${info.issuer}/v1/widget/exchange { code, client_id }.`);

  return `${header(info, "custom")}

## API contract
${steps.join("\n")}

## Rules
- Store tokens in an httpOnly cookie via my backend; never in localStorage.
- Mirror the password policy in the UI (${info.passwordHint || "minimum 8 characters"}).
- Every return_to / callback URL must be one of the registered redirect URIs.

## Deliverables
- Sign-in, sign-up${a.primary_auth_mode !== "password" ? ", code entry" : ""}${a.mfa_policy !== "off" ? ", two-step verification" : ""} screens
- A small API client for the endpoints above
- Session handling, protected routes and sign-out

Implement step by step and match my stack's conventions.`;
}

function widgetPrompt(info: ProjectInfo) {
  return `${header(info, "widget")}

## What to build
Add the sooauth drop-in widget to my login page. It already supports the sign-in methods above.

\`\`\`html
<div id="sooauth-widget"></div>
<script src="${info.issuer}/v1/widget/embed.js" data-client-id="${info.clientId}" data-mode="signin" data-target="#sooauth-widget" async></script>
\`\`\`

- Listen for the \`sooauth:login\` DOM event: \`event.detail.access_token\` is the user's token. Send it to my backend, verify it with ${info.issuer}/oauth/userinfo, and create my own session cookie.
- The page that hosts the widget must be one of the registered redirect URIs (social login returns there).
- Sign-out: clear my session cookie.

Implement step by step and explain each file you change.`;
}

export function buildPrompt(info: ProjectInfo, type: AppType) {
  if (type === "custom") return customPrompt(info);
  if (type === "widget") return widgetPrompt(info);
  return oidcPrompt(info, type);
}

export function snippetFor(info: ProjectInfo, type: AppType) {
  const cb = callbackFor(info, type);
  if (type === "widget") {
    return `<div id="sooauth-widget"></div>
<script src="${info.issuer}/v1/widget/embed.js"
  data-client-id="${info.clientId}" data-mode="signin"
  data-target="#sooauth-widget" async></script>`;
  }
  if (type === "custom") {
    return `const config = await fetch("${info.issuer}/v1/widget/config?client_id=${info.clientId}").then(r => r.json());`;
  }
  return `SOOAUTH_ISSUER=${info.issuer}
SOOAUTH_CLIENT_ID=${info.clientId}
SOOAUTH_REDIRECT_URI=${cb}`;
}
