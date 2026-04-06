import { getApiBaseUrl } from "./base-url";

export type RedundancyPreset = "(6,4)" | "(8,4)" | "(10,8)";
export type ResetScope = "files" | "credentials" | "all_data"

export const REDUNDANCY_PRESETS = [
  { val: "(6,4)" as const, label: "(6,4)", name: "Standard", overhead: "1.5x overhead", desc: "4 data + 2 parity shards. Requires 4 to recover." },
  { val: "(8,4)" as const, label: "(8,4)", name: "High Reliability", overhead: "2.0x overhead", desc: "4 data + 4 parity shards. Requires 4 to recover." },
  { val: "(10,8)" as const, label: "(10,8)", name: "Efficient", overhead: "1.25x overhead", desc: "8 data + 2 parity shards. Requires 8 to recover." },
]

export interface AppSettings {
  redundancy: RedundancyPreset;
  encrypt_default: boolean;
  auto_delete: boolean;
}

export interface ResetStoredDataResponse {
  scope: ResetScope
  delete_shards: boolean
  file_summary?: {
    deleted_files: number
    deleted_shards: number
    failed_shard_deletes: number
    delete_shards: boolean
  }
  credential_summary?: {
    deleted_credentials: number
    deleted_tokens: number
    disconnected_providers: number
  }
}

export function parseRedundancyPreset(preset: RedundancyPreset): { n: number; k: number } {
  const matches = preset.match(/^\((\d+),(\d+)\)$/)
  if (!matches) {
    return { n: 6, k: 4 }
  }

  return {
    n: Number(matches[1]),
    k: Number(matches[2]),
  }
}

export async function getSettings(): Promise<AppSettings> {
  const API_URL = getApiBaseUrl();
  const res = await fetch(
    `${API_URL}/api/settings`,
    typeof window === "undefined"
      ? { next: { revalidate: 300 } }
      : { cache: "no-store" },
  );
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

export async function resetStoredData(scope: ResetScope, deleteShards = true): Promise<ResetStoredDataResponse> {
  const API_URL = getApiBaseUrl()
  const res = await fetch(`${API_URL}/api/settings/reset`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ scope, delete_shards: deleteShards }),
  })

  if (!res.ok) {
    throw new Error("Failed to reset stored data")
  }

  return res.json()
}
