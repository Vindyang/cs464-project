import { getDashboardData } from "../dashboard/componentsAction/actions";
import { cn } from "@/lib/utils";

function formatBytes(bytes: number): string {
  if (bytes === 0) return "0 B";
  const k = 1024;
  const sizes = ["B", "KB", "MB", "GB", "TB"];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(1))} ${sizes[i]}`;
}

export default async function NodesPage() {
  const { providers, files } = await getDashboardData();

  const totalShards = files.reduce(
    (s, f) => s + (f.health_status?.total_shards ?? 0),
    0
  );
  const healthyShards = files.reduce(
    (s, f) => s + (f.health_status?.healthy_shards ?? 0),
    0
  );
  const activeCount = providers.filter(
    (p) => p.status === "active" || p.status === "connected"
  ).length;

  // Build health map cells from providers + file health
  const healthCells = providers.map((p) => {
    if (p.status === "active" || p.status === "connected") return "bg-neutral-800";
    if (p.status === "degraded") return "bg-neutral-400";
    return "bg-neutral-200";
  });
  files.forEach((f) => {
    if (f.status === "UPLOADED") healthCells.push("bg-neutral-800");
    else if (f.status === "DEGRADED") healthCells.push("bg-neutral-400");
    else healthCells.push("bg-neutral-200");
  });

  return (
    <div className="space-y-5">
      {/* Header */}
      <div className="flex items-end justify-between border-b pb-4">
        <div>
          <p className="font-mono text-[10px] uppercase tracking-[0.15em] text-neutral-400 mb-0.5">
            Infrastructure
          </p>
          <h1 className="text-2xl font-semibold tracking-tight">Storage Nodes</h1>
          <p className="text-sm text-neutral-500 mt-1">
            {activeCount} active provider{activeCount !== 1 ? "s" : ""} · {totalShards} shards tracked
          </p>
        </div>
      </div>

      {/* Health Map */}
      <section className="border bg-white p-5">
        <div className="flex items-center justify-between mb-4">
          <span className="font-mono text-[10px] uppercase tracking-[0.08em] text-neutral-500">
            Node Health Map
          </span>
          <span className="font-mono text-[10px] text-neutral-400">
            {healthyShards}/{totalShards} shards healthy
          </span>
        </div>
        {healthCells.length > 0 ? (
          <div className="flex flex-wrap gap-1.5">
            {healthCells.map((color, i) => (
              <div key={i} className={cn("w-3 h-3", color)} />
            ))}
          </div>
        ) : (
          <div className="py-4 font-mono text-xs text-neutral-400">No data available.</div>
        )}
        <div className="flex items-center gap-4 mt-4">
          <div className="flex items-center gap-1.5">
            <div className="w-2.5 h-2.5 bg-neutral-800" />
            <span className="font-mono text-[9px] text-neutral-500">Healthy</span>
          </div>
          <div className="flex items-center gap-1.5">
            <div className="w-2.5 h-2.5 bg-neutral-400" />
            <span className="font-mono text-[9px] text-neutral-500">Degraded</span>
          </div>
          <div className="flex items-center gap-1.5">
            <div className="w-2.5 h-2.5 bg-neutral-200" />
            <span className="font-mono text-[9px] text-neutral-500">Offline</span>
          </div>
        </div>
      </section>

      {/* Provider Table */}
      <section className="border bg-white">
        <div className="grid grid-cols-[1.5fr_1fr_80px_80px_1.5fr] gap-4 px-5 py-3 border-b bg-neutral-50">
          {["Provider", "Region", "Status", "Latency", "Storage Capacity"].map((h) => (
            <span key={h} className="font-mono text-[9px] uppercase tracking-[0.08em] text-neutral-400">
              {h}
            </span>
          ))}
        </div>

        {providers.length === 0 ? (
          <div className="px-5 py-10 text-center font-mono text-xs text-neutral-400">
            No providers connected.
          </div>
        ) : (
          <div className="divide-y">
            {providers.map((p) => {
              const isActive = p.status === "active" || p.status === "connected";
              const isDegraded = p.status === "degraded";
              const usedPct =
                p.quotaTotalBytes > 0
                  ? Math.round((p.quotaUsedBytes / p.quotaTotalBytes) * 100)
                  : 0;

              return (
                <div
                  key={p.providerId}
                  className="grid grid-cols-[1.5fr_1fr_80px_80px_1.5fr] gap-4 px-5 py-4 items-center"
                >
                  <div className="font-mono text-[13px] font-medium text-neutral-800">
                    {p.displayName}
                  </div>
                  <div className="font-mono text-[13px] text-neutral-500">
                    {p.region || "—"}
                  </div>
                  <div className="flex items-center gap-1.5">
                    <div
                      className={cn(
                        "w-1.5 h-1.5",
                        isActive ? "bg-neutral-800" : isDegraded ? "bg-neutral-400" : "bg-neutral-200"
                      )}
                    />
                    <span className="font-mono text-[11px] text-neutral-600 capitalize">
                      {p.status}
                    </span>
                  </div>
                  <div className="font-mono text-[13px] text-neutral-500">
                    {p.latencyMs != null ? `${p.latencyMs}ms` : "—"}
                  </div>
                  <div className="space-y-1">
                    <div className="flex items-center justify-between">
                      <span className="font-mono text-[11px] text-neutral-700">
                        {formatBytes(p.quotaUsedBytes)}
                      </span>
                      <span className="font-mono text-[10px] text-neutral-400">
                        / {p.quotaTotalBytes > 0 ? formatBytes(p.quotaTotalBytes) : "—"}
                      </span>
                    </div>
                    <div className="h-1 w-full bg-neutral-100">
                      <div
                        className="h-full bg-neutral-900 transition-all"
                        style={{ width: `${usedPct}%` }}
                      />
                    </div>
                  </div>
                </div>
              );
            })}
          </div>
        )}
      </section>
    </div>
  );
}
