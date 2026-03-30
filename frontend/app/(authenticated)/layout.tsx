import { DashboardShell } from "@/components/dashboard/DashboardShell";
import { getProviders } from "@/lib/api/providers";
import { auth } from "@/lib/auth";
import { headers } from "next/headers";
import { getFiles } from "@/lib/api/files";

export default async function DashboardLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const session = await auth.api.getSession({ headers: await headers() });
  const userId = session?.user?.id ?? null;

  const [providers, files] = await Promise.all([
    getProviders().catch(() => []),
    userId ? getFiles(userId).catch(() => []) : Promise.resolve([]),
  ]);

  const totalStorageUsedBytes = files.reduce(
    (sum, f) => sum + (f.original_size ?? 0),
    0
  );
  const totalStorageTotalBytes = providers.reduce(
    (sum, p) => sum + (p.quotaTotalBytes ?? 0),
    0
  );

  return (
    <DashboardShell
      providers={providers.map((p) => ({
        providerId: p.providerId,
        displayName: p.displayName,
        status: p.status,
        quotaUsedBytes: p.quotaUsedBytes,
        quotaTotalBytes: p.quotaTotalBytes,
      }))}
      totalStorageUsedBytes={totalStorageUsedBytes}
      totalStorageTotalBytes={totalStorageTotalBytes}
    >
      {children}
    </DashboardShell>
  );
}
