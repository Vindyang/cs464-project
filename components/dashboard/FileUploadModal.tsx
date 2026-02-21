import { useState } from "react";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter } from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import { Label } from "@/components/ui/label";
import { Upload, FileIcon, Shield, X, Check } from "lucide-react";
import { cn } from "@/lib/utils";

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
      <DialogContent className="sm:max-w-md bg-bg-canvas border-border-color p-0 gap-0 overflow-hidden">
        <div className="p-6 border-b border-border-color bg-bg-subtle/30">
            <DialogHeader>
                <DialogTitle className="text-lg font-semibold tracking-tight text-text-main">Upload File</DialogTitle>
                <DialogDescription className="text-text-secondary">
                    Securely push data to your decentralized vault.
                </DialogDescription>
            </DialogHeader>
        </div>
        
        <div className="p-6 space-y-6">
            {!isUploading ? (
              <div className="space-y-4">
                <div className={cn(
                    "grid place-items-center border-2 border-dashed rounded-[4px] p-8 transition-colors cursor-pointer relative group",
                    file ? "border-accent-primary bg-accent-primary/5" : "border-border-color hover:border-accent-primary/50 hover:bg-bg-subtle"
                )}>
                   <input 
                      type="file" 
                      className="absolute inset-0 opacity-0 cursor-pointer z-10" 
                      onChange={handleFileChange}
                   />
                   <div className="flex flex-col items-center gap-3 text-center">
                        <div className={cn(
                            "w-12 h-12 rounded-full flex items-center justify-center transition-colors",
                            file ? "bg-accent-primary text-white" : "bg-bg-subtle text-text-tertiary group-hover:text-accent-primary group-hover:scale-110 duration-200"
                        )}>
                            {file ? <Check className="w-6 h-6" /> : <Upload className="w-6 h-6" />}
                        </div>
                        <div className="space-y-1">
                            <p className="font-medium text-text-main">
                                {file ? file.name : "Click to browse or drag file here"}
                            </p>
                            <p className="text-xs text-text-tertiary">
                                {file ? `${(file.size / 1024 / 1024).toFixed(2)} MB` : "Supports any file type up to 5GB"}
                            </p>
                        </div>
                   </div>
                   {file && (
                       <div className="absolute top-2 right-2 z-20">
                           <Button 
                              variant="ghost" 
                              size="icon" 
                              className="h-6 w-6 rounded-full hover:bg-red-100 hover:text-red-600"
                              onClick={(e) => {
                                  e.stopPropagation();
                                  setFile(null);
                              }}
                           >
                               <X className="w-3 h-3" />
                           </Button>
                       </div>
                   )}
                </div>

                <div className="flex items-start space-x-3 p-3 bg-bg-subtle rounded-[2px] border border-border-color">
                  <Checkbox 
                    id="encrypt" 
                    checked={encrypt} 
                    onCheckedChange={(c: boolean | "indeterminate") => setEncrypt(!!c)} 
                    className="mt-0.5 data-[state=checked]:bg-accent-primary data-[state=checked]:border-accent-primary"
                  />
                  <div className="space-y-1">
                      <Label htmlFor="encrypt" className="text-sm font-medium leading-none cursor-pointer text-text-main flex items-center gap-2">
                        Client-Side Encryption
                        <Shield className="w-3 h-3 text-accent-primary" />
                      </Label>
                      <p className="text-xs text-text-secondary">
                          Encrypts your file with AES-256 before it leaves your browser.
                      </p>
                  </div>
                </div>
              </div>
            ) : (
                <div className="py-8 space-y-6">
                     <div className="flex items-center gap-4">
                        <div className="w-10 h-10 bg-accent-primary/10 rounded-[2px] flex items-center justify-center text-accent-primary">
                            <FileIcon className="w-5 h-5" />
                        </div>
                        <div className="flex-1 space-y-1">
                            <div className="flex justify-between text-sm font-medium text-text-main">
                                <span>{file?.name}</span>
                                <span>Uploading...</span>
                            </div>
                            <div className="h-1.5 w-full bg-bg-subtle rounded-full overflow-hidden">
                                <div className="h-full bg-accent-primary w-2/3 animate-pulse" />
                            </div>
                        </div>
                     </div>
                     
                     <div className="space-y-2">
                         <div className="flex items-center gap-2 text-xs font-medium text-text-secondary uppercase tracking-wider">
                             <Shield className="w-3 h-3" />
                             <span>Encryption & Sharding Status</span>
                         </div>
                         <div className="grid grid-cols-4 gap-1">
                             {[1,2,3,4].map(i => (
                                 <div key={i} className="h-1 bg-accent-primary rounded-full opacity-60 animate-pulse" style={{ animationDelay: `${i * 100}ms` }} />
                             ))}
                         </div>
                     </div>
                </div>
            )}
        </div>
        
        <DialogFooter className="p-6 pt-0">
          <Button variant="ghost" onClick={() => onOpenChange(false)} className="text-text-secondary hover:text-text-main hover:bg-bg-subtle">Cancel</Button>
          <Button 
              onClick={handleUpload} 
              disabled={!file || isUploading}
              className="bg-accent-primary hover:bg-accent-primary-hover text-white rounded-[2px] min-w-[100px]"
          >
            {isUploading ? "Processing" : "Upload"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
