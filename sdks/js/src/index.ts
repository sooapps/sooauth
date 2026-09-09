export type SooauthClientOptions = {
  issuer: string;
  clientId: string;
  redirectUri: string;
};

export type DiscoveryDocument = {
  issuer: string;
  authorization_endpoint: string;
  token_endpoint: string;
  userinfo_endpoint: string;
  jwks_uri: string;
};

export type TokenSet = {
  access_token: string;
  refresh_token?: string;
  id_token?: string;
  expires_in: number;
  token_type: string;
};

export type Session = {
  accessToken: string;
  refreshToken?: string;
  idToken?: string;
  expiresAt: number;
};

const STORAGE_KEY = "sooauth.session";

export function createSooauthClient(options: SooauthClientOptions) {
  const issuer = options.issuer.replace(/\/$/, "");

  async function discover(): Promise<DiscoveryDocument> {
    const res = await fetch(`${issuer}/.well-known/openid-configuration`);
    if (!res.ok) throw new Error("discovery_failed");
    return res.json();
  }

  async function signInWithRedirect(state?: string, nonce?: string) {
    const doc = await discover();
    const { verifier, challenge } = await createPKCE();
    sessionStorage.setItem("sooauth.pkce", verifier);
    const params = new URLSearchParams({
      client_id: options.clientId,
      redirect_uri: options.redirectUri,
      response_type: "code",
      scope: "openid email profile",
      code_challenge: challenge,
      code_challenge_method: "S256",
    });
    if (state) params.set("state", state);
    if (nonce) params.set("nonce", nonce);
    location.href = `${doc.authorization_endpoint}?${params}`;
  }

  async function handleCallback(url: string = location.href): Promise<Session> {
    const parsed = new URL(url);
    const code = parsed.searchParams.get("code");
    if (!code) throw new Error("missing_code");
    const verifier = sessionStorage.getItem("sooauth.pkce");
    if (!verifier) throw new Error("missing_pkce");
    const doc = await discover();
    const body = new URLSearchParams({
      grant_type: "authorization_code",
      client_id: options.clientId,
      code,
      redirect_uri: options.redirectUri,
      code_verifier: verifier,
    });
    const res = await fetch(doc.token_endpoint, { method: "POST", body });
    if (!res.ok) throw new Error("token_exchange_failed");
    const tokens = (await res.json()) as TokenSet;
    const session: Session = {
      accessToken: tokens.access_token,
      refreshToken: tokens.refresh_token,
      idToken: tokens.id_token,
      expiresAt: Date.now() + tokens.expires_in * 1000,
    };
    sessionStorage.setItem(STORAGE_KEY, JSON.stringify(session));
    sessionStorage.removeItem("sooauth.pkce");
    return session;
  }

  function getSession(): Session | null {
    const raw = sessionStorage.getItem(STORAGE_KEY);
    if (!raw) return null;
    try {
      return JSON.parse(raw) as Session;
    } catch {
      return null;
    }
  }

  async function getUserInfo(accessToken?: string) {
    const doc = await discover();
    const token = accessToken ?? getSession()?.accessToken;
    if (!token) throw new Error("unauthenticated");
    const res = await fetch(doc.userinfo_endpoint, {
      headers: { Authorization: `Bearer ${token}` },
    });
    if (!res.ok) throw new Error("userinfo_failed");
    return res.json();
  }

  async function getPasswordStatus(accessToken?: string): Promise<{ has_password: boolean }> {
    const token = accessToken ?? getSession()?.accessToken;
    if (!token) throw new Error("unauthenticated");
    const res = await fetch(`${issuer}/auth/account/password`, {
      headers: { Authorization: `Bearer ${token}` },
    });
    const body = await res.json().catch(() => ({}));
    if (!res.ok) throw new Error(body.error || "password_status_failed");
    return body as { has_password: boolean };
  }

  async function changePassword(currentPassword: string, newPassword: string, accessToken?: string): Promise<{ message: string }> {
    const token = accessToken ?? getSession()?.accessToken;
    if (!token) throw new Error("unauthenticated");
    const res = await fetch(`${issuer}/auth/account/password`, {
      method: "PUT",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${token}`,
      },
      body: JSON.stringify({ current_password: currentPassword, new_password: newPassword }),
    });
    const body = await res.json().catch(() => ({}));
    if (!res.ok) throw new Error(body.message || body.error || "password_change_failed");
    return body as { message: string };
  }

  async function setPassword(newPassword: string, accessToken?: string): Promise<{ message: string }> {
    const token = accessToken ?? getSession()?.accessToken;
    if (!token) throw new Error("unauthenticated");
    const res = await fetch(`${issuer}/auth/account/password/set`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${token}`,
      },
      body: JSON.stringify({ new_password: newPassword }),
    });
    const body = await res.json().catch(() => ({}));
    if (!res.ok) throw new Error(body.message || body.error || "password_set_failed");
    return body as { message: string };
  }

  async function signOut() {
    sessionStorage.removeItem(STORAGE_KEY);
    sessionStorage.removeItem("sooauth.pkce");
  }

  return {
    issuer,
    clientId: options.clientId,
    discoveryUrl: `${issuer}/.well-known/openid-configuration`,
    discover,
    signInWithRedirect,
    handleCallback,
    getSession,
    getUserInfo,
    getPasswordStatus,
    changePassword,
    setPassword,
    signOut,
  };
}

export type SooauthClient = ReturnType<typeof createSooauthClient>;

async function createPKCE() {
  const array = new Uint8Array(32);
  crypto.getRandomValues(array);
  const verifier = base64Url(array);
  const digest = await crypto.subtle.digest("SHA-256", new TextEncoder().encode(verifier));
  const challenge = base64Url(new Uint8Array(digest));
  return { verifier, challenge };
}

function base64Url(bytes: Uint8Array) {
  let binary = "";
  bytes.forEach((b) => (binary += String.fromCharCode(b)));
  return btoa(binary).replace(/\+/g, "-").replace(/\//g, "_").replace(/=+$/, "");
}

export * from "./middleware.js";
export * from "./react.js";
