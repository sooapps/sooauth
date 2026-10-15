import { useEffect, useState } from "react";
import { request, type Data } from "../api";
import { normalizeAuthConfig } from "../pages/auth/SignInMethods";
import type { ProjectInfo } from "./prompts";

// loadProjectInfo gathers everything the prompts and snippets need.
export async function loadProjectInfo(): Promise<ProjectInfo & { raw: { project: Data; integration: Data } }> {
  const [project, integration, providers] = await Promise.all([
    request("/project"),
    request("/integration"),
    request("/providers").catch(() => ({ providers: [] })),
  ]);
  return {
    projectName: project.name,
    issuer: integration.issuer,
    clientId: integration.client_id,
    tenantId: integration.tenant_id,
    redirectUris: integration.redirect_uris || [],
    dashboardUrl: `${location.origin}/dashboard/`,
    providers: (providers.providers || []).filter((p: Data) => p.enabled).map((p: Data) => p.provider),
    auth: normalizeAuthConfig(project.auth_config),
    fields: project.registration_schema || [],
    passwordHint: project.password_hint || "",
    emailVerifyRequired: Boolean(project.email_verify_required),
    emailVerifyDelivery: project.email_verify_delivery || "link",
    raw: { project, integration },
  };
}

export function useProjectInfo(deps: unknown[] = []) {
  const [info, setInfo] = useState<Awaited<ReturnType<typeof loadProjectInfo>> | null>(null);
  const [error, setError] = useState("");
  const reload = () => loadProjectInfo().then(setInfo).catch((e) => setError((e as Error).message));
  useEffect(() => { reload(); }, deps); // eslint-disable-line react-hooks/exhaustive-deps
  return { info, error, reload };
}
