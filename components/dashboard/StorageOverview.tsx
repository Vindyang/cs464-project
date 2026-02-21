import { DashboardCard } from "./DashboardCard";

export function StorageOverview() {
  return (
    <DashboardCard className="col-span-1 row-span-1">
      <div className="flex justify-between items-baseline mb-5">
        <span className="font-mono text-[11px] uppercase tracking-[0.05em] text-text-secondary">
            Storage Used
        </span>
        <span className="font-mono text-[11px] uppercase tracking-[0.05em] text-accent-primary">
            Healthy
        </span>
      </div>
      
      <div className="text-5xl font-semibold tracking-[-0.04em] leading-none mb-1">
        4.2<span className="text-xl text-text-tertiary ml-1">TB</span>
      </div>
      <div className="text-[13px] text-text-secondary mt-1">
        of 12.0 TB Total Capacity
      </div>

      <div className="mt-6">
        <div className="flex justify-between text-xs mb-2">
          <span>Usage Distribution</span>
          <span className="font-mono">35%</span>
        </div>
        <div className="h-1 bg-grid-line relative overflow-hidden">
          <div className="h-full bg-accent-primary w-[35%] transition-all duration-1000 ease-out" />
        </div>
      </div>

      <div className="grid grid-cols-2 gap-4 mt-6 pt-6 border-t border-grid-line">
        <div>
          <div className="font-mono text-[11px] uppercase tracking-[0.05em] text-text-secondary mb-1">
            Objects
          </div>
          <div className="text-2xl font-medium tracking-[-0.03em]">842k</div>
        </div>
        <div>
          <div className="font-mono text-[11px] uppercase tracking-[0.05em] text-text-secondary mb-1">
            Providers
          </div>
          <div className="text-2xl font-medium tracking-[-0.03em]">5</div>
        </div>
      </div>
    </DashboardCard>
  );
}
