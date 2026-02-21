import { DashboardCard } from "./DashboardCard";

export function SystemHealth() {
  // 9 data shards, 6 parity shards (from reference)
  const shards = [
    ...Array(9).fill('data'),
    ...Array(6).fill('parity')
  ];

  return (
    <DashboardCard className="col-span-1 row-span-1">
      <div className="flex justify-between items-baseline mb-5">
        <span className="font-mono text-[11px] uppercase tracking-[0.05em] text-text-secondary">
          Shard Health (R-S 9,6)
        </span>
        <span className="font-mono text-[11px] uppercase tracking-[0.05em] text-accent-primary">
          Optimal
        </span>
      </div>

      <div className="grid grid-cols-5 gap-1 my-5">
        {shards.map((type, index) => (
          <div 
            key={index} 
            className="aspect-square relative"
            style={{
              background: type === 'data' ? 'rgba(0, 78, 235, 0.05)' : 'rgba(255, 136, 102, 0.05)',
              border: `1px solid ${type === 'data' ? 'var(--accent-primary)' : 'var(--accent-secondary)'}`,
            }}
          >
            <div 
                className="absolute top-0.5 right-0.5 w-0.5 h-0.5 opacity-50"
                style={{ backgroundColor: 'currentColor' }} 
            />
          </div>
        ))}
      </div>
      
      <div className="text-xs text-text-secondary mt-3 leading-relaxed">
        Encryption: <span className="font-mono text-text-main">AES-256-GCM</span><br />
        Client-side verification active.
      </div>
    </DashboardCard>
  );
}
