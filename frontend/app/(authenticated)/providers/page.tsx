import { getCredentials } from "@/lib/api/credentials";
import { getProviders } from "@/lib/api/providers";
import { ProvidersClient } from "./componentsAction/ProvidersClient";
import { Suspense } from "react";

export default async function ProvidersPage() {
  const [providers, credentials] = await Promise.all([
    getProviders().catch(() => []),
    getCredentials().catch(() => []),
  ]);
  return (
    <Suspense>
      <ProvidersClient
        initialProviders={providers}
        initialConfiguredCredentialProviders={credentials.map((credential) => credential.provider_id)}
      />
    </Suspense>
  );
}
