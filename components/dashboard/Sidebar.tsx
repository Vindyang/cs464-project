import Link from "next/link";
import { usePathname } from "next/navigation";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import { 
  Folder, 
  Cloud, 
  Settings, 
  LayoutDashboard, 
  Plus, 
  HardDrive,
  Activity
} from "lucide-react";

interface SidebarProps {
  className?: string;
  onUploadClick: () => void;
}

export function Sidebar({ className, onUploadClick }: SidebarProps) {
  return (
    <div className={cn("pb-12 w-64 border-r border-sidebar-border bg-sidebar flex flex-col", className)}>
      <div className="space-y-4 py-4 flex-1">
        <div className="px-3 py-2">
          <div className="mb-6 px-4 flex items-center gap-2 font-semibold tracking-tight text-sidebar-foreground">
            <div className="relative flex h-6 w-6 items-center justify-center border border-sidebar-foreground/20 rounded-sm">
               <div className="h-2 w-2 bg-sidebar-primary rounded-[1px]" />
            </div>
            <span>ZERO-STORE</span>
          </div>
          
          <div className="space-y-1">
             <Button 
                className="w-full justify-center mb-6 bg-sidebar-primary text-sidebar-primary-foreground hover:bg-sidebar-primary/90 rounded-[2px] h-9 text-sm font-medium" 
                onClick={onUploadClick}
             >
                <Plus className="mr-2 h-4 w-4" />
                Upload File
             </Button>
             
            <NavItem href="/dashboard" icon={LayoutDashboard}>
              Dashboard
            </NavItem>
            <NavItem href="/nodes" icon={Activity}>
              Nodes
            </NavItem>
            <NavItem href="/files" icon={Folder}>
              My Files
            </NavItem>
            <NavItem href="/providers" icon={HardDrive}>
              Providers
            </NavItem>
            <NavItem href="/settings" icon={Settings}>
              Settings
            </NavItem>
          </div>
        </div>
        
        <div className="px-3 py-2 mt-auto">
          <div className="rounded-sm border border-sidebar-border bg-sidebar-accent/50 p-4">
            <h3 className="text-[10px] uppercase tracking-wider text-muted-foreground font-mono mb-2">
                Storage Used
            </h3>
            <div className="flex items-baseline gap-1">
                <span className="text-2xl font-semibold text-foreground">4.2</span>
                <span className="text-sm text-muted-foreground">TB</span>
            </div>
            <div className="w-full bg-sidebar-border h-1 mt-2 rounded-full overflow-hidden">
                <div className="h-full bg-sidebar-primary w-[35%]" />
            </div>
            <p className="text-[10px] text-muted-foreground mt-2">
                35% of 12.0 TB Used
            </p>
          </div>
        </div>
      </div>
    </div>
  );
}

function NavItem({ href, icon: Icon, children }: { href: string; icon: any; children: React.ReactNode }) {
  const pathname = usePathname();
  // Simple active check: exact match or starts with (for nested routes)
  // Special case for dashboard root
  const isActive = href === "/dashboard" 
    ? pathname === "/dashboard" || pathname === "/"
    : pathname.startsWith(href);
  
  return (
    <Link 
      href={href}
      className={cn(
        "flex items-center gap-3 rounded-[2px] px-3 py-2 text-sm transition-all",
        isActive 
          ? "bg-sidebar-accent text-sidebar-foreground font-medium" 
          : "text-muted-foreground hover:text-foreground hover:bg-sidebar-accent/50"
      )}
    >
      <Icon className={cn("h-4 w-4", isActive ? "text-sidebar-foreground" : "text-muted-foreground")} />
      {children}
    </Link>
  );
}
