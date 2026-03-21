"use client";

import { useState } from "react";
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { getGDriveAuthorizeURL } from "@/lib/api/providers";
import { toast } from "sonner";

interface ConnectProviderModalProps {
  open: boolean;
  onClose: () => void;
}

const PROVIDERS = [
  {
    id: "googleDrive",
    name: "Google Drive",
    description: "Connect your Google Drive account to store shards.",
    available: true,
  },
  {
    id: "awsS3",
    name: "Amazon S3",
    description: "Coming soon.",
    available: false,
  },
  {
    id: "oneDrive",
    name: "OneDrive",
    description: "Coming soon.",
    available: false,
  },
];

export function ConnectProviderModal({ open, onClose }: ConnectProviderModalProps) {
  const [loading, setLoading] = useState<string | null>(null);

  const handleConnect = async (providerId: string) => {
    if (providerId !== "googleDrive") return;
    setLoading(providerId);
    try {
      const { authURL } = await getGDriveAuthorizeURL();
      window.location.assign(authURL);
    } catch {
      toast.error("Failed to start Google Drive connection");
      setLoading(null);
    }
  };

  return (
    <Dialog open={open} onOpenChange={(v) => !v && onClose()}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Connect a Provider</DialogTitle>
        </DialogHeader>
        <div className="space-y-3 pt-2">
          {PROVIDERS.map((p) => (
            <div
              key={p.id}
              className="flex items-center justify-between rounded border border-border px-4 py-3"
            >
              <div>
                <p className={`text-sm font-medium ${p.available ? "text-foreground" : "text-muted-foreground"}`}>
                  {p.name}
                </p>
                <p className="text-xs text-muted-foreground">{p.description}</p>
              </div>
              <Button
                size="sm"
                disabled={!p.available || loading === p.id}
                onClick={() => handleConnect(p.id)}
              >
                {loading === p.id ? "Redirecting..." : "Connect"}
              </Button>
            </div>
          ))}
        </div>
      </DialogContent>
    </Dialog>
  );
}
