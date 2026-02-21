import { Progress } from "@/components/ui/progress";
import { cn } from "@/lib/utils";
import { Check, Upload, Loader2, Pause, X } from "lucide-react";

export interface ShardStatus {
  index: number;
  provider: string;
  status: "pending" | "uploading" | "complete" | "failed" | "paused";
  progress: number;
}

interface ShardProgressBarProps {
  shards: ShardStatus[];
  label?: string;
}

export function ShardProgressBar({ shards, label }: ShardProgressBarProps) {
  return (
    <div className="space-y-3">
      {label && <h4 className="text-sm font-medium leading-none">{label}</h4>}
      <div className="grid gap-2">
        {shards.map((shard) => (
          <div
            key={shard.index}
            className="flex items-center gap-3 text-sm p-2 rounded-md bg-muted/40 border border-border/50"
          >
            <StatusIcon status={shard.status} />
            <div className="flex-1 min-w-0">
              <div className="flex justify-between mb-1">
                <span className="font-medium truncate">
                  Shard {shard.index + 1}
                  <span className="text-muted-foreground font-normal ml-1">
                    → {shard.provider}
                  </span>
                </span>
                <span className="text-muted-foreground tabular-nums">
                  {shard.progress}%
                </span>
              </div>
              <Progress value={shard.progress} className="h-1.5" />
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

function StatusIcon({ status }: { status: ShardStatus["status"] }) {
  const styles = {
    pending: "text-muted-foreground bg-muted",
    uploading: "text-primary bg-primary/10",
    complete: "text-primary bg-primary/10",
    failed: "text-destructive bg-destructive/10",
    paused: "text-muted-foreground bg-muted/50",
  };

  const icons = {
    pending: Upload,
    uploading: Loader2,
    complete: Check,
    failed: X,
    paused: Pause,
  };

  const Icon = icons[status];

  return (
    <div className={cn("flex items-center justify-center w-8 h-8 rounded-full", styles[status])}>
      <Icon className={cn("w-4 h-4", status === "uploading" && "animate-spin")} />
    </div>
  );
}
