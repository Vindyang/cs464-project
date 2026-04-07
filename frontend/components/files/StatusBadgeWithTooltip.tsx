"use client";

import { Info } from "lucide-react";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";
import { FileHealthStatus } from "@/lib/api/files";

interface StatusBadgeWithTooltipProps {
  status: string;
  healthStatus?: FileHealthStatus;
}

function getStatusLabel(status: string): string {
  if (status === "UPLOADED") return "HEALTHY";
  return status;
}

function getStatusTooltip(status: string, healthStatus?: FileHealthStatus): string {
  const healthy = healthStatus?.healthy_shards ?? 0;
  const total = healthStatus?.total_shards ?? 6;
  switch (status) {
    case "PENDING":
      return "File is being uploaded and processed";
    case "UPLOADED":
      return "All shards are healthy and fully replicated";
    case "DEGRADED":
      return `Some shards are missing but file is still recoverable (${healthy}/${total} shards healthy)`;
    case "CORRUPTED":
      return "Too many shards lost — file cannot be recovered";
    default:
      return status;
  }
}

export function StatusBadgeWithTooltip({ status, healthStatus }: StatusBadgeWithTooltipProps) {
  return (
    <div className="inline-flex items-center gap-1.5">
      <span className="font-mono text-sm text-neutral-700 dark:text-neutral-200">
        {getStatusLabel(status)}
      </span>
      <Tooltip>
        <TooltipTrigger asChild>
          <button type="button" className="text-neutral-400 hover:text-neutral-600 dark:hover:text-neutral-300">
            <Info className="h-3.5 w-3.5" />
          </button>
        </TooltipTrigger>
        <TooltipContent>
          <p>{getStatusTooltip(status, healthStatus)}</p>
        </TooltipContent>
      </Tooltip>
    </div>
  );
}
