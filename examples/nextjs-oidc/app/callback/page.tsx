"use client";

import { useEffect, useState } from "react";
import { createSooauthClient } from "@sooauth/sdk-js";

const client = createSooauthClient({
  issuer: process.env.NEXT_PUBLIC_SOOAUTH_ISSUER || "https://auth.sooauth.com",
  clientId: process.env.NEXT_PUBLIC_SOOAUTH_CLIENT_ID || "replace-me",
  redirectUri: process.env.NEXT_PUBLIC_SOOAUTH_REDIRECT_URI || "http://localhost:3000/callback",
});

export default function CallbackPage() {
  const [message, setMessage] = useState("Completing sign-in...");
  useEffect(() => {
    client.handleCallback().then(() => setMessage("Signed in. Return to the home page to load userinfo.")).catch(() => setMessage("Sign-in could not be completed."));
  }, []);
  return <main><h1>{message}</h1><a href="/">Back to example</a></main>;
}
