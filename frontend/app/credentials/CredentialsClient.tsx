"use client";

import { useMemo, useState } from "react";
import { useRouter } from "next/navigation";
import {
  deleteCredential,
  ProviderCredential,
  saveCredential,
} from "@/lib/api/credentials";
import { toast } from "sonner";

const PROVIDERS = [
  { id: "googleDrive", label: "Google Drive" },
  { id: "awsS3", label: "AWS S3" },
  { id: "oneDrive", label: "OneDrive" },
];

interface CredentialsClientProps {
  initialCredentials: ProviderCredential[];
}

export function CredentialsClient({ initialCredentials }: CredentialsClientProps) {
  const router = useRouter();
  const [credentials, setCredentials] = useState(initialCredentials);
  const [providerId, setProviderId] = useState("googleDrive");
  const [jsonText, setJsonText] = useState("{}");
  const [saving, setSaving] = useState(false);

  const selectedName = useMemo(
    () => PROVIDERS.find((p) => p.id === providerId)?.label ?? providerId,
    [providerId],
  );

  async function handleSave() {
    let payload: unknown;
    try {
      payload = JSON.parse(jsonText);
    } catch {
      toast.error("Payload must be valid JSON");
      return;
    }

    setSaving(true);
    try {
      await saveCredential(providerId, payload);
      const next = credentials.filter((c) => c.providerId !== providerId);
      next.push({
        providerId,
        payload,
        updatedAt: new Date().toISOString(),
      });
      next.sort((a, b) => a.providerId.localeCompare(b.providerId));
      setCredentials(next);
      toast.success(`${selectedName} credentials saved`);
      router.refresh();
    } catch {
      toast.error("Failed to save credentials");
    } finally {
      setSaving(false);
    }
  }

  async function handleDelete(id: string) {
    try {
      await deleteCredential(id);
      setCredentials((prev) => prev.filter((c) => c.providerId !== id));
      toast.success("Credentials deleted");
      router.refresh();
    } catch {
      toast.error("Failed to delete credentials");
    }
  }

  return (
    <main className="mx-auto min-h-screen w-full max-w-5xl p-6">
      <div className="mb-6 border-b pb-4">
        <p className="font-mono text-[11px] uppercase tracking-widest text-neutral-500">
          Setup
        </p>
        <h1 className="text-2xl font-semibold tracking-tight">Provider Credentials</h1>
        <p className="mt-1 text-sm text-neutral-600">
          Add at least one provider credential to unlock dashboard, upload, and download operations.
        </p>
      </div>

      <section className="grid gap-4 lg:grid-cols-[1.2fr_1fr]">
        <div className="border p-4">
          <h2 className="font-mono text-xs uppercase tracking-wider text-neutral-500">
            Add or Update
          </h2>
          <div className="mt-3 space-y-3">
            <label className="block">
              <span className="mb-1 block font-mono text-[10px] uppercase tracking-wider text-neutral-500">
                Provider
              </span>
              <select
                value={providerId}
                onChange={(e) => setProviderId(e.target.value)}
                className="w-full border bg-white px-3 py-2 font-mono text-xs outline-none focus:ring-1 focus:ring-black"
              >
                {PROVIDERS.map((p) => (
                  <option key={p.id} value={p.id}>
                    {p.label}
                  </option>
                ))}
              </select>
            </label>

            <label className="block">
              <span className="mb-1 block font-mono text-[10px] uppercase tracking-wider text-neutral-500">
                JSON Payload
              </span>
              <textarea
                value={jsonText}
                onChange={(e) => setJsonText(e.target.value)}
                rows={12}
                className="w-full border bg-white px-3 py-2 font-mono text-xs outline-none focus:ring-1 focus:ring-black"
              />
            </label>

            <button
              onClick={handleSave}
              disabled={saving}
              className="font-mono text-[10px] uppercase tracking-wider border border-black bg-black px-4 py-2 text-white transition-colors hover:bg-neutral-800 disabled:cursor-not-allowed disabled:opacity-50"
            >
              {saving ? "Saving..." : "Save Credentials"}
            </button>
          </div>
        </div>

        <div className="border p-4">
          <div className="mb-3 flex items-center justify-between">
            <h2 className="font-mono text-xs uppercase tracking-wider text-neutral-500">
              Stored Credentials
            </h2>
            <span className="font-mono text-[10px] text-neutral-400">
              {credentials.length} total
            </span>
          </div>

          {credentials.length === 0 ? (
            <p className="font-mono text-xs text-neutral-400">No credentials saved yet.</p>
          ) : (
            <ul className="space-y-2">
              {credentials.map((cred) => (
                <li key={cred.providerId} className="border px-3 py-2">
                  <div className="flex items-center justify-between">
                    <span className="font-mono text-xs">{cred.providerId}</span>
                    <button
                      onClick={() => handleDelete(cred.providerId)}
                      className="font-mono text-[10px] uppercase tracking-wider text-neutral-400 hover:text-black"
                    >
                      Delete
                    </button>
                  </div>
                  <p className="mt-1 font-mono text-[10px] text-neutral-400">
                    Updated {new Date(cred.updatedAt).toLocaleString()}
                  </p>
                </li>
              ))}
            </ul>
          )}

          <button
            onClick={() => router.push("/dashboard")}
            disabled={credentials.length === 0}
            className="mt-4 font-mono text-[10px] uppercase tracking-wider border px-4 py-2 transition-colors hover:bg-black hover:text-white disabled:cursor-not-allowed disabled:border-neutral-200 disabled:text-neutral-300 disabled:hover:bg-transparent"
          >
            Continue to Dashboard
          </button>
        </div>
      </section>
    </main>
  );
}
