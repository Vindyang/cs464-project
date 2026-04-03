import { getProviders } from "@/lib/api/providers";
import { ProvidersClient } from "./componentsAction/ProvidersClient";
import { Suspense } from "react";

export default async function ProvidersPage() {
  const providers = await getProviders().catch(() => []);
  return (
    <Suspense>
      <ProvidersClient initialProviders={providers} />
    </Suspense>
  );
}
