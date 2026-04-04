import { AppSidebar } from "@/components/app-sidebar";
import {
  SidebarInset,
  SidebarProvider,
  SidebarTrigger,
} from "@/components/ui/sidebar";
import { getProviders } from "@/lib/api/providers";
import { getFiles } from "@/lib/api/files";
import { getCredentialStatus } from "@/lib/api/credentials";

export default async function CredentialsLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const credentialStatus = await getCredentialStatus().catch(() => ({
    configured: false,
  }));

  if (!credentialStatus.configured) {
    return (
      <div className="mx-auto min-h-screen w-full max-w-5xl p-6">
        {children}
      </div>
    );
  }

  const [providers, files] = await Promise.all([
    getProviders().catch(() => []),
    getFiles().catch(() => []),
  ]);
  const totalStorageUsedBytes = files.reduce(
    (sum, f) => sum + (f.original_size ?? 0),
    0,
  );
  const totalStorageTotalBytes = providers.reduce(
    (sum, p) => sum + (p.quotaTotalBytes ?? 0),
    0,
  );

  return (
    <SidebarProvider className="h-svh !min-h-0">
      <AppSidebar
        totalStorageUsedBytes={totalStorageUsedBytes}
        totalStorageTotalBytes={totalStorageTotalBytes}
      />
      <SidebarInset className="min-h-0 overflow-hidden">
        <header className="flex h-14 shrink-0 items-center gap-2.5 border-b px-4">
          <SidebarTrigger className="-ml-1" />
          <div className="h-5 w-px bg-neutral-200" />
          <span className="font-mono text-[12px] font-medium uppercase tracking-widest text-neutral-500">
            Omnishard
          </span>
        </header>
        <div className="min-h-0 flex-1 overflow-y-auto p-6">{children}</div>
      </SidebarInset>
    </SidebarProvider>
  );
}
