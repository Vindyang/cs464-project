"use client";

import { useState, useEffect } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { ProviderMetadata, getProviders, disconnectGDrive, getGDriveAuthorizeURL } from "@/lib/api/providers";
import { cn } from "@/lib/utils";
import { toast } from "sonner";

const CONNECT_OPTIONS = [
  { id: "googleDrive", name: "Google Drive", description: "Connect your Google Drive account.", available: true },
  { id: "awsS3", name: "Amazon S3", description: "Coming soon.", available: false },
  { id: "oneDrive", name: "OneDrive", description: "Coming soon.", available: false },
];

interface ProvidersClientProps {
  initialProviders: ProviderMetadata[];
}

export function ProvidersClient({ initialProviders }: ProvidersClientProps) {
  const router = useRouter();
  const searchParams = useSearchParams();
  const [providers, setProviders] = useState(initialProviders);
  const [search, setSearch] = useState("");
  const [modalOpen, setModalOpen] = useState(false);
  const [connecting, setConnecting] = useState<string | null>(null);

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

  const filtered = providers.filter((p) =>
    p.displayName.toLowerCase().includes(search.toLowerCase())
  );

  const handleConnect = async (providerId: string) => {
    if (providerId !== "googleDrive") return;
    setConnecting(providerId);
    try {
      const { authURL } = await getGDriveAuthorizeURL();
      window.location.assign(authURL);
    } catch {
      toast.error("Failed to start Google Drive connection");
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
          onClick={() => setModalOpen(true)}
          className="ml-auto font-mono text-[10px] uppercase tracking-wider border px-3 py-2 hover:bg-black hover:text-white transition-colors"
        >
          + Connect
        </button>
      </div>

      {/* Table */}
      {providers.length === 0 ? (
        <EmptyState onConnect={() => setModalOpen(true)} />
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

      {/* Connect Modal */}
      {modalOpen && (
        <ConnectModal
          connecting={connecting}
          onConnect={handleConnect}
          onClose={() => setModalOpen(false)}
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
                {connecting === p.id ? "Redirecting..." : "Connect"}
              </button>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
