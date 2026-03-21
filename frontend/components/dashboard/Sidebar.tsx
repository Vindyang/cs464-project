"use client";

import React from "react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import {
  Folder,
  Settings,
  LayoutDashboard,
  Plus,
  HardDrive,
  Activity,
} from "lucide-react";

interface SidebarProps {
  className?: string;
  onUploadClick: () => void;
}

export function Sidebar({ className, onUploadClick }: SidebarProps) {
  return (
    <div
      className={cn(
        "flex w-64 flex-col border-r bg-white",
        className
      )}
    >
      <div className="flex flex-1 flex-col space-y-4 py-4">
        {/* Logo */}
        <div className="px-6 py-2">
          <div className="mb-6 flex items-center gap-2">
            <div className="flex h-6 w-6 items-center justify-center border-2 border-black">
              <div className="h-2 w-2 bg-black" />
            </div>
            <span className="font-mono text-sm font-bold tracking-wider">
              NEBULA_DRIVE
            </span>
          </div>

          {/* Upload Button */}
          <Button
            className="mb-6 w-full font-mono text-xs"
            onClick={onUploadClick}
          >
            <Plus className="mr-2 h-4 w-4" />
            UPLOAD FILE
          </Button>

          {/* Navigation */}
          <div className="space-y-1">
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

        {/* Storage Stats */}
        <div className="mt-auto px-6">
          <div className="border bg-neutral-50 p-4">
            <h3 className="mb-3 font-mono text-[10px] uppercase tracking-wider text-neutral-500">
              Storage Used
            </h3>

            <div className="mb-2 flex items-baseline gap-1">
              <span className="font-mono text-2xl font-bold">4.2</span>
              <span className="font-mono text-xs text-neutral-500">TB</span>
              <span className="ml-auto font-mono text-xs text-neutral-500">
                / 12.0 TB
              </span>
            </div>

            {/* Progress bar */}
            <div className="mb-2 h-1.5 w-full overflow-hidden bg-neutral-200">
              <div
                className="h-full bg-black"
                style={{ width: "35%" }}
              />
            </div>

            <div className="flex items-center justify-between">
              <p className="font-mono text-[10px] text-neutral-500">
                35% utilized
              </p>
              <div className="flex items-center gap-1">
                <div className="h-1.5 w-1.5 bg-black" />
                <span className="font-mono text-[9px] uppercase tracking-wider">
                  OK
                </span>
              </div>
            </div>
          </div>

          {/* Provider Status */}
          <div className="mt-4 space-y-2">
            <h4 className="font-mono text-[10px] uppercase tracking-wider text-neutral-500">
              Provider Status
            </h4>
            <ProviderStatus name="Google Drive" status="online" />
            <ProviderStatus name="AWS S3" status="online" />
            <ProviderStatus name="Dropbox" status="offline" />
          </div>
        </div>
      </div>
    </div>
  );
}

function NavItem({
  href,
  icon: Icon,
  children,
}: {
  href: string;
  icon: React.ComponentType<{ className?: string }>;
  children: React.ReactNode;
}) {
  const pathname = usePathname();
  const isActive =
    href === "/dashboard"
      ? pathname === "/dashboard" || pathname === "/"
      : pathname.startsWith(href);

  return (
    <Link
      href={href}
      className={cn(
        "flex items-center gap-3 px-3 py-2.5 font-mono text-xs uppercase tracking-wider transition-all",
        isActive
          ? "bg-black text-white"
          : "text-neutral-600 hover:bg-neutral-100"
      )}
    >
      <Icon className="h-4 w-4" />
      {children}
    </Link>
  );
}

function ProviderStatus({
  name,
  status,
}: {
  name: string;
  status: "online" | "offline";
}) {
  return (
    <div className="flex items-center justify-between">
      <span className="font-mono text-[10px] text-neutral-600">{name}</span>
      <div className="flex items-center gap-1.5">
        <div
          className={cn(
            "h-1.5 w-1.5",
            status === "online" ? "bg-black" : "bg-neutral-300"
          )}
        />
        <span
          className={cn(
            "font-mono text-[9px] uppercase tracking-wider",
            status === "online" ? "text-black" : "text-neutral-400"
          )}
        >
          {status === "online" ? "OK" : "OFF"}
        </span>
      </div>
    </div>
  );
}
