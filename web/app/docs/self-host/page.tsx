import type { ReactNode } from "react";
import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "Self-host Sooauth",
  description: "Run the open-source Sooauth Community Edition on your own VPS with Docker Compose.",
};

function Step({
  n,
  title,
  children,
}: {
  n: number;
  title: string;
  children: ReactNode;
}) {
  return (
    <section className="grid gap-3 border-t border-line py-8 first:border-t-0 first:pt-0">
      <p className="font-mono text-xs text-fg-muted">Step {n}</p>
      <h2 className="font-sans text-[20px] font-semibold text-fg">{title}</h2>
      <div className="text-[15px] leading-[1.6] text-fg-muted">{children}</div>
    </section>
  );
}

export default function SelfHostDocsPage() {
  return (
    <article>
      <h1 className="font-sans text-[32px] font-semibold tracking-tight text-fg">
        Self-host Sooauth
      </h1>
      <p className="mt-3 max-w-[52ch] text-[16px] leading-[1.6] text-fg-muted">
        Sooauth is designed for both hosted and self-managed deployments. Run the
        Community Edition on your own infrastructure when you need control over
        data, deployment, or residency. The hosted service at
        <span className="font-mono text-[14px] text-fg"> auth.sooauth.com</span>
        is available when you want managed operations.
      </p>

      <div className="mt-10">
        <Step n={1} title="Requirements">
          <ul className="grid gap-2">
            <li>· Docker + Docker Compose</li>
            <li>· VPS with ~1 GB RAM free</li>
            <li>· SMTP for production email</li>
            <li>· A domain (e.g. auth.yourcompany.com)</li>
          </ul>
        </Step>

        <Step n={2} title="Deploy stack">
          <p>Use the monorepo compose file:</p>
          <pre className="mt-4 overflow-x-auto border border-line bg-bg-subtle px-4 py-3 font-mono text-[13px] leading-[1.7] text-fg">{`git clone https://github.com/sooapps/sooauth
cp .env.coolify.example .env
# set POSTGRES_PASSWORD, SMTP_URL, APP_URL, ADMIN_EMAILS
docker compose -f docker-compose.coolify.yml up -d`}</pre>
        </Step>

        <Step n={3} title="Domain & TLS">
          <p>
            Point your auth subdomain to the server. Coolify/Traefik or Caddy
            terminates TLS. Set <span className="font-mono text-[13px]">APP_URL</span> to
            your public HTTPS origin.
          </p>
        </Step>

        <Step n={4} title="Verify">
          <pre className="mt-4 overflow-x-auto border border-line bg-bg-subtle px-4 py-3 font-mono text-[13px] leading-[1.7] text-fg">{`curl https://auth.yourcompany.com/healthz
curl https://auth.yourcompany.com/.well-known/openid-configuration`}</pre>
        </Step>
      </div>
    </article>
  );
}
