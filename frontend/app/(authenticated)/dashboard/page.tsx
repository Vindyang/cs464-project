import { getDashboardData } from "./componentsAction/actions";
import { cn, formatBytes } from "@/lib/utils";
import Link from "next/link";
import { ReactNode } from "react";

export default async function DashboardPage() {
  const { files, providers } = await getDashboardData();
  const recentFileDateFormatter = new Intl.DateTimeFormat(undefined, {
    month: "short",
    day: "numeric",
  });

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
    <div className="space-y-6">
      {/* ── Page Header ── */}
      <div className="flex flex-wrap items-end justify-between gap-4 border-b pb-5">
        <div>
          <p className="mb-1 font-mono text-[12px] font-medium uppercase tracking-[0.15em] text-neutral-500">
            System
          </p>
          <h1 className="text-3xl font-semibold tracking-tight">Overview</h1>
        </div>
        <Link
          href="/files"
          className="bg-black px-5 py-2.5 font-mono text-[12px] uppercase tracking-wider text-white transition-colors hover:bg-neutral-800"
        >
          My Files
        </Link>
      </div>

      {/* ── 4 Stat Cards ── */}
      <div className="grid grid-cols-1 gap-5 sm:grid-cols-2 xl:grid-cols-4">
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
      <div className="grid gap-5 xl:grid-cols-[minmax(0,1fr)_320px]">
        {/* Provider Storage */}
        <DashCard>
          <SectionHeader label="Provider Storage" href="/providers" linkLabel="Manage" />

          {providers.length === 0 ? (
            <div className="py-12 text-center font-mono text-sm text-neutral-400">
              No providers connected.{" "}
              <Link href="/providers" className="underline hover:text-neutral-900 transition-colors">
                Connect one
              </Link>
            </div>
          ) : (
            <>
              <div className="overflow-x-auto">
                <div className="min-w-[620px]">
                  {/* Table header */}
                  <div className="mb-1 grid gap-3 border-b px-0 pb-2.5" style={{ gridTemplateColumns: "10px 1fr 88px 96px 190px" }}>
                    {["", "Provider", "Status", "Latency", "Usage"].map((h) => (
                      <span key={h} className="font-mono text-[11px] font-medium uppercase tracking-[0.08em] text-neutral-500">
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
                          className="grid items-center gap-3 py-4"
                          style={{ gridTemplateColumns: "10px 1fr 88px 96px 190px" }}
                        >
                          {/* Status dot */}
                          <div
                            className={cn(
                              "h-1.5 w-1.5",
                              isActive ? "bg-neutral-900" : "bg-neutral-300"
                            )}
                          />

                          {/* Name + region */}
                          <div className="min-w-0">
                            <div className="truncate text-base font-medium leading-snug">{p.displayName}</div>
                            {p.region && (
                              <div className="font-mono text-[12px] font-medium text-neutral-500">{p.region}</div>
                            )}
                          </div>

                          {/* Status label */}
                          <span
                            className={cn(
                              "font-mono text-[11px] font-medium uppercase tracking-wider",
                              isActive ? "text-neutral-700" : "text-neutral-500"
                            )}
                          >
                            {p.status}
                          </span>

                          {/* Latency */}
                          <span className="font-mono text-[12px] text-neutral-500">
                            {p.latencyMs != null ? `${p.latencyMs}ms` : "—"}
                          </span>

                      {/* Usage bar + numbers */}
                      <div className="flex min-w-0 items-center gap-2.5">
                        <div className="flex-1 h-1 bg-neutral-100 overflow-hidden">
                          <div
                            className="h-full bg-neutral-900 transition-all"
                            style={{ width: `${usedPct}%` }}
                          />
                        </div>
                        <span className="w-9 shrink-0 text-right font-mono text-[12px] font-medium tabular-nums text-neutral-600">
                          {usedPct}%
                        </span>
                        <span className="shrink-0 truncate font-mono text-[12px] font-medium text-neutral-500">
                          {formatBytes(p.quotaUsedBytes)}{p.quotaTotalBytes > 0 ? ` / ${formatBytes(p.quotaTotalBytes)}` : " / —"}
                        </span>
                      </div>
                    </div>
                  );
                })}
              </div>
            </div>
          </div>
            </>
          )}
        </DashCard>

        {/* Recent Files */}
        <DashCard>
          <SectionHeader label="Recent Files" href="/files" linkLabel="All files" />

          {recentFiles.length === 0 ? (
            <div className="py-10 text-center font-mono text-sm text-neutral-400">
              No files uploaded yet.
            </div>
          ) : (
            <ul className="divide-y">
              {recentFiles.map((f) => {
                const pct = f.health_status?.health_percent ?? 100;
                const isDegraded = f.status === "DEGRADED";
                const dateLabel = recentFileDateFormatter.format(new Date(f.created_at));
                return (
                  <li key={f.file_id}>
                    <Link
                      href={`/files/${f.file_id}`}
                      className="group -mx-5 flex items-start gap-3 px-5 py-3.5 transition-colors hover:bg-neutral-50"
                    >
                      <div
                        className={cn(
                          "mt-1.5 w-1.5 h-1.5 shrink-0",
                          isDegraded ? "bg-neutral-300" : "bg-neutral-900"
                        )}
                      />
                      <div className="min-w-0 flex-1">
                        <div className="truncate text-sm font-medium leading-snug group-hover:underline">
                          {f.original_name}
                        </div>
                        <div className="mt-1 flex items-center gap-1.5 font-mono text-[12px] font-medium text-neutral-500">
                          <span>{formatBytes(f.original_size ?? 0)}</span>
                          <span className="text-neutral-200">·</span>
                          <span>{dateLabel}</span>
                        </div>
                      </div>
                      <div className="shrink-0 text-right">
                        <div className="font-mono text-[12px] font-semibold tabular-nums">
                          {pct}%
                        </div>
                        <div
                          className={cn(
                            "mt-0.5 font-mono text-[11px] font-medium uppercase tracking-wider",
                            isDegraded ? "text-amber-700" : "text-neutral-500"
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
          <div className="mt-2 grid grid-cols-2 gap-4 sm:grid-cols-3 lg:grid-cols-4">
            {degradedFiles.map((f) => {
              const pct = f.health_status?.health_percent ?? 0;
              const recoverable = f.health_status?.recoverable ?? false;
              return (
                <Link
                  key={f.file_id}
                  href={`/files/${f.file_id}`}
                  className="group block border p-4 hover:border-neutral-900 transition-colors"
                >
                  <div className="mb-2.5 truncate font-mono text-[12px] font-medium text-neutral-500">
                    {f.original_name}
                  </div>
                  <div className="mb-2.5 flex items-baseline gap-1">
                    <span className="text-2xl font-semibold tabular-nums">{pct}</span>
                    <span className="font-mono text-[12px] font-medium text-neutral-500">%</span>
                  </div>
                  <div className="h-0.5 w-full bg-neutral-100">
                    <div className="h-full bg-neutral-900" style={{ width: `${pct}%` }} />
                  </div>
                  <div
                    className={cn(
                      "mt-2.5 font-mono text-[11px] font-medium uppercase tracking-wider",
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
    <section className="border bg-white p-6">
      <p className="mb-3.5 font-mono text-[12px] font-medium uppercase tracking-[0.08em] text-neutral-500">
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
      <p className="mt-2 font-mono text-[12px] font-medium leading-relaxed text-neutral-500">{sub}</p>
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
    <section className={cn("border bg-white p-6", className)}>{children}</section>
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
    <div className="mb-6 flex items-center justify-between">
      <span className="font-mono text-[12px] font-medium uppercase tracking-[0.08em] text-neutral-600">
        {label}
      </span>
      <Link
        href={href}
        className="font-mono text-[12px] font-medium uppercase tracking-wider text-neutral-500 transition-colors hover:text-neutral-900"
      >
        {linkLabel} →
      </Link>
    </div>
  );
}
