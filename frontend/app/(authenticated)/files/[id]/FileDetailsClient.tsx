"use client";

import { useRouter } from "next/navigation";
import { ShardMap, deleteFile } from "@/lib/api/files";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";
import { ShardProgressBar, ShardStatus } from "@/components/dashboard/ShardProgressBar";
import { ArrowLeft, Download, Trash2, Key, Share2 } from "lucide-react";
import Link from "next/link";
import { toast } from "sonner";

function formatBytes(bytes: number): string {
  if (bytes >= 1e12) return `${(bytes / 1e12).toFixed(1)} TB`;
  if (bytes >= 1e9) return `${(bytes / 1e9).toFixed(1)} GB`;
  return `${(bytes / 1e6).toFixed(1)} MB`;
}

function toShardStatus(status: string): ShardStatus["status"] {
  switch (status) {
    case "HEALTHY": return "complete";
    case "MISSING": return "failed";
    case "CORRUPTED": return "failed";
    default: return "pending";
  }
}

interface FileDetailsClientProps {
  file: ShardMap;
}

export function FileDetailsClient({ file }: FileDetailsClientProps) {
  const router = useRouter();

  const shards: ShardStatus[] = file.shards.map((s) => ({
    index: s.shard_index,
    providerId: s.provider,
    status: toShardStatus(s.status),
    progress: s.status === "HEALTHY" ? 100 : 0,
  }));

  const healthyCount = file.shards.filter((s) => s.status === "HEALTHY").length;
  const isRecoverable = healthyCount >= file.k;

  const handleDownload = () => {
    toast.promise(
      fetch(`${process.env.NEXT_PUBLIC_API_URL}/api/v1/files/${file.file_id}/download`),
      {
        loading: "Reconstructing file from shards...",
        success: "Download started",
        error: "Download failed",
      }
    );
  };

  const handleDelete = () => {
    toast("Are you sure?", {
      description: "This will permanently delete the file and all its shards.",
      action: {
        label: "Delete",
        onClick: async () => {
          try {
            await deleteFile(file.file_id);
            toast.success("File deleted");
            router.push("/files");
          } catch {
            toast.error("Failed to delete file");
          }
        },
      },
    });
  };

  const handleShare = () => {
    navigator.clipboard.writeText(window.location.href);
    toast.success("Link copied to clipboard");
  };

  return (
    <div className="space-y-6 max-w-5xl mx-auto">
      <div className="flex items-center gap-4">
        <Button variant="ghost" size="icon" asChild>
          <Link href="/files">
            <ArrowLeft className="w-4 h-4" />
          </Link>
        </Button>
        <h1 className="text-xl font-semibold truncate flex-1">{file.original_name}</h1>
        <div className="flex gap-2">
          <Button variant="outline" onClick={handleDownload}>
            <Download className="w-4 h-4 mr-2" />
            Download
          </Button>
          <Button variant="destructive" onClick={handleDelete}>
            <Trash2 className="w-4 h-4 mr-2" />
            Delete
          </Button>
        </div>
      </div>

      <div className="grid gap-6 md:grid-cols-3">
        {/* Main Content: Shard Distribution */}
        <div className="md:col-span-2 space-y-6">
          <Card>
            <CardHeader>
              <CardTitle className="text-base">
                Shard Distribution ({file.n},{file.k})
              </CardTitle>
            </CardHeader>
            <CardContent>
              <ShardProgressBar shards={shards} />
              <div className="mt-4 p-4 bg-muted/50 rounded-md border text-sm">
                {isRecoverable ? (
                  <p className="font-medium text-primary flex items-center gap-2">
                    ✅ File is healthy and fully recoverable
                  </p>
                ) : (
                  <p className="font-medium text-destructive flex items-center gap-2">
                    ⚠️ File may not be recoverable — too many shards missing
                  </p>
                )}
                <p className="text-muted-foreground mt-1">
                  {healthyCount} of {file.n} shards available (need {file.k} to recover).
                </p>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle className="text-base flex items-center gap-2">
                <Key className="w-4 h-4" /> Client-Side Encryption
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="grid grid-cols-[100px_1fr] gap-4 text-sm">
                <div className="text-muted-foreground">Algorithm</div>
                <div className="font-mono">AES-256-GCM</div>
                <div className="text-muted-foreground">Key Status</div>
                <div className="text-primary font-medium">Stored in Browser (Session)</div>
              </div>
              <Separator />
              <Button variant="outline" size="sm" className="w-full">
                Export Decryption Key
              </Button>
            </CardContent>
          </Card>
        </div>

        {/* Sidebar: Metadata */}
        <div className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle className="text-base">Metadata</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4 text-sm">
              <div className="space-y-1">
                <div className="text-muted-foreground">Size</div>
                <div>{formatBytes(file.original_size)}</div>
              </div>
              <div className="space-y-1">
                <div className="text-muted-foreground">Original Name</div>
                <div className="break-all">{file.original_name}</div>
              </div>
              <div className="space-y-1">
                <div className="text-muted-foreground">Status</div>
                <div>{file.status}</div>
              </div>
              <div className="space-y-1">
                <div className="text-muted-foreground">Redundancy</div>
                <div className="font-mono text-xs">
                  n={file.n}, k={file.k}
                </div>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle className="text-base">Actions</CardTitle>
            </CardHeader>
            <CardContent className="space-y-2">
              <Button
                variant="secondary"
                className="w-full justify-start"
                onClick={handleShare}
              >
                <Share2 className="w-4 h-4 mr-2" />
                Share File
              </Button>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  );
}
