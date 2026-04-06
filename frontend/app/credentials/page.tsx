import { getCredentials } from "@/lib/api/credentials";
import { CredentialsClient } from "./componentsAction/CredentialsClient";

export default async function CredentialsPage() {
  const credentials = await getCredentials().catch(() => []);
  return <CredentialsClient initialCredentials={credentials} />;
}
