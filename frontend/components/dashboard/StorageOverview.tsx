import { DashboardCard } from "./DashboardCard";
import { Database, Layers } from "lucide-react";

export function StorageOverview() {
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
        <span className="font-mono text-5xl font-bold">4.2</span>
        <span className="font-mono text-xl text-neutral-500">TB</span>
        <span className="ml-auto font-mono text-sm text-neutral-500">
          / 12.0 TB
        </span>
      </div>

      <div className="mb-6">
        <div className="mb-2 flex justify-between font-mono text-xs">
          <span className="text-neutral-600">UTILIZATION</span>
          <span className="font-bold">35.0%</span>
        </div>
        <div className="h-2 overflow-hidden bg-neutral-200">
          <div className="h-full bg-black" style={{ width: "35%" }} />
        </div>
      </div>

      <div className="grid grid-cols-2 gap-4 border-t pt-4">
        <StatItem icon={Database} label="Files" value="1,247" />
        <StatItem icon={Layers} label="Shards" value="7,482" />
      </div>
    </DashboardCard>
  );
}

function StatItem({
  icon: Icon,
  label,
  value,
}: {
  icon: any;
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
