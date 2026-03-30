import { getDashboardData } from "./componentsAction/actions";
import { cn } from "@/lib/utils";
import Link from "next/link";

function formatBytes(bytes: number): string {
  if (bytes >= 1e12) return `${(bytes / 1e12).toFixed(1)} TB`;
  if (bytes >= 1e9) return `${(bytes / 1e9).toFixed(1)} GB`;
  if (bytes >= 1e6) return `${(bytes / 1e6).toFixed(1)} MB`;
  return `${(bytes / 1e3).toFixed(1)} KB`;
}

export default async function DashboardPage() {
  const { files, providers } = await getDashboardData();

  const totalStorageBytes = files.reduce((s, f) => s + (f.original_size ?? 0), 0);
  const degradedCount = files.filter((f) => f.status === "DEGRADED").length;
  const providersOnline = providers.filter(
    (p) => p.status === "connected" || p.status === "online"
  ).length;

  const filesAtRisk = files.filter((f) => f.status === "DEGRADED");

  const schemeCounts: Record<string, number> = {};
  for (const f of files) {
    if (f.n && f.k) {
      const key = `(${f.n},${f.k})`;
      schemeCounts[key] = (schemeCounts[key] ?? 0) + 1;
    }
  }
  const schemeEntries = Object.entries(schemeCounts).sort((a, b) => b[1] - a[1]);

  return (
    <div className="space-y-8 max-w-4xl">
      {/* Stat Row */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-px bg-border">
        <StatCell label="Total Files" value={String(files.length)} />
        <StatCell label="Storage Used" value={formatBytes(totalStorageBytes)} />
        <StatCell
          label="Files at Risk"
          value={String(degradedCount)}
          muted={degradedCount === 0}
        />
        <StatCell
          label="Providers Online"
          value={`${providersOnline} / ${providers.length}`}
        />
      </div>

      {/* Files at Risk */}
      {filesAtRisk.length > 0 && (
        <section>
          <SectionLabel>Files at Risk</SectionLabel>
          <div className="border divide-y">
            {filesAtRisk.map((f) => {
              const pct = f.health_status?.health_percent ?? 0;
              const recoverable = f.health_status?.recoverable ?? false;
              return (
                <Link
                  key={f.file_id}
                  href={`/files/${f.file_id}`}
                  className="flex items-center gap-4 px-4 py-3 hover:bg-neutral-50 transition-colors"
                >
                  <span className="font-mono text-xs flex-1 truncate">{f.original_name}</span>
                  <div className="flex items-center gap-2 shrink-0">
                    <div className="w-24 h-0.5 bg-neutral-200">
                      <div className="h-full bg-black" style={{ width: `${pct}%` }} />
                    </div>
                    <span className="font-mono text-[10px] text-neutral-500 w-8 text-right">
                      {pct}%
                    </span>
                  </div>
                  <span
                    className={cn(
                      "font-mono text-[9px] uppercase tracking-wider px-2 py-0.5 border shrink-0",
                      recoverable
                        ? "border-black text-black"
                        : "border-neutral-300 text-neutral-400"
                    )}
                  >
                    {recoverable ? "Recoverable" : "At Risk"}
                  </span>
                </Link>
              );
            })}
          </div>
        </section>
      )}

      {/* Provider Grid */}
      <section>
        <SectionLabel>Provider Status</SectionLabel>
        {providers.length === 0 ? (
          <div className="border px-4 py-8 text-center">
            <p className="font-mono text-xs text-neutral-400">No providers connected.</p>
            <Link
              href="/providers"
              className="mt-2 inline-block font-mono text-[10px] uppercase tracking-wider underline text-neutral-500 hover:text-black"
            >
              Connect a provider
            </Link>
          </div>
        ) : (
          <div className="border divide-y">
            {providers.map((p) => {
              const usedPct =
                p.quotaTotalBytes > 0
                  ? Math.round((p.quotaUsedBytes / p.quotaTotalBytes) * 100)
                  : 0;
              const isOnline = p.status === "connected" || p.status === "online";
              return (
                <div key={p.providerId} className="flex items-center gap-4 px-4 py-3">
                  <span className="font-mono text-xs w-32 shrink-0 truncate">
                    {p.displayName}
                  </span>
                  <span
                    className={cn(
                      "font-mono text-[9px] uppercase tracking-wider px-2 py-0.5 border shrink-0",
                      isOnline
                        ? "border-black text-black"
                        : "border-neutral-300 text-neutral-400"
                    )}
                  >
                    {isOnline ? "Online" : p.status}
                  </span>
                  <div className="flex-1 flex items-center gap-2 min-w-0">
                    <div className="flex-1 h-0.5 bg-neutral-200">
                      <div className="h-full bg-black" style={{ width: `${usedPct}%` }} />
                    </div>
                    <span className="font-mono text-[10px] text-neutral-400 w-32 text-right shrink-0">
                      {formatBytes(p.quotaUsedBytes)} / {formatBytes(p.quotaTotalBytes)}
                    </span>
                  </div>
                  <span className="font-mono text-[10px] text-neutral-400 w-14 text-right shrink-0">
                    {p.latencyMs}ms
                  </span>
                </div>
              );
            })}
          </div>
        )}
      </section>

      {/* Redundancy Summary */}
      <section>
        <SectionLabel>Redundancy Distribution</SectionLabel>
        {schemeEntries.length === 0 ? (
          <div className="border px-4 py-4">
            <p className="font-mono text-xs text-neutral-400">No files uploaded yet.</p>
          </div>
        ) : (
          <div className="border divide-y">
            {schemeEntries.map(([scheme, count]) => (
              <div key={scheme} className="flex items-center gap-4 px-4 py-2.5">
                <span className="font-mono text-xs w-16 shrink-0">{scheme}</span>
                <div className="flex-1 h-0.5 bg-neutral-200">
                  <div
                    className="h-full bg-black"
                    style={{ width: `${(count / files.length) * 100}%` }}
                  />
                </div>
                <span className="font-mono text-[10px] text-neutral-500 shrink-0">
                  {count} {count === 1 ? "file" : "files"}
                </span>
              </div>
            ))}
          </div>
        )}
      </section>
    </div>
  );
}

function StatCell({
  label,
  value,
  muted = false,
}: {
  label: string;
  value: string;
  muted?: boolean;
}) {
  return (
    <div className="bg-white px-5 py-4">
      <p className="font-mono text-[9px] uppercase tracking-widest text-neutral-500 mb-1.5">
        {label}
      </p>
      <p className={cn("font-mono text-2xl font-bold", muted && "text-neutral-300")}>
        {value}
      </p>
    </div>
  );
}

function SectionLabel({ children }: { children: React.ReactNode }) {
  return (
    <h2 className="font-mono text-[9px] uppercase tracking-widest text-neutral-500 mb-3">
      {children}
    </h2>
  );
}
