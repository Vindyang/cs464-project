"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { ProviderCard, ProviderCardProps } from "@/components/dashboard/ProviderCard";
import { Button } from "@/components/ui/button";
import { mockProviders } from "@/lib/mocks/providers";
import { CheckCircle2, ArrowRight, ShieldCheck } from "lucide-react";
import { GridBackground } from "@/components/ui/grid-background";

export default function OnboardingPage() {
  const router = useRouter();
  
  // Initialize providers as disconnected for onboarding demo
  const [providers, setProviders] = useState(
    mockProviders.map(p => ({
      ...p,
      status: "disconnected" as ProviderCardProps["status"],
      used: "0 GB",
      percentage: 0,
      shardCount: 0,
      lastCheck: "Never"
    }))
  );

  const connectedCount = providers.filter(p => p.status === "connected").length;
  const minProviders = 2;
  const isReady = connectedCount >= minProviders;

  const handleConnect = (id: string) => {
    // Simulate connection delay
    setProviders(prev => prev.map(p => {
      if (p.id === id) {
        return { 
          ...p, 
          status: "connected",
          lastCheck: "Just now" 
        };
      }
      return p;
    }));
  };

  const handleDisconnect = (id: string) => {
    setProviders(prev => prev.map(p => 
      p.id === id ? { ...p, status: "disconnected" } : p
    ));
  };

  const handleContinue = () => {
    if (isReady) {
      router.push("/dashboard");
    }
  };

  return (
    <div className="min-h-screen bg-bg-canvas relative isolate flex flex-col">
      <GridBackground />
      
      <div className="container max-w-5xl py-12 flex-1 relative z-10">
        <div className="max-w-2xl mx-auto text-center mb-12">
           <div className="inline-flex items-center justify-center w-12 h-12 bg-accent-primary/10 text-accent-primary rounded-full mb-6">
              <ShieldCheck className="w-6 h-6" />
           </div>
           <h1 className="text-3xl font-semibold tracking-[-0.02em] text-text-main mb-4">
              Initialize Your Vault
           </h1>
           <p className="text-text-secondary text-lg leading-relaxed">
              To ensure redundancy and fault tolerance, please connect at least <span className="font-semibold text-text-main">2 storage providers</span>.
              Your data will be sharded across them.
           </p>
        </div>

        <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
          {providers.map((provider) => (
            <ProviderCard
              key={provider.id}
              {...provider}
              onConnect={() => handleConnect(provider.id)}
              onDisconnect={() => handleDisconnect(provider.id)}
              onRefresh={() => {}}
            />
          ))}
        </div>
      </div>

      <div className="sticky bottom-0 left-0 right-0 p-4 bg-bg-canvas/80 backdrop-blur-md border-t border-border-color z-50">
        <div className="container max-w-5xl flex items-center justify-between">
          <div className="flex items-center gap-4">
            <div className={`flex items-center justify-center w-10 h-10 rounded-full border-2 transition-colors ${isReady ? "border-accent-primary bg-accent-primary text-white" : "border-border-color text-text-tertiary"}`}>
              {isReady ? <CheckCircle2 className="w-5 h-5" /> : <span className="text-sm font-bold">{connectedCount}/{minProviders}</span>}
            </div>
            <div className="flex flex-col">
              <span className={`font-medium text-sm ${isReady ? "text-text-main" : "text-text-secondary"}`}>
                {isReady ? "Ready to initialize" : `Connect ${minProviders - connectedCount} more provider${minProviders - connectedCount === 1 ? '' : 's'}`}
              </span>
              <span className="text-xs text-text-tertiary">
                {connectedCount} of {minProviders} required
              </span>
            </div>
          </div>
          
          <Button 
             onClick={handleContinue} 
             disabled={!isReady} 
             className="h-10 px-6 bg-accent-primary text-white hover:bg-accent-primary-hover rounded-[2px]"
          >
            Enter Dashboard
            <ArrowRight className="ml-2 w-4 h-4" />
          </Button>
        </div>
      </div>
    </div>
  );
}
