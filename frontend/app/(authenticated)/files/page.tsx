"use client";

import { FileTable } from "@/components/dashboard/FileTable";
import { mockFiles } from "@/lib/mocks/files";
import { useState } from "react";
import { Input } from "@/components/ui/input";
import { Search, Filter, Plus } from "lucide-react";
import { Button } from "@/components/ui/button";
import { toast } from "sonner";
import { useRouter } from "next/navigation";

export default function FilesPage() {
  const router = useRouter();
  const [search, setSearch] = useState("");
  // Determine status based on shard availability (mock logic)
  const files = mockFiles.map(f => ({
    id: f.id,
    name: f.name,
    size: f.size,
    date: f.uploadedAt,
    providerCount: f.shardsAvailable, // Using available shards as proxy for "redundancy level"
    status: (f.shardsAvailable === f.shardsTotal ? "synced" : f.shardsAvailable > 3 ? "syncing" : "error") as "synced" | "syncing" | "error"
  })).filter(f => 
    f.name.toLowerCase().includes(search.toLowerCase())
  );

  const handleDownload = (id: string) => {
    toast.promise(
        new Promise((resolve) => setTimeout(resolve, 1000)),
        {
            loading: "Reconstructing file from shards...",
            success: "Download started",
            error: "Download failed"
        }
    );
  };

  const handleDelete = (id: string) => {
      toast("Are you sure?", {
          description: "This will permanently delete the file from all providers.",
          action: {
              label: "Delete",
              onClick: () => {
                  toast.success("File deletion scheduled");
              }
          }
      });
  };
  
  const handleDetails = (id: string) => {
      // router.push(`/dashboard/files/${id}`);
      toast.info("File details view not implemented yet");
  }

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
            <Button variant="outline" size="sm" className="flex-1 sm:flex-none border-border-color text-text-main hover:bg-bg-subtle">
                <Filter className="mr-2 h-4 w-4" />
                Filter
            </Button>
            <Button size="sm" className="flex-1 sm:flex-none bg-accent-primary text-white hover:bg-accent-primary-hover">
                <Plus className="mr-2 h-4 w-4" />
                Upload
            </Button>
         </div>
      </div>

      <FileTable 
        files={files}
        onDownload={handleDownload}
        onDelete={handleDelete}
        onDetails={handleDetails}
      />
    </div>
  );
}
