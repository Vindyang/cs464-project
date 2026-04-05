"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { type FileMetadata } from "@/lib/api/files";
import { ProvidersUploadFilesModal } from "@/components/dashboard/ProvidersUploadFilesModal";
import { FilesTableClient } from "./FilesTableClient";
import { toast } from "sonner";

interface FilesPageClientProps {
  initialFiles: FileMetadata[];
}

export function FilesPageClient({ initialFiles }: FilesPageClientProps) {
  const router = useRouter();
  const [uploadModalOpen, setUploadModalOpen] = useState(false);
  const [refreshingAll, setRefreshingAll] = useState(false);

  async function refreshAllHealth() {
    if (refreshingAll) return;
    setRefreshingAll(true);
    try {
      const res = await fetch("/api/files/health/refresh", { method: "POST" });
      const data = await res.json().catch(() => ({}));
      if (!res.ok) {
        throw new Error(data?.details || data?.message || data?.error || "Failed to refresh health");
      }
      const filesScanned = data?.files_scanned ?? 0;
      const missing = data?.marked_missing ?? 0;
      const skipped = data?.skipped_errors ?? 0;
      toast.success(`All files refreshed (${filesScanned} files, ${missing} missing, ${skipped} skipped)`);
      router.refresh();
    } catch (error) {
      const message = error instanceof Error ? error.message : "Failed to refresh health";
      toast.error(message);
    } finally {
      setRefreshingAll(false);
    }
  }

  return (
    <div className="space-y-5">
      <div className="flex items-end justify-between border-b pb-4">
        <div>
          <p className="mb-0.5 font-mono text-[11px] uppercase tracking-[0.15em] text-neutral-400">
            Storage
          </p>
          <h1 className="text-2xl font-semibold tracking-tight">Files</h1>
        </div>
        <div className="flex items-center gap-2">
          <button
            type="button"
            onClick={refreshAllHealth}
            disabled={refreshingAll}
            className="font-mono text-[11px] uppercase tracking-wider border px-4 py-2 transition-colors hover:bg-black hover:text-white disabled:cursor-not-allowed disabled:opacity-50"
          >
            {refreshingAll ? "Refreshing..." : "Refresh All Health"}
          </button>
          <button
            type="button"
            onClick={() => setUploadModalOpen(true)}
            className="font-mono text-[11px] uppercase tracking-wider border px-4 py-2 transition-colors hover:bg-black hover:text-white"
          >
            Upload More
          </button>
        </div>
      </div>

      <FilesTableClient initialFiles={initialFiles} />

      <ProvidersUploadFilesModal
        open={uploadModalOpen}
        onOpenChange={setUploadModalOpen}
        onUploadSuccess={() => router.refresh()}
      />
    </div>
  );
}
