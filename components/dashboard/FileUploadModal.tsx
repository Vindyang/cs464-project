import { useState } from "react";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from "@/components/ui/dialog";
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

export function FileUploadModal({
  open,
  onOpenChange,
  onUpload,
}: FileUploadModalProps) {
  const [file, setFile] = useState<File | null>(null);
  const [encrypt, setEncrypt] = useState(true);
  const [isUploading, setIsUploading] = useState(false);

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
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="gap-0 overflow-hidden border-2 border-black bg-white p-0 sm:max-w-md">
        <div className="border-b-2 border-black bg-neutral-50 p-6">
          <DialogHeader>
            <DialogTitle className="font-mono text-lg font-bold uppercase tracking-wider">
              Upload File
            </DialogTitle>
            <DialogDescription className="font-mono text-xs text-neutral-600">
              Encrypt, shard, and distribute to providers
            </DialogDescription>
          </DialogHeader>
        </div>

        <div className="space-y-6 p-6">
          {!isUploading ? (
            <div className="space-y-4">
              <div
                className={cn(
                  "group relative grid cursor-pointer place-items-center border-2 border-dashed p-8 transition-colors",
                  file
                    ? "border-black bg-neutral-50"
                    : "border-neutral-300 hover:border-black hover:bg-neutral-50"
                )}
              >
                <input
                  type="file"
                  className="absolute inset-0 z-10 cursor-pointer opacity-0"
                  onChange={handleFileChange}
                />
                <div className="flex flex-col items-center gap-3 text-center">
                  <div
                    className={cn(
                      "flex h-12 w-12 items-center justify-center transition-transform duration-200",
                      file
                        ? "bg-black text-white"
                        : "bg-neutral-100 text-neutral-400 group-hover:scale-110 group-hover:text-black"
                    )}
                  >
                    {file ? (
                      <Check className="h-6 w-6" />
                    ) : (
                      <Upload className="h-6 w-6" />
                    )}
                  </div>
                  <div className="space-y-1">
                    <p className="font-mono text-sm font-medium">
                      {file ? file.name : "Click to browse or drag file here"}
                    </p>
                    <p className="font-mono text-xs text-neutral-500">
                      {file
                        ? `${(file.size / 1024 / 1024).toFixed(2)} MB`
                        : "Supports any file type up to 5GB"}
                    </p>
                  </div>
                </div>
                {file && (
                  <div className="absolute right-2 top-2 z-20">
                    <Button
                      variant="ghost"
                      size="icon"
                      className="h-6 w-6 rounded-none hover:bg-black hover:text-white"
                      onClick={(e) => {
                        e.stopPropagation();
                        setFile(null);
                      }}
                    >
                      <X className="h-3 w-3" />
                    </Button>
                  </div>
                )}
              </div>

              <div className="flex items-start space-x-3 border bg-neutral-50 p-3">
                <Checkbox
                  id="encrypt"
                  checked={encrypt}
                  onCheckedChange={(c: boolean | "indeterminate") =>
                    setEncrypt(!!c)
                  }
                  className="mt-0.5"
                />
                <div className="space-y-1">
                  <Label
                    htmlFor="encrypt"
                    className="flex cursor-pointer items-center gap-2 font-mono text-sm font-medium leading-none"
                  >
                    Client-Side Encryption
                    <Shield className="h-3 w-3" />
                  </Label>
                  <p className="font-mono text-xs text-neutral-600">
                    Encrypts your file with AES-256 before upload
                  </p>
                </div>
              </div>

              {file && (
                <div className="border bg-neutral-50 p-4">
                  <h4 className="mb-3 font-mono text-xs uppercase tracking-wider text-neutral-500">
                    Distribution Preview
                  </h4>
                  <div className="space-y-2 font-mono text-xs text-neutral-700">
                    <div>• Split into 6 shards (4 data + 2 parity)</div>
                    <div>• Need any 4 shards to recover</div>
                    <div>• Distributed across providers</div>
                  </div>
                </div>
              )}
            </div>
          ) : (
            <div className="space-y-6 py-8">
              <div className="flex items-center gap-4">
                <div className="flex h-10 w-10 items-center justify-center border-2 border-black bg-neutral-50">
                  <FileIcon className="h-5 w-5" />
                </div>
                <div className="flex-1 space-y-1">
                  <div className="flex justify-between font-mono text-sm font-medium">
                    <span>{file?.name}</span>
                    <span>Uploading...</span>
                  </div>
                  <div className="h-1.5 w-full overflow-hidden bg-neutral-200">
                    <div className="h-full w-2/3 animate-pulse bg-black" />
                  </div>
                </div>
              </div>

              <div className="space-y-2">
                <div className="flex items-center gap-2 font-mono text-xs font-medium uppercase tracking-wider text-neutral-600">
                  <Shield className="h-3 w-3" />
                  <span>Encryption & Sharding</span>
                </div>
                <div className="grid grid-cols-4 gap-1">
                  {[1, 2, 3, 4].map((i) => (
                    <div
                      key={i}
                      className="h-1 animate-pulse bg-black opacity-60"
                      style={{ animationDelay: `${i * 100}ms` }}
                    />
                  ))}
                </div>
              </div>
            </div>
          )}
        </div>

        <DialogFooter className="border-t-2 border-black bg-neutral-50 p-6">
          <Button
            variant="ghost"
            onClick={() => onOpenChange(false)}
            className="font-mono text-xs"
          >
            Cancel
          </Button>
          <Button
            onClick={handleUpload}
            disabled={!file || isUploading}
            className="min-w-[100px] font-mono text-xs"
          >
            {isUploading ? "Processing" : "Upload"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
