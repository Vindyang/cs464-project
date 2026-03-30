import { auth } from "@/lib/auth";
import { headers } from "next/headers";
import { SettingsClient } from "./SettingsClient";

export default async function SettingsPage() {
  const session = await auth.api.getSession({ headers: await headers() });
  const email = session?.user?.email ?? "";
  const authProvider = session?.user?.name ? "OAuth" : "Email";

  return <SettingsClient email={email} authProvider={authProvider} />;
}
