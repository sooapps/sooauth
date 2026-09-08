export function buildAiSetupPrompt(opts?: {
  issuer?: string;
  discoveryUrl?: string;
  dashboardUrl?: string;
}) {
  const issuer = opts?.issuer ?? "https://auth.sooauth.com";
  const discovery =
    opts?.discoveryUrl ?? `${issuer}/.well-known/openid-configuration`;
  const dashboard = opts?.dashboardUrl ?? `${issuer}/dashboard/`;

  return `Integrate sooauth (hosted OpenID Connect) into my application.

## Auth provider (sooauth)
- Issuer: ${issuer}
- OIDC discovery: ${discovery}
- Dashboard (copy client_id + redirect URIs): ${dashboard}

## Rules
1. Use OIDC authorization code flow with PKCE (public client — no client secret in the browser).
2. Redirect users to sooauth hosted sign-in; do not build a custom login form unless needed.
3. Add an exact callback/redirect URI in sooauth dashboard → Integration before testing.
4. Exchange the authorization code at the token endpoint; use access token on /oauth/userinfo or validate JWT via JWKS.
5. Do NOT put Google, GitHub, Facebook, or X OAuth client IDs in my app. Social login is enabled in sooauth dashboard only.

## My app
- Stack: [e.g. Next.js 15 App Router / Hono / Go / React SPA]
- Callback URL: [e.g. https://myapp.com/api/auth/callback/oidc]
- Where auth is needed: [e.g. protected dashboard routes]

## Deliverables
- Environment variables list
- Callback route handler
- Session/token storage approach
- Protected route middleware
- Sign-in and sign-out UX

Implement step by step and explain each file you change.`;
}

export function buildAiCustomAuthPrompt(opts?: {
  issuer?: string;
  dashboardUrl?: string;
}) {
  const issuer = opts?.issuer ?? "https://auth.sooauth.com";
  const dashboard = opts?.dashboardUrl ?? `${issuer}/dashboard/`;

  return `Integrate sooauth **custom auth UI** into my application (do NOT use embed.js — I want my own login/sign-up pages).

## Auth provider (sooauth)
- Issuer / API base: ${issuer}
- Dashboard (client_id, redirect URIs, password policy, social toggles): ${dashboard}

## API contract
1. **Config (on page load)** — \`GET ${issuer}/v1/widget/config?client_id=APP_CLIENT_ID\`
   - Returns: \`providers\`, \`tenant_id\`, \`accent_color\`, \`brand_name\`, \`email_verify_required\`, \`email_verify_delivery\` (\`link\` | \`code\` | \`both\`), \`password_policy\` (min_length, require_uppercase, require_number, require_special, hint).
2. **Email sign-up** — \`POST ${issuer}/auth/sign-up\` JSON \`{ email, password, client_id, return_to? }\`
   - \`return_to\`: where to send the user after they click the verification link (must match a redirect URI host; prefer a \`/login\` URL).
   - If verification required: no \`access_token\` — show check-email UI.
3. **Email sign-in** — \`POST ${issuer}/auth/sign-in\` JSON \`{ email, password, client_id }\`
   - Success: \`{ access_token, email, expires_in, … }\`
   - Weak password: \`{ error: "weak_password", message: "<policy hint>" }\`
4. **Verify email (link)** — user clicks link in mail → \`${issuer}/auth/verify?token=…&client_id=…&return_to=…\` → redirect to app with \`?email_verified=1\`.
5. **Verify email (6-digit code)** — when \`email_verify_delivery\` is \`code\` or \`both\`:
   - \`POST ${issuer}/auth/verify-code\` JSON \`{ email, code, client_id }\`
   - Success: \`{ access_token, email, expires_in, message }\` (auto sign-in for app users).
   - Resend: \`POST ${issuer}/auth/resend-verification\` JSON \`{ email, client_id, return_to? }\`.
6. **Google / GitHub** — link to:
   \`${issuer}/auth/social/{provider}/start?tenant_id=TENANT_ID&client_id=APP_CLIENT_ID&return_to=CALLBACK_URL\`
   - After redirect, URL has \`?widget_code=…\` → \`POST ${issuer}/v1/widget/exchange\` \`{ code, client_id }\` → access_token
7. Add exact \`return_to\` / callback hosts in dashboard → Integration redirect URIs (include app \`/login\` and \`/register\` for email verify + social, not only OIDC callback).
8. Do NOT put Google/GitHub OAuth secrets in my app — only in sooauth dashboard → Social login.

## UI requirements
- Fetch widget config server-side or via my API proxy (cache: no-store).
- **Sign-up page**: live password checklist from \`password_policy\` (✓/✗ per rule); disable submit until valid; mirror server rules.
- **Check-email page** (after sign-up when \`email_verify_required\`): read \`email_verify_delivery\` — show link-only message, 6-digit code input, or both. On code submit call \`/auth/verify-code\` with \`client_id\`; create app session from \`access_token\`.
- **Sign-in page**: email + password + social buttons from \`providers\`; handle \`?email_verified=1\` success banner after link verification.
- Show \`password_policy.hint\` or API \`message\` on validation errors.
- After success: create my app session (httpOnly cookie or backend session) from \`access_token\`.

## My app
- Stack: [e.g. Next.js 15 App Router / Vue 3 / React Native]
- client_id: [app_… from dashboard]
- Sign-in URL: [e.g. /login]
- Sign-up URL: [e.g. /signup]
- Social callback URL: [e.g. https://myapp.com/api/auth/callback/social]
- Protected routes: [e.g. /app/*]

## Deliverables
- Env vars (\`SOOAUTH_ISSUER\`, \`SOOAUTH_CLIENT_ID\`, callback URLs)
- Optional BFF/proxy routes (\`/api/auth/widget/config\`, sign-in, sign-up, verify-code, resend-verification, social callback)
- Login + signup + check-email components (code input when delivery is code/both)
- Social OAuth redirect + widget_code exchange handler
- Session middleware for protected routes
- Sign-out

Implement step by step. Match my stack conventions. Do not use embed.js.`;
}
