import { cn } from "@/lib/utils";
import { Cloud, Activity } from "lucide-react";
import { Button } from "@/components/ui/button";
import { useProviderStore } from "@/lib/store/providerStore";
import { useEffect } from "react";

const statusConfig = {
  online: {
    color: "#000000",
    label: "ONLINE",
    bg: "#ffffff",
  },
  connected: {
    color: "#000000",
    label: "ONLINE",
    bg: "#ffffff",
  },
  warning: {
    color: "#737373",
    label: "DEGRADED",
    bg: "#fafafa",
  },
  degraded: {
    color: "#737373",
    label: "DEGRADED",
    bg: "#fafafa",
  },
  offline: {
    color: "#d4d4d4",
    label: "OFFLINE",
    bg: "#f5f5f5",
  },
  disconnected: {
    color: "#d4d4d4",
    label: "OFFLINE",
    bg: "#f5f5f5",
  },
  error: {
    color: "#737373",
    label: "ERROR",
    bg: "#fafafa",
  },
};

const providerIcons: Record<string, string> = {
  googleDrive: "GD",
  awsS3: "S3",
  dropbox: "DB",
  backblazeB2: "B2",
  oneDrive: "OD",
};

export function ProviderMatrix() {
  const { providers, fetchProviders } = useProviderStore();

  useEffect(() => {
    fetchProviders();
  }, []);

  return (
    <div className="col-span-1 row-span-2 flex flex-col gap-6">
      <div className="mb-2 flex items-center justify-between">
        <div className="flex items-center gap-3">
          <Cloud className="h-5 w-5" />
          <h2 className="font-mono text-sm font-bold uppercase tracking-wider">
            Provider Matrix
          </h2>
        </div>
        <Button variant="outline" size="sm" className="font-mono text-xs">
          Manage
        </Button>
      </div>

      <div className="border bg-white">
        {/* Table Header */}
        <div className="grid grid-cols-[50px_2fr_120px_80px_100px_80px] items-center gap-4 border-b bg-neutral-50 px-4 py-3">
          <div className="font-mono text-[9px] uppercase tracking-wider text-neutral-500">
            #
          </div>
          <div className="font-mono text-[9px] uppercase tracking-wider text-neutral-500">
            Provider
          </div>
          <div className="font-mono text-[9px] uppercase tracking-wider text-neutral-500">
            Region
          </div>
          <div className="font-mono text-[9px] uppercase tracking-wider text-neutral-500">
            Latency
          </div>
          <div className="font-mono text-[9px] uppercase tracking-wider text-neutral-500">
            Usage
          </div>
          <div className="font-mono text-[9px] uppercase tracking-wider text-neutral-500">
            Shards
          </div>
        </div>

        {/* Table Rows */}
        {providers.map((provider, index) => {
          const config =
            statusConfig[provider.status as keyof typeof statusConfig] || statusConfig.online;
          
          const usagePercent = provider.quotaTotalBytes > 0 
            ? Math.round((provider.quotaUsedBytes / provider.quotaTotalBytes) * 100) 
            : 0;

          return (
            <div
              key={provider.providerId}
              className={cn(
                "grid grid-cols-[50px_2fr_120px_80px_100px_80px] items-center gap-4 px-4 py-4 transition-all hover:bg-neutral-50",
                index < providers.length - 1 && "border-b"
              )}
            >
              {/* Icon */}
              <div
                className="flex h-8 w-8 items-center justify-center border font-mono text-[10px] font-bold"
                style={{
                  borderColor: config.color,
                  backgroundColor: config.bg,
                  color: config.color,
                }}
              >
                {providerIcons[provider.providerId] || provider.displayName.substring(0, 2).toUpperCase()}
              </div>

              {/* Name */}
              <div>
                <div className="font-mono text-sm font-medium">
                  {provider.displayName}
                </div>
                <div
                  className="mt-0.5 inline-flex items-center gap-1.5"
                  style={{ color: config.color }}
                >
                  <div
                    className="h-1.5 w-1.5"
                    style={{ backgroundColor: config.color }}
                  />
                  <span className="font-mono text-[9px] uppercase tracking-wider">
                    {config.label}
                  </span>
                </div>
              </div>

              {/* Region */}
              <div className="font-mono text-xs text-neutral-600">
                {provider.region}
              </div>

              {/* Latency */}
              <div className="flex items-center gap-1.5">
                <Activity className="h-3 w-3 text-neutral-400" />
                <span className="font-mono text-xs text-neutral-600">
                  {provider.latencyMs}ms
                </span>
              </div>

              {/* Usage */}
              <div>
                <div className="mb-1 font-mono text-[10px] text-neutral-500">
                  {usagePercent}%
                </div>
                <div className="h-1 w-full overflow-hidden bg-neutral-200">
                  <div
                    className="h-full bg-black"
                    style={{ width: `${usagePercent}%` }}
                  />
                </div>
              </div>

              {/* Shards */}
              <div className="font-mono text-xs text-neutral-400">—</div>
            </div>
          );
        })}
      </div>

      {/* Summary Stats */}
      {(() => {
        const active = providers.filter(p => p.status === 'connected').length;
        const degraded = providers.filter(p => p.status === 'degraded').length;
        const offline = providers.filter(p => p.status === 'disconnected' || p.status === 'error' || p.status === 'offline').length;
        return (
          <div className="grid grid-cols-3 gap-3">
            <div className="border bg-white p-3">
              <div className="mb-1 font-mono text-[9px] uppercase tracking-wider text-neutral-500">Active</div>
              <div className="font-mono text-xl font-bold">{active}</div>
            </div>
            <div className="border bg-neutral-50 p-3">
              <div className="mb-1 font-mono text-[9px] uppercase tracking-wider text-neutral-500">Degraded</div>
              <div className="font-mono text-xl font-bold text-neutral-600">{degraded}</div>
            </div>
            <div className="border bg-neutral-100 p-3">
              <div className="mb-1 font-mono text-[9px] uppercase tracking-wider text-neutral-500">Offline</div>
              <div className="font-mono text-xl font-bold text-neutral-400">{offline}</div>
            </div>
          </div>
        );
      })()}
    </div>
  );
}
