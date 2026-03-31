"use client";

import { useState } from "react";
import { Sidebar, SidebarProviderData } from "@/components/dashboard/Sidebar";
import { FileUploadModal } from "@/components/dashboard/FileUploadModal";
import { GridBackground } from "@/components/ui/grid-background";
import Link from "next/link";

interface DashboardShellProps {
  children: React.ReactNode;
  providers: SidebarProviderData[];
  totalStorageUsedBytes: number;
  totalStorageTotalBytes: number;
  pageTitle?: string;
}

export function DashboardShell({
  children,
  providers,
  totalStorageUsedBytes,
  totalStorageTotalBytes,
  pageTitle = "Dashboard",
}: DashboardShellProps) {
  const [uploadModalOpen, setUploadModalOpen] = useState(false);

  const handleUpload = (file: File, encrypt: boolean) => {
    console.log("Upload started:", file.name, "Encrypted:", encrypt);
    setTimeout(() => {
      setUploadModalOpen(false);
    }, 1000);
  };

  return (
    <div className="flex h-screen overflow-hidden bg-background relative isolate">
      <GridBackground />
      <Sidebar
        className="hidden md:flex z-10"
        onUploadClick={() => setUploadModalOpen(true)}
        providers={providers}
        totalStorageUsedBytes={totalStorageUsedBytes}
        totalStorageTotalBytes={totalStorageTotalBytes}
      />

      <div className="flex-1 flex flex-col min-w-0 overflow-hidden relative z-0">
        <header className="flex items-center justify-between px-6 py-3 border-b bg-background/80 backdrop-blur-sm z-10 relative">
          <span className="font-mono text-[11px] uppercase tracking-widest text-neutral-500">
            {pageTitle}
          </span>
          <Link
            href="/settings"
            className="font-mono text-[10px] uppercase tracking-widest text-neutral-500 hover:text-black transition-colors"
          >
            Settings
          </Link>
        </header>

        <main className="flex-1 overflow-y-auto p-6 relative">{children}</main>
      </div>

      <FileUploadModal
        open={uploadModalOpen}
        onOpenChange={setUploadModalOpen}
        onUpload={handleUpload}
      />
    </div>
  );
}
