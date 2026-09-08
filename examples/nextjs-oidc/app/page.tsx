"use client";

import { useState } from "react";
import { createSooauthClient } from "@sooauth/sdk-js";

const client = createSooauthClient({
  issuer: process.env.NEXT_PUBLIC_SOOAUTH_ISSUER || "https://auth.sooauth.com",
  clientId: process.env.NEXT_PUBLIC_SOOAUTH_CLIENT_ID || "replace-me",
  redirectUri: process.env.NEXT_PUBLIC_SOOAUTH_REDIRECT_URI || "http://localhost:3000/callback",
});

export default function Home() {
  const [message, setMessage] = useState("Not signed in");
  return (
    <main style={{ maxWidth: 640 }}>
      <p style={{ fontSize: 12, letterSpacing: "0.12em", textTransform: "uppercase" }}>Sooauth example</p>
      <h1>OIDC login in one small client.</h1>
      <p>Redirect to hosted Sooauth authentication, then return to your app.</p>
      <button type="button" onClick={() => client.signInWithRedirect()}>
        Sign in with Sooauth
      </button>
      <p>{message}</p>
      <button type="button" onClick={async () => setMessage(JSON.stringify(await client.getUserInfo()))}>
        Load userinfo
      </button>
    </main>
  );
}
