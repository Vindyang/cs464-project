"use client";

import { useEffect, useId, useRef, useState } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { ProviderMetadata, getProviders, disconnectGDrive, getGDriveAuthorizeURL, connectS3, disconnectS3, getOneDriveAuthorizeURL, disconnectOneDrive } from "@/lib/api/providers";
import { cn } from "@/lib/utils";
import { toast } from "sonner";
import { helpToast } from "@/lib/help/help-toast";
import { ProvidersUploadFilesModal } from "@/components/dashboard/ProvidersUploadFilesModal";

const CONNECT_OPTIONS = [
  { id: "googleDrive", name: "Google Drive", description: "Connect your Google Drive account.", available: true },
  { id: "awsS3", name: "Amazon S3", description: "Connect using Access Key credentials.", available: true },
  { id: "oneDrive", name: "OneDrive", description: "Connect your Microsoft OneDrive account.", available: true },
];

interface ProvidersClientProps {
  initialProviders: ProviderMetadata[];
  initialConfiguredCredentialProviders: string[];
}

const PROVIDER_LABELS: Record<string, string> = {
  googleDrive: "Google Drive",
  awsS3: "Amazon S3",
  oneDrive: "OneDrive",
};

const PROVIDER_CONNECT_ERRORS: Record<string, string> = {
  credentials_missing: "The selected provider is missing local credentials. Add them on the Credentials page before connecting.",
  oauth_failed: "The provider rejected the sign-in request. Try again and confirm the OAuth app settings are correct.",
  save_failed: "The provider connected, but the local token could not be saved. Check backend storage and try again.",
  adapter_failed: "The provider authorization completed, but the backend could not finish the connection. Check the provider and adapter logs.",
};

export function ProvidersClient({
  initialProviders,
  initialConfiguredCredentialProviders,
}: ProvidersClientProps) {
  const router = useRouter();
  const searchParams = useSearchParams();
  const [providers, setProviders] = useState(initialProviders);
  const [configuredCredentialProviders] = useState(initialConfiguredCredentialProviders);
  const [search, setSearch] = useState("");
  const [connectModalOpen, setConnectModalOpen] = useState(false);
  const [uploadModalOpen, setUploadModalOpen] = useState(false);
  const [connecting, setConnecting] = useState<string | null>(null);
  const providerSearchInputId = useId();

  const refresh = async () => {
    const fresh = await getProviders().catch(() => providers);
    setProviders(fresh);
  };

  useEffect(() => {
    const connected = searchParams.get("connected");
    const error = searchParams.get("error");
    const upload = searchParams.get("upload");
    if (connected === "googleDrive") {
      toast.success("Google Drive connected successfully");
      router.replace("/providers");
      refresh();
    } else if (connected === "oneDrive") {
      toast.success("OneDrive connected successfully");
      router.replace("/providers");
      refresh();
    } else if (error) {
      helpToast({
        error: PROVIDER_CONNECT_ERRORS[error] ?? "Failed to connect provider. Please try again.",
        code: error === "credentials_missing" ? "PROVIDER_NOT_FOUND" : "PROVIDER_AUTH_EXPIRED",
      });
      router.replace("/providers");
    } else if (upload === "1") {
      setUploadModalOpen(true);
      router.replace("/providers");
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const filtered = providers.filter((p) =>
    p.displayName.toLowerCase().includes(search.toLowerCase())
  );

  const handleConnect = async (providerId: string) => {
    if (!configuredCredentialProviders.includes(providerId)) {
      helpToast({
        error: `Missing stored credentials for ${PROVIDER_LABELS[providerId] ?? providerId}. Add them on the Credentials page before connecting.`,
        code: "PROVIDER_NOT_FOUND",
      });
      return;
    }

    setConnecting(providerId);
    if (providerId === "googleDrive") {
      try {
        const { authURL } = await getGDriveAuthorizeURL();
        window.location.assign(authURL);
      } catch {
        helpToast({ error: "Failed to start Google Drive connection", code: "PROVIDER_AUTH_EXPIRED" });
        setConnecting(null);
      }
    } else if (providerId === "awsS3") {
      try {
        await connectS3();
        toast.success("AWS S3 connected successfully");
        refresh();
      } catch (err) {
        helpToast(err instanceof Error ? { error: err.message, code: "PROVIDER_AUTH_EXPIRED" } : { error: "Failed to connect S3", code: "PROVIDER_AUTH_EXPIRED" });
      } finally {
        setConnecting(null);
      }
    } else if (providerId === "oneDrive") {
      try {
        const { authURL } = await getOneDriveAuthorizeURL();
        window.location.assign(authURL);
      } catch {
        helpToast({ error: "Failed to start OneDrive connection", code: "PROVIDER_AUTH_EXPIRED" });
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
            else if (providerId === "oneDrive") await disconnectOneDrive();
            toast.success("Provider disconnected");
            refresh();
          } catch {
            helpToast({ error: "Failed to disconnect provider", code: "UNKNOWN_ERROR" });
          }
        },
      },
    });
  };

  return (
    <div className="space-y-6">
      <div className="flex flex-wrap items-center gap-3.5">
        <label htmlFor={providerSearchInputId} className="sr-only">
          Search providers
        </label>
        <input
          id={providerSearchInputId}
          type="text"
          placeholder="Search providers..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          className="w-full max-w-xs border border-neutral-200 bg-white px-3.5 py-2.5 font-mono text-sm text-neutral-900 outline-none placeholder:text-neutral-400 focus:border-sky-500 focus:ring-1 focus:ring-sky-500 dark:border-neutral-800 dark:bg-neutral-950 dark:text-neutral-100 dark:placeholder:text-neutral-500 dark:focus:border-sky-500 dark:focus:ring-sky-500"
        />
        <span className="font-mono text-[12px] font-medium text-neutral-500 dark:text-neutral-400">
          {filtered.length} {filtered.length === 1 ? "provider" : "providers"}
        </span>
        <button
          onClick={() => setUploadModalOpen(true)}
          className="ml-auto border border-neutral-900 px-4 py-2.5 font-mono text-[12px] uppercase tracking-wider text-neutral-900 transition-colors hover:bg-neutral-900 hover:text-white dark:border-neutral-700 dark:text-neutral-100 dark:hover:bg-neutral-100 dark:hover:text-neutral-950"
        >
          Upload File
        </button>
        <button
          onClick={() => setConnectModalOpen(true)}
          className="border border-neutral-900 px-4 py-2.5 font-mono text-[12px] uppercase tracking-wider text-neutral-900 transition-colors hover:bg-neutral-900 hover:text-white dark:border-neutral-700 dark:text-neutral-100 dark:hover:bg-neutral-100 dark:hover:text-neutral-950"
        >
          + Connect
        </button>
      </div>

      {providers.length === 0 ? (
        <EmptyState onConnect={() => setConnectModalOpen(true)} />
      ) : filtered.length === 0 ? (
        <div className="border border-neutral-200 px-4 py-10 text-center dark:border-neutral-800 dark:bg-neutral-950">
          <p className="font-mono text-sm font-medium text-neutral-500 dark:text-neutral-400">No providers match &ldquo;{search}&rdquo;</p>
        </div>
      ) : (
        <div className="relative border border-neutral-200 dark:border-neutral-800 dark:bg-neutral-950">
          <span className="pointer-events-none absolute -left-px -top-px h-1.5 w-1.5 border-l border-t border-neutral-400 opacity-40 dark:border-neutral-700" />
          <span className="pointer-events-none absolute -right-px -top-px h-1.5 w-1.5 border-r border-t border-neutral-400 opacity-40 dark:border-neutral-700" />
          <span className="pointer-events-none absolute -bottom-px -left-px h-1.5 w-1.5 border-b border-l border-neutral-400 opacity-40 dark:border-neutral-700" />
          <span className="pointer-events-none absolute -bottom-px -right-px h-1.5 w-1.5 border-b border-r border-neutral-400 opacity-40 dark:border-neutral-700" />
          <div className="overflow-x-auto">
            <div className="min-w-[740px]">
              <div className="grid grid-cols-[1fr_96px_220px_96px_72px] gap-4 border-b border-neutral-200 bg-neutral-50 px-4 py-2.5 dark:border-neutral-800 dark:bg-neutral-900/70">
                {["Name", "Status", "Storage", "Latency", ""].map((h) => (
                  <span key={h} className="font-mono text-[11px] font-medium uppercase tracking-widest text-neutral-500 dark:text-neutral-400">
                    {h}
                  </span>
                ))}
              </div>
              <div className="divide-y divide-neutral-200 dark:divide-neutral-800">
                {filtered.map((p) => {
                  const isOnline = p.status === "connected" || p.status === "online";
                  const usedPct = p.quotaTotalBytes > 0
                    ? Math.round((p.quotaUsedBytes / p.quotaTotalBytes) * 100)
                    : 0;
                  return (
                    <div
                      key={p.providerId}
                      className="grid grid-cols-[1fr_96px_220px_96px_72px] gap-4 px-4 py-3.5 transition-colors hover:bg-neutral-50 dark:hover:bg-neutral-900/70"
                    >
                      <span className="truncate font-mono text-sm font-medium text-neutral-900 dark:text-neutral-100">{p.displayName}</span>
                      <span
                        className={cn(
                          "self-center border px-2 py-0.5 text-center font-mono text-[11px] font-medium uppercase tracking-wider",
                          isOnline
                            ? "border-emerald-300 text-emerald-700 dark:border-emerald-800 dark:text-emerald-300"
                            : "border-neutral-300 text-neutral-500 dark:border-neutral-700 dark:text-neutral-400"
                        )}
                      >
                        {isOnline ? "Online" : p.status}
                      </span>
                      <div className="flex items-center gap-2.5">
                        <div className="h-0.5 flex-1 bg-neutral-200 dark:bg-neutral-800">
                          <div className="h-full bg-neutral-900 dark:bg-sky-400" style={{ width: `${usedPct}%` }} />
                        </div>
                        <span className="w-9 shrink-0 text-right font-mono text-[11px] font-medium text-neutral-500 dark:text-neutral-400">
                          {usedPct}%
                        </span>
                      </div>
                      <span className="self-center font-mono text-[12px] text-neutral-500 dark:text-neutral-400">
                        {p.latencyMs != null ? `${p.latencyMs}ms` : "—"}
                      </span>
                      <button
                        onClick={() => handleDisconnect(p.providerId)}
                        className="self-center text-right font-mono text-[12px] text-neutral-300 transition-colors hover:text-neutral-950 dark:text-neutral-600 dark:hover:text-white"
                        title="Disconnect"
                        aria-label={`Disconnect ${p.displayName}`}
                      >
                        ×
                      </button>
                    </div>
                  );
                })}
              </div>
            </div>
          </div>
        </div>
      )}

      <ProvidersUploadFilesModal
        open={uploadModalOpen}
        onOpenChange={setUploadModalOpen}
        onUploadSuccess={() => {
          router.refresh()
        }}
      />

      {connectModalOpen && (
        <ConnectModal
          connectedProviderIds={new Set(providers.map((p) => p.providerId))}
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
    <div className="border border-neutral-200 px-4 py-16 text-center dark:border-neutral-800 dark:bg-neutral-950">
      <p className="mb-1 font-mono text-sm font-medium text-neutral-500 dark:text-neutral-400">No providers connected.</p>
      <button
        onClick={onConnect}
        className="mt-2 font-mono text-[12px] uppercase tracking-wider text-neutral-500 underline transition-colors hover:text-black dark:text-neutral-400 dark:hover:text-white"
      >
        Connect a provider
      </button>
    </div>
  );
}

function ConnectModal({
  connectedProviderIds,
  connecting,
  onConnect,
  onClose,
}: {
  connectedProviderIds: Set<string>;
  connecting: string | null;
  onConnect: (id: string) => void;
  onClose: () => void;
}) {
  const connectDialogTitleId = useId();
  const closeButtonRef = useRef<HTMLButtonElement | null>(null);

  useEffect(() => {
    closeButtonRef.current?.focus();
    const onKeyDown = (event: KeyboardEvent) => {
      if (event.key === "Escape") onClose();
    };
    window.addEventListener("keydown", onKeyDown);
    return () => window.removeEventListener("keydown", onKeyDown);
  }, [onClose]);

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4 backdrop-blur-sm"
      onClick={onClose}
    >
      <div
        className="w-full max-w-sm border border-neutral-200 bg-white shadow-2xl dark:border-neutral-800 dark:bg-neutral-950"
        onClick={(e) => e.stopPropagation()}
        role="dialog"
        aria-modal="true"
        aria-labelledby={connectDialogTitleId}
      >
        <div className="flex items-center justify-between border-b border-neutral-200 px-5 py-4 dark:border-neutral-800">
          <h2 id={connectDialogTitleId} className="font-mono text-[12px] font-medium uppercase tracking-widest text-neutral-900 dark:text-neutral-100">
            Connect a Provider
          </h2>
          <button
            onClick={onClose}
            className="font-mono text-sm text-neutral-400 transition-colors hover:text-black dark:text-neutral-500 dark:hover:text-white"
            aria-label="Close connect provider modal"
            ref={closeButtonRef}
          >
            ×
          </button>
        </div>
        <div className="divide-y divide-neutral-200 dark:divide-neutral-800">
          {CONNECT_OPTIONS.map((p) => (
            <div key={p.id} className="flex items-center justify-between gap-4 px-5 py-4">
              {(() => {
                const isConnected = connectedProviderIds.has(p.id);
                const isDisabled = !p.available || connecting === p.id || isConnected;
                const label = isConnected
                  ? "Connected"
                  : connecting === p.id
                    ? "Connecting..."
                    : "Connect";
                return (
                  <>
                    <div>
                      <p className={cn("font-mono text-sm font-medium text-neutral-900 dark:text-neutral-100", !p.available && "text-neutral-500 dark:text-neutral-500")}>
                        {p.name}
                      </p>
                      <p className="mt-1 font-mono text-[12px] font-medium text-neutral-500 dark:text-neutral-400">{p.description}</p>
                    </div>
                    <button
                      disabled={isDisabled}
                      onClick={() => onConnect(p.id)}
                      className={cn(
                        "ml-4 shrink-0 border px-3.5 py-2 font-mono text-[12px] uppercase tracking-wider transition-colors",
                        isConnected
                          ? "border-neutral-300 bg-neutral-100 text-neutral-500 dark:border-neutral-700 dark:bg-neutral-900 dark:text-neutral-400"
                          : p.available && connecting !== p.id
                            ? "border-neutral-900 text-neutral-900 hover:bg-neutral-900 hover:text-white dark:border-neutral-700 dark:text-neutral-100 dark:hover:bg-neutral-100 dark:hover:text-neutral-950"
                            : "cursor-not-allowed border-neutral-200 text-neutral-300 dark:border-neutral-800 dark:text-neutral-600"
                      )}
                    >
                      {label}
                    </button>
                  </>
                );
              })()}
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
