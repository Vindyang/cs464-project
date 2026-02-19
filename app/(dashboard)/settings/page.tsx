"use client";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";

export default function SettingsPage() {
  return (
    <div className="space-y-6 max-w-4xl">
      <div>
        <h2 className="text-lg font-medium">Settings</h2>
        <p className="text-sm text-muted-foreground">
          Manage your account, redundancy, and storage preferences.
        </p>
      </div>

      <Tabs defaultValue="account" className="space-y-4">
        <TabsList>
          <TabsTrigger value="account">Account</TabsTrigger>
          <TabsTrigger value="redundancy">Redundancy</TabsTrigger>
          <TabsTrigger value="storage">Storage</TabsTrigger>
        </TabsList>

        {/* Account Settings */}
        <TabsContent value="account">
          <Card>
            <CardHeader>
              <CardTitle>Account Information</CardTitle>
              <CardDescription>
                Make changes to your account here.
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="email">Email</Label>
                <Input id="email" defaultValue="user@example.com" disabled />
              </div>
              <div className="space-y-2">
                <Label htmlFor="auth-method">Authentication Method</Label>
                <Input id="auth-method" defaultValue="Google OAuth" disabled />
              </div>
            </CardContent>
            <CardFooter className="border-t px-6 py-4">
               <Button variant="destructive">Delete Account</Button>
            </CardFooter>
          </Card>
        </TabsContent>

        {/* Redundancy Configuration */}
        <TabsContent value="redundancy">
          <Card>
            <CardHeader>
              <CardTitle>Reed-Solomon Configuration</CardTitle>
              <CardDescription>
                Configure how your files are split and distributed.
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
               <div className="p-4 border rounded-md bg-muted/50">
                  <div className="flex justify-between items-center mb-2">
                     <span className="font-semibold">Current Scheme: (6,4)</span>
                     <span className="text-xs px-2 py-1 rounded bg-secondary text-secondary-foreground">Balanced</span>
                  </div>
                  <p className="text-sm text-muted-foreground">
                     Files are split into 6 shards. You need any 4 shards to recover the file. 
                     This adds 50% storage overhead.
                  </p>
               </div>
               
               <div className="space-y-2">
                  <Label>Preset Configurations</Label>
                  <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                     <Button variant="secondary" className="justify-start h-auto py-3 px-4 flex-col items-start gap-1">
                        <span className="font-medium">Balanced (6,4)</span>
                        <span className="text-xs text-muted-foreground">50% Overhead</span>
                     </Button>
                     <Button variant="outline" className="justify-start h-auto py-3 px-4 flex-col items-start gap-1">
                        <span className="font-medium">High Redundancy (8,4)</span>
                        <span className="text-xs text-muted-foreground">100% Overhead</span>
                     </Button>
                     <Button variant="outline" className="justify-start h-auto py-3 px-4 flex-col items-start gap-1">
                        <span className="font-medium">Low Overhead (10,8)</span>
                        <span className="text-xs text-muted-foreground">25% Overhead</span>
                     </Button>
                  </div>
               </div>
               
               <div className="flex items-center gap-2 text-muted-foreground text-sm bg-muted p-3 rounded-md">
                   Changing this configuration will only affect new uploads. Existing files will remain on their current scheme.
               </div>
            </CardContent>
             <CardFooter className="border-t px-6 py-4">
               <Button>Save Changes</Button>
            </CardFooter>
          </Card>
        </TabsContent>

        {/* Storage Preferences */}
        <TabsContent value="storage">
          <Card>
            <CardHeader>
              <CardTitle>Storage Preferences</CardTitle>
              <CardDescription>
                 Manage default behaviors for your decentralized storage.
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-6">
              <div className="flex items-center justify-between space-x-2">
                <div className="space-y-0.5">
                   <Label className="text-base">Default Encryption</Label>
                   <p className="text-sm text-muted-foreground">
                      Automatically check "Encrypt file" on new uploads.
                   </p>
                </div>
                <Switch defaultChecked />
              </div>
              <Separator />
               <div className="flex items-center justify-between space-x-2">
                <div className="space-y-0.5">
                   <Label className="text-base">Auto-Delete Old Files</Label>
                   <p className="text-sm text-muted-foreground">
                      Automatically remove files older than 30 days.
                   </p>
                </div>
                <Switch />
              </div>
            </CardContent>
            <CardFooter className="border-t px-6 py-4">
               <Button>Save Changes</Button>
            </CardFooter>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  );
}
