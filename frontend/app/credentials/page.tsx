import { getCredentials } from "@/lib/api/credentials";
import { CredentialsClient } from "./componentsAction/CredentialsClient";

export const dynamic = 'force-dynamic';

export default async function CredentialsPage() {
  const credentials = await getCredentials().catch(() => []);
  return <CredentialsClient initialCredentials={credentials} />;
}
