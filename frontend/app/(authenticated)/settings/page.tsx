"use client";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { DashboardCard } from "@/components/dashboard/DashboardCard";
import { Separator } from "@/components/ui/separator";
import { useState } from "react";
import { toast } from "sonner";
import { cn } from "@/lib/utils";
import { Shield, HardDrive, User, Trash2, Save } from "lucide-react";

export default function SettingsPage() {
  const [redundancyScheme, setRedundancyScheme] = useState<"(6,4)" | "(8,4)" | "(10,8)">("(6,4)");
  const [encryptDefault, setEncryptDefault] = useState(true);
  const [autoDelete, setAutoDelete] = useState(false);

  const handleSavePreferences = () => {
    toast.promise(
        new Promise((resolve) => setTimeout(resolve, 800)),
        {
            loading: "Saving system preferences...",
            success: "System configuration updated",
            error: "Failed to save preferences"
        }
    );
  };

  const handleDeleteAccount = () => {
      toast("Delete account?", {
          description: "This action cannot be undone. All shards will be marked for deletion.",
          action: {
              label: "Delete",
              onClick: () => toast.error("Account deletion is disabled in demo mode")
          }
      })
  }

  return (
    <div className="space-y-6 max-w-5xl">
      <div>
        <h2 className="text-xl font-semibold tracking-tight text-text-main">System Settings</h2>
        <p className="text-sm text-text-secondary mt-1">
          Manage your vault configuration, redundancy strategies, and access controls.
        </p>
      </div>

      <Tabs defaultValue="account" className="space-y-6">
        <TabsList className="bg-bg-subtle border border-border-color p-1 rounded-[2px] h-auto">
          <TabsTrigger value="account" className="rounded-[1px] data-[state=active]:bg-bg-canvas data-[state=active]:text-text-main data-[state=active]:shadow-sm">Account</TabsTrigger>
          <TabsTrigger value="redundancy" className="rounded-[1px] data-[state=active]:bg-bg-canvas data-[state=active]:text-text-main data-[state=active]:shadow-sm">Redundancy</TabsTrigger>
          <TabsTrigger value="storage" className="rounded-[1px] data-[state=active]:bg-bg-canvas data-[state=active]:text-text-main data-[state=active]:shadow-sm">Storage</TabsTrigger>
        </TabsList>

        {/* Account Settings */}
        <TabsContent value="account">
          <DashboardCard>
            <div className="p-6 border-b border-border-color">
                <div className="flex items-center gap-2 mb-1">
                    <User className="w-5 h-5 text-accent-primary" />
                    <h3 className="text-lg font-medium text-text-main">Account Information</h3>
                </div>
                <p className="text-sm text-text-secondary">Manage your identity and access credentials.</p>
            </div>
            
            <div className="p-6 space-y-6">
              <div className="grid gap-2">
                <Label htmlFor="email">Email Address</Label>
                <Input id="email" defaultValue="user@example.com" disabled className="bg-bg-subtle border-border-color" />
              </div>
              <div className="grid gap-2">
                <Label htmlFor="auth-method">Authentication Provider</Label>
                <Input id="auth-method" defaultValue="Google OAuth" disabled className="bg-bg-subtle border-border-color" />
              </div>
              
              <div className="pt-4">
                 <Button variant="destructive" onClick={handleDeleteAccount} className="bg-red-600 hover:bg-red-700 text-white rounded-[2px]">
                    <Trash2 className="w-4 h-4 mr-2" />
                    Delete Account
                 </Button>
              </div>
            </div>
          </DashboardCard>
        </TabsContent>

        {/* Redundancy Configuration */}
        <TabsContent value="redundancy">
          <DashboardCard>
            <div className="p-6 border-b border-border-color">
                <div className="flex items-center gap-2 mb-1">
                    <Shield className="w-5 h-5 text-accent-primary" />
                    <h3 className="text-lg font-medium text-text-main">Reed-Solomon Configuration</h3>
                </div>
                <p className="text-sm text-text-secondary">Configure erasure coding parameters for file distribution.</p>
            </div>
            
            <div className="p-6 space-y-6">
               <div className="p-4 border border-border-color rounded-[2px] bg-bg-subtle/30">
                  <div className="flex justify-between items-center mb-3">
                     <span className="font-semibold text-text-main">Current Scheme: {redundancyScheme}</span>
                     <span className={cn(
                         "text-[10px] uppercase font-bold px-2 py-1 rounded-[2px] tracking-wider",
                         redundancyScheme === "(6,4)" ? "bg-accent-primary/10 text-accent-primary border border-accent-primary/20" :
                         redundancyScheme === "(8,4)" ? "bg-blue-500/10 text-blue-600 border border-blue-500/20" :
                         "bg-emerald-500/10 text-emerald-600 border border-emerald-500/20"
                     )}>
                         {redundancyScheme === "(6,4)" ? "Standard" : redundancyScheme === "(8,4)" ? "High Reliability" : "Efficiency Mode"}
                     </span>
                  </div>
                  <p className="text-sm text-text-secondary leading-relaxed">
                     {redundancyScheme === "(6,4)" && "Files are split into 6 shards (4 data + 2 parity). Requires 4 shards to reconstruct. 1.5x storage overhead."}
                     {redundancyScheme === "(8,4)" && "Files are split into 8 shards (4 data + 4 parity). Requires 4 shards to reconstruct. 2.0x storage overhead."}
                     {redundancyScheme === "(10,8)" && "Files are split into 10 shards (8 data + 2 parity). Requires 8 shards to reconstruct. 1.25x storage overhead."}
                  </p>
               </div>
               
               <div className="space-y-4">
                  <Label>Preset Configurations</Label>
                  <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                     {[
                        { val: "(6,4)", label: "Standard (6,4)", sub: "1.5x Overhead" },
                        { val: "(8,4)", label: "High Rel (8,4)", sub: "2.0x Overhead" },
                        { val: "(10,8)", label: "Efficient (10,8)", sub: "1.25x Overhead" }
                     ].map((opt) => (
                         <Button 
                            key={opt.val}
                            variant={redundancyScheme === opt.val ? "default" : "outline"} 
                            className={cn(
                                "justify-start h-auto py-4 px-4 flex-col items-start gap-1 rounded-[2px] border-border-color",
                                redundancyScheme === opt.val ? "bg-accent-primary hover:bg-accent-primary-hover text-white border-transparent" : "hover:bg-bg-subtle text-text-main"
                            )}
                            onClick={() => setRedundancyScheme(opt.val as "(6,4)" | "(8,4)" | "(10,8)")}
                         >
                            <span className="font-semibold">{opt.label}</span>
                            <span className={cn("text-xs opacity-80", redundancyScheme === opt.val ? "text-white" : "text-text-secondary")}>{opt.sub}</span>
                         </Button>
                     ))}
                  </div>
               </div>
               
               <div className="flex items-center gap-2 text-text-tertiary text-xs bg-bg-subtle p-3 rounded-[2px] border border-border-color">
                   Note: Changing this configuration will only affect future uploads. Existing shards remain unchanged.
               </div>
               
               <div className="pt-2 flex justify-end">
                   <Button onClick={handleSavePreferences} className="bg-text-main text-bg-canvas hover:bg-text-main/90 rounded-[2px]">
                       <Save className="w-4 h-4 mr-2" />
                       Save Configuration
                   </Button>
               </div>
            </div>
          </DashboardCard>
        </TabsContent>

        {/* Storage Preferences */}
        <TabsContent value="storage">
          <DashboardCard>
            <div className="p-6 border-b border-border-color">
                <div className="flex items-center gap-2 mb-1">
                    <HardDrive className="w-5 h-5 text-accent-primary" />
                    <h3 className="text-lg font-medium text-text-main">Storage Behaviors</h3>
                </div>
                <p className="text-sm text-text-secondary">Manage default behaviors for your decentralized storage.</p>
            </div>
            
            <div className="p-6 space-y-8">
              <div className="flex items-center justify-between space-x-4">
                <div className="space-y-1">
                   <Label className="text-base font-medium">Default Encryption</Label>
                   <p className="text-sm text-text-secondary">
                      Automatically encrypt all files client-side before sharding.
                   </p>
                </div>
                <Switch checked={encryptDefault} onCheckedChange={setEncryptDefault} />
              </div>
              <Separator className="bg-border-color" />
               <div className="flex items-center justify-between space-x-4">
                <div className="space-y-1">
                   <Label className="text-base font-medium">Auto-Delete Stale Files</Label>
                   <p className="text-sm text-text-secondary">
                      Automatically remove files that haven&apos;t been accessed in 30 days.
                   </p>
                </div>
                <Switch checked={autoDelete} onCheckedChange={setAutoDelete} />
              </div>
              
              <div className="pt-4 flex justify-end">
                   <Button onClick={handleSavePreferences} className="bg-text-main text-bg-canvas hover:bg-text-main/90 rounded-[2px]">
                       <Save className="w-4 h-4 mr-2" />
                       Save Preferences
                   </Button>
               </div>
            </div>
          </DashboardCard>
        </TabsContent>
      </Tabs>
    </div>
  );
}
