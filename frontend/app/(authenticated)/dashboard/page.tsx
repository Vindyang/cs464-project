import { getDashboardData } from "./componentsAction/actions";
import { cn, formatBytes } from "@/lib/utils";
import Link from "next/link";
import { ReactNode } from "react";

export default async function DashboardPage() {
  const { files, providers } = await getDashboardData();

  const totalStorageBytes = files.reduce((s, f) => s + (f.original_size ?? 0), 0);
  const totalCapacityBytes = providers.reduce((s, p) => s + (p.quotaTotalBytes ?? 0), 0);

  const activeProviders = providers.filter(
    (p) => p.status === "active" || p.status === "connected"
  );
  const degradedFiles = files.filter((f) => f.status === "DEGRADED");

  const totalHealthyShards = files.reduce(
    (s, f) => s + (f.health_status?.healthy_shards ?? 0),
    0
  );
  const totalShards = files.reduce(
    (s, f) => s + (f.health_status?.total_shards ?? 0),
    0
  );
  const healthPct = totalShards > 0 ? Math.round((totalHealthyShards / totalShards) * 100) : 100;

  const recentFiles = [...files]
    .sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime())
    .slice(0, 7);

  return (
    <div className="space-y-5">
      {/* ── Page Header ── */}
      <div className="flex items-end justify-between border-b pb-4">
        <div>
          <p className="font-mono text-[10px] uppercase tracking-[0.15em] text-neutral-400 mb-0.5">
            System
          </p>
          <h1 className="text-2xl font-semibold tracking-tight">Overview</h1>
        </div>
        <Link
          href="/files"
          className="font-mono text-[11px] uppercase tracking-wider bg-black text-white px-4 py-2 hover:bg-neutral-800 transition-colors"
        >
          My Files
        </Link>
      </div>

      {/* ── 4 Stat Cards ── */}
      <div className="grid grid-cols-4 gap-4">
        <StatCard
          label="Total Storage"
          value={formatBytes(totalStorageBytes)}
          sub={
            totalCapacityBytes > 0
              ? `of ${formatBytes(totalCapacityBytes)} capacity`
              : "capacity unknown"
          }
        />
        <StatCard
          label="Active Providers"
          value={String(activeProviders.length)}
          sub={`${providers.length - activeProviders.length} inactive · ${providers.length} total`}
        />
        <StatCard
          label="Total Files"
          value={String(files.length)}
          sub={`${degradedFiles.length} degraded`}
        />
        <StatCard
          label="Shard Health"
          value={`${healthPct}%`}
          sub={`${totalShards} shards tracked`}
          warn={healthPct < 100}
        />
      </div>

      {/* ── Main 2-column ── */}
      <div className="grid gap-4" style={{ gridTemplateColumns: "1fr 300px" }}>
        {/* Provider Storage */}
        <DashCard>
          <SectionHeader label="Provider Storage" href="/providers" linkLabel="Manage" />

          {providers.length === 0 ? (
            <div className="py-12 text-center font-mono text-xs text-neutral-400">
              No providers connected.{" "}
              <Link href="/providers" className="underline hover:text-neutral-900 transition-colors">
                Connect one
              </Link>
            </div>
          ) : (
            <>
              {/* Table header */}
              <div className="grid gap-3 px-0 pb-2 border-b mb-1"
                style={{ gridTemplateColumns: "10px 1fr 70px 80px 160px" }}>
                {["", "Provider", "Status", "Latency", "Usage"].map((h) => (
                  <span key={h} className="font-mono text-[9px] uppercase tracking-[0.08em] text-neutral-400">
                    {h}
                  </span>
                ))}
              </div>

              {/* Provider rows */}
              <div className="divide-y">
                {providers.map((p) => {
                  const isActive = p.status === "active" || p.status === "connected";
                  const usedPct =
                    p.quotaTotalBytes > 0
                      ? Math.round((p.quotaUsedBytes / p.quotaTotalBytes) * 100)
                      : 0;
                  return (
                    <div
                      key={p.providerId}
                      className="grid items-center gap-3 py-3.5"
                      style={{ gridTemplateColumns: "10px 1fr 70px 80px 160px" }}
                    >
                      {/* Status dot */}
                      <div
                        className={cn(
                          "w-1.5 h-1.5",
                          isActive ? "bg-neutral-900" : "bg-neutral-300"
                        )}
                      />

                      {/* Name + region */}
                      <div className="min-w-0">
                        <div className="text-sm font-medium truncate">{p.displayName}</div>
                        {p.region && (
                          <div className="font-mono text-[10px] text-neutral-400">{p.region}</div>
                        )}
                      </div>

                      {/* Status label */}
                      <span
                        className={cn(
                          "font-mono text-[9px] uppercase tracking-wider",
                          isActive ? "text-neutral-700" : "text-neutral-400"
                        )}
                      >
                        {p.status}
                      </span>

                      {/* Latency */}
                      <span className="font-mono text-[11px] text-neutral-500">
                        {p.latencyMs ?? "—"}ms
                      </span>

                      {/* Usage bar + numbers */}
                      <div className="flex items-center gap-2 min-w-0">
                        <div className="flex-1 h-1 bg-neutral-100 overflow-hidden">
                          <div
                            className="h-full bg-neutral-900 transition-all"
                            style={{ width: `${usedPct}%` }}
                          />
                        </div>
                        <span className="font-mono text-[10px] text-neutral-500 shrink-0 tabular-nums w-8 text-right">
                          {usedPct}%
                        </span>
                        <span className="font-mono text-[10px] text-neutral-400 shrink-0 truncate">
                          {formatBytes(p.quotaUsedBytes)} / {formatBytes(p.quotaTotalBytes)}
                        </span>
                      </div>
                    </div>
                  );
                })}
              </div>
            </>
          )}
        </DashCard>

        {/* Recent Files */}
        <DashCard>
          <SectionHeader label="Recent Files" href="/files" linkLabel="All files" />

          {recentFiles.length === 0 ? (
            <div className="py-10 text-center font-mono text-xs text-neutral-400">
              No files uploaded yet.
            </div>
          ) : (
            <ul className="divide-y">
              {recentFiles.map((f) => {
                const pct = f.health_status?.health_percent ?? 100;
                const isDegraded = f.status === "DEGRADED";
                const date = new Date(f.created_at);
                const dateLabel = date.toLocaleDateString("en-US", {
                  month: "short",
                  day: "numeric",
                });
                return (
                  <li key={f.file_id}>
                    <Link
                      href={`/files/${f.file_id}`}
                      className="flex items-start gap-2.5 py-3 -mx-5 px-5 hover:bg-neutral-50 transition-colors group"
                    >
                      <div
                        className={cn(
                          "mt-1.5 w-1.5 h-1.5 shrink-0",
                          isDegraded ? "bg-neutral-300" : "bg-neutral-900"
                        )}
                      />
                      <div className="min-w-0 flex-1">
                        <div className="text-[13px] font-medium truncate group-hover:underline leading-snug">
                          {f.original_name}
                        </div>
                        <div className="font-mono text-[10px] text-neutral-400 mt-0.5 flex items-center gap-1.5">
                          <span>{formatBytes(f.original_size ?? 0)}</span>
                          <span className="text-neutral-200">·</span>
                          <span>{dateLabel}</span>
                        </div>
                      </div>
                      <div className="shrink-0 text-right">
                        <div className="font-mono text-[11px] font-semibold tabular-nums">
                          {pct}%
                        </div>
                        <div
                          className={cn(
                            "font-mono text-[9px] uppercase tracking-wider mt-0.5",
                            isDegraded ? "text-neutral-400" : "text-neutral-400"
                          )}
                        >
                          {f.status === "UPLOADED" ? "OK" : f.status === "DEGRADED" ? "RISK" : f.status}
                        </div>
                      </div>
                    </Link>
                  </li>
                );
              })}
            </ul>
          )}
        </DashCard>
      </div>

      {/* ── Files at Risk (conditional) ── */}
      {degradedFiles.length > 0 && (
        <DashCard>
          <SectionHeader
            label={`Files at Risk — ${degradedFiles.length} degraded`}
            href="/files"
            linkLabel="View all"
          />
          <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 gap-3 mt-1">
            {degradedFiles.map((f) => {
              const pct = f.health_status?.health_percent ?? 0;
              const recoverable = f.health_status?.recoverable ?? false;
              return (
                <Link
                  key={f.file_id}
                  href={`/files/${f.file_id}`}
                  className="group border p-3.5 hover:border-neutral-900 transition-colors block"
                >
                  <div className="font-mono text-[10px] text-neutral-400 truncate mb-2">
                    {f.original_name}
                  </div>
                  <div className="flex items-baseline gap-1 mb-2">
                    <span className="text-2xl font-semibold tabular-nums">{pct}</span>
                    <span className="font-mono text-[10px] text-neutral-400">%</span>
                  </div>
                  <div className="h-0.5 w-full bg-neutral-100">
                    <div className="h-full bg-neutral-900" style={{ width: `${pct}%` }} />
                  </div>
                  <div
                    className={cn(
                      "mt-2 font-mono text-[9px] uppercase tracking-wider",
                      recoverable ? "text-neutral-500" : "text-neutral-900 font-semibold"
                    )}
                  >
                    {recoverable ? "Recoverable" : "Critical"}
                  </div>
                </Link>
              );
            })}
          </div>
        </DashCard>
      )}
    </div>
  );
}

// ── Shared UI primitives ────────────────────────────────────────────────────

function StatCard({
  label,
  value,
  sub,
  warn = false,
}: {
  label: string;
  value: string;
  sub: string;
  warn?: boolean;
}) {
  return (
    <section className="border bg-white p-5">
      <p className="font-mono text-[10px] uppercase tracking-[0.08em] text-neutral-400 mb-3">
        {label}
      </p>
      <p
        className={cn(
          "text-3xl font-semibold tracking-tight leading-none tabular-nums",
          warn && "text-neutral-500"
        )}
      >
        {value}
      </p>
      <p className="font-mono text-[11px] text-neutral-400 mt-1.5">{sub}</p>
    </section>
  );
}

function DashCard({
  children,
  className,
}: {
  children: ReactNode;
  className?: string;
}) {
  return (
    <section className={cn("border bg-white p-5", className)}>{children}</section>
  );
}

function SectionHeader({
  label,
  href,
  linkLabel,
}: {
  label: string;
  href: string;
  linkLabel: string;
}) {
  return (
    <div className="flex items-center justify-between mb-5">
      <span className="font-mono text-[10px] uppercase tracking-[0.08em] text-neutral-500">
        {label}
      </span>
      <Link
        href={href}
        className="font-mono text-[10px] uppercase tracking-wider text-neutral-400 hover:text-neutral-900 transition-colors"
      >
        {linkLabel} →
      </Link>
    </div>
  );
}
