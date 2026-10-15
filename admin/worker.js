const proxiedPrefixes = ["/auth/", "/dashboard/api/", "/v1/", "/oauth/", "/.well-known/"];

// Serves the dashboard SPA. Assets are built for the /dashboard/ base path;
// every other /dashboard/* URL (e.g. /dashboard/p/<id>/users) gets index.html
// so client-side routes work on reload and when shared.
export default {
  async fetch(request, env) {
    const url = new URL(request.url);
    if (proxiedPrefixes.some((prefix) => url.pathname.startsWith(prefix))) {
      const origin = new URL(env.API_ORIGIN || "https://api.sooauth.com");
      url.protocol = origin.protocol;
      url.hostname = origin.hostname;
      url.port = origin.port;
      return fetch(new Request(url, request));
    }
    if (url.pathname === "/dashboard" || url.pathname.startsWith("/dashboard/")) {
      const rest = url.pathname.slice("/dashboard".length) || "/";
      const isAsset = /\.[a-z0-9]+$/i.test(rest);
      if (isAsset) {
        url.pathname = rest;
        return env.ASSETS.fetch(new Request(url, request));
      }
      url.pathname = "/index.html";
      const response = await env.ASSETS.fetch(new Request(url, request));
      const headers = new Headers(response.headers);
      headers.set("Cache-Control", "no-store");
      return new Response(response.body, { status: response.status, headers });
    }
    return env.ASSETS.fetch(request);
  },
};
