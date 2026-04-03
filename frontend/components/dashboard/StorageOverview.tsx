import { DashboardCard } from "./DashboardCard";
import { Database, Layers } from "lucide-react";
import { ProviderMetadata } from "@/lib/api/providers";

function formatBytes(bytes: number): { value: string; unit: string } {
  if (bytes >= 1e12) return { value: (bytes / 1e12).toFixed(1), unit: "TB" };
  if (bytes >= 1e9) return { value: (bytes / 1e9).toFixed(1), unit: "GB" };
  return { value: (bytes / 1e6).toFixed(1), unit: "MB" };
}

interface StorageOverviewProps {
  providers: ProviderMetadata[];
  fileCount: number;
}

export function StorageOverview({ providers, fileCount }: StorageOverviewProps) {
  const totalBytes = providers.reduce((sum, p) => sum + p.quotaTotalBytes, 0);
  const usedBytes = providers.reduce((sum, p) => sum + p.quotaUsedBytes, 0);
  const utilizationPercent = totalBytes > 0 ? Math.round((usedBytes / totalBytes) * 100) : 0;

  const used = formatBytes(usedBytes);
  const total = formatBytes(totalBytes);

  return (
    <DashboardCard className="col-span-1 row-span-1">
      <div className="mb-5 flex items-baseline justify-between">
        <span className="font-mono text-[10px] uppercase tracking-wider text-neutral-500">
          Storage Capacity
        </span>
        <div className="flex items-center gap-1.5">
          <div className="h-1.5 w-1.5 bg-black" />
          <span className="font-mono text-[9px] uppercase tracking-wider">
            HEALTHY
          </span>
        </div>
      </div>

      <div className="mb-2 flex items-baseline gap-2">
        <span className="font-mono text-5xl font-bold">{used.value}</span>
        <span className="font-mono text-xl text-neutral-500">{used.unit}</span>
        <span className="ml-auto font-mono text-sm text-neutral-500">
          / {totalBytes > 0 ? `${total.value} ${total.unit}` : "—"}
        </span>
      </div>

      <div className="mb-6">
        <div className="mb-2 flex justify-between font-mono text-xs">
          <span className="text-neutral-600">UTILIZATION</span>
          <span className="font-bold">{utilizationPercent}%</span>
        </div>
        <div className="h-2 overflow-hidden bg-neutral-200">
          <div className="h-full bg-black" style={{ width: `${utilizationPercent}%` }} />
        </div>
      </div>

      <div className="grid grid-cols-2 gap-4 border-t pt-4">
        <StatItem icon={Database} label="Files" value={String(fileCount)} />
        <StatItem icon={Layers} label="Providers" value={String(providers.length)} />
      </div>
    </DashboardCard>
  );
}

function StatItem({
  icon: Icon,
  label,
  value,
}: {
  icon: React.ComponentType<{ className?: string }>;
  label: string;
  value: string;
}) {
  return (
    <div>
      <div className="mb-2 flex items-center gap-2">
        <Icon className="h-3 w-3" />
        <span className="font-mono text-[10px] uppercase tracking-wider text-neutral-500">
          {label}
        </span>
      </div>
      <div className="font-mono text-2xl font-bold">{value}</div>
    </div>
  );
}
