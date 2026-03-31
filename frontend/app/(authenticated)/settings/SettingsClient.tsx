"use client";

import { useState } from "react";
import { Switch } from "@/components/ui/switch";

import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { cn } from "@/lib/utils";
import { toast } from "sonner";

const REDUNDANCY_PRESETS = [
  { val: "(6,4)" as const, label: "(6,4)", name: "Standard", overhead: "1.5× overhead", desc: "4 data + 2 parity shards. Requires 4 to recover." },
  { val: "(8,4)" as const, label: "(8,4)", name: "High Reliability", overhead: "2.0× overhead", desc: "4 data + 4 parity shards. Requires 4 to recover." },
  { val: "(10,8)" as const, label: "(10,8)", name: "Efficient", overhead: "1.25× overhead", desc: "8 data + 2 parity shards. Requires 8 to recover." },
];

export function SettingsClient() {
  const [redundancy, setRedundancy] = useState<"(6,4)" | "(8,4)" | "(10,8)">("(6,4)");
  const [encryptDefault, setEncryptDefault] = useState(true);
  const [autoDelete, setAutoDelete] = useState(false);

  const handleSave = () => {
    toast.promise(
      new Promise((resolve) => setTimeout(resolve, 600)),
      {
        loading: "Saving preferences...",
        success: "Preferences saved",
        error: "Failed to save",
      }
    );
  };

  return (
    <div className="space-y-5">
      <div>
        <h2 className="font-mono text-sm font-bold uppercase tracking-widest">System Settings</h2>
        <p className="font-mono text-[11px] text-neutral-500 mt-1">
          Vault configuration, redundancy strategy, and storage behaviors.
        </p>
      </div>

      <Tabs defaultValue="redundancy" className="space-y-6">
        <TabsList className="bg-neutral-50 border p-1 rounded-none h-auto gap-0.5">
          {["redundancy", "storage"].map((t) => (
            <TabsTrigger
              key={t}
              value={t}
              className="rounded-none font-mono text-[10px] uppercase tracking-wider data-[state=active]:bg-white data-[state=active]:shadow-none data-[state=active]:border data-[state=inactive]:text-neutral-400"
            >
              {t}
            </TabsTrigger>
          ))}
        </TabsList>

        {/* ── Redundancy ── */}
        <TabsContent value="redundancy">
          <SettingsCard>
            <SettingsCardHeader
              title="Reed-Solomon Configuration"
              subtitle="Erasure coding preset for future uploads. Does not affect existing files."
              tag="Preference only"
            />
            <div className="p-5 space-y-5">
              <div className="space-y-2">
                {REDUNDANCY_PRESETS.map((preset) => {
                  const active = redundancy === preset.val;
                  return (
                    <button
                      key={preset.val}
                      onClick={() => setRedundancy(preset.val)}
                      className={cn(
                        "w-full text-left border px-4 py-3 transition-colors relative",
                        active ? "border-neutral-900 bg-white" : "border-neutral-200 bg-neutral-50 hover:border-neutral-400"
                      )}
                    >
                      {active && (
                        <>
                          <span className="absolute -top-px -left-px w-1.5 h-1.5 border-t border-l border-neutral-600 pointer-events-none" />
                          <span className="absolute -top-px -right-px w-1.5 h-1.5 border-t border-r border-neutral-600 pointer-events-none" />
                          <span className="absolute -bottom-px -left-px w-1.5 h-1.5 border-b border-l border-neutral-600 pointer-events-none" />
                          <span className="absolute -bottom-px -right-px w-1.5 h-1.5 border-b border-r border-neutral-600 pointer-events-none" />
                        </>
                      )}
                      <div className="flex items-baseline justify-between mb-0.5">
                        <span className="font-mono text-xs font-bold">{preset.label}</span>
                        <span className="font-mono text-[9px] uppercase tracking-widest text-neutral-400">{preset.name}</span>
                      </div>
                      <div className="font-mono text-[10px] text-neutral-500">{preset.desc}</div>
                      <div className="font-mono text-[9px] text-neutral-400 mt-0.5">{preset.overhead}</div>
                    </button>
                  );
                })}
              </div>
              <div className="flex justify-end pt-2">
                <button
                  onClick={handleSave}
                  className="font-mono text-[10px] uppercase tracking-wider border border-neutral-900 bg-neutral-900 text-white px-4 py-2 hover:bg-neutral-700 transition-colors"
                >
                  Save Configuration
                </button>
              </div>
            </div>
          </SettingsCard>
        </TabsContent>

        {/* ── Storage ── */}
        <TabsContent value="storage">
          <SettingsCard>
            <SettingsCardHeader
              title="Storage Behaviors"
              subtitle="Default behaviors for uploads and retention. Does not affect existing files."
              tag="Preference only"
            />
            <div className="p-5 space-y-0 divide-y">
              <ToggleRow
                label="Default Encryption"
                description="Encrypt all files client-side (AES-256-GCM) before sharding."
                checked={encryptDefault}
                onCheckedChange={setEncryptDefault}
              />
              <ToggleRow
                label="Auto-Delete Stale Files"
                description="Remove files not accessed in 30 days."
                checked={autoDelete}
                onCheckedChange={setAutoDelete}
              />
            </div>
            <div className="px-5 pb-5 flex justify-end">
              <button
                onClick={handleSave}
                className="font-mono text-[10px] uppercase tracking-wider border border-neutral-900 bg-neutral-900 text-white px-4 py-2 hover:bg-neutral-700 transition-colors"
              >
                Save Preferences
              </button>
            </div>
          </SettingsCard>
        </TabsContent>
      </Tabs>
    </div>
  );
}

function SettingsCard({ children }: { children: React.ReactNode }) {
  return (
    <div className="border relative">
      <span className="absolute -top-px -left-px w-1.5 h-1.5 border-t border-l border-neutral-400 opacity-50 pointer-events-none" />
      <span className="absolute -top-px -right-px w-1.5 h-1.5 border-t border-r border-neutral-400 opacity-50 pointer-events-none" />
      <span className="absolute -bottom-px -left-px w-1.5 h-1.5 border-b border-l border-neutral-400 opacity-50 pointer-events-none" />
      <span className="absolute -bottom-px -right-px w-1.5 h-1.5 border-b border-r border-neutral-400 opacity-50 pointer-events-none" />
      {children}
    </div>
  );
}

function SettingsCardHeader({
  title,
  subtitle,
  tag,
}: {
  title: string;
  subtitle: string;
  tag?: string;
}) {
  return (
    <div className="px-5 py-4 border-b bg-neutral-50 flex items-start justify-between gap-4">
      <div>
        <h3 className="font-mono text-xs font-bold uppercase tracking-wider">{title}</h3>
        <p className="font-mono text-[10px] text-neutral-500 mt-0.5">{subtitle}</p>
      </div>
      {tag && (
        <span className="font-mono text-[9px] uppercase tracking-widest border px-2 py-0.5 text-neutral-400 shrink-0">
          {tag}
        </span>
      )}
    </div>
  );
}

function ToggleRow({
  label,
  description,
  checked,
  onCheckedChange,
}: {
  label: string;
  description: string;
  checked: boolean;
  onCheckedChange: (v: boolean) => void;
}) {
  return (
    <div className="flex items-center justify-between gap-4 py-4">
      <div>
        <div className="font-mono text-xs font-medium">{label}</div>
        <div className="font-mono text-[10px] text-neutral-500 mt-0.5">{description}</div>
      </div>
      <Switch checked={checked} onCheckedChange={onCheckedChange} />
    </div>
  );
}
