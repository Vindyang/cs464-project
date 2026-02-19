"use client";

import { FileCard } from "@/components/dashboard/FileCard";
import { EmptyState } from "@/components/dashboard/EmptyState";
import { mockFiles } from "@/lib/mocks/files";
import { useState } from "react";
import { Input } from "@/components/ui/input";
import { Search, Filter } from "lucide-react";
import { Button } from "@/components/ui/button";

export default function DashboardPage() {
  const [search, setSearch] = useState("");
  const [files, setFiles] = useState(mockFiles); // In real app, this comes from store/query

  const filteredFiles = files.filter(f => 
    f.name.toLowerCase().includes(search.toLowerCase())
  );

  const handleDownload = (id: string) => {
    console.log("Download file", id);
  };

  const handleDelete = (id: string) => {
      console.log("Delete file", id);
      setFiles(prev => prev.filter(f => f.id !== id));
  };
  
  const handleDetails = (id: string) => {
      console.log("View details for", id);
  }

  // TODO: Add actual empty state check when store is connected
  // if (files.length === 0) return <EmptyState ... />;

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
         <div className="relative w-72">
            <Search className="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
            <Input 
               placeholder="Search files..." 
               className="pl-8" 
               value={search}
               onChange={(e) => setSearch(e.target.value)}
            />
         </div>
         <Button variant="outline" size="sm">
            <Filter className="mr-2 h-4 w-4" />
            Filter
         </Button>
      </div>

      {filteredFiles.length > 0 ? (
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {filteredFiles.map((file) => (
            <FileCard
              key={file.id}
              {...file}
              onDownload={() => handleDownload(file.id)}
              onDelete={() => handleDelete(file.id)}
              onDetails={() => handleDetails(file.id)}
            />
          ))}
        </div>
      ) : (
          <div className="py-20 text-center text-muted-foreground">
             No files found matching "{search}"
          </div>
      )}
    </div>
  );
}
