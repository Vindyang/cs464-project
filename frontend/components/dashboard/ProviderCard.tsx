import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Cloud, Check, Loader2, AlertTriangle, RefreshCw, HardDrive } from "lucide-react";
import { cn } from "@/lib/utils";
import { DashboardCard } from "../dashboard/DashboardCard";

export interface ProviderCardProps {
  providerId: string;
  displayName: string;
  status: "connected" | "degraded" | "disconnected" | "error";
  region: string;
  latencyMs: number;
  quotaUsedBytes: number;
  quotaTotalBytes: number;
  shardCount: number;
  lastHealthCheckAt: string;
  onConnect: () => void;
  onDisconnect: () => void;
  onRefresh: () => void;
}

export function ProviderCard({
  displayName,
  status,
  quotaUsedBytes,
  quotaTotalBytes,
  shardCount,
  latencyMs,
  onConnect,
  onDisconnect,
  onRefresh,
}: ProviderCardProps) {
  const percentage = quotaTotalBytes > 0 ? Math.round((quotaUsedBytes / quotaTotalBytes) * 100) : 0;
  const formattedUsed = (quotaUsedBytes / (1024 * 1024 * 1024)).toFixed(1) + " GB";
  const formattedTotal = (quotaTotalBytes / (1024 * 1024 * 1024)).toFixed(1) + " GB";
  return (
    <DashboardCard className="p-0 border-border-color bg-bg-canvas">
      <div className="p-4 border-b border-border-color flex justify-between items-center">
        <div className="flex items-center gap-3">
          <div className="w-8 h-8 bg-bg-subtle rounded-[2px] flex items-center justify-center border border-border-color">
            <Cloud className="w-4 h-4 text-text-secondary" />
          </div>
          <span className="font-semibold text-sm tracking-tight">{displayName}</span>
        </div>
        <StatusBadge status={status} />
      </div>

      <div className="p-4">
        {status === "connected" ? (
          <div className="space-y-4">
            <div>
              <div className="flex justify-between text-[11px] uppercase tracking-wider text-text-secondary mb-2 font-mono">
                <span>Usage</span>
                <span>{percentage}%</span>
              </div>
              <div className="h-1 bg-bg-subtle overflow-hidden rounded-full">
                <div
                  className="h-full bg-accent-primary transition-all duration-500"
                  style={{ width: `${percentage}%` }}
                />
              </div>
              <div className="flex justify-between text-xs text-text-secondary mt-1.5">
                <span>{formattedUsed} Used</span>
                <span>{formattedTotal} Total</span>
              </div>
            </div>

            <div className="grid grid-cols-2 gap-px bg-border-color border border-border-color rounded-[2px] overflow-hidden">
              <div className="bg-bg-subtle/30 p-2 text-center">
                <div className="text-lg font-medium leading-none mb-1">{shardCount}</div>
                <div className="text-[10px] text-text-secondary uppercase tracking-wider">Shards</div>
              </div>
              <div className="bg-bg-subtle/30 p-2 text-center">
                <div className="text-lg font-medium leading-none mb-1">{latencyMs}ms</div>
                <div className="text-[10px] text-text-secondary uppercase tracking-wider">Ping</div>
              </div>
            </div>

            <div className="flex gap-2">
              <Button variant="outline" size="sm" className="flex-1 h-8 text-xs border-border-color text-text-main hover:bg-bg-subtle" onClick={onDisconnect}>
                Disconnect
              </Button>
              <Button variant="ghost" size="icon" className="h-8 w-8 text-text-secondary hover:text-text-main" onClick={onRefresh}>
                <RefreshCw className="w-3.5 h-3.5" />
              </Button>
            </div>
          </div>
        ) : (
          <div className="py-2 text-center">
            <p className="text-xs text-text-secondary mb-4 leading-relaxed">
              Connect your {displayName} account to list it as an available node.
            </p>
            <Button size="sm" className="w-full h-8 bg-text-main text-bg-canvas hover:bg-text-main/90" onClick={onConnect}>
              Connect Provider
            </Button>
          </div>
        )}
      </div>
    </DashboardCard>
  );
}

function StatusBadge({ status }: { status: ProviderCardProps["status"] }) {
  const styles = {
    connected: "text-emerald-600 bg-emerald-50 border-emerald-200",
    degraded: "text-amber-600 bg-amber-50 border-amber-200",
    disconnected: "text-text-tertiary bg-bg-subtle border-border-color",
    error: "text-red-600 bg-red-50 border-red-200",
  };

  const icons = {
    connected: Check,
    degraded: AlertTriangle,
    disconnected: HardDrive,
    error: AlertTriangle,
  };

  const Icon = icons[status];

  return (
    <div className={cn("inline-flex items-center gap-1.5 px-2 py-0.5 rounded-[2px] border text-[10px] font-medium uppercase tracking-wide", styles[status])}>
      <Icon className="w-3 h-3" />
      <span>{status}</span>
    </div>
  );
}
