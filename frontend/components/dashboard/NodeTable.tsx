import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Cloud, Check, Loader2, AlertTriangle, MoreHorizontal, Settings, Globe, Trash2 } from "lucide-react";
import { DashboardCard } from "./DashboardCard";
import { cn } from "@/lib/utils";

import { ProviderData } from "@/lib/mocks/providers";

interface NodeTableProps {
  providers: ProviderData[];
  onConfig: (id: string) => void;
  onRemove: (id: string) => void;
}

export function NodeTable({ providers, onConfig, onRemove }: NodeTableProps) {
  return (
    <DashboardCard className="p-0 overflow-hidden bg-bg-canvas border-border-color">
      <div className="overflow-x-auto">
        <table className="w-full text-left text-sm">
          <thead>
            <tr className="bg-bg-subtle border-b border-border-color">
              <th className="h-10 px-4 font-mono text-[11px] uppercase tracking-wider text-text-tertiary font-medium">Provider</th>
              <th className="h-10 px-4 font-mono text-[11px] uppercase tracking-wider text-text-tertiary font-medium">Type</th>
              <th className="h-10 px-4 font-mono text-[11px] uppercase tracking-wider text-text-tertiary font-medium">Status</th>
              <th className="h-10 px-4 font-mono text-[11px] uppercase tracking-wider text-text-tertiary font-medium">Region</th>
              <th className="h-10 px-4 font-mono text-[11px] uppercase tracking-wider text-text-tertiary font-medium">Usage</th>
               <th className="h-10 px-4 font-mono text-[11px] uppercase tracking-wider text-text-tertiary font-medium">Latency</th>
              <th className="h-10 px-4 font-mono text-[11px] uppercase tracking-wider text-text-tertiary font-medium text-right">Actions</th>
            </tr>
          </thead>
          <tbody>
            {providers.map((provider) => {
              const usagePercent = provider.quotaTotalBytes > 0 
                ? Math.round((provider.quotaUsedBytes / provider.quotaTotalBytes) * 100) 
                : 0;
              
              const formattedUsed = (provider.quotaUsedBytes / (1024 * 1024 * 1024)).toFixed(1) + " GB";

              return (
                <tr 
                  key={provider.providerId} 
                  className="border-b border-border-color last:border-0 hover:bg-bg-subtle/30 transition-colors group"
                >
                  <td className="p-4 align-middle">
                     <div className="flex items-center gap-3">
                        <div className="w-8 h-8 flex items-center justify-center bg-bg-subtle rounded-[2px] text-text-secondary border border-border-color">
                           <Cloud className="w-4 h-4" />
                        </div>
                        <span className="font-medium text-text-main">
                           {provider.displayName}
                        </span>
                     </div>
                  </td>
                  <td className="p-4 align-middle font-mono text-xs text-text-secondary">
                     Object Storage
                  </td>
                  <td className="p-4 align-middle">
                     <StatusBadge status={provider.status} />
                  </td>
                  <td className="p-4 align-middle">
                     <div className="flex items-center gap-1.5 text-xs text-text-secondary">
                        <Globe className="w-3.5 h-3.5 text-text-tertiary" />
                        <span>{provider.region}</span>
                     </div>
                  </td>
                  <td className="p-4 align-middle w-48">
                     <div className="flex flex-col gap-1">
                        <div className="flex justify-between text-[10px] text-text-secondary uppercase tracking-wider">
                           <span>{usagePercent}%</span>
                           <span>{formattedUsed}</span>
                        </div>
                        <div className="h-1 w-full bg-bg-subtle rounded-full overflow-hidden">
                           <div 
                              className="h-full bg-accent-primary" 
                              style={{ width: `${usagePercent}%` }} 
                           />
                        </div>
                     </div>
                  </td>
                  <td className="p-4 align-middle font-mono text-xs text-text-primary">
                      {provider.latencyMs}ms
                  </td>
                  <td className="p-4 align-middle text-right">
                     <div className="flex justify-end gap-2 opacity-0 group-hover:opacity-100 transition-opacity">
                        <Button variant="ghost" size="icon" className="h-8 w-8 hover:bg-bg-subtle hover:text-accent-primary" onClick={() => onConfig(provider.providerId)}>
                           <Settings className="w-4 h-4" />
                        </Button>
                        <Button variant="ghost" size="icon" className="h-8 w-8 hover:bg-bg-subtle hover:text-destructive" onClick={() => onRemove(provider.providerId)}>
                           <Trash2 className="w-4 h-4" />
                        </Button>
                     </div>
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>
      {providers.length === 0 && (
         <div className="p-12 text-center text-text-secondary">
            No providers connected. Connect a provider to get started.
         </div>
      )}
    </DashboardCard>
  );
}

function StatusBadge({ status }: { status: string }) {
  const styles: Record<string, string> = {
    connected: "text-emerald-600 bg-emerald-50 border-emerald-200",
    degraded: "text-amber-600 bg-amber-50 border-amber-200",
    disconnected: "text-text-tertiary bg-bg-subtle border-border-color",
    error: "text-red-600 bg-red-50 border-red-200",
  };

  return (
    <div className={cn("inline-flex items-center gap-1.5 px-2 py-0.5 rounded-[2px] border text-[10px] font-medium uppercase tracking-wide", styles[status])}>
      <div className={`w-1.5 h-1.5 rounded-full ${
        status === 'connected' ? 'bg-emerald-500' : 
        status === 'degraded' ? 'bg-amber-500' :
        status === 'error' ? 'bg-red-500' : 
        'bg-gray-400'}`} 
      />
      <span>{status}</span>
    </div>
  );
}
