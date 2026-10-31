import { useEffect, useState } from "react";
import { Table } from "../../ui";
import { useI18n } from "../../context";
import { decodeJwt } from "../../lib/pkce";
import { TEST_MESSAGE, TEST_STORAGE_PREFIX, type TestResult } from "./LiveTest";

async function finish(): Promise<TestResult> {
  const params = new URLSearchParams(location.search);
  const error = params.get("error");
  if (error) return { ok: false, error: params.get("error_description") || error };
  const code = params.get("code");
  const state = params.get("state") || "";
  const stored = localStorage.getItem(TEST_STORAGE_PREFIX + state);
  localStorage.removeItem(TEST_STORAGE_PREFIX + state);
  if (!code || !stored) return { ok: false, error: "missing code or state" };
  const { verifier, clientId, redirectUri, issuer } = JSON.parse(stored);

  const tokenRes = await fetch(`${issuer}/oauth/token`, {
    method: "POST",
    headers: { "Content-Type": "application/x-www-form-urlencoded" },
    body: new URLSearchParams({ grant_type: "authorization_code", code, redirect_uri: redirectUri, client_id: clientId, code_verifier: verifier }),
  });
  const tokens = await tokenRes.json().catch(() => ({}));
  if (!tokenRes.ok) return { ok: false, error: tokens.error_description || tokens.error || `token ${tokenRes.status}` };

  const infoRes = await fetch(`${issuer}/oauth/userinfo`, { headers: { Authorization: `Bearer ${tokens.access_token}` } });
  const userinfo = await infoRes.json().catch(() => ({}));
  return { ok: true, claims: decodeJwt(tokens.id_token || "") || {}, userinfo: infoRes.ok ? userinfo : undefined };
}

// TestCallback is the redirect target of the dashboard live test. It finishes
// the code exchange in the popup and reports the result to the opener.
export default function TestCallback() {
  const { tr } = useI18n();
  const [result, setResult] = useState<TestResult | null>(null);
  useEffect(() => {
    finish()
      .catch((e) => ({ ok: false, error: (e as Error).message }) as TestResult)
      .then((r) => {
        setResult(r);
        history.replaceState(null, "", location.pathname);
        const channel = new BroadcastChannel(TEST_MESSAGE);
        channel.postMessage({ type: TEST_MESSAGE, result: r });
        channel.close();
        if (window.opener) {
          window.opener.postMessage({ type: TEST_MESSAGE, result: r }, location.origin);
          if (r.ok) setTimeout(() => window.close(), 1200);
        }
      });
  }, []);
  return (
    <div className="callback-page">
      <div className="card">
        {!result && <p>{tr("test.finishing")}</p>}
        {result?.ok && (
          <>
            <h2>✓ {tr("test.success")}</h2>
            {window.opener ? <p>{tr("test.closeHint")}</p> : (
              <Table
                headers={[tr("test.claim"), tr("test.value")]}
                rows={Object.entries({ ...(result.claims || {}), ...(result.userinfo || {}) })
                  .filter(([k]) => !["aud", "iat", "exp", "nbf", "jti", "nonce"].includes(k))
                  .map(([k, v]) => [k, typeof v === "object" ? JSON.stringify(v) : String(v)])}
              />
            )}
          </>
        )}
        {result && !result.ok && <><h2>{tr("test.failedTitle")}</h2><p>{result.error}</p></>}
        {!window.opener && result && <a className="button" href="/dashboard/">{tr("test.back")}</a>}
      </div>
    </div>
  );
}
