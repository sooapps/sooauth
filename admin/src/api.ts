export type Data = Record<string, any>;

export const API = import.meta.env.VITE_API_URL || "";

export function csrf() {
  return decodeURIComponent(document.cookie.match(/(?:^|; )sooauth_csrf=([^;]+)/)?.[1] || "");
}

export async function request(path: string, options: RequestInit = {}) {
  const method = (options.method || "GET").toString().toUpperCase();
  const headers = new Headers(options.headers);
  if (method !== "GET") {
    headers.set("Content-Type", "application/json");
    headers.set("X-CSRF-Token", csrf());
  }
  const response = await fetch(`${API}/dashboard/api${path}`, { ...options, headers, credentials: "include" });
  const body = await response.json().catch(() => ({}));
  if (response.status === 401) {
    const back = encodeURIComponent(location.pathname + location.search);
    window.location.href = `${window.location.origin}/auth/sign-in?return_to=${back}`;
    throw new Error("Authentication required");
  }
  if (!response.ok) throw new Error(body.message || body.error || "Request failed");
  return body;
}

export function put(path: string, body: Data) {
  return request(path, { method: "PUT", body: JSON.stringify(body) });
}

export function post(path: string, body?: Data) {
  return request(path, { method: "POST", body: body ? JSON.stringify(body) : undefined });
}

export type AuthMode = "password" | "otp" | "both";
export type MFAPolicy = "off" | "optional" | "required";

export type AuthConfig = {
  allowed_identifiers: string[];
  primary_auth_mode: AuthMode;
  mfa_policy: MFAPolicy;
  mfa_methods: string[];
  phone_otp_via_webhook: boolean;
};

export type SelectOption = { label: string; value: string };
export type OptionsSource = {
  type: "static" | "dynamic_api";
  options?: SelectOption[];
  url?: string;
  method?: string;
  headers?: Record<string, string>;
  items_path?: string;
  label_key?: string;
  value_key?: string;
  cache_ttl_seconds?: number;
};
export type RegistrationField = {
  id: string;
  label: string;
  type: "text" | "textarea" | "number" | "select" | "checkbox";
  required: boolean;
  placeholder?: string;
  description?: string;
  options_source?: OptionsSource;
};

export type Onboarding = {
  dismissed: boolean;
  app_type: string;
  steps: { id: string; done: boolean }[];
  done_count: number;
  total: number;
  complete: boolean;
};
