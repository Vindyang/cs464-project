import { cn } from "@/lib/utils";

interface DashboardCardProps {
  children: React.ReactNode;
  className?: string;
}

export function DashboardCard({ children, className }: DashboardCardProps) {
  return (
    <div className={cn("relative bg-bg-canvas border border-border-color p-5 flex flex-col", className)}>
      {/* Top Left Corner Mark */}
      <div className="absolute -top-px -left-px w-1.5 h-1.5 border-t border-l border-text-tertiary/50 pointer-events-none" />
      {/* Top Right Corner Mark */}
      <div className="absolute -top-px -right-px w-1.5 h-1.5 border-t border-r border-text-tertiary/50 pointer-events-none" />
      {/* Bottom Left Corner Mark */}
      <div className="absolute -bottom-px -left-px w-1.5 h-1.5 border-b border-l border-text-tertiary/50 pointer-events-none" />
      {/* Bottom Right Corner Mark */}
      <div className="absolute -bottom-px -right-px w-1.5 h-1.5 border-b border-r border-text-tertiary/50 pointer-events-none" />
      
      {children}
    </div>
  );
}
