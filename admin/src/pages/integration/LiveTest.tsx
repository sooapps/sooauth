import { useEffect, useState } from "react";
import { API, request, type Data } from "../../api";
import { useApp, useI18n } from "../../context";
import { challengeFor, randomString } from "../../lib/pkce";
import { Button, Card, Loading, Table, useLoad } from "../../ui";

export type TestResult = { ok: boolean; error?: string; claims?: Data; userinfo?: Data };

export const TEST_STORAGE_PREFIX = "sooauth_test_";
export const TEST_MESSAGE = "sooauth-live-test";

function issuerBase() {
  return API || location.origin;
}

// LiveTest runs a real OIDC sign-in against this project in a popup: the
// hosted sign-in page, the code exchange with PKCE and /oauth/userinfo. It
// proves the project works before the owner writes any code.
export function LiveTest({ compact }: { compact?: boolean }) {
  const { tr } = useI18n();
  const { notify, refreshOnboarding } = useApp();
  const [integration] = useLoad<Data>(() => request("/integration"));
  const [result, setResult] = useState<TestResult | null>(null);
  const [waiting, setWaiting] = useState(false);

  useEffect(() => {
    const onMessage = (e: MessageEvent) => {
      if (e.origin !== location.origin || e.data?.type !== TEST_MESSAGE) return;
      setWaiting(false);
      setResult(e.data.result);
      if (e.data.result?.ok) refreshOnboarding();
    };
    // A test user who verifies by email finishes in a new tab, without an opener.
    const channel = new BroadcastChannel(TEST_MESSAGE);
    const onChannel = (e: MessageEvent) => onMessage(new MessageEvent("message", { origin: location.origin, data: e.data }));
    channel.addEventListener("message", onChannel);
    window.addEventListener("message", onMessage);
    return () => {
      channel.close();
      window.removeEventListener("message", onMessage);
    };
  }, [refreshOnboarding]);

  if (!integration) return <Loading />;

  const start = async () => {
    const verifier = randomString(48);
    const state = randomString(16);
    const challenge = await challengeFor(verifier);
    const redirectUri = integration.test_redirect_uri;
    localStorage.setItem(TEST_STORAGE_PREFIX + state, JSON.stringify({ verifier, clientId: integration.client_id, redirectUri, issuer: issuerBase(), created: Date.now() }));
    const q = new URLSearchParams({
      client_id: integration.client_id,
      redirect_uri: redirectUri,
      response_type: "code",
      scope: "openid email profile",
      state,
      nonce: randomString(12),
      code_challenge: challenge,
      code_challenge_method: "S256",
      // Always show the sign-in page, even if a test user is still signed in.
      prompt: "login",
    });
    const popup = window.open(`${issuerBase()}/oauth/authorize?${q}`, "sooauth-live-test", "width=480,height=720");
    if (!popup) {
      notify(tr("test.popupBlocked"));
      return;
    }
    setResult(null);
    setWaiting(true);
  };

  const claims = result?.claims || {};
  return (
    <div className="stack">
      {!compact && <p className="section-desc">{tr("test.explain")}</p>}
      <ol className="how-list">
        <li>{tr("test.step1")}</li>
        <li>{tr("test.step2")}</li>
        <li>{tr("test.step3")}</li>
      </ol>
      <div className="actions">
        <Button primary onClick={start}>{waiting ? tr("test.waiting") : tr("test.start")}</Button>
      </div>
      {result && !result.ok && <div className="notice">{tr("test.failed", { error: result.error || "unknown" })}</div>}
      {result?.ok && (
        <div className="success-box">
          <strong>✓ {tr("test.success")}</strong>
          <Table
            headers={[tr("test.claim"), tr("test.value")]}
            rows={Object.entries({ ...claims, ...(result.userinfo || {}) })
              .filter(([k]) => !["aud", "iat", "exp", "nbf", "jti", "nonce"].includes(k))
              .map(([k, v]) => [k, typeof v === "object" ? JSON.stringify(v) : String(v)])}
          />
        </div>
      )}
    </div>
  );
}

export default function LiveTestPage() {
  const { tr } = useI18n();
  return (
    <Card title={tr("test.title")} description={tr("test.desc")}>
      <LiveTest />
    </Card>
  );
}
