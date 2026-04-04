"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { type FileMetadata } from "@/lib/api/files";
import { ProvidersUploadFilesModal } from "@/components/dashboard/ProvidersUploadFilesModal";
import { FilesTableClient } from "./FilesTableClient";

interface FilesPageClientProps {
  initialFiles: FileMetadata[];
}

export function FilesPageClient({ initialFiles }: FilesPageClientProps) {
  const router = useRouter();
  const [uploadModalOpen, setUploadModalOpen] = useState(false);

  return (
    <div className="space-y-5">
      <div className="flex items-end justify-between border-b pb-4">
        <div>
          <p className="mb-0.5 font-mono text-[11px] uppercase tracking-[0.15em] text-neutral-400">
            Storage
          </p>
          <h1 className="text-2xl font-semibold tracking-tight">Files</h1>
        </div>
        <button
          type="button"
          onClick={() => setUploadModalOpen(true)}
          className="font-mono text-[11px] uppercase tracking-wider border px-4 py-2 transition-colors hover:bg-black hover:text-white"
        >
          Upload More
        </button>
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
