// Builds the AI setup prompt on the docs from the choices a reader makes. Every
// value is either chosen or explicitly left for the assistant to ask about, so
// no [placeholder] can slip through unedited.

export type Stack = "react" | "vue" | "svelte" | "nextjs" | "node" | "go" | "python" | "php" | "other";
export type Design = "hosted" | "widget" | "own";
export type SignIn = "password" | "otp" | "both";
export type TwoStep = "off" | "optional" | "required";
export type EmailVerify = "link" | "code" | "both" | "off";
export type ColorScheme = "system" | "light" | "dark";

export type PromptOptions = {
  issuer: string;
  stack: Stack;
  otherStack: string;
  design: Design;
  matchExisting: boolean;
  accent: string;
  colorScheme: ColorScheme;
  signIn: SignIn;
  twoStep: TwoStep;
  emailVerify: EmailVerify;
  social: string[];
  extraFields: boolean;
  clientId: string;
  appUrl: string;
  language: string;
};

export const stackLabels: Record<Stack, string> = {
  react: "React (Vite)",
  vue: "Vue",
  svelte: "Svelte",
  nextjs: "Next.js",
  node: "Node / Express",
  go: "Go",
  python: "Python",
  php: "PHP / Laravel",
  other: "Other",
};

export const defaultAppUrl: Record<Stack, string> = {
  react: "http://localhost:5173",
  vue: "http://localhost:5173",
  svelte: "http://localhost:5173",
  nextjs: "http://localhost:3000",
  node: "http://localhost:3000",
  go: "http://localhost:8000",
  python: "http://localhost:8000",
  php: "http://localhost:8000",
  other: "http://localhost:3000",
};

// Stacks with no server of their own have to finish sign-in in the browser.
export function isBrowserOnly(stack: Stack) {
  return stack === "react" || stack === "vue" || stack === "svelte";
}

const ASK_CLIENT_ID = "NOT SET: ask me for it before writing code (dashboard → Keys & URLs)";

export function callbackPath(o: Pick<PromptOptions, "stack" | "design">) {
  if (o.design !== "hosted") return "/login";
  if (o.stack === "nextjs") return "/api/auth/callback/oidc";
  return isBrowserOnly(o.stack) ? "/callback" : "/auth/callback";
}

export function callbackUrl(o: PromptOptions) {
  return `${o.appUrl.trim().replace(/\/+$/, "")}${callbackPath(o)}`;
}

function stackName(o: PromptOptions) {
  return o.stack === "other" ? o.otherStack.trim() || "NOT SET: ask me what stack I use" : stackLabels[o.stack];
}

function dashboardSteps(o: PromptOptions) {
  const steps = [
    `Keys & URLs: add ${callbackUrl(o)} as a redirect URI, one for each environment.`,
    `Sign-in methods: primary method "${{ password: "Password", otp: "One-time code only", both: "Password or code" }[o.signIn]}", two-step verification "${o.twoStep}".`,
  ];
  if (o.emailVerify === "off") steps.push(`Passwords & verification: turn off required email verification.`);
  else steps.push(`Passwords & verification: require email verification, delivery "${o.emailVerify}".`);
  if (o.social.length) steps.push(`Social login: enable ${o.social.join(" and ")}.`);
  if (o.extraFields) steps.push(`Sign-up fields: add the extra fields I want to collect.`);
  if (o.design !== "own") steps.push(`Theme: set the accent color to ${o.accent}. The ${o.design === "hosted" ? "hosted pages" : "widget"} read it from there.`);
  return steps.map((s, i) => `${i + 1}. ${s}`).join("\n");
}

function sessionRules(o: PromptOptions) {
  return isBrowserOnly(o.stack)
    ? "This app has no server: keep tokens in memory or sessionStorage (never localStorage) and add a route guard that sends signed-out users to sign-in."
    : "Keep tokens on the server in an httpOnly cookie session and add middleware that protects private routes.";
}

function hostedFlow(o: PromptOptions) {
  const i = o.issuer;
  const browser = isBrowserOnly(o.stack);
  const lines = [
    `Use the OIDC authorization code flow with PKCE. It is a public client, so there is no client secret.`,
    `Send users to the sooauth hosted sign-in page. It already does sign-in, sign-up${o.twoStep !== "off" ? ", two-step verification" : ""} and password reset. Do not build a login form.`,
    `Authorize at ${i}/oauth/authorize with client_id, response_type=code, scope=openid email profile, redirect_uri, code_challenge, code_challenge_method=S256 and a random state.`,
    `Exchange the code at ${i}/oauth/token (form-encoded: grant_type=authorization_code, code, redirect_uri, client_id, code_verifier). ${browser ? "Do this from the browser; the endpoint allows cross-origin requests." : "Do this on the server."}`,
    `Read the user from ${i}/oauth/userinfo with the access token, or validate the id_token against ${i}/.well-known/jwks.json.`,
    `Access tokens last 15 minutes. Refresh with grant_type=refresh_token, refresh_token and client_id at the token endpoint.`,
  ];
  if (browser) lines.push(`Keep the PKCE verifier and state in localStorage across the redirect, not sessionStorage: the email verification link can finish sign-in in a new tab.`);
  if (o.emailVerify !== "off") lines.push(`New users verify their email on the hosted pages, and the link brings them back to this sign-in. My app needs no email handling. If the callback finds no matching state (for example the link was opened on another device), start sign-in again instead of showing an error.`);
  if (o.social.length) lines.push(`Social login needs no keys in my app; the hosted page shows the buttons.`);
  return lines;
}

function widgetFlow(o: PromptOptions) {
  const i = o.issuer;
  return [
    `Add this to my login page at ${callbackUrl(o)}:\n\n<div id="sooauth-widget"></div>\n<script src="${i}/v1/widget/embed.js" data-client-id="${o.clientId.trim() || "CLIENT_ID"}" data-mode="signin" data-target="#sooauth-widget" async></script>\n`,
    ...(isBrowserOnly(o.stack) || o.stack === "nextjs" ? [`This is a component framework, so load the script once from an effect after the target element mounts instead of pasting the tag.`] : []),
    `Listen for the sooauth:login DOM event. event.detail.access_token is the user's token.`,
    isBrowserOnly(o.stack)
      ? `Keep that token and read the user from ${i}/oauth/userinfo.`
      : `Send that token to my server, verify it with ${i}/oauth/userinfo and create my own session.`,
  ];
}

function ownFlow(o: PromptOptions) {
  const i = o.issuer;
  const client = o.clientId.trim() || "CLIENT_ID";
  const back = callbackUrl(o);
  const lines = [
    `On page load GET ${i}/v1/widget/config?client_id=${client}. Build the screens from it: auth_config, password_policy, registration_schema, providers, email_verify_required and email_verify_delivery.`,
  ];
  if (o.signIn !== "otp") {
    lines.push(`Sign-up: POST ${i}/auth/sign-up { email, password, client_id, return_to: "${back}", metadata }. Without an access_token in the response, show a check-email screen.`);
    lines.push(`Sign-in: POST ${i}/auth/sign-in { identifier, password, client_id } returns { access_token, refresh_token, expires_in }. 403 email_not_verified means show the check-email screen.`);
    lines.push(`Forgot password: POST ${i}/auth/forgot-password { email, client_id, return_to: "${back}" }.`);
  }
  if (o.signIn !== "password") {
    lines.push(`One-time code: POST ${i}/auth/otp/send { identifier, client_id }, then POST ${i}/auth/otp/verify { identifier, code, client_id, metadata }. 422 registration_required lists the fields a new user must also send.`);
  }
  if (o.signIn === "both") lines.push(`Show a switch on the sign-in form so users can pick password or code.`);
  if (o.twoStep !== "off") {
    lines.push(`Sign-in may return 401 { error: "mfa_required", challenge, method }. Ask for the code and POST ${i}/auth/mfa/verify { challenge, code, method }. For method "totp_setup" first POST ${i}/auth/mfa/setup { challenge } and show the qr_code. Resend email codes with POST ${i}/auth/mfa/resend { challenge }.`);
  }
  if (o.emailVerify !== "off") {
    lines.push(`Email verification: the link returns to ${back}?email_verified=1, so show a success message there; it is not an error.${o.emailVerify !== "link" ? ` Codes go to POST ${i}/auth/verify-code { email, code, client_id }.` : ""} Resend with POST ${i}/auth/resend-verification { email, client_id, return_to: "${back}" }.`);
  }
  if (o.social.length) {
    lines.push(`Social buttons link to ${i}/auth/social/{provider}/start?tenant_id=TENANT_ID&client_id=${client}&return_to=${encodeURIComponent(back)} (tenant_id comes from the config). On return, exchange ?widget_code= with POST ${i}/v1/widget/exchange { code, client_id }.`);
  }
  if (o.extraFields) lines.push(`Render registration_schema on the sign-up form and send the values as metadata. They come back as custom_claims from ${i}/oauth/userinfo.`);
  lines.push(`Every endpoint above allows cross-origin requests from the browser.`);
  return lines;
}

function designBrief(o: PromptOptions) {
  if (o.design === "hosted") return "";
  if (o.design === "widget") return `\n## Design\nThe widget brings its own form. Place it in a centered column on my login page and keep the rest of the page in my app's style.\n`;
  const look = o.matchExisting
    ? "Match my app's existing design system and components. If it has none, use the accent color below."
    : "Design the screens from scratch.";
  return `\n## Design\n- ${look}\n- Accent color: ${o.accent}\n- Color scheme: ${o.colorScheme === "system" ? "follow the system setting" : `${o.colorScheme} only`}\n- Screens: sign-in, sign-up, check email${o.signIn !== "password" ? ", code entry" : ""}${o.twoStep !== "off" ? ", two-step verification" : ""}${o.signIn !== "otp" ? ", forgot password" : ""}\n- Show a live password checklist from password_policy and the API's message on errors.\n`;
}

export function missingValues(o: PromptOptions) {
  const missing: string[] = [];
  if (!o.clientId.trim()) missing.push("client ID");
  if (o.stack === "other" && !o.otherStack.trim()) missing.push("stack");
  return missing;
}

export function buildPrompt(o: PromptOptions) {
  const flow = o.design === "hosted" ? hostedFlow(o) : o.design === "widget" ? widgetFlow(o) : ownFlow(o);
  const title = { hosted: "the ready-made hosted sign-in page", widget: "the drop-in sign-in widget", own: "my own sign-in and sign-up screens over the API" }[o.design];
  const language = o.language.trim();
  return `Add sign-in to my application with sooauth, using ${title}.

## My project
- Stack: ${stackName(o)}
- Issuer: ${o.issuer}
- OIDC discovery: ${o.issuer}/.well-known/openid-configuration
- client_id: ${o.clientId.trim() || ASK_CLIENT_ID}
- App URL: ${o.appUrl.trim()}
- ${o.design === "hosted" ? "Callback" : "Login page"} URL: ${callbackUrl(o)}

## How it works
${flow.map((l, i) => `${i + 1}. ${l}`).join("\n")}
${designBrief(o)}
## Rules
- ${sessionRules(o)}
- Do not create user or password tables; sooauth stores them.
- Do not put Google, GitHub or other OAuth keys in my app. Social login lives in the sooauth dashboard.
- Read the issuer and client_id from environment variables.${language ? `\n- Write all UI text in ${language}.` : ""}

## Tell me to do these in the sooauth dashboard
You cannot reach the dashboard, so list these as steps for me at the end:
${dashboardSteps(o)}

Implement step by step and explain each file you change.`;
}
