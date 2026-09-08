# @sooauth/sdk-js

Official JavaScript and TypeScript client for [Sooauth](https://sooauth.com), the open-source authentication layer for vibecoders and modern stacks.

## Install

```bash
npm install @sooauth/sdk-js
```

## Quick Start

```typescript
import { createSooauthClient } from "@sooauth/sdk-js";

export const auth = createSooauthClient({
  issuer: process.env.NEXT_PUBLIC_SOOAUTH_ISSUER || "https://auth.sooauth.com",
  clientId: process.env.NEXT_PUBLIC_SOOAUTH_CLIENT_ID!,
  redirectUri: process.env.NEXT_PUBLIC_SOOAUTH_REDIRECT_URI!,
});

// 1. Trigger login
await auth.signInWithRedirect();

// 2. Handle callback on redirect route
const session = await auth.handleCallback();
console.log("Logged in user:", session.user);
```

## Route Middleware (Node / Next.js)

```typescript
import { createRemoteJWKSet, jwtVerify } from "jose";

const JWKS = createRemoteJWKSet(new URL("https://auth.sooauth.com/.well-known/jwks.json"));

export async function verifyAuth(req: Request) {
  const authHeader = req.headers.get("Authorization");
  if (!authHeader?.startsWith("Bearer ")) throw new Error("Unauthorized");
  const token = authHeader.substring(7);
  const { payload } = await jwtVerify(token, JWKS, {
    issuer: "https://auth.sooauth.com",
  });
  return payload;
}
```

## React Hook

```tsx
import { useSession } from "@sooauth/sdk-js/react";

export function Profile() {
  const { session, isLoading } = useSession(auth);

  if (isLoading) return <div>Loading...</div>;
  if (!session) return <div>Not logged in</div>;

  return <div>Welcome, {session.user.email}</div>;
}
```

## License

AGPL-3.0-or-later. Built with care by [Sooapps](https://sooapps.com).
