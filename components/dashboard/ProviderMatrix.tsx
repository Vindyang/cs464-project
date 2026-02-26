import { DashboardCard } from "./DashboardCard";
import { cn } from "@/lib/utils";
import { Cloud, Activity, HardDrive } from "lucide-react";
import { Button } from "@/components/ui/button";

const providers = [
  {
    icon: "GD",
    name: "Google Drive",
    region: "global",
    latency: 45,
    usage: 42,
    status: "online",
    shards: 248,
  },
  {
    icon: "S3",
    name: "AWS S3",
    region: "us-east-1",
    latency: 24,
    usage: 78,
    status: "online",
    shards: 312,
  },
  {
    icon: "DB",
    name: "Dropbox",
    region: "eu-west",
    latency: 112,
    usage: 15,
    status: "warning",
    shards: 89,
  },
  {
    icon: "B2",
    name: "Backblaze B2",
    region: "us-west",
    latency: 89,
    usage: 91,
    status: "online",
    shards: 402,
  },
  {
    icon: "OD",
    name: "OneDrive",
    region: "us-east",
    latency: 320,
    usage: 5,
    status: "offline",
    shards: 0,
  },
];

const statusConfig = {
  online: {
    color: "#000000",
    label: "ONLINE",
    bg: "#ffffff",
  },
  warning: {
    color: "#737373",
    label: "DEGRADED",
    bg: "#fafafa",
  },
  offline: {
    color: "#d4d4d4",
    label: "OFFLINE",
    bg: "#f5f5f5",
  },
};

export function ProviderMatrix() {
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
            statusConfig[provider.status as keyof typeof statusConfig];

          return (
            <div
              key={index}
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
                {provider.icon}
              </div>

              {/* Name */}
              <div>
                <div className="font-mono text-sm font-medium">
                  {provider.name}
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
                  {provider.latency}ms
                </span>
              </div>

              {/* Usage */}
              <div>
                <div className="mb-1 font-mono text-[10px] text-neutral-500">
                  {provider.usage}%
                </div>
                <div className="h-1 w-full overflow-hidden bg-neutral-200">
                  <div
                    className="h-full bg-black"
                    style={{ width: `${provider.usage}%` }}
                  />
                </div>
              </div>

              {/* Shards */}
              <div className="flex items-center gap-1.5">
                <HardDrive className="h-3 w-3 text-neutral-400" />
                <span className="font-mono text-xs text-neutral-600">
                  {provider.shards}
                </span>
              </div>
            </div>
          );
        })}
      </div>

      {/* Summary Stats */}
      <div className="grid grid-cols-3 gap-3">
        <div className="border bg-white p-3">
          <div className="mb-1 font-mono text-[9px] uppercase tracking-wider text-neutral-500">
            Active
          </div>
          <div className="font-mono text-xl font-bold">3</div>
        </div>
        <div className="border bg-neutral-50 p-3">
          <div className="mb-1 font-mono text-[9px] uppercase tracking-wider text-neutral-500">
            Degraded
          </div>
          <div className="font-mono text-xl font-bold text-neutral-600">1</div>
        </div>
        <div className="border bg-neutral-100 p-3">
          <div className="mb-1 font-mono text-[9px] uppercase tracking-wider text-neutral-500">
            Offline
          </div>
          <div className="font-mono text-xl font-bold text-neutral-400">1</div>
        </div>
      </div>
    </div>
  );
}
