import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

// The dashboard is served under /dashboard/ by both the Cloudflare worker and
// the Go server (self-hosted), so assets are built for that base path.
export default defineConfig({
  base: "/dashboard/",
  plugins: [react()],
  server: {
    proxy: {
      "/dashboard/api": "http://localhost:8080",
      "/auth": "http://localhost:8080",
      "/oauth": "http://localhost:8080",
      "/v1": "http://localhost:8080",
      "/.well-known": "http://localhost:8080",
    },
  },
  build: {
    outDir: "dist",
    emptyOutDir: true,
  },
});
