import { SettingsClient } from "./SettingsClient";
import { getSettings } from "@/lib/api/settings";

export default async function SettingsPage() {
  const initialSettings = await getSettings().catch(() => ({
    redundancy: "(6,4)" as const,
    encrypt_default: true,
    auto_delete: false,
  }));
  return <SettingsClient initialSettings={initialSettings} />;
}
