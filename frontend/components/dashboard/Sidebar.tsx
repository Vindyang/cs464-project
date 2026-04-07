"use client";

import React from "react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { Button } from "@/components/ui/button";
import { cn, formatBytes } from "@/lib/utils";

export interface SidebarProviderData {
  providerId: string;
  displayName: string;
  status: string;
  quotaUsedBytes: number;
  quotaTotalBytes: number;
}

interface SidebarProps {
  className?: string;
  onUploadClick: () => void;
  providers?: SidebarProviderData[];
  totalStorageUsedBytes?: number;
  totalStorageTotalBytes?: number;
}

const NAV_ITEMS = [
  { href: "/dashboard", label: "Dashboard" },
  { href: "/files", label: "Files" },
  { href: "/providers", label: "Providers" },
  { href: "/settings", label: "Settings" },
];

export function Sidebar({
  className,
  onUploadClick,
  providers = [],
  totalStorageUsedBytes = 0,
  totalStorageTotalBytes = 0,
}: SidebarProps) {
  const usedPct =
    totalStorageTotalBytes > 0
      ? Math.round((totalStorageUsedBytes / totalStorageTotalBytes) * 100)
      : 0;

  return (
    <div className={cn("flex w-56 flex-col border-r bg-white", className)}>
      <div className="flex flex-1 flex-col py-6">
        {/* Logo */}
        <div className="mb-6 px-5">
          <div className="flex items-center gap-2">
            <div className="flex h-5 w-5 items-center justify-center border-2 border-black dark:border-neutral-100 transition-colors">
              <div className="h-1.5 w-1.5 bg-black dark:bg-neutral-100 transition-colors" />
            </div>
            <span className="font-mono text-xs font-bold tracking-widest uppercase text-neutral-900 dark:text-neutral-100 transition-colors">
              Omnishard
            </span>
          </div>
        </div>

        {/* Upload Button */}
        <div className="mb-6 px-5">
          <Button
            className="w-full font-mono text-[11px] tracking-wider"
            onClick={onUploadClick}
          >
            + UPLOAD FILE
          </Button>
        </div>

        {/* Navigation */}
        <nav className="flex-1 px-2 space-y-0.5">
          {NAV_ITEMS.map((item) => (
            <NavItem key={item.href} href={item.href}>
              {item.label}
            </NavItem>
          ))}
        </nav>

        {/* Storage Stats */}
        <div className="mt-auto px-5 space-y-4">
          <div className="border bg-neutral-50 p-3 relative">
            <span className="absolute -top-px -left-px w-1.5 h-1.5 border-t border-l border-neutral-400 opacity-40 pointer-events-none" />
            <span className="absolute -top-px -right-px w-1.5 h-1.5 border-t border-r border-neutral-400 opacity-40 pointer-events-none" />
            <span className="absolute -bottom-px -left-px w-1.5 h-1.5 border-b border-l border-neutral-400 opacity-40 pointer-events-none" />
            <span className="absolute -bottom-px -right-px w-1.5 h-1.5 border-b border-r border-neutral-400 opacity-40 pointer-events-none" />
            <p className="mb-2 font-mono text-[9px] uppercase tracking-widest text-neutral-500">
              Storage
            </p>
            <div className="mb-1.5 flex items-baseline gap-1">
              <span className="font-mono text-lg font-bold">
                {formatBytes(totalStorageUsedBytes)}
              </span>
              <span className="ml-auto font-mono text-[10px] text-neutral-500">
                /{" "}
                {totalStorageTotalBytes > 0
                  ? formatBytes(totalStorageTotalBytes)
                  : "—"}
              </span>
            </div>
            <div className="mb-1.5 h-1 w-full bg-neutral-200">
              <div
                className="h-full bg-black transition-all"
                style={{ width: `${usedPct}%` }}
              />
            </div>
            <p className="font-mono text-[9px] text-neutral-500">
              {usedPct}% utilized
            </p>
          </div>

          {/* Provider Status */}
          {providers.length > 0 && (
            <div className="space-y-1.5">
              <p className="font-mono text-[9px] uppercase tracking-widest text-neutral-500">
                Providers
              </p>
              {providers.map((p) => (
                <ProviderRow
                  key={p.providerId}
                  name={p.displayName}
                  status={p.status}
                />
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

function NavItem({
  href,
  children,
}: {
  href: string;
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
        "block px-3 py-2 font-mono text-[11px] uppercase tracking-wider transition-colors",
        isActive
          ? "bg-black text-white"
          : "text-neutral-500 hover:bg-neutral-100 hover:text-black",
      )}
    >
      {children}
    </Link>
  );
}

function ProviderRow({ name, status }: { name: string; status: string }) {
  const isOnline = status === "connected" || status === "online";
  return (
    <div className="flex items-center justify-between">
      <span className="font-mono text-[9px] text-neutral-600 truncate max-w-[100px]">
        {name}
      </span>
      <div className="flex items-center gap-1">
        <div
          className={cn(
            "h-1.5 w-1.5",
            isOnline ? "bg-black" : "bg-neutral-300",
          )}
        />
        <span
          className={cn(
            "font-mono text-[8px] uppercase tracking-wider",
            isOnline ? "text-black" : "text-neutral-400",
          )}
        >
          {isOnline ? "OK" : "OFF"}
        </span>
      </div>
    </div>
  );
}
