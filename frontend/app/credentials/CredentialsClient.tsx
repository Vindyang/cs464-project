"use client";

import { useMemo, useState } from "react";
import { useRouter } from "next/navigation";
import {
  CREDENTIAL_PROVIDERS,
  deleteCredential,
  getCredentials,
  getCredentialStatus,
  ProviderCredential,
  saveCredential,
} from "@/lib/api/credentials";
import { toast } from "sonner";

const PROVIDERS = [
  { id: "googleDrive", label: "Google Drive" },
  { id: "awsS3", label: "AWS S3" },
  { id: "oneDrive", label: "OneDrive" },
];

const PROVIDER_FIELDS: Record<string, {
  field1: { label: string; placeholder: string };
  field2: { label: string; placeholder: string };
  field3: { label: string; placeholder: string; default: string };
  field4?: { label: string; placeholder: string };
}> = {
  googleDrive: {
    field1: { label: "Client ID", placeholder: "Google OAuth client_id" },
    field2: { label: "Client Secret", placeholder: "Google OAuth client_secret" },
    field3: { label: "Redirect URI", placeholder: "http://localhost:8080/api/oauth/gdrive/callback", default: "http://localhost:8080/api/oauth/gdrive/callback" },
  },
  awsS3: {
    field1: { label: "Access Key ID", placeholder: "AKIAIOSFODNN7EXAMPLE" },
    field2: { label: "Secret Access Key", placeholder: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY" },
    field3: { label: "Region", placeholder: "us-east-1", default: "ap-southeast-1" },
  },
  oneDrive: {
    field1: { label: "Client ID", placeholder: "Azure app (client) ID" },
    field2: { label: "Client Secret", placeholder: "Azure app client secret" },
    field3: { label: "Redirect URI", placeholder: "http://localhost:8080/api/oauth/onedrive/callback", default: "" },
  },
};

interface CredentialsClientProps {
  initialCredentials: ProviderCredential[];
}

export function CredentialsClient({ initialCredentials }: CredentialsClientProps) {
  const router = useRouter();
  const [credentials, setCredentials] = useState(initialCredentials);
  const [providerId, setProviderId] = useState("googleDrive");
  const [clientId, setClientId] = useState("");
  const [clientSecret, setClientSecret] = useState("");
  const [redirectUri, setRedirectUri] = useState("http://localhost:8080/api/oauth/gdrive/callback");

  const fields = PROVIDER_FIELDS[providerId] ?? PROVIDER_FIELDS.googleDrive;
  const [saving, setSaving] = useState(false);

  const selectedName = useMemo(
    () => PROVIDERS.find((p) => p.id === providerId)?.label ?? providerId,
    [providerId],
  );

  async function handleSave() {
    if (!clientId || !clientSecret || !redirectUri) {
      toast.error(`${fields.field1.label}, ${fields.field2.label}, and ${fields.field3.label} are required`);
      return;
    }
    setSaving(true);
    try {
      await saveCredential(providerId, {
        clientId,
        clientSecret,
        redirectUri,
      });
      const persisted = await getCredentials();
      setCredentials(persisted);
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
      const persisted = await getCredentials();
      setCredentials(persisted);
      toast.success("Credentials deleted");
      router.refresh();
    } catch {
      toast.error("Failed to delete credentials");
    }
  }

  async function handleContinue() {
    try {
      const status = await getCredentialStatus();
      if (!status.configured) {
        toast.error("Please save at least one credential first");
        return;
      }
      router.push("/dashboard");
    } catch {
      toast.error("Unable to verify credentials status");
    }
  }

  return (
    <main className="max-w-5xl">
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
                onChange={(e) => {
                  const id = e.target.value;
                  setProviderId(id);
                  setClientId("");
                  setClientSecret("");
                  setRedirectUri(PROVIDER_FIELDS[id]?.field3.default ?? "");
                }}
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
                {fields.field1.label}
              </span>
              <input
                value={clientId}
                onChange={(e) => setClientId(e.target.value)}
                className="w-full border bg-white px-3 py-2 font-mono text-xs outline-none focus:ring-1 focus:ring-black"
                placeholder={fields.field1.placeholder}
              />
            </label>

            <label className="block">
              <span className="mb-1 block font-mono text-[10px] uppercase tracking-wider text-neutral-500">
                {fields.field2.label}
              </span>
              <input
                type="password"
                value={clientSecret}
                onChange={(e) => setClientSecret(e.target.value)}
                className="w-full border bg-white px-3 py-2 font-mono text-xs outline-none focus:ring-1 focus:ring-black"
                placeholder={fields.field2.placeholder}
              />
            </label>

            <label className="block">
              <span className="mb-1 block font-mono text-[10px] uppercase tracking-wider text-neutral-500">
                {fields.field3.label}
              </span>
              <input
                value={redirectUri}
                onChange={(e) => setRedirectUri(e.target.value)}
                className="w-full border bg-white px-3 py-2 font-mono text-xs outline-none focus:ring-1 focus:ring-black"
                placeholder={fields.field3.placeholder}
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
                <li key={cred.provider_id} className="border px-3 py-2">
                  <div className="flex items-center justify-between">
                    <span className="font-mono text-xs">{cred.provider_id}</span>
                    <button
                      onClick={() => handleDelete(cred.provider_id)}
                      className="font-mono text-[10px] uppercase tracking-wider text-neutral-400 hover:text-black"
                    >
                      Delete
                    </button>
                  </div>
                  <p className="mt-1 font-mono text-[10px] text-neutral-400">
                    {cred.client_id.slice(0, 14)}... · {cred.redirect_uri}
                  </p>
                  <p className="mt-0.5 font-mono text-[10px] text-neutral-300">
                    Updated {new Date(cred.updated_at).toLocaleString()}
                  </p>
                </li>
              ))}
            </ul>
          )}

          <button
            onClick={handleContinue}
            disabled={credentials.length === 0}
            className="mt-4 font-mono text-[10px] uppercase tracking-wider border px-4 py-2 transition-colors hover:bg-black hover:text-white disabled:cursor-not-allowed disabled:border-neutral-200 disabled:text-neutral-300 disabled:hover:bg-transparent"
          >
            Continue to Dashboard
          </button>
        </div>
      </section>
      <p className="mt-4 font-mono text-[10px] text-neutral-400">
        Supported providers in backend right now: {CREDENTIAL_PROVIDERS.join(", ")}
      </p>
    </main>
  );
}
