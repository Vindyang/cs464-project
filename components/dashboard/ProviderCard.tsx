import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Progress } from "@/components/ui/progress";
import { Cloud, Check, Loader2, AlertTriangle, RefreshCw } from "lucide-react";
import { cn } from "@/lib/utils";

interface ProviderCardProps {
  id: string;
  name: string;
  status: "connected" | "disconnected" | "error";
  used: string;
  total: string;
  percentage: number;
  shardCount: number;
  lastCheck: string;
  onConnect: () => void;
  onDisconnect: () => void;
  onRefresh: () => void;
}

export function ProviderCard({
  name,
  status,
  used,
  total,
  percentage,
  shardCount,
  lastCheck,
  onConnect,
  onDisconnect,
  onRefresh,
}: ProviderCardProps) {
  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between pb-2">
        <CardTitle className="text-sm font-medium flex items-center gap-2">
          <Cloud className="w-4 h-4" />
          {name}
        </CardTitle>
        <StatusBadge status={status} />
      </CardHeader>
      <CardContent>
        {status === "connected" ? (
          <div className="space-y-4">
            <div className="space-y-2">
              <div className="flex justify-between text-xs text-muted-foreground">
                <span>Storage Used</span>
                <span>{used} / {total}</span>
              </div>
              <Progress value={percentage} className="h-2" />
            </div>
            
            <div className="grid grid-cols-2 gap-2 text-xs">
              <div className="p-2 bg-muted rounded-md text-center">
                <div className="font-semibold text-foreground">{shardCount}</div>
                <div className="text-muted-foreground">Shards</div>
              </div>
              <div className="p-2 bg-muted rounded-md text-center">
                <div className="font-semibold text-foreground">Health</div>
                <div className="text-muted-foreground">{lastCheck}</div>
              </div>
            </div>

            <div className="flex gap-2">
              <Button variant="outline" size="sm" className="flex-1" onClick={onDisconnect}>
                Disconnect
              </Button>
              <Button variant="ghost" size="icon" className="h-8 w-8" onClick={onRefresh}>
                <RefreshCw className="w-4 h-4" />
              </Button>
            </div>
          </div>
        ) : (
          <div className="flex flex-col items-center justify-center py-6 gap-3">
             <p className="text-sm text-muted-foreground text-center">
                Connect your account to enable redundancy
             </p>
             <Button size="sm" onClick={onConnect}>
                Connect {name}
             </Button>
          </div>
        )}
      </CardContent>
    </Card>
  );
}

function StatusBadge({ status }: { status: ProviderCardProps["status"] }) {
  const styles = {
    connected: "bg-primary/10 text-primary hover:bg-primary/20",
    disconnected: "bg-muted text-muted-foreground hover:bg-muted/80",
    error: "bg-destructive/10 text-destructive hover:bg-destructive/20",
  };

  const icons = {
    connected: Check,
    disconnected: Loader2, // Or a plug icon
    error: AlertTriangle,
  };

  const Icon = icons[status];

  return (
    <Badge variant="outline" className={cn("gap-1.5", styles[status])}>
      <Icon className="w-3 h-3" />
      <span className="capitalize">{status}</span>
    </Badge>
  );
}
