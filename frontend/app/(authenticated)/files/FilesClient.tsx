"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { FileMetadata, deleteFile } from "@/lib/api/files";
import { cn } from "@/lib/utils";
import { toast } from "sonner";

function formatBytes(bytes: number): string {
  if (bytes >= 1e12) return `${(bytes / 1e12).toFixed(1)} TB`;
  if (bytes >= 1e9) return `${(bytes / 1e9).toFixed(1)} GB`;
  return `${(bytes / 1e6).toFixed(1)} MB`;
}

function formatDate(iso: string): string {
  return new Date(iso).toLocaleDateString("en-US", {
    year: "numeric",
    month: "short",
    day: "numeric",
  });
}

interface FilesClientProps {
  initialFiles: FileMetadata[];
}

export function FilesClient({ initialFiles }: FilesClientProps) {
  const router = useRouter();
  const [files, setFiles] = useState(initialFiles);
  const [search, setSearch] = useState("");

  const filtered = files.filter((f) =>
    f.original_name.toLowerCase().includes(search.toLowerCase())
  );

  const handleDelete = (id: string) => {
    toast("Are you sure?", {
      description: "This will permanently delete the file from all providers.",
      action: {
        label: "Delete",
        onClick: async () => {
          try {
            await deleteFile(id);
            setFiles((prev) => prev.filter((f) => f.file_id !== id));
            toast.success("File deleted");
          } catch {
            toast.error("Failed to delete file");
          }
        },
      },
    });
  };

  return (
    <div className="space-y-4 max-w-5xl">
      {/* Search */}
      <div className="flex items-center gap-3">
        <input
          type="text"
          placeholder="Search files..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          className="font-mono text-xs border bg-white px-3 py-2 w-64 outline-none focus:ring-1 focus:ring-black placeholder:text-neutral-400"
        />
        <span className="font-mono text-[10px] text-neutral-400">
          {filtered.length} {filtered.length === 1 ? "file" : "files"}
        </span>
      </div>

      {/* Table */}
      {files.length === 0 ? (
        <EmptyState />
      ) : filtered.length === 0 ? (
        <div className="border px-4 py-10 text-center">
          <p className="font-mono text-xs text-neutral-400">No files match &ldquo;{search}&rdquo;</p>
        </div>
      ) : (
        <div className="border">
          {/* Header */}
          <div className="grid grid-cols-[1fr_80px_60px_120px_80px_100px_36px] gap-4 px-4 py-2 border-b bg-neutral-50">
            {["Name", "Size", "n/k", "Health", "Status", "Created", ""].map((h) => (
              <span key={h} className="font-mono text-[9px] uppercase tracking-widest text-neutral-400">
                {h}
              </span>
            ))}
          </div>

          {/* Rows */}
          <div className="divide-y">
            {filtered.map((f) => {
              const pct = f.health_status?.health_percent ?? 100;
              const status =
                f.status === "UPLOADED"
                  ? "Healthy"
                  : f.status === "DEGRADED"
                  ? "Degraded"
                  : f.status === "UPLOADING"
                  ? "Uploading"
                  : f.status;

              const statusClass =
                f.status === "UPLOADED"
                  ? "border-black text-black"
                  : f.status === "DEGRADED"
                  ? "border-neutral-400 text-neutral-500"
                  : "border-neutral-200 text-neutral-400";

              return (
                <div
                  key={f.file_id}
                  className="grid grid-cols-[1fr_80px_60px_120px_80px_100px_36px] gap-4 px-4 py-3 hover:bg-neutral-50 transition-colors"
                >
                  <span
                    className="font-mono text-xs truncate cursor-pointer hover:underline"
                    onClick={() => router.push(`/files/${f.file_id}`)}
                  >
                    {f.original_name}
                  </span>
                  <span className="font-mono text-[11px] text-neutral-600">
                    {formatBytes(f.original_size)}
                  </span>
                  <span className="font-mono text-[11px] text-neutral-500">
                    ({f.n},{f.k})
                  </span>
                  <div className="flex items-center gap-2">
                    <div className="flex-1 h-0.5 bg-neutral-200">
                      <div className="h-full bg-black" style={{ width: `${pct}%` }} />
                    </div>
                    <span className="font-mono text-[9px] text-neutral-400 w-8 text-right shrink-0">
                      {pct}%
                    </span>
                  </div>
                  <span
                    className={cn(
                      "font-mono text-[9px] uppercase tracking-wider px-2 py-0.5 border self-center text-center",
                      statusClass
                    )}
                  >
                    {status}
                  </span>
                  <span className="font-mono text-[10px] text-neutral-400">
                    {formatDate(f.created_at)}
                  </span>
                  <button
                    onClick={(e) => { e.stopPropagation(); handleDelete(f.file_id); }}
                    className="font-mono text-[10px] text-neutral-300 hover:text-black transition-colors self-center"
                    title="Delete file"
                  >
                    ×
                  </button>
                </div>
              );
            })}
          </div>
        </div>
      )}
    </div>
  );

  function EmptyState() {
    return (
      <div className="border px-4 py-16 text-center">
        <p className="font-mono text-xs text-neutral-400 mb-1">No files uploaded yet.</p>
        <p className="font-mono text-[10px] text-neutral-300">
          Use the Upload button in the sidebar to get started.
        </p>
      </div>
    );
  }
}
