"use client";

import { useState } from "react";
import { Download, Loader2 } from "lucide-react";
import { toast } from "sonner";

import type { FileHealthStatus } from "@/lib/api/files";

interface DownloadFileButtonProps {
  fileId: string;
  fileName: string;
  requiredShards: number;
  healthStatus?: FileHealthStatus;
  variant?: "primary" | "icon";
}

function extractFilename(contentDisposition: string | null, fallback: string): string {
  if (!contentDisposition) {
    return fallback;
  }

  const utf8Match = contentDisposition.match(/filename\*=UTF-8''([^;]+)/i);
  if (utf8Match?.[1]) {
    return decodeURIComponent(utf8Match[1]);
  }

  const plainMatch = contentDisposition.match(/filename="?([^";]+)"?/i);
  if (plainMatch?.[1]) {
    return plainMatch[1];
  }

  return fallback;
}

function triggerBrowserDownload(blob: Blob, fileName: string) {
  const blobUrl = URL.createObjectURL(blob);
  const link = document.createElement("a");
  link.href = blobUrl;
  link.download = fileName;
  document.body.appendChild(link);
  link.click();
  link.remove();
  URL.revokeObjectURL(blobUrl);
}

function buildRecoverabilityMessage(fileName: string, healthyShards: number, requiredShards: number) {
  return `Cannot download ${fileName}: only ${healthyShards} healthy shards are currently available, but ${requiredShards} are required to reconstruct the file.`;
}

export function DownloadFileButton({
  fileId,
  fileName,
  requiredShards,
  healthStatus,
  variant = "icon",
}: DownloadFileButtonProps) {
  const [downloading, setDownloading] = useState(false);

  async function handleDownload() {
    if (downloading) return;

    if (healthStatus && healthStatus.recoverable === false) {
      toast.error(buildRecoverabilityMessage(fileName, healthStatus.healthy_shards, requiredShards));
      return;
    }

    setDownloading(true);
    try {
      const response = await fetch(`/api/download/${fileId}`, {
        method: "GET",
        cache: "no-store",
      });

      if (!response.ok) {
        const payload = await response.json().catch(() => ({}));
        throw new Error(payload?.details || payload?.error || `Failed to download ${fileName}`);
      }

      const blob = await response.blob();
      const resolvedName = extractFilename(response.headers.get("Content-Disposition"), fileName);
      triggerBrowserDownload(blob, resolvedName);
    } catch (error) {
      const message = error instanceof Error ? error.message : `Failed to download ${fileName}`;
      toast.error(message);
    } finally {
      setDownloading(false);
    }
  }

  if (variant === "primary") {
    return (
      <button
        type="button"
        onClick={handleDownload}
        disabled={downloading}
        className="font-mono text-[11px] uppercase tracking-wider border border-black bg-black px-3 py-2 text-white hover:bg-neutral-800 transition-colors disabled:cursor-not-allowed disabled:opacity-50"
      >
        {downloading ? "Downloading..." : "Download"}
      </button>
    );
  }

  return (
    <button
      type="button"
      onClick={handleDownload}
      disabled={downloading}
      className="inline-flex h-8 w-8 items-center justify-center border border-transparent text-neutral-500 transition-colors hover:border-neutral-200 hover:text-black disabled:cursor-not-allowed disabled:opacity-50"
      aria-label={`Download ${fileName}`}
      title={downloading ? "Downloading..." : "Download"}
    >
      {downloading ? <Loader2 className="h-4 w-4 animate-spin" /> : <Download className="h-4 w-4" />}
    </button>
  );
}