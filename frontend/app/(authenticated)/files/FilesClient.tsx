"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { FileTable } from "@/components/dashboard/FileTable";
import { FileMetadata, deleteFile } from "@/lib/api/files";
import { Input } from "@/components/ui/input";
import { Search, Filter, Plus } from "lucide-react";
import { Button } from "@/components/ui/button";
import { toast } from "sonner";

function formatBytes(bytes: number): string {
  if (bytes >= 1e12) return `${(bytes / 1e12).toFixed(1)} TB`;
  if (bytes >= 1e9) return `${(bytes / 1e9).toFixed(1)} GB`;
  return `${(bytes / 1e6).toFixed(1)} MB`;
}

interface FilesClientProps {
  initialFiles: FileMetadata[];
}

export function FilesClient({ initialFiles }: FilesClientProps) {
  const router = useRouter();
  const [files, setFiles] = useState(initialFiles);
  const [search, setSearch] = useState("");

  const tableFiles = files
    .map((f) => ({
      id: f.file_id,
      name: f.original_name,
      size: formatBytes(f.original_size),
      date: new Date(f.created_at).toLocaleDateString(),
      providerCount: f.health_status?.total_shards ?? 0,
      status: (
        f.status === "UPLOADED" ? "synced" : f.status === "DEGRADED" ? "error" : "syncing"
      ) as "synced" | "syncing" | "error",
    }))
    .filter((f) => f.name.toLowerCase().includes(search.toLowerCase()));

  const handleDownload = (id: string) => {
    toast.promise(
      fetch(`${process.env.NEXT_PUBLIC_API_URL}/api/v1/files/${id}/download`),
      {
        loading: "Reconstructing file from shards...",
        success: "Download started",
        error: "Download failed",
      }
    );
  };

  const handleDelete = (id: string) => {
    toast("Are you sure?", {
      description: "This will permanently delete the file from all providers.",
      action: {
        label: "Delete",
        onClick: async () => {
          try {
            await deleteFile(id);
            setFiles((prev) => prev.filter((f) => f.file_id !== id));
            toast.success("File deleted");
          } catch {
            toast.error("Failed to delete file");
          }
        },
      },
    });
  };

  const handleDetails = (id: string) => {
    router.push(`/files/${id}`);
  };

  return (
    <div className="space-y-6">
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
        <div className="relative w-full sm:w-72">
          <Search className="absolute left-2 top-2.5 h-4 w-4 text-text-secondary" />
          <Input
            placeholder="Search files..."
            className="pl-8 bg-bg-subtle border-border-color focus-visible:ring-accent-primary"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
          />
        </div>
        <div className="flex gap-2 w-full sm:w-auto">
          <Button
            variant="outline"
            size="sm"
            className="flex-1 sm:flex-none border-border-color text-text-main hover:bg-bg-subtle"
          >
            <Filter className="mr-2 h-4 w-4" />
            Filter
          </Button>
          <Button
            size="sm"
            className="flex-1 sm:flex-none bg-accent-primary text-white hover:bg-accent-primary-hover"
          >
            <Plus className="mr-2 h-4 w-4" />
            Upload
          </Button>
        </div>
      </div>

      <FileTable
        files={tableFiles}
        onDownload={handleDownload}
        onDelete={handleDelete}
        onDetails={handleDetails}
      />
    </div>
  );
}
