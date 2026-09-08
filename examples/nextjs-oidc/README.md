# Sooauth Next.js example

This is the smallest browser-side OIDC example for Sooauth. It uses the
published JavaScript client and PKCE; the same issuer and client ID work with
any framework that supports OpenID Connect.

## Run

```bash
npm install @sooauth/sdk-js
npm run dev
```

Set these values in `.env.local`:

```env
NEXT_PUBLIC_SOOAUTH_ISSUER=https://auth.sooauth.com
NEXT_PUBLIC_SOOAUTH_CLIENT_ID=your_client_id
NEXT_PUBLIC_SOOAUTH_REDIRECT_URI=http://localhost:3000/callback
```

Register the exact callback URL in Dashboard -> Integration before testing.

## Production note

The sample stores the demo session in browser session storage for clarity. Use
an HTTP-only application session and server-side token handling for a
production application.
