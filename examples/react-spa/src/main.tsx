import { useEffect, useState, type ReactNode } from "react";
import { createRoot } from "react-dom/client";
import { api, clearSession, clientId, issuer, finishHostedSignIn, refreshSession, saveSession, session, startHostedSignIn, userinfo, type ApiError } from "./auth";
import "./styles.css";

function Shell({ children }: { children: ReactNode }) {
  return <main className="shell"><a className="logo" href="/">Northwind</a>{children}</main>;
}

function Home() {
  return (
    <Shell>
      <h1>Sign in to Northwind</h1>
      <p>Three ways to do the same thing.</p>
      <div className="row">
        <button onClick={startHostedSignIn}>Ready-made page (OIDC)</button>
        <a className="button ghost" href="/widget">Ready-made widget</a>
        <a className="button ghost" href="/login">Our own design</a>
      </div>
      {session() && <p><a href="/app">Continue to the app →</a></p>}
    </Shell>
  );
}

function Callback() {
  const [error, setError] = useState("");
  useEffect(() => {
    finishHostedSignIn().then(() => { if (session()) location.replace("/app"); }).catch((e) => setError(e.message));
  }, []);
  return <Shell>{error ? <><h1>Sign-in failed</h1><p className="error">{error}</p><a href="/">Back</a></> : <p>Finishing sign-in…</p>}</Shell>;
}

function Protected() {
  const [info, setInfo] = useState<unknown>(null);
  const [error, setError] = useState("");
  useEffect(() => {
    if (!session()) { location.replace("/"); return; }
    userinfo().then(setInfo).catch((e) => setError(e.message));
  }, []);
  return (
    <Shell>
      <h1>You are in</h1>
      {error && <p className="error">{error}</p>}
      <pre>{JSON.stringify(info, null, 2)}</pre>
      <div className="row">
        <button className="ghost" onClick={() => refreshSession().then(() => userinfo()).then(setInfo).catch((e) => setError(e.message))}>Refresh token</button>
        <button className="ghost" onClick={() => { clearSession(); location.href = "/"; }}>Sign out</button>
      </div>
    </Shell>
  );
}

// The drop-in widget: mount the target, then load embed.js once.
function WidgetLogin() {
  useEffect(() => {
    const onLogin = (e: Event) => {
      saveSession((e as CustomEvent).detail);
      location.href = "/app";
    };
    window.addEventListener("sooauth:login", onLogin);
    const script = document.createElement("script");
    script.src = `${issuer}/v1/widget/embed.js`;
    script.dataset.clientId = clientId;
    script.dataset.mode = "signin";
    script.dataset.target = "#sooauth-widget";
    document.body.append(script);
    return () => {
      window.removeEventListener("sooauth:login", onLogin);
      script.remove();
    };
  }, []);
  return <Shell><h1>Sign in to Northwind</h1><div id="sooauth-widget" /></Shell>;
}

type Mode = "signin" | "signup" | "code" | "check-email" | "forgot" | "mfa";

function CustomLogin() {
  const [config, setConfig] = useState<any>(null);
  const [mode, setMode] = useState<Mode>("signin");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [code, setCode] = useState("");
  const [codeSent, setCodeSent] = useState(false);
  const [metadata, setMetadata] = useState<Record<string, string>>({});
  const [challenge, setChallenge] = useState<{ challenge: string; method: string } | null>(null);
  const [message, setMessage] = useState(new URLSearchParams(location.search).get("email_verified") ? "Email verified. You can sign in now." : "");
  const [error, setError] = useState("");

  useEffect(() => {
    api(`/v1/widget/config?client_id=${clientId}`).then(setConfig).catch((e) => setError(e.message));
  }, []);

  if (!config) return <Shell><p className={error ? "error" : ""}>{error || "Loading…"}</p></Shell>;

  const signedIn = (res: any) => {
    saveSession(res);
    location.href = "/app";
  };
  const run = async (fn: () => Promise<void>) => {
    setError("");
    try {
      await fn();
    } catch (e) {
      const err = e as ApiError;
      if (err.body?.error === "mfa_required") {
        setChallenge({ challenge: err.body.challenge, method: err.body.method });
        setMode("mfa");
      } else setError(err.message);
    }
  };
  const authMode = config.auth_config.primary_auth_mode;
  const returnTo = `${location.origin}/login`;

  return (
    <Shell>
      <div className="card" style={{ borderTopColor: config.accent_color }}>
        <h1>{{ signin: "Welcome back", signup: "Create your account", code: "Sign in with a code", "check-email": "Check your email", forgot: "Reset your password", mfa: "Two-step verification" }[mode]}</h1>
        {message && <p className="ok">{message}</p>}
        {error && <p className="error">{error}</p>}

        {mode !== "check-email" && mode !== "mfa" && (
          <label>Email<input type="email" value={email} onChange={(e) => setEmail(e.target.value)} /></label>
        )}
        {(mode === "signin" || mode === "signup") && authMode !== "otp" && (
          <label>Password<input type="password" value={password} onChange={(e) => setPassword(e.target.value)} /></label>
        )}
        {mode === "signup" && <p className="hint">{config.password_policy.hint}</p>}
        {mode === "signup" && (config.registration_schema || []).map((f: any) => (
          <label key={f.id}>{f.label}<input value={metadata[f.id] || ""} onChange={(e) => setMetadata({ ...metadata, [f.id]: e.target.value })} /></label>
        ))}

        {mode === "signin" && authMode !== "otp" && (
          <button onClick={() => run(async () => signedIn(await api("/auth/sign-in", { identifier: email, password, client_id: clientId })))}>Sign in</button>
        )}
        {mode === "signup" && (
          <button onClick={() => run(async () => {
            const res = await api("/auth/sign-up", { email, password, client_id: clientId, return_to: returnTo, metadata });
            if (res.access_token) signedIn(res); else setMode("check-email");
          })}>Create account</button>
        )}
        {mode === "code" && !codeSent && (
          <button onClick={() => run(async () => { await api("/auth/otp/send", { identifier: email, client_id: clientId }); setCodeSent(true); })}>Send code</button>
        )}
        {mode === "code" && codeSent && (
          <>
            <label>6-digit code<input value={code} onChange={(e) => setCode(e.target.value)} /></label>
            <button onClick={() => run(async () => signedIn(await api("/auth/otp/verify", { identifier: email, code, client_id: clientId, metadata })))}>Verify</button>
          </>
        )}
        {mode === "check-email" && (
          <>
            <p>We sent a message to <b>{email}</b>.</p>
            {config.email_verify_delivery !== "link" && (
              <>
                <label>6-digit code<input value={code} onChange={(e) => setCode(e.target.value)} /></label>
                <button onClick={() => run(async () => signedIn(await api("/auth/verify-code", { email, code, client_id: clientId })))}>Verify</button>
              </>
            )}
            <button className="ghost" onClick={() => run(async () => { await api("/auth/resend-verification", { email, client_id: clientId, return_to: returnTo }); setMessage("Sent again."); })}>Resend</button>
          </>
        )}
        {mode === "forgot" && (
          <button onClick={() => run(async () => { await api("/auth/forgot-password", { email, client_id: clientId, return_to: returnTo }); setMessage("If that account exists, a reset message is on its way."); })}>Send reset message</button>
        )}
        {mode === "mfa" && challenge && (
          <>
            <label>Code ({challenge.method})<input value={code} onChange={(e) => setCode(e.target.value)} /></label>
            <button onClick={() => run(async () => signedIn(await api("/auth/mfa/verify", { ...challenge, code })))}>Verify</button>
          </>
        )}

        <div className="links">
          {mode !== "signin" && authMode !== "otp" && <a onClick={() => setMode("signin")}>Sign in with a password</a>}
          {mode !== "signup" && authMode !== "otp" && <a onClick={() => setMode("signup")}>Create an account</a>}
          {mode !== "code" && authMode !== "password" && <a onClick={() => setMode("code")}>Use a one-time code</a>}
          {mode !== "forgot" && authMode !== "otp" && <a onClick={() => setMode("forgot")}>Forgot password?</a>}
        </div>
      </div>
    </Shell>
  );
}

const pages: Record<string, () => ReactNode> = { "/": Home, "/callback": Callback, "/app": Protected, "/login": CustomLogin, "/widget": WidgetLogin };
const Page = pages[location.pathname] || Home;
createRoot(document.getElementById("root")!).render(<Page />);
