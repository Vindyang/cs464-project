import { getApiBaseUrl } from "./base-url";

export type RedundancyPreset = "(6,4)" | "(8,4)" | "(10,8)";

export interface AppSettings {
  redundancy: RedundancyPreset;
  encrypt_default: boolean;
  auto_delete: boolean;
}

export async function getSettings(): Promise<AppSettings> {
  const API_URL = getApiBaseUrl();
  const res = await fetch(`${API_URL}/api/settings`, { cache: "no-store" });
  if (!res.ok) {
    throw new Error("Failed to fetch settings");
  }
  return res.json();
}

export async function saveSettings(payload: AppSettings): Promise<void> {
  const API_URL = getApiBaseUrl();
  const res = await fetch(`${API_URL}/api/settings`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  });
  if (!res.ok) {
    throw new Error("Failed to save settings");
  }
}
