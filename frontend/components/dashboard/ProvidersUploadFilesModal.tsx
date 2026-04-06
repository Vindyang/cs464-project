"use client";

import { useEffect, useId, useRef, useState } from "react";
import { Upload, X } from "lucide-react";
import { toast } from "sonner";
import { cn } from "@/lib/utils";
import { FileMetadata, getFiles, uploadFile } from "@/lib/api/files";
import { getSettings, parseRedundancyPreset, REDUNDANCY_PRESETS, RedundancyPreset } from "@/lib/api/settings";

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
  const [defaultRedundancy, setDefaultRedundancy] = useState<RedundancyPreset>("(6,4)")
  const [selectedRedundancy, setSelectedRedundancy] = useState<RedundancyPreset>("(6,4)")
  const [loadingSettings, setLoadingSettings] = useState(false)

  const refreshUploadHistory = async () => {
    const files = await getFiles().catch(() => [] as FileMetadata[]);
    const next = [...files]
      .sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime())
      .slice(0, 8)
      .map(toHistoryItem);
    setHistory(next);
  };

  const loadUploadDefaults = async () => {
    setLoadingSettings(true)
    try {
      const settings = await getSettings()
      setDefaultRedundancy(settings.redundancy)
      setSelectedRedundancy(settings.redundancy)
    } catch {
      setDefaultRedundancy("(6,4)")
      setSelectedRedundancy("(6,4)")
    } finally {
      setLoadingSettings(false)
    }
  }

  useEffect(() => {
    return () => {
      if (progressTimer.current) clearInterval(progressTimer.current);
    };
  }, []);

  useEffect(() => {
    if (!open) return;
    closeButtonRef.current?.focus();
    refreshUploadHistory();
    loadUploadDefaults();
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

    const { k, n } = parseRedundancyPreset(selectedRedundancy)

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
      await uploadFile(selectedFile, k, n);
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
    <div className="fixed inset-0 z-50 flex items-start justify-center overflow-y-auto bg-black/45 p-3 backdrop-blur-sm sm:items-center sm:p-4" onClick={() => onOpenChange(false)}>
      <div
        className="flex max-h-[calc(100vh-1.5rem)] w-full max-w-5xl flex-col overflow-hidden border border-neutral-200 bg-white shadow-2xl dark:border-neutral-800 dark:bg-neutral-950 sm:max-h-[calc(100vh-2rem)]"
        onClick={(e) => e.stopPropagation()}
        role="dialog"
        aria-modal="true"
        aria-labelledby={uploadDialogTitleId}
        aria-describedby={uploadDialogDescriptionId}
      >
        <div className="flex items-center justify-between border-b border-neutral-200 px-4 py-4 sm:px-6 dark:border-neutral-800">
          <h2 id={uploadDialogTitleId} className="font-mono text-[12px] font-medium uppercase tracking-[0.12em] text-neutral-600 dark:text-neutral-300">
            Providers / Upload Files
          </h2>
          <button
            onClick={() => onOpenChange(false)}
            className="text-neutral-400 transition-colors hover:text-black dark:text-neutral-500 dark:hover:text-white"
            aria-label="Close upload modal"
            ref={closeButtonRef}
          >
            <X className="h-4 w-4" />
          </button>
        </div>

        <div className="grid min-h-0 gap-4 overflow-y-auto p-4 sm:p-6 xl:grid-cols-[minmax(0,1.35fr)_minmax(300px,0.9fr)]">
          <div className="min-w-0 space-y-4">
            <div className="border border-neutral-200 p-4 dark:border-neutral-800 dark:bg-neutral-950 sm:p-5">
              <div className="flex flex-wrap items-start justify-between gap-3">
                <div>
                  <p className="mb-2 font-mono text-[12px] font-medium uppercase tracking-[0.12em] text-neutral-500 dark:text-neutral-400">
                    Upload Configuration
                  </p>
                  <p id={uploadDialogDescriptionId} className="font-mono text-[12px] font-medium leading-relaxed text-neutral-600 dark:text-neutral-300">
                    Files are sharded before upload. The Reed-Solomon preset defaults from Settings, but you can override it for this upload.
                  </p>
                </div>
                <div className="min-w-[220px] max-w-full space-y-2">
                  <label className="block font-mono text-[11px] uppercase tracking-[0.1em] text-neutral-500 dark:text-neutral-400">
                    Reed-Solomon Preset
                  </label>
                  <select
                    value={selectedRedundancy}
                    onChange={(event) => setSelectedRedundancy(event.target.value as RedundancyPreset)}
                    className="w-full border border-neutral-200 bg-white px-3 py-2 font-mono text-sm text-neutral-900 outline-none focus:border-sky-500 focus:ring-1 focus:ring-sky-500 dark:border-neutral-800 dark:bg-neutral-950 dark:text-neutral-100"
                  >
                    {REDUNDANCY_PRESETS.map((preset) => (
                      <option key={preset.val} value={preset.val}>
                        {preset.label} · {preset.name}
                      </option>
                    ))}
                  </select>
                  <div className="font-mono text-[11px] text-neutral-500 dark:text-neutral-400">
                    {loadingSettings ? "Loading defaults..." : `Default from settings: ${defaultRedundancy}`}
                  </div>
                  <div className="font-mono text-[11px] text-neutral-500 dark:text-neutral-400">
                    {(() => {
                      const { k, n } = parseRedundancyPreset(selectedRedundancy)
                      return `${k} data shards, ${n - k} parity shards, ${n} total shards`
                    })()}
                  </div>
                </div>
              </div>

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
                  "mt-4 grid min-h-[180px] place-items-center border border-dashed px-4 py-8 text-center transition-colors sm:min-h-[220px]",
                  dragActive
                    ? "border-sky-500 bg-sky-50 dark:bg-sky-950/20"
                    : "border-neutral-300 dark:border-neutral-700",
                )}
              >
                <div className="space-y-3">
                  <div className="mx-auto grid h-10 w-10 place-items-center border border-neutral-300 text-neutral-500 dark:border-neutral-700 dark:text-neutral-400">
                    <Upload className="h-5 w-5" />
                  </div>
                  <div>
                    <p className="max-w-full truncate font-mono text-sm text-neutral-700 dark:text-neutral-200">
                      {selectedFile ? selectedFile.name : "Click or drag a file to upload"}
                    </p>
                    <p className="mt-1 font-mono text-[12px] font-medium text-neutral-500 dark:text-neutral-400">
                      {selectedFile ? formatBytes(selectedFile.size) : "Responsive modal, tuned for smaller laptop screens"}
                    </p>
                  </div>
                </div>
              </div>
            </div>

            <div className="border border-neutral-200 p-4 dark:border-neutral-800 dark:bg-neutral-950 sm:p-5">
              <p className="mb-3 font-mono text-[12px] font-medium uppercase tracking-[0.12em] text-neutral-500 dark:text-neutral-400">
                Recent Upload History
              </p>
              <div className="overflow-x-auto">
                <div className="min-w-[520px]">
                  <div className="grid grid-cols-[1.6fr_0.9fr_0.9fr_1fr_0.8fr] gap-3 border-b border-neutral-200 pb-2.5 font-mono text-[11px] font-medium uppercase tracking-[0.08em] text-neutral-500 dark:border-neutral-800 dark:text-neutral-400">
                    <span>Filename</span>
                    <span>Size</span>
                    <span>Provider</span>
                    <span>Date</span>
                    <span>Status</span>
                  </div>
                  <div className="space-y-1 pt-2.5">
                    {history.length === 0 ? (
                      <p className="py-6 text-center font-mono text-sm text-neutral-400 dark:text-neutral-500">
                        No uploads found yet.
                      </p>
                    ) : (
                      history.slice(0, 5).map((item) => (
                        <div
                          key={item.id}
                          className="grid grid-cols-[1.6fr_0.9fr_0.9fr_1fr_0.8fr] gap-3 py-2 font-mono text-sm text-neutral-700 dark:text-neutral-200"
                        >
                          <span className="block min-w-0 truncate">{item.filename}</span>
                          <span className="text-neutral-600 dark:text-neutral-400">{item.sizeLabel}</span>
                          <span className="text-neutral-600 dark:text-neutral-400">{item.provider}</span>
                          <span className="text-neutral-500 dark:text-neutral-400">{item.date}</span>
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

          <div className="flex min-w-0 flex-col gap-4 border border-neutral-200 p-4 dark:border-neutral-800 dark:bg-neutral-950 sm:p-5">
            <p className="font-mono text-[12px] font-medium uppercase tracking-[0.12em] text-neutral-500 dark:text-neutral-400">
              Active Transfers
            </p>
            {activeTransfer ? (
              <div className="space-y-4">
                <div>
                  <div className="mb-1.5 flex items-end justify-between gap-2">
                    <div className="min-w-0 flex-1">
                      <span className="block truncate font-mono text-sm text-neutral-700 dark:text-neutral-200">
                        {activeTransfer.filename}
                      </span>
                    </div>
                    <span className="shrink-0 font-mono text-[12px] text-blue-600">
                      {isUploading ? "Uploading" : "Queued"}
                    </span>
                  </div>
                  <div className="mb-1.5 flex items-center justify-between font-mono text-[12px] font-medium text-neutral-500 dark:text-neutral-400">
                    <span>
                      {formatBytes(activeTransfer.uploaded)} / {formatBytes(activeTransfer.total)}
                    </span>
                    <span>{progressPct}%</span>
                  </div>
                  <div className="h-1 bg-neutral-200 dark:bg-neutral-800">
                    <div className="h-full bg-blue-600 transition-all" style={{ width: `${progressPct}%` }} />
                  </div>
                </div>
              </div>
            ) : (
              <p className="py-10 text-center font-mono text-sm text-neutral-400 dark:text-neutral-500">
                No active transfers
              </p>
            )}

            <button
              onClick={startUpload}
              disabled={!selectedFile || isUploading}
              className={cn(
                "mt-auto w-full border px-4 py-2.5 font-mono text-[12px] transition-colors",
                !selectedFile || isUploading
                  ? "cursor-not-allowed border-neutral-200 text-neutral-300 dark:border-neutral-800 dark:text-neutral-600"
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
