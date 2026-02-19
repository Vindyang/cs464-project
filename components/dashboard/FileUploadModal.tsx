import { useState } from "react";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription } from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import { Label } from "@/components/ui/label";
import { ShardProgressBar, ShardStatus } from "./ShardProgressBar";
import { Upload, FileIcon } from "lucide-react";

interface FileUploadModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onUpload: (file: File, encrypt: boolean) => void;
}

export function FileUploadModal({ open, onOpenChange, onUpload }: FileUploadModalProps) {
  const [file, setFile] = useState<File | null>(null);
  const [encrypt, setEncrypt] = useState(true);
  const [isUploading, setIsUploading] = useState(false); // Placeholder for local state

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files && e.target.files[0]) {
      setFile(e.target.files[0]);
    }
  };
  
  const handleUpload = () => {
      if (file) {
          setIsUploading(true);
          onUpload(file, encrypt);
      }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Upload File</DialogTitle>
          <DialogDescription>
            Securely upload into distributed storage.
          </DialogDescription>
        </DialogHeader>
        
        {!isUploading ? (
          <div className="grid gap-4 py-4">
            <div className="grid place-items-center border-2 border-dashed rounded-lg p-6 hover:bg-muted/50 transition-colors cursor-pointer relative">
               <input 
                  type="file" 
                  className="absolute inset-0 opacity-0 cursor-pointer" 
                  onChange={handleFileChange}
               />
               <Upload className="w-8 h-8 text-muted-foreground mb-2" />
               <p className="text-sm text-muted-foreground">
                  {file ? file.name : "Drag & drop or click to browse"}
               </p>
            </div>

            <div className="flex items-center space-x-2">
              <Checkbox 
                id="encrypt" 
                checked={encrypt} 
                onCheckedChange={(c: boolean | "indeterminate") => setEncrypt(!!c)} 
              />
              <Label htmlFor="encrypt" className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70">
                Encrypt file before upload
              </Label>
            </div>
            
             {file && (
                <div className="flex items-center gap-2 p-2 bg-muted rounded text-sm">
                   <FileIcon className="w-4 h-4" />
                   <span className="truncate flex-1">{file.name}</span>
                   <span className="text-muted-foreground">{(file.size / 1024 / 1024).toFixed(2)} MB</span>
                </div>
             )}
          </div>
        ) : (
            <div className="py-4 space-y-4">
                 <div className="flex items-center gap-2">
                    <FileIcon className="w-4 h-4 text-primary" />
                    <span className="font-medium">{file?.name}</span>
                 </div>
                 {/* TODO: Connect with real upload state */}
            </div>
        )}

        <div className="flex justify-end gap-2">
          <Button variant="outline" onClick={() => onOpenChange(false)}>Cancel</Button>
          <Button onClick={handleUpload} disabled={!file || isUploading}>
            {isUploading ? "Uploading..." : "Upload File"}
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  );
}
