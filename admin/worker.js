const proxiedPrefixes = ["/auth/", "/dashboard/api/", "/v1/", "/oauth/", "/.well-known/"];

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
    if (url.pathname === "/dashboard" || url.pathname === "/dashboard/") {
      url.pathname = "/index.html";
      const response = await env.ASSETS.fetch(new Request(url, request));
      const headers = new Headers(response.headers);
      headers.set("Cache-Control", "no-store");
      return new Response(response.body, { status: response.status, headers });
    }
    return env.ASSETS.fetch(request);
  },
};
