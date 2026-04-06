import { getApiBaseUrl } from "./base-url";

export const CREDENTIAL_PROVIDERS = ["googleDrive", "awsS3", "oneDrive"] as const;
export type CredentialProvider = (typeof CREDENTIAL_PROVIDERS)[number];

export interface ProviderCredential {
  provider_id: string;
  client_id: string;
  redirect_uri: string;
  updated_at: string;
}

export interface CredentialStatus {
  configured: boolean;
  count: number;
  providers: string[];
}

export async function getCredentials(): Promise<ProviderCredential[]> {
  const API_URL = getApiBaseUrl();
  const res = await fetch(`${API_URL}/api/credentials`, { cache: "no-store" });
  if (!res.ok) {
    throw new Error("Failed to fetch credentials");
  }
  return res.json();
}

export async function getCredentialStatus(): Promise<CredentialStatus> {
  const API_URL = getApiBaseUrl();
  const res = await fetch(`${API_URL}/api/credentials/status`, { cache: "no-store" });
  if (!res.ok) {
    throw new Error("Failed to fetch credential status");
  }
  return res.json();
}

export async function saveCredential(
  providerId: string,
  payload: {
    clientId: string;
    clientSecret: string;
    redirectUri: string;
  },
): Promise<void> {
  const API_URL = getApiBaseUrl();
  const res = await fetch(`${API_URL}/api/credentials/${providerId}`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      client_id: payload.clientId,
      client_secret: payload.clientSecret,
      redirect_uri: payload.redirectUri,
    }),
  });
  if (!res.ok) {
    throw new Error("Failed to save credentials");
  }
}

export async function deleteCredential(providerId: string): Promise<void> {
  const API_URL = getApiBaseUrl();
  const res = await fetch(`${API_URL}/api/credentials/${providerId}`, {
    method: "DELETE",
  });
  if (!res.ok) {
    throw new Error("Failed to delete credentials");
  }
}
