import { getFiles, FileMetadata } from "@/lib/api/files";
import { getProviders, ProviderMetadata } from "@/lib/api/providers";

export interface DashboardData {
  files: FileMetadata[];
  providers: ProviderMetadata[];
}

export async function getDashboardData(): Promise<DashboardData> {
  const [files, providers] = await Promise.all([
    getFiles().catch(() => [] as FileMetadata[]),
    getProviders().catch(() => [] as ProviderMetadata[]),
  ]);

  return { files, providers };
}
