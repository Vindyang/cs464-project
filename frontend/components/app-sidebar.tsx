"use client";

import * as React from "react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { cn } from "@/lib/utils";
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupContent,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarRail,
} from "@/components/ui/sidebar";

const navItems = [
  { title: "Dashboard", href: "/dashboard" },
  { title: "Files", href: "/files" },
  { title: "Nodes", href: "/nodes" },
  { title: "Providers", href: "/providers" },
  { title: "Credentials", href: "/credentials" },
  { title: "Settings", href: "/settings" },
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
  const usedPct =
    totalStorageTotalBytes > 0
      ? Math.round((totalStorageUsedBytes / totalStorageTotalBytes) * 100)
      : 0;

  return (
    <Sidebar collapsible="offcanvas" {...props}>
      {/* Logo */}
      <SidebarHeader className="h-12 flex-row items-center border-b px-4 !py-0">
        <div className="flex items-center gap-2">
          <div className="flex h-5 w-5 items-center justify-center border-2 border-neutral-900 shrink-0">
            <div className="h-1.5 w-1.5 bg-neutral-900" />
          </div>
          <span className="font-mono text-[11px] font-bold uppercase tracking-[0.15em] text-neutral-900">
            Nebula Drive
          </span>
        </div>
      </SidebarHeader>

      {/* Navigation */}
      <SidebarContent className="px-2 py-3">
        <SidebarGroup>
          <SidebarGroupContent>
            <SidebarMenu>
              {navItems.map((item) => {
                const isActive =
                  item.href === "/dashboard"
                    ? pathname === "/dashboard" || pathname === "/"
                    : pathname.startsWith(item.href);
                return (
                  <SidebarMenuItem key={item.href}>
                    <SidebarMenuButton
                      asChild
                      isActive={isActive}
                      className={cn(
                        "font-mono text-[11px] uppercase tracking-[0.08em] rounded-none",
                        "text-neutral-500 hover:text-neutral-900 hover:bg-neutral-100",
                        "data-[active=true]:bg-neutral-900 data-[active=true]:text-white data-[active=true]:hover:bg-neutral-800",
                      )}
                    >
                      <Link href={item.href}>{item.title}</Link>
                    </SidebarMenuButton>
                  </SidebarMenuItem>
                );
              })}
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>
      </SidebarContent>

      {/* Storage footer */}
      <SidebarFooter className="border-t p-4">
        <div className="space-y-2">
          <p className="font-mono text-[9px] uppercase tracking-[0.1em] text-neutral-400">
            Storage Used
          </p>
          <div className="flex items-baseline gap-1">
            <span className="font-mono text-sm font-semibold tabular-nums">
              {formatBytes(totalStorageUsedBytes)}
            </span>
            {totalStorageTotalBytes > 0 && (
              <span className="font-mono text-[10px] text-neutral-400">
                / {formatBytes(totalStorageTotalBytes)}
              </span>
            )}
          </div>
          <div className="h-1 w-full bg-neutral-100">
            <div
              className="h-full bg-neutral-900 transition-all"
              style={{ width: `${usedPct}%` }}
            />
          </div>
          <p className="font-mono text-[9px] text-neutral-400">
            {usedPct}% utilized
          </p>
        </div>
      </SidebarFooter>

      <SidebarRail />
    </Sidebar>
  );
}
