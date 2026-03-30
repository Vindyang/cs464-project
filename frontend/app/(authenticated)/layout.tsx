import { AppSidebar } from "@/components/app-sidebar";
import { SidebarInset, SidebarProvider, SidebarTrigger } from "@/components/ui/sidebar";
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
    <SidebarProvider className="h-svh !min-h-0">
      <AppSidebar
        totalStorageUsedBytes={totalStorageUsedBytes}
        totalStorageTotalBytes={totalStorageTotalBytes}
      />
      <SidebarInset className="min-h-0 overflow-hidden">
        <header className="flex h-12 shrink-0 items-center gap-2 border-b px-4">
          <SidebarTrigger className="-ml-1" />
          <div className="h-4 w-px bg-neutral-200" />
          <span className="font-mono text-[10px] uppercase tracking-widest text-neutral-400">
            Nebula Drive
          </span>
        </header>
        <div className="min-h-0 flex-1 overflow-y-auto p-6">{children}</div>
      </SidebarInset>
    </SidebarProvider>
  );
}
