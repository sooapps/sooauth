// A small sooauth client for a browser-only app: OIDC with PKCE for the hosted
// page, plain fetch calls for the custom forms.

export const issuer = import.meta.env.VITE_SOOAUTH_ISSUER as string;
export const clientId = import.meta.env.VITE_SOOAUTH_CLIENT_ID as string;
const redirectUri = `${location.origin}/callback`;

const b64url = (bytes: ArrayBuffer | Uint8Array) =>
  btoa(String.fromCharCode(...new Uint8Array(bytes))).replace(/\+/g, "-").replace(/\//g, "_").replace(/=+$/, "");

const random = () => b64url(crypto.getRandomValues(new Uint8Array(32)));

export async function startHostedSignIn() {
  const verifier = random();
  const state = random();
  // localStorage, not sessionStorage: an email verification link opens a new tab.
  localStorage.setItem("pkce", JSON.stringify({ verifier, state }));
  const challenge = b64url(await crypto.subtle.digest("SHA-256", new TextEncoder().encode(verifier)));
  const q = new URLSearchParams({
    client_id: clientId,
    response_type: "code",
    scope: "openid email profile",
    redirect_uri: redirectUri,
    code_challenge: challenge,
    code_challenge_method: "S256",
    state,
  });
  location.href = `${issuer}/oauth/authorize?${q}`;
}

export async function finishHostedSignIn() {
  const params = new URLSearchParams(location.search);
  const stored = JSON.parse(localStorage.getItem("pkce") || "{}");
  localStorage.removeItem("pkce");
  if (params.get("error")) throw new Error(params.get("error_description") || params.get("error") || "sign-in failed");
  // No matching state: the flow finished on another device, so start it again.
  if (!params.get("code") || params.get("state") !== stored.state) return startHostedSignIn();
  const res = await fetch(`${issuer}/oauth/token`, {
    method: "POST",
    headers: { "Content-Type": "application/x-www-form-urlencoded" },
    body: new URLSearchParams({
      grant_type: "authorization_code",
      code: params.get("code") || "",
      redirect_uri: redirectUri,
      client_id: clientId,
      code_verifier: stored.verifier,
    }),
  });
  const tokens = await res.json();
  if (!res.ok) throw new Error(tokens.error_description || tokens.error);
  saveSession(tokens);
}

export function saveSession(tokens: { access_token: string; refresh_token?: string }) {
  sessionStorage.setItem("session", JSON.stringify(tokens));
}

export function session(): { access_token: string; refresh_token?: string } | null {
  return JSON.parse(sessionStorage.getItem("session") || "null");
}

export function clearSession() {
  sessionStorage.removeItem("session");
}

export async function refreshSession() {
  const res = await fetch(`${issuer}/oauth/token`, {
    method: "POST",
    headers: { "Content-Type": "application/x-www-form-urlencoded" },
    body: new URLSearchParams({ grant_type: "refresh_token", refresh_token: session()?.refresh_token || "", client_id: clientId }),
  });
  const tokens = await res.json();
  if (!res.ok) throw new Error(tokens.error_description || tokens.error);
  saveSession(tokens);
}

export async function userinfo() {
  const res = await fetch(`${issuer}/oauth/userinfo`, { headers: { Authorization: `Bearer ${session()?.access_token}` } });
  if (!res.ok) throw new Error(`userinfo ${res.status}`);
  return res.json();
}

export type ApiError = Error & { status: number; body: Record<string, any> };

export async function api(path: string, body?: unknown) {
  const res = await fetch(`${issuer}${path}`, body === undefined ? undefined : {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });
  const data = await res.json().catch(() => ({}));
  if (!res.ok) throw Object.assign(new Error(data.message || data.error || `HTTP ${res.status}`), { status: res.status, body: data }) as ApiError;
  return data;
}
