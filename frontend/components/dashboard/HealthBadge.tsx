import { Badge } from "@/components/ui/badge";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";
import { Check, AlertTriangle, XCircle } from "lucide-react";
import { cn } from "@/lib/utils";

interface HealthBadgeProps {
  shardsAvailable: number;
  shardsTotal: number;
  minShards: number;
}

export function HealthBadge({ shardsAvailable, shardsTotal, minShards }: HealthBadgeProps) {
  let status: "healthy" | "warning" | "critical" = "healthy";
  if (shardsAvailable < minShards) {
    status = "critical";
  } else if (shardsAvailable < shardsTotal) {
    status = "warning";
  }

  const variants = {
    healthy: "bg-primary/10 text-primary hover:bg-primary/25 border-primary/20",
    warning: "bg-muted text-muted-foreground hover:bg-muted/80 border-border",
    critical: "bg-destructive/10 text-destructive hover:bg-destructive/25 border-destructive/20",
  };

  const icons = {
    healthy: Check,
    warning: AlertTriangle,
    critical: XCircle,
  };

  const Icon = icons[status];

  return (
    <Tooltip>
      <TooltipTrigger asChild>
        <Badge variant="outline" className={cn("gap-1.5 pr-2.5", variants[status])}>
          <Icon className="w-3.5 h-3.5" />
          <span>
            {shardsAvailable}/{shardsTotal} Shards
          </span>
        </Badge>
      </TooltipTrigger>
      <TooltipContent>
        <p className="text-xs font-medium">
          {status === "healthy" && "File is fully replicated and healthy"}
          {status === "warning" && `File is recoverable but missing ${shardsTotal - shardsAvailable} shards`}
          {status === "critical" && "File cannot be recovered (below threshold)"}
        </p>
      </TooltipContent>
    </Tooltip>
  );
}
