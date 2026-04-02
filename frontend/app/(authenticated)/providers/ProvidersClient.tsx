"use client";

import { useEffect, useRef, useState } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { ProviderMetadata, getProviders, disconnectGDrive, getGDriveAuthorizeURL, connectS3, disconnectS3 } from "@/lib/api/providers";
import { FileMetadata, getFiles, uploadFile } from "@/lib/api/files";
import { cn } from "@/lib/utils";
import { toast } from "sonner";
import { Upload, X } from "lucide-react";

const CONNECT_OPTIONS = [
  { id: "googleDrive", name: "Google Drive", description: "Connect your Google Drive account.", available: true },
  { id: "awsS3", name: "Amazon S3", description: "Connect using Access Key credentials.", available: true },
  { id: "oneDrive", name: "OneDrive", description: "Coming soon.", available: false },
];

interface ProvidersClientProps {
  initialProviders: ProviderMetadata[];
}

interface UploadHistoryItem {
  id: string;
  filename: string;
  sizeLabel: string;
  provider: string;
  date: string;
  status: "Success" | "Failed";
}

function formatBytes(value: number) {
  if (value <= 0) return "0 B";
  const units = ["B", "KB", "MB", "GB", "TB"];
  const idx = Math.min(Math.floor(Math.log(value) / Math.log(1024)), units.length - 1);
  const scaled = value / 1024 ** idx;
  const digits = idx === 0 ? 0 : scaled >= 10 ? 1 : 2;
  return `${scaled.toFixed(digits)} ${units[idx]}`;
}

function toHistoryItem(file: FileMetadata): UploadHistoryItem {
  return {
    id: file.file_id,
    filename: file.original_name,
    sizeLabel: formatBytes(file.original_size ?? 0),
    provider: "Distributed",
    date: new Date(file.created_at).toLocaleString("en-US", {
      year: "numeric",
      month: "2-digit",
      day: "2-digit",
      hour: "2-digit",
      minute: "2-digit",
    }),
    status: file.status === "FAILED" ? "Failed" : "Success",
  };
}

export function ProvidersClient({ initialProviders }: ProvidersClientProps) {
  const router = useRouter();
  const searchParams = useSearchParams();
  const [providers, setProviders] = useState(initialProviders);
  const [search, setSearch] = useState("");
  const [connectModalOpen, setConnectModalOpen] = useState(false);
  const [uploadModalOpen, setUploadModalOpen] = useState(false);
  const [connecting, setConnecting] = useState<string | null>(null);
  const [uploadHistory, setUploadHistory] = useState<UploadHistoryItem[]>([]);

  const refreshUploadHistory = async () => {
    const files = await getFiles().catch(() => [] as FileMetadata[]);
    const history = [...files]
      .sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime())
      .slice(0, 8)
      .map(toHistoryItem);
    setUploadHistory(history);
  };

  const refresh = async () => {
    const fresh = await getProviders().catch(() => providers);
    setProviders(fresh);
  };

  useEffect(() => {
    const connected = searchParams.get("connected");
    const error = searchParams.get("error");
    if (connected === "googleDrive") {
      toast.success("Google Drive connected successfully");
      router.replace("/providers");
      refresh();
    } else if (error) {
      toast.error("Failed to connect provider. Please try again.");
      router.replace("/providers");
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  useEffect(() => {
    if (!uploadModalOpen) return;
    refreshUploadHistory();
  }, [uploadModalOpen]);

  const filtered = providers.filter((p) =>
    p.displayName.toLowerCase().includes(search.toLowerCase())
  );

  const handleConnect = async (providerId: string) => {
    setConnecting(providerId);
    if (providerId === "googleDrive") {
      try {
        const { authURL } = await getGDriveAuthorizeURL();
        window.location.assign(authURL);
      } catch {
        toast.error("Failed to start Google Drive connection");
        setConnecting(null);
      }
    } else if (providerId === "awsS3") {
      try {
        await connectS3();
        toast.success("AWS S3 connected successfully");
        refresh();
      } catch (err) {
        toast.error(err instanceof Error ? err.message : "Failed to connect S3");
      } finally {
        setConnecting(null);
      }
    } else {
      setConnecting(null);
    }
  };

  const handleDisconnect = (providerId: string) => {
    toast("Disconnect provider?", {
      description: "This will stop shards from being stored on this provider.",
      action: {
        label: "Disconnect",
        onClick: async () => {
          try {
            if (providerId === "googleDrive") await disconnectGDrive();
            else if (providerId === "awsS3") await disconnectS3();
            toast.success("Provider disconnected");
            refresh();
          } catch {
            toast.error("Failed to disconnect provider");
          }
        },
      },
    });
  };

  return (
    <div className="space-y-5">
      {/* Toolbar */}
      <div className="flex items-center gap-3">
        <input
          type="text"
          placeholder="Search providers..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          className="font-mono text-xs border bg-white px-3 py-2 w-64 outline-none focus:ring-1 focus:ring-black placeholder:text-neutral-400"
        />
        <span className="font-mono text-[10px] text-neutral-400">
          {filtered.length} {filtered.length === 1 ? "provider" : "providers"}
        </span>
        <button
          onClick={() => setUploadModalOpen(true)}
          className="ml-auto font-mono text-[10px] uppercase tracking-wider border px-3 py-2 transition-colors hover:bg-black hover:text-white"
        >
          Upload File
        </button>
        <button
          onClick={() => setConnectModalOpen(true)}
          className="font-mono text-[10px] uppercase tracking-wider border px-3 py-2 hover:bg-black hover:text-white transition-colors"
        >
          + Connect
        </button>
      </div>

      {/* Table */}
      {providers.length === 0 ? (
        <EmptyState onConnect={() => setConnectModalOpen(true)} />
      ) : filtered.length === 0 ? (
        <div className="border px-4 py-10 text-center">
          <p className="font-mono text-xs text-neutral-400">No providers match &ldquo;{search}&rdquo;</p>
        </div>
      ) : (
        <div className="border relative">
          <span className="absolute -top-px -left-px w-1.5 h-1.5 border-t border-l border-neutral-400 opacity-40 pointer-events-none" />
          <span className="absolute -top-px -right-px w-1.5 h-1.5 border-t border-r border-neutral-400 opacity-40 pointer-events-none" />
          <span className="absolute -bottom-px -left-px w-1.5 h-1.5 border-b border-l border-neutral-400 opacity-40 pointer-events-none" />
          <span className="absolute -bottom-px -right-px w-1.5 h-1.5 border-b border-r border-neutral-400 opacity-40 pointer-events-none" />
          <div className="grid grid-cols-[1fr_80px_180px_80px_60px] gap-4 px-4 py-2 border-b bg-neutral-50">
            {["Name", "Status", "Storage", "Latency", ""].map((h) => (
              <span key={h} className="font-mono text-[9px] uppercase tracking-widest text-neutral-400">
                {h}
              </span>
            ))}
          </div>
          <div className="divide-y">
            {filtered.map((p) => {
              const isOnline = p.status === "connected" || p.status === "online";
              const usedPct = p.quotaTotalBytes > 0
                ? Math.round((p.quotaUsedBytes / p.quotaTotalBytes) * 100)
                : 0;
              return (
                <div
                  key={p.providerId}
                  className="grid grid-cols-[1fr_80px_180px_80px_60px] gap-4 px-4 py-3 hover:bg-neutral-50 transition-colors"
                >
                  <span className="font-mono text-xs truncate">{p.displayName}</span>
                  <span
                    className={cn(
                      "font-mono text-[9px] uppercase tracking-wider px-2 py-0.5 border self-center text-center",
                      isOnline
                        ? "border-black text-black"
                        : "border-neutral-300 text-neutral-400"
                    )}
                  >
                    {isOnline ? "Online" : p.status}
                  </span>
                  <div className="flex items-center gap-2">
                    <div className="flex-1 h-0.5 bg-neutral-200">
                      <div className="h-full bg-black" style={{ width: `${usedPct}%` }} />
                    </div>
                    <span className="font-mono text-[9px] text-neutral-400 w-8 text-right shrink-0">
                      {usedPct}%
                    </span>
                  </div>
                  <span className="font-mono text-[11px] text-neutral-500 self-center">
                    {p.latencyMs}ms
                  </span>
                  <button
                    onClick={() => handleDisconnect(p.providerId)}
                    className="font-mono text-[10px] text-neutral-300 hover:text-black transition-colors self-center text-right"
                    title="Disconnect"
                  >
                    ×
                  </button>
                </div>
              );
            })}
          </div>
        </div>
      )}

      {uploadModalOpen && (
        <UploadFilesModal
          history={uploadHistory}
          onRefreshHistory={refreshUploadHistory}
          onClose={() => setUploadModalOpen(false)}
          onUploadSuccess={(filename) => {
            toast.success(`Uploaded ${filename}`);
          }}
        />
      )}

      {/* Connect Modal */}
      {connectModalOpen && (
        <ConnectModal
          connecting={connecting}
          onConnect={handleConnect}
          onClose={() => setConnectModalOpen(false)}
        />
      )}
    </div>
  );
}

function EmptyState({ onConnect }: { onConnect: () => void }) {
  return (
    <div className="border px-4 py-16 text-center">
      <p className="font-mono text-xs text-neutral-400 mb-1">No providers connected.</p>
      <button
        onClick={onConnect}
        className="mt-2 font-mono text-[10px] uppercase tracking-wider underline text-neutral-500 hover:text-black transition-colors"
      >
        Connect a provider
      </button>
    </div>
  );
}

function UploadFilesModal({
  history,
  onRefreshHistory,
  onClose,
  onUploadSuccess,
}: {
  history: UploadHistoryItem[];
  onRefreshHistory: () => Promise<void>;
  onClose: () => void;
  onUploadSuccess: (filename: string) => void;
}) {
  const fileInputRef = useRef<HTMLInputElement | null>(null);
  const progressTimer = useRef<ReturnType<typeof setInterval> | null>(null);
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [dragActive, setDragActive] = useState(false);
  const [isUploading, setIsUploading] = useState(false);
  const [uploadedBytes, setUploadedBytes] = useState(0);

  useEffect(() => {
    return () => {
      if (progressTimer.current) clearInterval(progressTimer.current);
    };
  }, []);

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
      await onRefreshHistory();
      onUploadSuccess(selectedFile.name);
      setSelectedFile(null);
      setTimeout(() => onClose(), 300);
    } catch (err) {
      const message = err instanceof Error ? err.message : "Failed to upload file";
      toast.error(message);
    } finally {
      setIsUploading(false);
      if (progressTimer.current) clearInterval(progressTimer.current);
      progressTimer.current = null;
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/20 p-4" onClick={onClose}>
      <div
        className="w-full max-w-5xl border bg-white"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="flex items-center justify-between border-b px-5 py-4">
          <span className="font-mono text-[10px] uppercase tracking-[0.12em] text-neutral-500">
            Providers / Upload Files
          </span>
          <button
            onClick={onClose}
            className="text-neutral-400 transition-colors hover:text-black"
            aria-label="Close upload modal"
          >
            <X className="h-4 w-4" />
          </button>
        </div>

        <div className="grid gap-4 p-5 lg:grid-cols-[1.6fr_1fr]">
          <div className="space-y-4">
            <div className="border p-4">
              <p className="mb-3 font-mono text-[10px] uppercase tracking-[0.12em] text-neutral-400">
                Upload Configuration
              </p>
              <p className="font-mono text-[11px] text-neutral-500">
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
                  "mt-4 grid min-h-[270px] place-items-center border border-dashed px-4 py-10 text-center transition-colors",
                  dragActive ? "border-black bg-neutral-50" : "border-neutral-300"
                )}
              >
                <div className="space-y-3">
                  <div className="mx-auto grid h-10 w-10 place-items-center border border-neutral-300 text-neutral-500">
                    <Upload className="h-5 w-5" />
                  </div>
                  <div>
                    <p className="font-mono text-xs text-neutral-700">
                      {selectedFile ? selectedFile.name : "Click or drag files to upload"}
                    </p>
                    <p className="mt-1 font-mono text-[10px] text-neutral-400">
                      {selectedFile ? formatBytes(selectedFile.size) : "Max file size: 500GB"}
                    </p>
                  </div>
                </div>
              </div>
            </div>

            <div className="border p-4">
              <p className="mb-3 font-mono text-[10px] uppercase tracking-[0.12em] text-neutral-400">
                Recent Upload History
              </p>
              <div className="grid grid-cols-[1.6fr_0.9fr_0.9fr_1fr_0.8fr] gap-3 border-b pb-2 font-mono text-[9px] uppercase tracking-[0.08em] text-neutral-400">
                <span>Filename</span>
                <span>Size</span>
                <span>Provider</span>
                <span>Date</span>
                <span>Status</span>
              </div>
              <div className="space-y-1 pt-2">
                {history.length === 0 ? (
                  <p className="py-6 text-center font-mono text-xs text-neutral-400">
                    No uploads found yet.
                  </p>
                ) : (
                  history.slice(0, 3).map((item) => (
                    <div
                      key={item.id}
                      className="grid grid-cols-[1.6fr_0.9fr_0.9fr_1fr_0.8fr] gap-3 py-1.5 font-mono text-xs"
                    >
                      <span className="truncate">{item.filename}</span>
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

          <div className="space-y-4 border p-4">
            <p className="font-mono text-[10px] uppercase tracking-[0.12em] text-neutral-400">
              Active Transfers
            </p>
            {activeTransfer ? (
              <div className="space-y-4">
                <div>
                  <div className="mb-1 flex items-end justify-between gap-2">
                    <span className="truncate font-mono text-xs text-neutral-700">{activeTransfer.filename}</span>
                    <span className="font-mono text-[10px] text-blue-600">
                      {isUploading ? "Uploading" : "Queued"}
                    </span>
                  </div>
                  <div className="mb-1 flex items-center justify-between font-mono text-[10px] text-neutral-400">
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
              <p className="py-10 text-center font-mono text-xs text-neutral-400">
                No active transfers
              </p>
            )}

            <button
              onClick={startUpload}
              disabled={!selectedFile || isUploading}
              className={cn(
                "mt-auto w-full border px-3 py-2 font-mono text-[11px] transition-colors",
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

function ConnectModal({
  connecting,
  onConnect,
  onClose,
}: {
  connecting: string | null;
  onConnect: (id: string) => void;
  onClose: () => void;
}) {
  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/20"
      onClick={onClose}
    >
      <div
        className="bg-white border w-full max-w-sm mx-4"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="flex items-center justify-between px-5 py-4 border-b">
          <span className="font-mono text-[10px] uppercase tracking-widest">Connect a Provider</span>
          <button
            onClick={onClose}
            className="font-mono text-neutral-400 hover:text-black transition-colors text-sm"
          >
            ×
          </button>
        </div>
        <div className="divide-y">
          {CONNECT_OPTIONS.map((p) => (
            <div key={p.id} className="flex items-center justify-between px-5 py-4">
              <div>
                <p className={cn("font-mono text-xs", !p.available && "text-neutral-400")}>
                  {p.name}
                </p>
                <p className="font-mono text-[10px] text-neutral-400 mt-0.5">{p.description}</p>
              </div>
              <button
                disabled={!p.available || connecting === p.id}
                onClick={() => onConnect(p.id)}
                className={cn(
                  "font-mono text-[10px] uppercase tracking-wider border px-3 py-1.5 transition-colors ml-4 shrink-0",
                  p.available && connecting !== p.id
                    ? "hover:bg-black hover:text-white"
                    : "text-neutral-300 border-neutral-200 cursor-not-allowed"
                )}
              >
                {connecting === p.id ? "Connecting..." : "Connect"}
              </button>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
