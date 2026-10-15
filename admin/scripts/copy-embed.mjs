// Copies the built dashboard into the Go server so self-hosted builds serve
// the same React dashboard at /dashboard/.
import { cpSync, existsSync, mkdirSync, readdirSync, rmSync } from "node:fs";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";

const here = dirname(fileURLToPath(import.meta.url));
const dist = join(here, "..", "dist");
const target = join(here, "..", "..", "server", "internal", "adminui", "dist");

if (!existsSync(join(dist, "index.html"))) {
  console.error("admin/dist/index.html not found — run vite build first");
  process.exit(1);
}
mkdirSync(target, { recursive: true });
for (const entry of readdirSync(target)) {
  if (entry !== ".keep") rmSync(join(target, entry), { recursive: true, force: true });
}
cpSync(dist, target, { recursive: true });
console.log(`copied dashboard build to ${target}`);
