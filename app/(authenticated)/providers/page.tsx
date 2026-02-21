"use client";

import { NodeTable } from "@/components/dashboard/NodeTable";
import { mockProviders } from "@/lib/mocks/providers";
import { useState } from "react";
import { Input } from "@/components/ui/input";
import { Search, Filter, Plus } from "lucide-react";
import { Button } from "@/components/ui/button";
import { toast } from "sonner";
import { useRouter } from "next/navigation";

export default function ProvidersPage() {
  const router = useRouter();
  const [search, setSearch] = useState("");
  
  const providers = mockProviders.filter(p => 
    p.name.toLowerCase().includes(search.toLowerCase())
  );

  const handleConfig = (id: string) => {
    // Navigate to provider details/config
    router.push(`/dashboard/providers/${id}`);
  };

  const handleRemove = (id: string) => {
      toast("Disconnect Provider?", {
          description: "This will stop shards from being stored on this provider.",
          action: {
              label: "Disconnect",
              onClick: () => {
                  toast.success("Provider disconnected");
              }
          }
      });
  };
  
  return (
    <div className="space-y-6">
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
         <div className="relative w-full sm:w-72">
            <Search className="absolute left-2 top-2.5 h-4 w-4 text-text-secondary" />
            <Input 
               placeholder="Search providers..." 
               className="pl-8 bg-bg-subtle border-border-color focus-visible:ring-accent-primary" 
               value={search}
               onChange={(e) => setSearch(e.target.value)}
            />
         </div>
         <div className="flex gap-2 w-full sm:w-auto">
            <Button variant="outline" size="sm" className="flex-1 sm:flex-none border-border-color text-text-main hover:bg-bg-subtle">
                <Filter className="mr-2 h-4 w-4" />
                Filter
            </Button>
            <Button size="sm" className="flex-1 sm:flex-none bg-accent-primary text-white hover:bg-accent-primary-hover">
                <Plus className="mr-2 h-4 w-4" />
                Connect New
            </Button>
         </div>
      </div>

      <NodeTable 
        providers={providers}
        onConfig={handleConfig}
        onRemove={handleRemove}
      />
    </div>
  );
}
