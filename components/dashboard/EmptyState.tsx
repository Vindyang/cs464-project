import { Button } from "@/components/ui/button";
import { Upload, Cloud } from "lucide-react";

interface EmptyStateProps {
  onUpload: () => void;
  storageUsage?: string;
  providerCount?: number;
}

export function EmptyState({
  onUpload,
  storageUsage = "0GB",
  providerCount = 0,
}: EmptyStateProps) {
  return (
    <div className="flex flex-col items-center justify-center p-8 text-center border-2 border-dashed rounded-lg border-muted-foreground/25 bg-muted/50 h-[60vh]">
      <div className="p-4 mb-4 rounded-full bg-background ring-1 ring-border shadow-sm">
        <Cloud className="w-12 h-12 text-muted-foreground" />
      </div>
      <h3 className="text-2xl font-semibold tracking-tight">
        No files uploaded yet
      </h3>
      <p className="mt-2 text-muted-foreground max-w-sm">
        Combine multiple cloud providers into one fault-tolerant drive. Files are
        encrypted and distributed across your providers.
      </p>
      
      <div className="mt-6 flex flex-col gap-2">
         <p className="text-sm text-muted-foreground">
            Combined storage: {storageUsage} across {providerCount} providers
         </p>
        <Button onClick={onUpload} size="lg" className="mt-2">
          <Upload className="w-4 h-4 mr-2" />
          Upload Your First File
        </Button>
      </div>
    </div>
  );
}
