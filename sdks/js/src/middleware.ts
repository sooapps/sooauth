import { createRemoteJWKSet, jwtVerify, type JWTPayload } from "jose";

export type MiddlewareOptions = {
  issuer: string;
  audience?: string;
};

export function createAuthMiddleware(options: MiddlewareOptions) {
  const issuer = options.issuer.replace(/\/$/, "");
  const jwks = createRemoteJWKSet(new URL(`${issuer}/.well-known/jwks.json`));

  return async function verifyBearer(authorization?: string): Promise<JWTPayload> {
    if (!authorization?.startsWith("Bearer ")) {
      throw new Error("missing_bearer");
    }
    const token = authorization.slice("Bearer ".length).trim();
    const { payload } = await jwtVerify(token, jwks, {
      issuer,
      audience: options.audience,
    });
    return payload;
  };
}
