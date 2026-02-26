import { DashboardCard } from "./DashboardCard";
import { Shield, Lock } from "lucide-react";

export function SystemHealth() {
  // 6 total shards (4 data + 2 parity) as per project spec
  const shards = [
    { type: "data", status: "healthy", index: 1 },
    { type: "data", status: "healthy", index: 2 },
    { type: "data", status: "healthy", index: 3 },
    { type: "data", status: "healthy", index: 4 },
    { type: "parity", status: "healthy", index: 5 },
    { type: "parity", status: "healthy", index: 6 },
  ];

  return (
    <DashboardCard className="col-span-1 row-span-1">
      <div className="mb-5 flex items-baseline justify-between">
        <span className="font-mono text-[10px] uppercase tracking-wider text-neutral-500">
          Reed-Solomon (6,4)
        </span>
        <div className="flex items-center gap-1.5">
          <Shield className="h-3 w-3" />
          <span className="font-mono text-[9px] uppercase tracking-wider">
            OPTIMAL
          </span>
        </div>
      </div>

      {/* Shard visualization */}
      <div className="mb-6 grid grid-cols-3 gap-2">
        {shards.map((shard) => (
          <div
            key={shard.index}
            className="group relative aspect-square border transition-all hover:scale-105 hover:border-black"
            style={{
              backgroundColor:
                shard.type === "data" ? "#fafafa" : "#f5f5f5",
            }}
          >
            {/* Shard index */}
            <div className="absolute left-1 top-1 font-mono text-[9px] text-neutral-400">
              {shard.index}
            </div>

            {/* Status indicator */}
            <div
              className="absolute bottom-1 right-1 h-1 w-1"
              style={{
                backgroundColor: shard.status === "healthy" ? "#000" : "#d4d4d4",
              }}
            />

            {/* Shard type label (on hover) */}
            <div className="absolute inset-0 flex items-center justify-center opacity-0 transition-opacity group-hover:opacity-100">
              <span className="font-mono text-[8px] uppercase tracking-wider text-black">
                {shard.type}
              </span>
            </div>
          </div>
        ))}
      </div>

      {/* Legend */}
      <div className="mb-4 flex items-center justify-center gap-4 border-t pt-4">
        <div className="flex items-center gap-2">
          <div className="h-2 w-2 border bg-neutral-50" />
          <span className="font-mono text-[10px] text-neutral-600">DATA (4)</span>
        </div>
        <div className="flex items-center gap-2">
          <div className="h-2 w-2 border bg-neutral-100" />
          <span className="font-mono text-[10px] text-neutral-600">PARITY (2)</span>
        </div>
      </div>

      {/* Encryption info */}
      <div className="space-y-2 border bg-neutral-50 p-3">
        <div className="flex items-center gap-2">
          <Lock className="h-3 w-3" />
          <span className="font-mono text-[10px] uppercase tracking-wider text-neutral-500">
            Encryption
          </span>
        </div>
        <div className="font-mono text-xs font-bold">AES-256-GCM</div>
        <div className="font-mono text-[10px] text-neutral-600">
          Client-side • Zero-knowledge
        </div>
      </div>
    </DashboardCard>
  );
}
