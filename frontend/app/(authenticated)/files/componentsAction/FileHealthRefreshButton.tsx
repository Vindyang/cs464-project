"use client";

import { useState } from "react";
import { RefreshCw } from "lucide-react";
import { toast } from "sonner";
import { helpToast } from "@/lib/help/help-toast";

interface FileHealthRefreshButtonProps {
  fileId: string;
  fileName: string;
}

export function FileHealthRefreshButton({ fileId, fileName }: FileHealthRefreshButtonProps) {
  const [refreshing, setRefreshing] = useState(false);

  async function refreshHealth() {
    if (refreshing) return;
    setRefreshing(true);
    try {
      const res = await fetch(`/api/files/${fileId}/health/refresh`, {
        method: "POST",
      });
      const data = await res.json().catch(() => ({}));
      if (!res.ok) {
        helpToast(data);
        return;
      }
      const missing = data?.marked_missing ?? 0;
      const skipped = data?.skipped_errors ?? 0;
      toast.success(`Health refreshed for ${fileName} (${missing} missing, ${skipped} skipped)`);
      window.location.reload();
    } catch (error) {
      helpToast(error);
    } finally {
      setRefreshing(false);
    }
  }

  return (
    <button
      type="button"
      disabled={refreshing}
      onClick={refreshHealth}
      className="inline-flex items-center gap-1.5 border border-sky-600 bg-sky-600 px-3 py-2 font-mono text-[11px] uppercase tracking-wider text-white transition-colors hover:bg-sky-700 disabled:cursor-not-allowed disabled:opacity-50"
    >
      <RefreshCw className={`h-3.5 w-3.5 ${refreshing ? "animate-spin" : ""}`} />
      {refreshing ? "Refreshing..." : "Refresh Health"}
    </button>
  );
}
