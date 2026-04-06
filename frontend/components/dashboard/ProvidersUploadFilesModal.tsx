"use client";

import { useEffect, useId, useRef, useState } from "react";
import { Upload, X } from "lucide-react";
import { toast } from "sonner";
import { cn } from "@/lib/utils";
import { FileMetadata, getFiles, uploadFile } from "@/lib/api/files";

interface UploadHistoryItem {
  id: string;
  filename: string;
  sizeLabel: string;
  provider: string;
  date: string;
  status: "Success" | "Failed";
}

interface ProvidersUploadFilesModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onUploadSuccess?: (filename: string) => void;
}

function formatBytes(value: number) {
  if (value <= 0) return "0 B";
  const units = ["B", "KB", "MB", "GB", "TB"];
  const idx = Math.min(Math.floor(Math.log(value) / Math.log(1024)), units.length - 1);
  const scaled = value / 1024 ** idx;
  const digits = idx === 0 ? 0 : scaled >= 10 ? 1 : 2;
  return `${scaled.toFixed(digits)} ${units[idx]}`;
}

const uploadHistoryDateFormatter = new Intl.DateTimeFormat(undefined, {
  year: "numeric",
  month: "2-digit",
  day: "2-digit",
  hour: "2-digit",
  minute: "2-digit",
});

function toHistoryItem(file: FileMetadata): UploadHistoryItem {
  return {
    id: file.file_id,
    filename: file.original_name,
    sizeLabel: formatBytes(file.original_size ?? 0),
    provider: "Distributed",
    date: uploadHistoryDateFormatter.format(new Date(file.created_at)),
    status: file.status === "FAILED" ? "Failed" : "Success",
  };
}

export function ProvidersUploadFilesModal({
  open,
  onOpenChange,
  onUploadSuccess,
}: ProvidersUploadFilesModalProps) {
  const fileInputRef = useRef<HTMLInputElement | null>(null);
  const closeButtonRef = useRef<HTMLButtonElement | null>(null);
  const progressTimer = useRef<ReturnType<typeof setInterval> | null>(null);
  const uploadDialogTitleId = useId();
  const uploadDialogDescriptionId = useId();
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [dragActive, setDragActive] = useState(false);
  const [isUploading, setIsUploading] = useState(false);
  const [uploadedBytes, setUploadedBytes] = useState(0);
  const [history, setHistory] = useState<UploadHistoryItem[]>([]);

  const refreshUploadHistory = async () => {
    const files = await getFiles().catch(() => [] as FileMetadata[]);
    const next = [...files]
      .sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime())
      .slice(0, 8)
      .map(toHistoryItem);
    setHistory(next);
  };

  useEffect(() => {
    return () => {
      if (progressTimer.current) clearInterval(progressTimer.current);
    };
  }, []);

  useEffect(() => {
    if (!open) return;
    closeButtonRef.current?.focus();
    refreshUploadHistory();
    const onKeyDown = (event: KeyboardEvent) => {
      if (event.key === "Escape") onOpenChange(false);
    };
    window.addEventListener("keydown", onKeyDown);
    return () => window.removeEventListener("keydown", onKeyDown);
  }, [open, onOpenChange]);

  const activeTransfer = selectedFile
    ? {
        filename: selectedFile.name,
        total: selectedFile.size,
        uploaded: Math.min(uploadedBytes, selectedFile.size),
      }
    : null;

  const progressPct = activeTransfer
    ? Math.min(Math.round((activeTransfer.uploaded / Math.max(activeTransfer.total, 1)) * 100), 100)
    : 0;

  const pickFile = () => fileInputRef.current?.click();

  const onInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) {
      setSelectedFile(file);
      setUploadedBytes(0);
    }
  };

  const onDrop: React.DragEventHandler<HTMLDivElement> = (e) => {
    e.preventDefault();
    e.stopPropagation();
    setDragActive(false);
    const file = e.dataTransfer.files?.[0];
    if (file) {
      setSelectedFile(file);
      setUploadedBytes(0);
    }
  };

  const startUpload = async () => {
    if (!selectedFile || isUploading) return;

    setIsUploading(true);
    setUploadedBytes(0);

    progressTimer.current = setInterval(() => {
      setUploadedBytes((prev) => {
        const cap = Math.floor(selectedFile.size * 0.92);
        const next = prev + Math.max(Math.floor(selectedFile.size / 20), 64 * 1024);
        return Math.min(next, cap);
      });
    }, 220);

    try {
      await uploadFile(selectedFile);
      setUploadedBytes(selectedFile.size);
      await refreshUploadHistory();
      onUploadSuccess?.(selectedFile.name);
      toast.success(`Uploaded ${selectedFile.name}`);
      setSelectedFile(null);
      setTimeout(() => onOpenChange(false), 300);
    } catch (err) {
      const message = err instanceof Error ? err.message : "Failed to upload file";
      toast.error(message);
    } finally {
      setIsUploading(false);
      if (progressTimer.current) clearInterval(progressTimer.current);
      progressTimer.current = null;
    }
  };

  if (!open) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/20 p-4" onClick={() => onOpenChange(false)}>
      <div
        className="w-full max-w-6xl border bg-white"
        onClick={(e) => e.stopPropagation()}
        role="dialog"
        aria-modal="true"
        aria-labelledby={uploadDialogTitleId}
        aria-describedby={uploadDialogDescriptionId}
      >
        <div className="flex items-center justify-between border-b px-6 py-4">
          <h2 id={uploadDialogTitleId} className="font-mono text-[12px] font-medium uppercase tracking-[0.12em] text-neutral-600">
            Providers / Upload Files
          </h2>
          <button
            onClick={() => onOpenChange(false)}
            className="text-neutral-400 transition-colors hover:text-black"
            aria-label="Close upload modal"
            ref={closeButtonRef}
          >
            <X className="h-4 w-4" />
          </button>
        </div>

        <div className="grid gap-5 p-6 lg:grid-cols-[minmax(0,1.6fr)_minmax(0,1fr)]">
          <div className="min-w-0 space-y-5">
            <div className="border p-5">
              <p className="mb-3 font-mono text-[12px] font-medium uppercase tracking-[0.12em] text-neutral-500">
                Upload Configuration
              </p>
              <p id={uploadDialogDescriptionId} className="font-mono text-[12px] font-medium leading-relaxed text-neutral-600">
                Files are automatically sharded and distributed across currently connected providers.
              </p>

              <input ref={fileInputRef} type="file" className="hidden" onChange={onInputChange} />
              <div
                role="button"
                tabIndex={0}
                onClick={pickFile}
                onKeyDown={(e) => {
                  if (e.key === "Enter" || e.key === " ") {
                    e.preventDefault();
                    pickFile();
                  }
                }}
                onDrop={onDrop}
                onDragOver={(e) => {
                  e.preventDefault();
                  setDragActive(true);
                }}
                onDragLeave={(e) => {
                  e.preventDefault();
                  setDragActive(false);
                }}
                className={cn(
                  "mt-4 grid min-h-[280px] place-items-center border border-dashed px-4 py-10 text-center transition-colors",
                  dragActive ? "border-black bg-neutral-50" : "border-neutral-300"
                )}
              >
                <div className="space-y-3">
                  <div className="mx-auto grid h-10 w-10 place-items-center border border-neutral-300 text-neutral-500">
                    <Upload className="h-5 w-5" />
                  </div>
                  <div>
                    <p className="max-w-full truncate font-mono text-sm text-neutral-700">
                      {selectedFile ? selectedFile.name : "Click or drag files to upload"}
                    </p>
                    <p className="mt-1 font-mono text-[12px] font-medium text-neutral-500">
                      {selectedFile ? formatBytes(selectedFile.size) : "Max file size: 500GB"}
                    </p>
                  </div>
                </div>
              </div>
            </div>

            <div className="border p-5">
              <p className="mb-3 font-mono text-[12px] font-medium uppercase tracking-[0.12em] text-neutral-500">
                Recent Upload History
              </p>
              <div className="overflow-x-auto">
                <div className="min-w-[680px]">
                  <div className="grid grid-cols-[1.7fr_1fr_1fr_1.1fr_0.9fr] gap-3 border-b pb-2.5 font-mono text-[11px] font-medium uppercase tracking-[0.08em] text-neutral-500">
                    <span>Filename</span>
                    <span>Size</span>
                    <span>Provider</span>
                    <span>Date</span>
                    <span>Status</span>
                  </div>
                  <div className="space-y-1 pt-2.5">
                    {history.length === 0 ? (
                      <p className="py-6 text-center font-mono text-sm text-neutral-400">
                        No uploads found yet.
                      </p>
                    ) : (
                      history.slice(0, 3).map((item) => (
                        <div
                          key={item.id}
                          className="grid grid-cols-[1.7fr_1fr_1fr_1.1fr_0.9fr] gap-3 py-2 font-mono text-sm"
                        >
                          <span className="block min-w-0 truncate">{item.filename}</span>
                          <span className="text-neutral-600">{item.sizeLabel}</span>
                          <span className="text-neutral-600">{item.provider}</span>
                          <span className="text-neutral-500">{item.date}</span>
                          <span
                            className={cn(
                              item.status === "Success" ? "text-emerald-600" : "text-red-600"
                            )}
                          >
                            {item.status}
                          </span>
                        </div>
                      ))
                    )}
                  </div>
                </div>
              </div>
            </div>
          </div>

          <div className="min-w-0 space-y-5 border p-5">
            <p className="font-mono text-[12px] font-medium uppercase tracking-[0.12em] text-neutral-500">
              Active Transfers
            </p>
            {activeTransfer ? (
              <div className="space-y-4">
                <div>
                  <div className="mb-1.5 flex items-end justify-between gap-2">
                    <div className="min-w-0 flex-1">
                      <span className="block truncate font-mono text-sm text-neutral-700">
                        {activeTransfer.filename}
                      </span>
                    </div>
                    <span className="shrink-0 font-mono text-[12px] text-blue-600">
                      {isUploading ? "Uploading" : "Queued"}
                    </span>
                  </div>
                  <div className="mb-1.5 flex items-center justify-between font-mono text-[12px] font-medium text-neutral-500">
                    <span>
                      {formatBytes(activeTransfer.uploaded)} / {formatBytes(activeTransfer.total)}
                    </span>
                    <span>{progressPct}%</span>
                  </div>
                  <div className="h-1 bg-neutral-200">
                    <div className="h-full bg-blue-600 transition-all" style={{ width: `${progressPct}%` }} />
                  </div>
                </div>
              </div>
            ) : (
              <p className="py-10 text-center font-mono text-sm text-neutral-400">
                No active transfers
              </p>
            )}

            <button
              onClick={startUpload}
              disabled={!selectedFile || isUploading}
              className={cn(
                "mt-auto w-full border px-4 py-2.5 font-mono text-[12px] transition-colors",
                !selectedFile || isUploading
                  ? "cursor-not-allowed border-neutral-200 text-neutral-300"
                  : "bg-blue-600 text-white hover:bg-blue-700 border-blue-600"
              )}
            >
              {isUploading ? "Uploading..." : "Start All Uploads"}
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
