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
  const res = await fetch(`${process.env.NEXT_PUBLIC_API_URL}/api/providers`);
  if (!res.ok) throw new Error("Failed to fetch providers");
  return res.json();
}
