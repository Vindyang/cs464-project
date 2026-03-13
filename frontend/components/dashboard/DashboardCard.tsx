import { cn } from "@/lib/utils";

interface DashboardCardProps {
  children: React.ReactNode;
  className?: string;
}

export function DashboardCard({ children, className }: DashboardCardProps) {
  return (
    <div
      className={cn(
        "relative flex flex-col border bg-white p-5",
        className
      )}
    >
      {/* Corner marks for technical aesthetic */}
      <div className="pointer-events-none absolute -left-px -top-px h-1.5 w-1.5 border-l border-t border-neutral-300" />
      <div className="pointer-events-none absolute -right-px -top-px h-1.5 w-1.5 border-r border-t border-neutral-300" />
      <div className="pointer-events-none absolute -bottom-px -left-px h-1.5 w-1.5 border-b border-l border-neutral-300" />
      <div className="pointer-events-none absolute -bottom-px -right-px h-1.5 w-1.5 border-b border-r border-neutral-300" />

      {children}
    </div>
  );
}
