import { DashboardCard } from "./DashboardCard";

const specs = [
    { label: 'Protocol', value: 'ZK-Rollup' },
    { label: 'Redundancy', value: '1.5x' },
    { label: 'Block Size', value: '4MB' },
    { label: 'Active Nodes', value: '142', highlight: true },
    { label: 'API Version', value: 'v2.4.0' }
];

export function TechSpecs() {
  return (
    <DashboardCard className="col-span-1 row-span-1">
      <div className="flex justify-between items-baseline mb-5">
        <span className="font-mono text-[11px] uppercase tracking-[0.05em] text-text-secondary">
           Configuration
        </span>
      </div>
      
      {specs.map((spec, index) => (
        <div key={index} className={`flex justify-between py-3 text-[13px] ${index < specs.length - 1 ? 'border-b border-grid-line' : ''}`}>
          <span className="text-text-secondary">{spec.label}</span>
          <span className={`font-mono ${spec.highlight ? 'text-accent-primary' : 'text-text-main'}`}>
             {spec.value}
          </span>
        </div>
      ))}
    </DashboardCard>
  );
}
