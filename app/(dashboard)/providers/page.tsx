"use client";

import { ProviderCard } from "@/components/dashboard/ProviderCard";
import { mockProviders } from "@/lib/mocks/providers";
import { Button } from "@/components/ui/button";
import { Plus } from "lucide-react";

export default function ProvidersPage() {
  const handleConnect = (id: string) => {
     console.log("Connect provider", id);
  };

  const handleDisconnect = (id: string) => {
     console.log("Disconnect provider", id);
  };
  
  const handleRefresh = (id: string) => {
      console.log("Refresh provider", id);
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-lg font-medium">Cloud Providers</h2>
          <p className="text-sm text-muted-foreground">
            Connect at least 2 providers to ensure data redundancy.
          </p>
        </div>
        <Button>
           <Plus className="mr-2 h-4 w-4" />
           Add Provider
        </Button>
      </div>

      <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
        {mockProviders.map((provider) => (
          <ProviderCard
            key={provider.id}
            {...provider}
            onConnect={() => handleConnect(provider.id)}
            onDisconnect={() => handleDisconnect(provider.id)}
            onRefresh={() => handleRefresh(provider.id)}
          />
        ))}
      </div>
    </div>
  );
}
