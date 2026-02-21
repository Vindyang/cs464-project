import { DashboardCard } from "./DashboardCard";
import { cn } from "@/lib/utils";

const providers = [
  { icon: 'AWS', name: 'Amazon S3', region: 'us-east-1', latency: '24ms', usage: 78, status: 'green' },
  { icon: 'GCP', name: 'Google Drive', region: 'global', latency: '45ms', usage: 42, status: 'green' },
  { icon: 'DBX', name: 'Dropbox', region: 'eu-west', latency: '112ms', usage: 15, status: 'blue' },
  { icon: 'B2', name: 'Backblaze B2', region: 'us-west', latency: '89ms', usage: 91, status: 'green' },
  { icon: 'MS', name: 'OneDrive', region: 'us-east', latency: '320ms', usage: 5, status: 'orange' }
];

const statusColors: Record<string, string> = {
  green: '#10B981',
  blue: 'var(--accent-primary)',
  orange: 'var(--accent-secondary)'
};

export function ProviderMatrix() {
  return (
    <div className="col-span-1 row-span-2 flex flex-col gap-6">
      <div className="flex justify-between items-center mb-4">
        <h2 className="text-lg font-semibold tracking-[-0.02em]">Storage Providers</h2>
        <button className="inline-flex items-center justify-center px-5 h-8 text-xs font-medium transition-all rounded-[2px] border border-border-color text-text-main hover:bg-bg-subtle">
           Manage Providers
        </button>
      </div>

      <div className="border border-border-color bg-bg-canvas p-0">
         <div className="grid grid-cols-[40px_2fr_1fr_1fr_1fr] px-5 items-center bg-bg-subtle h-10 text-[11px] uppercase text-text-secondary tracking-[0.05em]">
            <div></div>
            <div>Provider</div>
            <div>Region</div>
            <div>Latency</div>
            <div>Usage</div>
         </div>

         {providers.map((provider, index) => (
            <div key={index} className={cn(
               "grid grid-cols-[40px_2fr_1fr_1fr_1fr] px-5 py-4 items-center transition-colors hover:bg-bg-subtle/50",
               index < providers.length - 1 ? "border-b border-border-color" : ""
            )}>
               <div className="w-6 h-6 bg-[#F0F0F0] rounded-[2px] flex items-center justify-center text-[10px] font-bold text-text-secondary">
                  {provider.icon}
               </div>
               <div className="font-medium text-[13px]">{provider.name}</div>
               <div className="font-mono text-xs text-text-secondary">{provider.region}</div>
               <div className="inline-flex items-center gap-1.5 text-xs text-text-secondary">
                  <div 
                     className="w-1.5 h-1.5 rounded-full" 
                     style={{ backgroundColor: statusColors[provider.status] }} 
                  />
                  {provider.latency}
               </div>
               <div className="w-[60px] h-1 bg-[#eee] relative overflow-hidden rounded-full">
                  <div 
                     className="h-full bg-accent-primary" 
                     style={{ width: `${provider.usage}%` }} 
                  />
               </div>
            </div>
         ))}
      </div>
    </div>
  );
}
