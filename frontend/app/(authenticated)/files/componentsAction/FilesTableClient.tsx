"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useState } from "react";
import { Download, Eye, Loader2, Trash2 } from "lucide-react";
import { deleteFile, FileMetadata } from "@/lib/api/files";
import { formatBytes } from "@/lib/utils";
import { toast } from "sonner";

interface FilesTableClientProps {
  initialFiles: FileMetadata[];
}

export function FilesTableClient({ initialFiles }: FilesTableClientProps) {
  const router = useRouter();
  const [deletingId, setDeletingId] = useState<string | null>(null);
  const [fileToDelete, setFileToDelete] = useState<FileMetadata | null>(null);

  async function confirmDelete() {
    if (!fileToDelete) return;
    setDeletingId(fileToDelete.file_id);
    try {
      await deleteFile(fileToDelete.file_id, true);
      toast.success("File deleted");
      setFileToDelete(null);
      router.refresh();
    } catch (error) {
      const message = error instanceof Error ? error.message : "Failed to delete file";
      toast.error(message);
    } finally {
      setDeletingId(null);
    }
  }

  return (
    <>
      <section className="border bg-white">
        <div className="grid grid-cols-[1.8fr_0.8fr_0.8fr_0.9fr_0.9fr] gap-4 px-5 py-3 border-b bg-neutral-50">
          {["File", "Size", "Status", "Created", "Actions"].map((h) => (
            <span key={h} className="font-mono text-[11px] uppercase tracking-[0.08em] text-neutral-400">
              {h}
            </span>
          ))}
        </div>

        {initialFiles.length === 0 ? (
          <div className="px-5 py-10 text-center font-mono text-sm text-neutral-400">
            No files uploaded yet.
          </div>
        ) : (
          <div className="divide-y">
            {initialFiles.map((file) => (
              <div
                key={file.file_id}
                className="grid grid-cols-[1.8fr_0.8fr_0.8fr_0.9fr_0.9fr] items-center gap-4 px-5 py-3"
              >
                <Link
                  href={`/files/${file.file_id}`}
                  className="truncate font-mono text-sm text-neutral-800 hover:underline"
                >
                  {file.original_name}
                </Link>
                <span className="font-mono text-sm text-neutral-500">
                  {formatBytes(file.original_size ?? 0)}
                </span>
                <span className="font-mono text-[11px] uppercase tracking-wider text-neutral-500">
                  {file.status}
                </span>
                <span className="font-mono text-[11px] text-neutral-500">
                  {new Date(file.created_at).toLocaleDateString("en-US")}
                </span>
                <div className="flex items-center gap-1">
                  <Link
                    href={`/files/${file.file_id}`}
                    className="inline-flex h-8 w-8 items-center justify-center border border-transparent text-neutral-500 transition-colors hover:border-neutral-200 hover:text-black"
                    aria-label={`View details for ${file.original_name}`}
                    title="Details"
                  >
                    <Eye className="h-4 w-4" />
                  </Link>
                  <a
                    href={`/api/download/${file.file_id}`}
                    className="inline-flex h-8 w-8 items-center justify-center border border-transparent text-neutral-500 transition-colors hover:border-neutral-200 hover:text-black"
                    aria-label={`Download ${file.original_name}`}
                    title="Download"
                  >
                    <Download className="h-4 w-4" />
                  </a>
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
            ))}
          </div>
        )}
      </section>

      {fileToDelete && (
        <div
          className="fixed inset-0 z-50 flex items-center justify-center bg-black/20 p-4"
          onClick={() => (deletingId ? null : setFileToDelete(null))}
        >
          <div
            className="w-full max-w-md border bg-white p-5"
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
    </>
  );
}
