const API_URL = process.env.NEXT_PUBLIC_API_URL;

export interface ProviderCredential {
  providerId: string;
  payload: unknown;
  updatedAt: string;
}

export interface CredentialStatus {
  configured: boolean;
  count: number;
  providers: string[];
}

export async function getCredentialStatus(): Promise<CredentialStatus> {
  const res = await fetch(`${API_URL}/api/credentials/status`, { cache: "no-store" });
  if (!res.ok) {
    throw new Error("Failed to fetch credential status");
  }
  return res.json();
}

export async function getCredentials(): Promise<ProviderCredential[]> {
  const res = await fetch(`${API_URL}/api/credentials`, { cache: "no-store" });
  if (!res.ok) {
    throw new Error("Failed to fetch credentials");
  }
  return res.json();
}

export async function saveCredential(providerId: string, payload: unknown): Promise<void> {
  const res = await fetch(`${API_URL}/api/credentials/${providerId}`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ payload }),
  });
  if (!res.ok) {
    throw new Error("Failed to save credentials");
  }
}

export async function deleteCredential(providerId: string): Promise<void> {
  const res = await fetch(`${API_URL}/api/credentials/${providerId}`, {
    method: "DELETE",
  });
  if (!res.ok) {
    throw new Error("Failed to delete credentials");
  }
}
