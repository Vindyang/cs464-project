import { auth } from "@/lib/auth";
import { headers } from "next/headers";
import { getFiles, FileMetadata } from "@/lib/api/files";
import { getProviders, ProviderMetadata } from "@/lib/api/providers";

export interface DashboardData {
  files: FileMetadata[];
  providers: ProviderMetadata[];
  userId: string | null;
}

export async function getDashboardData(): Promise<DashboardData> {
  const session = await auth.api.getSession({ headers: await headers() });
  const userId = session?.user?.id ?? null;

  if (!userId) {
    return { files: [], providers: [], userId: null };
  }

  const [files, providers] = await Promise.all([
    getFiles(userId).catch(() => [] as FileMetadata[]),
    getProviders().catch(() => [] as ProviderMetadata[]),
  ]);

  return { files, providers, userId };
}
