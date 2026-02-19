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
  HardDrive
} from "lucide-react";

interface SidebarProps {
  className?: string;
  onUploadClick: () => void;
}

export function Sidebar({ className, onUploadClick }: SidebarProps) {
  // We'll need to use a hook later for active route, 
  // but for now we can just mock or use next/navigation
  
  return (
    <div className={cn("pb-12 w-64 border-r bg-background", className)}>
      <div className="space-y-4 py-4">
        <div className="px-3 py-2">
          <h2 className="mb-2 px-4 text-lg font-semibold tracking-tight flex items-center gap-2">
            <Cloud className="w-5 h-5 text-foreground" />
            Nebula Drive
          </h2>
          <div className="space-y-1">
             <Button variant="secondary" className="w-full justify-start mt-4 mb-4" onClick={onUploadClick}>
                <Plus className="mr-2 h-4 w-4" />
                New Upload
             </Button>
             
            <NavItem href="/dashboard" icon={LayoutDashboard}>
              Dashboard
            </NavItem>
            <NavItem href="/dashboard/files" icon={Folder}>
              My Files
            </NavItem>
            <NavItem href="/dashboard/providers" icon={HardDrive}>
              Providers
            </NavItem>
            <NavItem href="/dashboard/settings" icon={Settings}>
              Settings
            </NavItem>
          </div>
        </div>
        
        <div className="px-3 py-2">
          <h2 className="mb-2 px-4 text-xs font-semibold tracking-tight text-muted-foreground">
            Storage Status
          </h2>
          {/* We can put the global storage widget here later */}
          <div className="px-4 text-sm text-muted-foreground">
             <p>2 Providers connected</p>
          </div>
        </div>
      </div>
    </div>
  );
}

function NavItem({ href, icon: Icon, children }: { href: string; icon: any; children: React.ReactNode }) {
  // In a real app, check isActive
  const isActive = false; 
  
  return (
    <Button variant={isActive ? "secondary" : "ghost"} className="w-full justify-start" asChild>
      <Link href={href}>
        <Icon className="mr-2 h-4 w-4" />
        {children}
      </Link>
    </Button>
  );
}
