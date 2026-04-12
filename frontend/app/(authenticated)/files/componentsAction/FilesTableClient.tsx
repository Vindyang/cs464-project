"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useState } from "react";
import { Eye, Loader2, Trash2, Upload, FileUp } from "lucide-react";
import { deleteFile, FileMetadata } from "@/lib/api/files";
import { cn, formatBytes, formatUtcDate, formatUtcDateTime } from "@/lib/utils";
import { toast } from "sonner";
import { helpToast } from "@/lib/help/help-toast";

import { DownloadFileButton } from "./DownloadFileButton";
import { StatusBadgeWithTooltip } from "@/components/files/StatusBadgeWithTooltip";
import { ProvidersUploadFilesModal } from "@/components/dashboard/ProvidersUploadFilesModal";

interface FilesTableClientProps {
  initialFiles: FileMetadata[];
}

function getHealthTone(file: FileMetadata) {
  if (file.status === "CORRUPTED") return "critical";
  const missingShards = file.health_status?.missing_shards ?? 0;
  const corruptedShards = file.health_status?.corrupted_shards ?? 0;
  if (missingShards > 0 || corruptedShards > 0 || file.status === "DEGRADED") {
    return "risk";
  }
  return "healthy";
}

export function FilesTableClient({ initialFiles }: FilesTableClientProps) {
  const router = useRouter();
  const [deletingId, setDeletingId] = useState<string | null>(null);
  const [fileToDelete, setFileToDelete] = useState<FileMetadata | null>(null);
  const [uploadModalOpen, setUploadModalOpen] = useState(false);

  // Multi-select state
  const [selectedIds, setSelectedIds] = useState<Set<string>>(new Set());
  const [bulkDeleteOpen, setBulkDeleteOpen] = useState(false);
  const [bulkDeleting, setBulkDeleting] = useState(false);

  const allSelected = initialFiles.length > 0 && selectedIds.size === initialFiles.length;
  const someSelected = selectedIds.size > 0 && !allSelected;

  function toggleAll() {
    if (allSelected) {
      setSelectedIds(new Set());
    } else {
      setSelectedIds(new Set(initialFiles.map((f) => f.file_id)));
    }
  }

  function toggleOne(id: string) {
    setSelectedIds((prev) => {
      const next = new Set(prev);
      if (next.has(id)) {
        next.delete(id);
      } else {
        next.add(id);
      }
      return next;
    });
  }

  async function confirmDelete() {
    if (!fileToDelete) return;
    setDeletingId(fileToDelete.file_id);
    try {
      await deleteFile(fileToDelete.file_id, true);
      toast.success("File deleted");
      setFileToDelete(null);
      router.refresh();
    } catch (error) {
      helpToast(error);
    } finally {
      setDeletingId(null);
    }
  }

  async function confirmBulkDelete() {
    setBulkDeleting(true);
    const ids = Array.from(selectedIds);
    const results = await Promise.allSettled(ids.map((id) => deleteFile(id, true)));
    const failed = results.filter((r) => r.status === "rejected").length;
    const succeeded = results.length - failed;

    if (failed === 0) {
      toast.success(`${succeeded} ${succeeded === 1 ? "file" : "files"} deleted`);
    } else {
      toast.error(`${failed} of ${ids.length} files failed to delete`);
    }

    setBulkDeleting(false);
    setBulkDeleteOpen(false);
    setSelectedIds(new Set());
    router.refresh();
  }

  return (
    <>
      {/* Bulk action toolbar */}
      {selectedIds.size > 0 && (
        <div className="mb-2 flex items-center justify-between border bg-neutral-50 px-4 py-2 dark:border-neutral-800 dark:bg-neutral-900/60">
          <span className="font-mono text-[11px] text-neutral-500">
            {selectedIds.size} {selectedIds.size === 1 ? "file" : "files"} selected
          </span>
          <button
            type="button"
            onClick={() => setBulkDeleteOpen(true)}
            className="flex items-center gap-1.5 border border-red-600 bg-red-600 px-3 py-1.5 font-mono text-[11px] uppercase tracking-[0.08em] text-white transition-colors hover:bg-red-700"
          >
            <Trash2 className="h-3 w-3" />
            Delete selected ({selectedIds.size})
          </button>
        </div>
      )}

      <section className="border bg-white dark:bg-neutral-950">
        <div className="grid grid-cols-[auto_1.8fr_0.8fr_0.9fr_1.1fr_1fr_0.9fr] gap-4 border-b bg-neutral-50 px-5 py-3 dark:border-neutral-800 dark:bg-neutral-900/60">
          {/* Select-all checkbox */}
          {initialFiles.length > 0 ? (
            <input
              type="checkbox"
              checked={allSelected}
              ref={(el) => {
                if (el) el.indeterminate = someSelected;
              }}
              onChange={toggleAll}
              className="h-4 w-4 cursor-pointer accent-sky-600"
              aria-label="Select all files"
            />
          ) : (
            <span />
          )}
          {["File", "Size", "Health", "Last Checked", "Created", "Actions"].map((h) => (
            <span key={h} className="font-mono text-[11px] uppercase tracking-[0.08em] text-neutral-400">
              {h}
            </span>
          ))}
        </div>

        {initialFiles.length === 0 ? (
          <div className="flex flex-col items-center justify-center p-12 text-center border-b border-neutral-200 dark:border-neutral-800">
            <div className="mb-4 flex h-14 w-14 items-center justify-center border border-neutral-200 bg-neutral-50 dark:border-neutral-800 dark:bg-neutral-900/60">
              <FileUp className="h-6 w-6 text-neutral-400" />
            </div>
            <h3 className="mb-1.5 text-base font-semibold text-neutral-900 dark:text-neutral-100">No files uploaded yet</h3>
            <p className="mb-6 font-mono text-[13px] text-neutral-500 max-w-sm">
              Upload files to securely shard and encrypt them across your connected providers.
            </p>
            <button
              type="button"
              onClick={() => setUploadModalOpen(true)}
              className="flex items-center gap-2 border border-sky-600 bg-sky-600 px-4 py-2 font-mono text-[11px] font-semibold uppercase tracking-[0.08em] text-white transition-colors hover:bg-sky-700"
            >
              <Upload className="h-3.5 w-3.5" />
              <span>Upload File</span>
            </button>
          </div>
        ) : (
          <div className="divide-y">
            {initialFiles.map((file) => (
              (() => {
                const tone = getHealthTone(file);
                const healthPercent = Math.round(file.health_status?.health_percent ?? 100);
                return (
              <div
                key={file.file_id}
                className={cn(
                  "grid grid-cols-[auto_1.8fr_0.8fr_0.9fr_1.1fr_1fr_0.9fr] items-center gap-4 px-5 py-3",
                  tone === "risk" && "bg-amber-50/70 dark:bg-amber-950/20",
                  tone === "critical" && "bg-red-50/70 dark:bg-red-950/20",
                )}
              >
                <input
                  type="checkbox"
                  checked={selectedIds.has(file.file_id)}
                  onChange={() => toggleOne(file.file_id)}
                  className="h-4 w-4 cursor-pointer accent-sky-600"
                  aria-label={`Select ${file.original_name}`}
                />
                <Link
                  href={`/files/${file.file_id}`}
                  className="truncate font-mono text-sm text-neutral-800 hover:underline dark:text-neutral-100"
                >
                  {file.original_name}
                </Link>
                <span className="font-mono text-sm text-neutral-500">
                  {formatBytes(file.original_size ?? 0)}
                </span>
                <div className="space-y-1">
                  <StatusBadgeWithTooltip status={file.status} healthStatus={file.health_status} />
                  <div className="font-mono text-[11px] text-neutral-500 dark:text-neutral-400">
                    {healthPercent}% healthy
                    {(file.health_status?.missing_shards ?? 0) > 0 && ` · ${file.health_status?.missing_shards} missing`}
                  </div>
                </div>
                <div className="font-mono text-[11px] text-neutral-500 dark:text-neutral-400">
                  <div>{formatUtcDateTime(file.last_health_refresh_at)}</div>
                </div>
                <span className="font-mono text-[11px] text-neutral-500 dark:text-neutral-400">
                  {formatUtcDate(file.created_at)}
                </span>
                <div className="flex items-center gap-1">
                  <Link
                    href={`/files/${file.file_id}`}
                    className="inline-flex h-8 w-8 cursor-pointer items-center justify-center border border-transparent text-neutral-500 transition-colors hover:border-neutral-200 hover:text-black dark:hover:border-neutral-700 dark:hover:text-white"
                    aria-label={`View details for ${file.original_name}`}
                    title="Details"
                  >
                    <Eye className="h-4 w-4" />
                  </Link>
                  <DownloadFileButton
                    fileId={file.file_id}
                    fileName={file.original_name}
                    requiredShards={file.k}
                    healthStatus={file.health_status}
                  />
                  <button
                    type="button"
                    disabled={deletingId === file.file_id}
                    onClick={() => setFileToDelete(file)}
                    className="inline-flex h-8 w-8 items-center justify-center border border-transparent text-red-500 transition-colors hover:border-red-200 hover:text-red-700 disabled:cursor-not-allowed disabled:opacity-50"
                    aria-label={`Delete ${file.original_name}`}
                    title="Delete"
                  >
                    {deletingId === file.file_id ? (
                      <Loader2 className="h-4 w-4 animate-spin" />
                    ) : (
                      <Trash2 className="h-4 w-4" />
                    )}
                  </button>
                </div>
              </div>
                );
              })()
            ))}
          </div>
        )}
      </section>

      {/* Single-file delete modal */}
      {fileToDelete && (
        <div
          className="fixed inset-0 z-50 flex items-center justify-center bg-black/20 p-4"
          onClick={() => (deletingId ? null : setFileToDelete(null))}
        >
          <div
            className="w-full max-w-md border bg-white p-5 dark:border-neutral-800 dark:bg-neutral-950"
            onClick={(e) => e.stopPropagation()}
          >
            <p className="font-mono text-[11px] uppercase tracking-widest text-neutral-400">
              Confirm Delete
            </p>
            <h3 className="mt-2 text-base font-semibold text-neutral-900">{fileToDelete.original_name}</h3>
            <p className="mt-2 font-mono text-sm text-neutral-500">
              This will remove file metadata and attempt to delete shards from connected providers.
            </p>
            <div className="mt-5 flex justify-end gap-2">
              <button
                type="button"
                disabled={Boolean(deletingId)}
                onClick={() => setFileToDelete(null)}
                className="font-mono text-[11px] uppercase tracking-wider border px-3 py-2 hover:bg-black hover:text-white disabled:cursor-not-allowed disabled:opacity-50"
              >
                Cancel
              </button>
              <button
                type="button"
                disabled={Boolean(deletingId)}
                onClick={confirmDelete}
                className="font-mono text-[11px] uppercase tracking-wider border border-red-600 bg-red-600 px-3 py-2 text-white hover:bg-red-700 disabled:cursor-not-allowed disabled:opacity-50"
              >
                {deletingId ? "Deleting..." : "Delete File"}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Bulk delete modal */}
      {bulkDeleteOpen && (
        <div
          className="fixed inset-0 z-50 flex items-center justify-center bg-black/20 p-4"
          onClick={() => (bulkDeleting ? null : setBulkDeleteOpen(false))}
        >
          <div
            className="w-full max-w-md border bg-white p-5 dark:border-neutral-800 dark:bg-neutral-950"
            onClick={(e) => e.stopPropagation()}
          >
            <p className="font-mono text-[11px] uppercase tracking-widest text-neutral-400">
              Confirm Delete
            </p>
            <h3 className="mt-2 text-base font-semibold text-neutral-900">
              Delete {selectedIds.size} {selectedIds.size === 1 ? "file" : "files"}?
            </h3>
            <p className="mt-2 font-mono text-sm text-neutral-500">
              This will remove file metadata and attempt to delete shards from all connected providers. This cannot be undone.
            </p>
            <div className="mt-5 flex justify-end gap-2">
              <button
                type="button"
                disabled={bulkDeleting}
                onClick={() => setBulkDeleteOpen(false)}
                className="font-mono text-[11px] uppercase tracking-wider border px-3 py-2 hover:bg-black hover:text-white disabled:cursor-not-allowed disabled:opacity-50"
              >
                Cancel
              </button>
              <button
                type="button"
                disabled={bulkDeleting}
                onClick={confirmBulkDelete}
                className="flex items-center gap-2 font-mono text-[11px] uppercase tracking-wider border border-red-600 bg-red-600 px-3 py-2 text-white hover:bg-red-700 disabled:cursor-not-allowed disabled:opacity-50"
              >
                {bulkDeleting && <Loader2 className="h-3 w-3 animate-spin" />}
                {bulkDeleting ? "Deleting..." : `Delete ${selectedIds.size} ${selectedIds.size === 1 ? "File" : "Files"}`}
              </button>
            </div>
          </div>
        </div>
      )}

      <ProvidersUploadFilesModal
        open={uploadModalOpen}
        onOpenChange={setUploadModalOpen}
        onUploadSuccess={() => router.refresh()}
      />
    </>
  );
}
