"use client";

import { use, useState, useEffect } from "react";
import { mockFiles } from "@/lib/mocks/files";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";
import { ShardProgressBar, ShardStatus } from "@/components/dashboard/ShardProgressBar";
import { ArrowLeft, Download, Trash2, Key, Share2 } from "lucide-react";
import Link from "next/link";
import { notFound, useRouter } from "next/navigation";
import { toast } from "sonner";

// Mock shard distribution for detail view
const mockShards: ShardStatus[] = [
  { index: 0, providerId: "googleDrive", status: "complete", progress: 100 },
  { index: 1, providerId: "googleDrive", status: "complete", progress: 100 },
  { index: 2, providerId: "awsS3", status: "complete", progress: 100 },
  { index: 3, providerId: "awsS3", status: "complete", progress: 100 },
  { index: 4, providerId: "googleDrive", status: "complete", progress: 100 },
  { index: 5, providerId: "awsS3", status: "complete", progress: 100 },
];

export default function FileDetailsPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = use(params); 
  const router = useRouter();
  
  const file = mockFiles.find((f) => f.id === id);

  if (!file) {
    return <div>File not found</div>;
  }

  const handleDownload = () => {
      toast.promise(
          new Promise((resolve) => setTimeout(resolve, 2000)),
          {
              loading: "Reconstructing file from shards...",
              success: "File downloaded successfully",
              error: "Download failed"
          }
      );
  };

  const handleDelete = () => {
      toast("Are you sure?", {
          description: "This will permanently delete the file and all its shards.",
          action: {
              label: "Delete",
              onClick: () => {
                  toast.success("File deleted");
                  router.push("/dashboard");
              }
          }
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
          <Link href="/dashboard">
            <ArrowLeft className="w-4 h-4" />
          </Link>
        </Button>
        <h1 className="text-xl font-semibold truncate flex-1">{file.name}</h1>
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
                 <CardTitle className="text-base">Shard Distribution (6,4)</CardTitle>
              </CardHeader>
              <CardContent>
                 <ShardProgressBar shards={mockShards} />
                 <div className="mt-4 p-4 bg-muted/50 rounded-md border text-sm">
                    <p className="font-medium text-primary flex items-center gap-2">
                       ✅ File is healthy and fully recoverable
                    </p>
                    <p className="text-muted-foreground mt-1">
                       You have partial redundancy across 2 providers.
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
                    <div>{file.size}</div>
                 </div>
                 <div className="space-y-1">
                    <div className="text-muted-foreground">Original Name</div>
                    <div className="break-all">{file.name}</div>
                 </div>
                  <div className="space-y-1">
                    <div className="text-muted-foreground">Uploaded</div>
                    <div>{file.uploadedAt}</div>
                 </div>
                  <div className="space-y-1">
                    <div className="text-muted-foreground">Verification</div>
                    <div className="font-mono text-xs">SHA-256: a3f2...8b1c</div>
                 </div>
              </CardContent>
           </Card>
           
           <Card>
              <CardHeader>
                 <CardTitle className="text-base">Actions</CardTitle>
              </CardHeader>
              <CardContent className="space-y-2">
                 <Button variant="secondary" className="w-full justify-start" onClick={handleShare}>
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
