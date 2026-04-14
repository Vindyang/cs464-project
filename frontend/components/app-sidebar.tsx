"use client";

import * as React from "react";
import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import {
  Database,
  FileText,
  FolderOpen,
  HelpCircle,
  History,
  KeyRound,
  LayoutDashboard,
  Settings,
  Upload,
} from "lucide-react";
import { cn } from "@/lib/utils";
import { ProvidersUploadFilesModal } from "@/components/dashboard/ProvidersUploadFilesModal";
import { ThemeToggle } from "@/components/theme-toggle";
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarRail,
} from "@/components/ui/sidebar";

const navSections = [
  {
    label: "Getting Started",
    items: [{ title: "Quick Start", href: "/quick-start", icon: FileText }],
  },
  {
    label: "Workspace",
    items: [
      { title: "Dashboard", href: "/dashboard", icon: LayoutDashboard },
      { title: "Files", href: "/files", icon: FolderOpen },
      { title: "History", href: "/history", icon: History },
    ],
  },
  {
    label: "Configuration",
    items: [
      { title: "Providers", href: "/providers", icon: Database },
      { title: "Credentials", href: "/credentials", icon: KeyRound },
      { title: "Settings", href: "/settings", icon: Settings },
    ],
  },
  {
    label: "Documentation",
    items: [{ title: "Help", href: "/help", icon: HelpCircle }],
  },
];

function formatBytes(bytes: number): string {
  if (bytes === 0) return "0 B";
  const k = 1024;
  const sizes = ["B", "KB", "MB", "GB", "TB"];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(1))} ${sizes[i]}`;
}

interface AppSidebarProps extends React.ComponentProps<typeof Sidebar> {
  totalStorageUsedBytes: number;
  totalStorageTotalBytes: number;
}

export function AppSidebar({
  totalStorageUsedBytes,
  totalStorageTotalBytes,
  ...props
}: AppSidebarProps) {
  const pathname = usePathname();
  const router = useRouter();
  const [uploadModalOpen, setUploadModalOpen] = React.useState(false);
  const usedPct =
    totalStorageTotalBytes > 0
      ? Math.round((totalStorageUsedBytes / totalStorageTotalBytes) * 100)
      : 0;

  return (
    <Sidebar collapsible="offcanvas" {...props}>
      {/* Logo */}
      <SidebarHeader className="h-14 flex-row items-center border-b px-4 !py-0">
        <div className="flex items-center gap-2.5">
          <div className="flex h-5 w-5 items-center justify-center border-2 border-neutral-900 dark:border-neutral-100 shrink-0 transition-colors">
            <div className="h-1.5 w-1.5 bg-neutral-900 dark:bg-neutral-100 transition-colors" />
          </div>
          <span className="font-mono text-[12px] font-bold uppercase tracking-[0.15em] text-neutral-900 dark:text-neutral-100 transition-colors">
            Omnishard
          </span>
        </div>
      </SidebarHeader>

      {/* Navigation */}
      <SidebarContent className="px-2 py-4">
        <SidebarGroup>
          <SidebarGroupContent>
            <button
              type="button"
              onClick={() => setUploadModalOpen(true)}
              className="flex h-10 w-full cursor-pointer items-center justify-between border border-sky-600 bg-sky-600 px-3 font-mono text-[11px] font-semibold uppercase tracking-[0.08em] text-white transition-colors hover:bg-sky-700"
            >
              <span>Upload Files</span>
              <Upload className="h-3.5 w-3.5" />
            </button>
          </SidebarGroupContent>
        </SidebarGroup>
        {navSections.map((section) => (
          <SidebarGroup key={section.label} className="pb-0">
            <SidebarGroupLabel className="h-6 px-2 font-mono text-[10px] font-semibold uppercase tracking-[0.12em] text-neutral-400 dark:text-neutral-500">
              {section.label}
            </SidebarGroupLabel>
            <SidebarGroupContent>
              <SidebarMenu>
                {section.items.map((item) => {
                  const isActive =
                    item.href === "/dashboard"
                      ? pathname === "/dashboard" || pathname === "/"
                      : pathname.startsWith(item.href);
                  const Icon = item.icon;
                  return (
                    <SidebarMenuItem key={item.href}>
                      <SidebarMenuButton
                        asChild
                        isActive={isActive}
                        className={cn(
                          "h-9 font-mono text-[12px] font-medium uppercase tracking-[0.08em] rounded-none",
                          "text-neutral-700 hover:text-neutral-950 hover:bg-neutral-100 dark:text-neutral-200 dark:hover:bg-sky-950/50 dark:hover:text-sky-100",
                          "data-[active=true]:bg-neutral-900 data-[active=true]:text-white data-[active=true]:hover:bg-neutral-800 dark:data-[active=true]:bg-sky-500 dark:data-[active=true]:text-white dark:data-[active=true]:hover:bg-sky-400",
                        )}
                      >
                        <Link href={item.href} className="cursor-pointer">
                          <Icon />
                          <span>{item.title}</span>
                        </Link>
                      </SidebarMenuButton>
                    </SidebarMenuItem>
                  );
                })}
              </SidebarMenu>
            </SidebarGroupContent>
          </SidebarGroup>
        ))}
      </SidebarContent>

      {/* Storage footer */}
      <SidebarFooter className="border-t p-5">
        <div className="space-y-2.5">
          <p className="font-mono text-[11px] font-medium uppercase tracking-[0.1em] text-neutral-500">
            Storage Used
          </p>
          <div className="flex items-baseline gap-1">
            <span className="font-mono text-base font-semibold tabular-nums">
              {formatBytes(totalStorageUsedBytes)}
            </span>
            {totalStorageTotalBytes > 0 && (
              <span className="font-mono text-[12px] font-medium text-neutral-500">
                / {formatBytes(totalStorageTotalBytes)}
              </span>
            )}
          </div>
          <div className="h-1 w-full bg-neutral-100">
            <div
              className="h-full bg-sky-600 transition-all"
              style={{ width: `${usedPct}%` }}
            />
          </div>
          <p className="font-mono text-[11px] font-medium text-neutral-500">
            {usedPct}% utilized
          </p>
        </div>
        <ThemeToggle />
      </SidebarFooter>

      <SidebarRail />

      <ProvidersUploadFilesModal
        open={uploadModalOpen}
        onOpenChange={setUploadModalOpen}
        onUploadSuccess={() => router.refresh()}
      />
    </Sidebar>
  );
}
