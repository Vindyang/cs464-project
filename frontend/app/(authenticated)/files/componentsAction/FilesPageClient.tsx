"use client";

import { useMemo, useState } from "react";
import { useRouter } from "next/navigation";
import { type FileMetadata } from "@/lib/api/files";
import { formatUtcDateTime } from "@/lib/utils";
import { FilesTableClient } from "./FilesTableClient";
import { toast } from "sonner";

interface FilesPageClientProps {
  initialFiles: FileMetadata[];
}

export function FilesPageClient({ initialFiles }: FilesPageClientProps) {
  const router = useRouter();
  const [refreshingAll, setRefreshingAll] = useState(false);

  const latestHealthRefreshAt = useMemo(() => {
    const timestamps = initialFiles
      .map((file) => file.last_health_refresh_at)
      .filter((value): value is string => Boolean(value))
      .sort((left, right) => new Date(right).getTime() - new Date(left).getTime());
    return timestamps[0] ?? null;
  }, [initialFiles]);

  const atRiskFiles = useMemo(
    () =>
      initialFiles.filter(
        (file) =>
          file.status === "DEGRADED" ||
          (file.health_status?.missing_shards ?? 0) > 0 ||
          (file.health_status?.corrupted_shards ?? 0) > 0,
      ),
    [initialFiles],
  );

  async function refreshAllHealth() {
    if (refreshingAll) return;
    setRefreshingAll(true);
    try {
      const res = await fetch("/api/files/health/refresh", { method: "POST" });
      const data = await res.json().catch(() => ({}));
      if (!res.ok) {
        throw new Error(data?.details || data?.message || data?.error || "Failed to refresh health");
      }
      const filesScanned = data?.files_scanned ?? 0;
      const skipped = data?.skipped_errors ?? 0;
      toast.success(`All files refreshed (${filesScanned} files, ${skipped} skipped)`);
      router.refresh();
    } catch (error) {
      const message = error instanceof Error ? error.message : "Failed to refresh health";
      toast.error(message);
    } finally {
      setRefreshingAll(false);
    }
  }

  return (
    <div className="space-y-5">
      <div className="flex items-end justify-between border-b pb-4">
        <div>
          <p className="mb-0.5 font-mono text-[11px] uppercase tracking-[0.15em] text-neutral-400">
            Storage
          </p>
          <h1 className="text-2xl font-semibold tracking-tight">Files</h1>
          <div className="mt-2 flex flex-wrap items-center gap-2 text-[11px]">
            <span className="border border-sky-200 bg-sky-50 px-2 py-1 font-mono uppercase tracking-[0.08em] text-sky-700 dark:border-sky-900 dark:bg-sky-950/60 dark:text-sky-300">
              Last Health Sync {formatUtcDateTime(latestHealthRefreshAt)}
            </span>
            {atRiskFiles.length > 0 && (
              <span className="border border-amber-200 bg-amber-50 px-2 py-1 font-mono uppercase tracking-[0.08em] text-amber-800 dark:border-amber-900 dark:bg-amber-950/60 dark:text-amber-300">
                {atRiskFiles.length} file{atRiskFiles.length === 1 ? "" : "s"} need attention
              </span>
            )}
          </div>
        </div>
        <div className="flex items-center gap-2">
          <button
            type="button"
            onClick={refreshAllHealth}
            disabled={refreshingAll}
            className="font-mono text-[11px] uppercase tracking-wider border border-sky-600 bg-sky-600 px-4 py-2 text-white transition-colors hover:bg-sky-700 disabled:cursor-not-allowed disabled:opacity-50"
          >
            {refreshingAll ? "Refreshing..." : "Refresh All Health"}
          </button>
        </div>
      </div>

      {atRiskFiles.length > 0 && (
        <section className="border border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-900 dark:border-amber-900 dark:bg-amber-950/40 dark:text-amber-200">
          <p className="font-mono text-[11px] uppercase tracking-[0.1em] text-amber-700 dark:text-amber-300">
            Health Attention
          </p>
          <p className="mt-1">
            Missing or corrupted shards were detected. Refresh health after provider recovery and inspect the affected files below.
          </p>
        </section>
      )}

      <FilesTableClient initialFiles={initialFiles} />
    </div>
  );
}
