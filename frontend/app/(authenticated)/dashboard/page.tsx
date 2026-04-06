import { getDashboardData } from "./componentsAction/actions";
import { cn, formatBytes, formatDateTime, formatRelativeTime, roundHealthPercent } from "@/lib/utils";
import { getCredentialStatus } from "@/lib/api/credentials";
import Link from "next/link";
import { ReactNode } from "react";

export default async function DashboardPage() {
  const [{ files, providers }, credentialStatus] = await Promise.all([
    getDashboardData(),
    getCredentialStatus().catch(() => ({ configured: false, count: 0, providers: [] })),
  ]);
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
  const latestHealthRefreshAt = [...files]
    .map((file) => file.last_health_refresh_at)
    .filter((value): value is string => Boolean(value))
    .sort((left, right) => new Date(right).getTime() - new Date(left).getTime())[0] ?? null;

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
      <div className="flex flex-wrap items-end justify-between gap-4 border-b border-neutral-200 pb-5 dark:border-neutral-800">
        <div>
          <p className="mb-1 font-mono text-[12px] font-medium uppercase tracking-[0.15em] text-neutral-500 dark:text-neutral-400">
            System
          </p>
          <h1 className="text-3xl font-semibold tracking-tight text-neutral-950 dark:text-neutral-100">Overview</h1>
          <div className="mt-2 flex flex-wrap items-center gap-2 text-[11px]">
            <span className="border border-sky-200 bg-sky-50 px-2 py-1 font-mono uppercase tracking-[0.08em] text-sky-700 dark:border-sky-900 dark:bg-sky-950/60 dark:text-sky-300">
              Last Health Sync {formatRelativeTime(latestHealthRefreshAt)}
            </span>
            <span className="font-mono text-neutral-500 dark:text-neutral-400">
              {formatDateTime(latestHealthRefreshAt)}
            </span>
          </div>
        </div>
        <Link
          href="/files"
          className="bg-sky-600 px-5 py-2.5 font-mono text-[12px] uppercase tracking-wider text-white transition-colors hover:bg-sky-700"
        >
          My Files
        </Link>
      </div>

      {!credentialStatus.configured && (
        <section className="flex flex-wrap items-start justify-between gap-4 border border-sky-200 bg-sky-50 px-4 py-3 text-sm text-sky-950 dark:border-sky-900 dark:bg-sky-950/40 dark:text-sky-100">
          <div>
            <p className="font-mono text-[11px] uppercase tracking-[0.1em] text-sky-700 dark:text-sky-300">
              Setup Required
            </p>
            <p className="mt-1 max-w-2xl">
              No provider credentials are configured right now. Add credentials to begin uploading files, connecting providers, and using the service.
            </p>
          </div>
          <Link
            href="/credentials"
            className="border border-sky-600 bg-sky-600 px-4 py-2 font-mono text-[11px] uppercase tracking-wider text-white transition-colors hover:bg-sky-700"
          >
            Set Credentials
          </Link>
        </section>
      )}

      {degradedFiles.length > 0 && (
        <section className="border border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-900 dark:border-amber-900 dark:bg-amber-950/40 dark:text-amber-200">
          <p className="font-mono text-[11px] uppercase tracking-[0.1em] text-amber-700 dark:text-amber-300">
            Action Needed
          </p>
          <p className="mt-1">
            {degradedFiles.length} file{degradedFiles.length === 1 ? " is" : "s are"} degraded. Review missing shards and refresh health from the files view after provider recovery.
          </p>
        </section>
      )}

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
          tone={healthPct < 100 ? "warning" : "healthy"}
        />
      </div>

      {/* ── Main 2-column ── */}
      <div className="grid gap-5 xl:grid-cols-[minmax(0,1fr)_320px]">
        {/* Provider Storage */}
        <DashCard>
          <SectionHeader label="Provider Storage" href="/providers" linkLabel="Manage" />

          {providers.length === 0 ? (
            <div className="py-12 text-center font-mono text-sm text-neutral-400 dark:text-neutral-500">
              No providers connected.{" "}
              <Link href="/providers" className="underline transition-colors hover:text-neutral-900 dark:hover:text-neutral-100">
                Connect one
              </Link>
            </div>
          ) : (
            <>
              <div className="overflow-x-auto">
                <div className="min-w-[620px]">
                  {/* Table header */}
                  <div className="mb-1 grid gap-3 border-b border-neutral-200 px-0 pb-2.5 dark:border-neutral-800" style={{ gridTemplateColumns: "10px 1fr 88px 96px 190px" }}>
                    {["", "Provider", "Status", "Latency", "Usage"].map((h) => (
                      <span key={h} className="font-mono text-[11px] font-medium uppercase tracking-[0.08em] text-neutral-500 dark:text-neutral-400">
                        {h}
                      </span>
                    ))}
                  </div>

                  {/* Provider rows */}
                  <div className="divide-y divide-neutral-200 dark:divide-neutral-800">
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
                              isActive ? "bg-neutral-900 dark:bg-sky-400" : "bg-neutral-300 dark:bg-neutral-700"
                            )}
                          />

                          {/* Name + region */}
                          <div className="min-w-0">
                            <div className="truncate text-base font-medium leading-snug text-neutral-950 dark:text-neutral-100">{p.displayName}</div>
                            {p.region && (
                              <div className="font-mono text-[12px] font-medium text-neutral-500 dark:text-neutral-400">{p.region}</div>
                            )}
                          </div>

                          {/* Status label */}
                          <span
                            className={cn(
                              "font-mono text-[11px] font-medium uppercase tracking-wider",
                              isActive ? "text-neutral-700 dark:text-neutral-200" : "text-neutral-500 dark:text-neutral-400"
                            )}
                          >
                            {p.status}
                          </span>

                          {/* Latency */}
                          <span className="font-mono text-[12px] text-neutral-500 dark:text-neutral-400">
                            {p.latencyMs != null ? `${p.latencyMs}ms` : "—"}
                          </span>

                      {/* Usage bar + numbers */}
                      <div className="flex min-w-0 items-center gap-2.5">
                        <div className="flex-1 h-1 overflow-hidden bg-neutral-100 dark:bg-neutral-800">
                          <div
                            className="h-full bg-neutral-900 transition-all dark:bg-sky-400"
                            style={{ width: `${usedPct}%` }}
                          />
                        </div>
                        <span className="w-9 shrink-0 text-right font-mono text-[12px] font-medium tabular-nums text-neutral-600 dark:text-neutral-300">
                          {usedPct}%
                        </span>
                        <span className="shrink-0 truncate font-mono text-[12px] font-medium text-neutral-500 dark:text-neutral-400">
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
            <div className="py-10 text-center font-mono text-sm text-neutral-400 dark:text-neutral-500">
              No files uploaded yet.
            </div>
          ) : (
            <ul className="divide-y divide-neutral-200 dark:divide-neutral-800">
              {recentFiles.map((f) => {
                const pct = roundHealthPercent(f.health_status?.health_percent, 100);
                const isDegraded = f.status === "DEGRADED";
                const dateLabel = recentFileDateFormatter.format(new Date(f.created_at));
                return (
                  <li key={f.file_id}>
                    <Link
                      href={`/files/${f.file_id}`}
                      className="group -mx-5 flex items-start gap-3 px-5 py-3.5 transition-colors hover:bg-neutral-50 dark:hover:bg-neutral-900/70"
                    >
                      <div
                        className={cn(
                          "mt-1.5 w-1.5 h-1.5 shrink-0",
                          isDegraded ? "bg-neutral-300 dark:bg-amber-400" : "bg-neutral-900 dark:bg-sky-400"
                        )}
                      />
                      <div className="min-w-0 flex-1">
                        <div className="truncate text-sm font-medium leading-snug text-neutral-950 group-hover:underline dark:text-neutral-100">
                          {f.original_name}
                        </div>
                        <div className="mt-1 flex items-center gap-1.5 font-mono text-[12px] font-medium text-neutral-500 dark:text-neutral-400">
                          <span>{formatBytes(f.original_size ?? 0)}</span>
                          <span className="text-neutral-200 dark:text-neutral-700">·</span>
                          <span>{dateLabel}</span>
                        </div>
                      </div>
                      <div className="shrink-0 text-right">
                        <div className="font-mono text-[12px] font-semibold tabular-nums text-neutral-950 dark:text-neutral-100">
                          {pct}%
                        </div>
                        <div
                          className={cn(
                            "mt-0.5 font-mono text-[11px] font-medium uppercase tracking-wider",
                            isDegraded ? "text-amber-700" : "text-emerald-700 dark:text-emerald-300"
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
              const pct = roundHealthPercent(f.health_status?.health_percent, 0);
              const recoverable = f.health_status?.recoverable ?? false;
              return (
                <Link
                  key={f.file_id}
                  href={`/files/${f.file_id}`}
                  className="group block border border-neutral-200 p-4 transition-colors hover:border-neutral-900 dark:border-neutral-800 dark:hover:border-neutral-600"
                >
                  <div className="mb-2.5 truncate font-mono text-[12px] font-medium text-neutral-500 dark:text-neutral-400">
                    {f.original_name}
                  </div>
                  <div className="mb-2.5 flex items-baseline gap-1">
                    <span className="text-2xl font-semibold tabular-nums text-neutral-950 dark:text-neutral-100">{pct}</span>
                    <span className="font-mono text-[12px] font-medium text-neutral-500 dark:text-neutral-400">%</span>
                  </div>
                  <div className="h-0.5 w-full bg-neutral-100 dark:bg-neutral-800">
                    <div className="h-full bg-neutral-900 dark:bg-amber-400" style={{ width: `${pct}%` }} />
                  </div>
                  <div
                    className={cn(
                      "mt-2.5 font-mono text-[11px] font-medium uppercase tracking-wider",
                      recoverable ? "text-amber-700 dark:text-amber-300" : "text-red-700 font-semibold dark:text-red-300"
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
  tone = "neutral",
}: {
  label: string;
  value: string;
  sub: string;
  warn?: boolean;
  tone?: "neutral" | "healthy" | "warning";
}) {
  return (
    <section
      className={cn(
        "border bg-white p-6 dark:border-neutral-800 dark:bg-neutral-950",
        tone === "healthy" && "border-emerald-200 dark:border-emerald-900",
        tone === "warning" && "border-amber-200 bg-amber-50/60 dark:border-amber-900 dark:bg-amber-950/20",
      )}
    >
      <p className="mb-3.5 font-mono text-[12px] font-medium uppercase tracking-[0.08em] text-neutral-500 dark:text-neutral-400">
        {label}
      </p>
      <p
        className={cn(
          "text-3xl font-semibold tracking-tight leading-none tabular-nums",
          warn && "text-amber-700 dark:text-amber-300"
        )}
      >
        {value}
      </p>
      <p className="mt-2 font-mono text-[12px] font-medium leading-relaxed text-neutral-500 dark:text-neutral-400">{sub}</p>
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
    <section className={cn("border border-neutral-200 bg-white p-6 dark:border-neutral-800 dark:bg-neutral-950", className)}>{children}</section>
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
      <span className="font-mono text-[12px] font-medium uppercase tracking-[0.08em] text-neutral-600 dark:text-neutral-300">
        {label}
      </span>
      <Link
        href={href}
        className="font-mono text-[12px] font-medium uppercase tracking-wider text-neutral-500 transition-colors hover:text-neutral-900 dark:text-neutral-400 dark:hover:text-neutral-100"
      >
        {linkLabel} →
      </Link>
    </div>
  );
}
