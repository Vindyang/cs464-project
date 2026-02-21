import { DashboardCard } from "./DashboardCard";

const activities = [
  { type: 'UPLOAD', time: '10:42 AM', description: 'project-alpha-v2.zip', highlight: false },
  { type: 'SELF-HEAL', time: '09:15 AM', description: 'Repaired shard on OneDrive', highlight: true },
  { type: 'DOWNLOAD', time: '08:30 AM', description: 'database-backup.sql', highlight: false },
  { type: 'SYNC', time: '08:00 AM', description: 'Provider latency check', highlight: false }
];

export function ActivityFeed() {
  return (
    <DashboardCard className="col-span-1 row-span-1">
      <div className="flex justify-between items-baseline mb-5">
        <span className="font-mono text-[11px] uppercase tracking-[0.05em] text-text-secondary">
           Recent Activity
        </span>
      </div>
      
      <ul className="mt-4 list-none">
        {activities.map((activity, index) => (
          <li key={index} className={`py-3 text-[13px] ${index < activities.length - 1 ? 'border-b border-grid-line' : ''}`}>
            <div className="flex justify-between text-text-tertiary text-[11px] mb-1 font-mono">
              <span style={{ color: activity.highlight ? 'var(--accent-secondary)' : 'inherit' }}>
                 {activity.type}
              </span>
              <span>{activity.time}</span>
            </div>
            <div>{activity.description}</div>
          </li>
        ))}
      </ul>
    </DashboardCard>
  );
}
