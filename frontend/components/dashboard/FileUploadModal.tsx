import { useEffect, useState, useRef } from "react";
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
  onUpload: (file: File, encrypt: boolean, signal?: AbortSignal) => Promise<void> | void;
}

export function FileUploadModal({
  open,
  onOpenChange,
  onUpload,
}: FileUploadModalProps) {
  const [files, setFiles] = useState<File[]>([]);
  const [encrypt, setEncrypt] = useState(true);
  const [isUploading, setIsUploading] = useState(false);
  const abortControllerRef = useRef<AbortController | null>(null);

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files && e.target.files.length > 0) {
      const newFiles = Array.from(e.target.files);
      setFiles((prev) => [...prev, ...newFiles]);
      e.target.value = "";
    }
  };

  const removeFile = (index: number) => {
    setFiles((prev) => prev.filter((_, i) => i !== index));
  };

  const handleUpload = async () => {
    if (files.length > 0) {
      setIsUploading(true);
      const abortController = new AbortController();
      abortControllerRef.current = abortController;
      try {
        for (const file of files) {
          if (abortController.signal.aborted) break;
          await onUpload(file, encrypt, abortController.signal);
        }
        setFiles([]);
        setEncrypt(true);
        onOpenChange(false);
      } catch (err: any) {
        if (err.name !== "AbortError") {
          // Parent handles user-facing errors; keep modal open for retry.
        }
      } finally {
        setIsUploading(false);
        abortControllerRef.current = null;
      }
    }
  };

  useEffect(() => {
    if (!open) {
      if (abortControllerRef.current) abortControllerRef.current.abort();
      setFiles([]);
      setEncrypt(true);
      setIsUploading(false);
    }
  }, [open]);

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
                  files.length > 0
                    ? "border-black bg-neutral-50"
                    : "border-neutral-300 hover:border-black hover:bg-neutral-50"
                )}
              >
                <input
                  type="file"
                  multiple
                  className="absolute inset-0 z-10 cursor-pointer opacity-0"
                  onChange={handleFileChange}
                />
                <div className="flex flex-col items-center gap-3 text-center pointer-events-none">
                  <div
                    className={cn(
                      "flex h-12 w-12 items-center justify-center transition-transform duration-200",
                      files.length > 0
                        ? "bg-black text-white"
                        : "bg-neutral-100 text-neutral-400 group-hover:scale-110 group-hover:text-black"
                    )}
                  >
                    {files.length > 0 ? (
                      <Check className="h-6 w-6" />
                    ) : (
                      <Upload className="h-6 w-6" />
                    )}
                  </div>
                  <div className="space-y-1">
                    <p className="font-mono text-sm font-medium">
                      {files.length > 0 ? "Click or drag to add more files" : "Click to browse or drag files here"}
                    </p>
                    <p className="font-mono text-xs text-neutral-500">
                      Supports multiple files up to 5GB each
                    </p>
                  </div>
                </div>
              </div>

              {files.length > 0 && (
                <div className="space-y-2">
                  <h4 className="font-mono text-xs uppercase tracking-wider text-neutral-500">
                    Selected Files ({files.length})
                  </h4>
                  <div className="max-h-[140px] overflow-y-auto space-y-1.5 pr-1">
                    {files.map((f, i) => (
                      <div key={i} className="flex items-center justify-between border border-neutral-200 bg-white p-2">
                        <div className="min-w-0 pr-4">
                           <p className="truncate font-mono text-sm">{f.name}</p>
                           <p className="font-mono text-[10px] text-neutral-500">{(f.size / 1024 / 1024).toFixed(2)} MB</p>
                        </div>
                        <Button
                          type="button"
                          variant="ghost"
                          size="icon"
                          onClick={(e) => { e.stopPropagation(); removeFile(i); }}
                          disabled={isUploading}
                          className="h-6 w-6 shrink-0 rounded-none hover:bg-red-50 hover:text-red-500 text-neutral-400"
                        >
                          <X className="h-4 w-4" />
                        </Button>
                      </div>
                    ))}
                  </div>
                </div>
              )}

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

              {files.length > 0 && (
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
                    <span>{files.length > 0 ? `Uploading ${files.length} file(s)...` : "Uploading..."}</span>
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
            onClick={() => {
              if (isUploading && abortControllerRef.current) {
                abortControllerRef.current.abort();
              } else {
                onOpenChange(false);
              }
            }}
            className={cn(
              "font-mono text-xs transition-colors",
              isUploading && "border border-red-600 text-red-600 hover:bg-red-50 hover:text-red-700"
            )}
          >
            {isUploading ? "Cancel Upload" : "Cancel"}
          </Button>
          <Button
            onClick={handleUpload}
            disabled={files.length === 0 || isUploading}
            className="min-w-[100px] font-mono text-xs"
          >
            {isUploading ? "Processing" : "Upload"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
