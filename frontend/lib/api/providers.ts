import { getApiBaseUrl } from "./base-url";

export interface ProviderMetadata {
  providerId: string;
  displayName: string;
  status: string;
  latencyMs: number;
  region: string;
  capabilities: Record<string, unknown>;
  quotaTotalBytes: number;
  quotaUsedBytes: number;
  lastHealthCheckAt: string;
}

export async function getProviders(): Promise<ProviderMetadata[]> {
  const API_URL = getApiBaseUrl();
  const res = await fetch(`${API_URL}/api/providers`);
  if (!res.ok) throw new Error("Failed to fetch providers");
  return res.json();
}

export async function getGDriveAuthorizeURL(): Promise<{ authURL: string }> {
  const API_URL = getApiBaseUrl();
  const res = await fetch(`${API_URL}/api/oauth/gdrive/authorize`);
  if (!res.ok) throw new Error("Failed to get authorization URL");
  return res.json();
}

export async function disconnectGDrive(): Promise<void> {
  const API_URL = getApiBaseUrl();
  const res = await fetch(`${API_URL}/api/oauth/gdrive/disconnect`, {
    method: "POST",
  });
  if (!res.ok) throw new Error("Failed to disconnect Google Drive");
}

export async function connectS3(): Promise<void> {
  const API_URL = getApiBaseUrl();
  const res = await fetch(`${API_URL}/api/providers/awsS3/connect`, { method: "POST" });
  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new Error((body as { error?: string }).error ?? "Failed to connect S3");
  }
}

export async function disconnectS3(): Promise<void> {
  const API_URL = getApiBaseUrl();
  const res = await fetch(`${API_URL}/api/providers/awsS3/disconnect`, { method: "POST" });
  if (!res.ok) throw new Error("Failed to disconnect S3");
}
