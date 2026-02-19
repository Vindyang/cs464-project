"use client";

import { Sidebar } from "@/components/dashboard/Sidebar";
import { FileUploadModal } from "@/components/dashboard/FileUploadModal";
import { useState } from "react";
import { Button } from "@/components/ui/button";
import { UserCircle } from "lucide-react";
import Link from "next/link";

export default function DashboardLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const [uploadModalOpen, setUploadModalOpen] = useState(false);

  const handleUpload = (file: File, encrypt: boolean) => {
    // This will be connected to the upload store later
    console.log("Upload started:", file.name, "Encrypted:", encrypt);
    // Mimic start
    setTimeout(() => {
        // In reality, valid upload flow would keep modal open or show global progress
        setUploadModalOpen(false);
    }, 1000);
  };

  return (
    <div className="flex h-screen overflow-hidden bg-background">
      <Sidebar 
          className="hidden md:block" 
          onUploadClick={() => setUploadModalOpen(true)}
      />
      
      <div className="flex-1 flex flex-col min-w-0 overflow-hidden">
        <header className="flex items-center justify-between px-6 py-4 border-b">
           <h1 className="text-xl font-semibold">Dashboard</h1>
           <div className="flex items-center gap-4">
              <span className="text-sm text-muted-foreground hidden sm:inline-block">
                 Connected to 2 providers
              </span>
              <Button variant="ghost" size="icon" asChild>
                 <Link href="/settings">
                    <UserCircle className="w-6 h-6" />
                 </Link>
              </Button>
           </div>
        </header>

        <main className="flex-1 overflow-y-auto p-6">
          {children}
        </main>
      </div>

      <FileUploadModal 
        open={uploadModalOpen} 
        onOpenChange={setUploadModalOpen}
        onUpload={handleUpload}
      />
    </div>
  );
}
