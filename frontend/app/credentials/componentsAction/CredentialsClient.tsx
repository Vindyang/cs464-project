"use client";

import { useMemo, useState } from "react";
import { useRouter } from "next/navigation";
import { Trash2 } from "lucide-react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
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
    field2: { label: "Client Secret", placeholder: "Azure client secret value" },
    field3: { label: "Redirect URI", placeholder: "http://localhost:8080/api/oauth/onedrive/callback", default: "http://localhost:8080/api/oauth/onedrive/callback" },
  },
};

interface CredentialsClientProps {
  initialCredentials: ProviderCredential[];
}

export function CredentialsClient({ initialCredentials }: CredentialsClientProps) {
  const router = useRouter();
  const [credentials, setCredentials] = useState(initialCredentials);
  const [revealedCredentialIds, setRevealedCredentialIds] = useState<string[]>([]);
  const [providerId, setProviderId] = useState("googleDrive");
  const [clientId, setClientId] = useState("");
  const [clientSecret, setClientSecret] = useState("");
  const [redirectUri, setRedirectUri] = useState("http://localhost:8080/api/oauth/gdrive/callback");
  const [pendingDeleteId, setPendingDeleteId] = useState<string | null>(null);
  const [deleting, setDeleting] = useState(false);

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
    setDeleting(true);
    try {
      await deleteCredential(id);
      const persisted = await getCredentials();
      setCredentials(persisted);
      toast.success("Credentials deleted");
      router.refresh();
      setPendingDeleteId(null);
    } catch {
      toast.error("Failed to delete credentials");
    } finally {
      setDeleting(false);
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

  function toggleReveal(provider: string) {
    setRevealedCredentialIds((current) =>
      current.includes(provider)
        ? current.filter((value) => value !== provider)
        : [...current, provider],
    );
  }

  function maskValue(value: string, provider: string) {
    if (revealedCredentialIds.includes(provider)) {
      return value;
    }
    if (value.length <= 8) {
      return "•".repeat(Math.max(value.length, 4));
    }
    return `${value.slice(0, 4)}••••${value.slice(-4)}`;
  }

  return (
    <main className="max-w-5xl">
      <div className="mb-6 border-b pb-4">
        <p className="font-mono text-[12px] uppercase tracking-widest text-neutral-500">
          Setup
        </p>
        <h1 className="text-2xl font-semibold tracking-tight">Provider Credentials</h1>
        <p className="mt-1 text-sm text-neutral-600">
          Add at least one provider credential to unlock dashboard, upload, and download operations.
        </p>
      </div>

      <section className="grid gap-4 lg:grid-cols-[1.2fr_1fr]">
        <div className="border p-4">
          <h2 className="font-mono text-sm uppercase tracking-wider text-neutral-500">
            Add or Update
          </h2>
          <div className="mt-3 space-y-3">
            <label className="block">
              <span className="mb-1 block font-mono text-[11px] uppercase tracking-wider text-neutral-500">
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
                className="w-full border bg-white px-3 py-2 font-mono text-sm outline-none focus:ring-1 focus:ring-black"
              >
                {PROVIDERS.map((p) => (
                  <option key={p.id} value={p.id}>
                    {p.label}
                  </option>
                ))}
              </select>
            </label>

            <label className="block">
              <span className="mb-1 block font-mono text-[11px] uppercase tracking-wider text-neutral-500">
                {fields.field1.label}
              </span>
              <input
                value={clientId}
                onChange={(e) => setClientId(e.target.value)}
                className="w-full border bg-white px-3 py-2 font-mono text-sm outline-none focus:ring-1 focus:ring-black"
                placeholder={fields.field1.placeholder}
              />
            </label>

            <label className="block">
              <span className="mb-1 block font-mono text-[11px] uppercase tracking-wider text-neutral-500">
                {fields.field2.label}
              </span>
              <input
                type="password"
                value={clientSecret}
                onChange={(e) => setClientSecret(e.target.value)}
                className="w-full border bg-white px-3 py-2 font-mono text-sm outline-none focus:ring-1 focus:ring-black"
                placeholder={fields.field2.placeholder}
              />
            </label>

            <label className="block">
              <span className="mb-1 block font-mono text-[11px] uppercase tracking-wider text-neutral-500">
                {fields.field3.label}
              </span>
              <input
                value={redirectUri}
                onChange={(e) => setRedirectUri(e.target.value)}
                className="w-full border bg-white px-3 py-2 font-mono text-sm outline-none focus:ring-1 focus:ring-black"
                placeholder={fields.field3.placeholder}
              />
            </label>

            <button
              onClick={handleSave}
              disabled={saving}
              className="font-mono text-[11px] uppercase tracking-wider border border-black bg-black px-4 py-2 text-white transition-colors hover:bg-neutral-800 disabled:cursor-not-allowed disabled:opacity-50"
            >
              {saving ? "Saving..." : "Save Credentials"}
            </button>
          </div>
        </div>

        <div className="border p-4">
          <div className="mb-3 flex items-center justify-between">
            <h2 className="font-mono text-sm uppercase tracking-wider text-neutral-500">
              Stored Credentials
            </h2>
            <span className="font-mono text-[11px] text-neutral-400">
              {credentials.length} total
            </span>
          </div>

          {credentials.length === 0 ? (
            <p className="font-mono text-sm text-neutral-400">No credentials saved yet.</p>
          ) : (
            <ul className="space-y-2">
              {credentials.map((cred) => (
                <li key={cred.provider_id} className="border px-3 py-3">
                  <div className="flex items-start justify-between gap-3">
                    <div>
                      <span className="font-mono text-sm">{cred.provider_id}</span>
                      <p className="mt-1 font-mono text-[11px] text-neutral-500">
                        Updated {new Date(cred.updated_at).toLocaleString()}
                      </p>
                    </div>
                    <div className="flex items-center gap-2">
                      <button
                        type="button"
                        onClick={() => toggleReveal(cred.provider_id)}
                        className="font-mono text-[11px] uppercase tracking-wider text-neutral-500 transition-colors hover:text-black"
                      >
                        {revealedCredentialIds.includes(cred.provider_id) ? "Hide" : "Reveal"}
                      </button>
                      <button
                        onClick={() => setPendingDeleteId(cred.provider_id)}
                        aria-label={`Delete ${cred.provider_id} credential`}
                        className="text-red-500 transition-colors hover:text-red-700"
                      >
                        <Trash2 className="h-4 w-4" />
                      </button>
                    </div>
                  </div>
                  <dl className="mt-3 space-y-2">
                    <div className="grid grid-cols-[96px_1fr] gap-2">
                      <dt className="font-mono text-[11px] uppercase tracking-wider text-neutral-400">
                        Client ID
                      </dt>
                      <dd className="font-mono text-sm break-all text-neutral-700">
                        {maskValue(cred.client_id, cred.provider_id)}
                      </dd>
                    </div>
                    <div className="grid grid-cols-[96px_1fr] gap-2">
                      <dt className="font-mono text-[11px] uppercase tracking-wider text-neutral-400">
                        {cred.provider_id === "awsS3" ? "Region" : "Redirect URI"}
                      </dt>
                      <dd className="font-mono text-sm break-all text-neutral-700">
                        {cred.redirect_uri}
                      </dd>
                    </div>
                  </dl>
                </li>
              ))}
            </ul>
          )}

          <button
            onClick={handleContinue}
            disabled={credentials.length === 0}
            className="mt-4 font-mono text-[11px] uppercase tracking-wider border px-4 py-2 transition-colors hover:bg-black hover:text-white disabled:cursor-not-allowed disabled:border-neutral-200 disabled:text-neutral-300 disabled:hover:bg-transparent"
          >
            Continue to Dashboard
          </button>
        </div>
      </section>
      <p className="mt-4 font-mono text-[11px] text-neutral-400">
        Supported providers in backend right now: {CREDENTIAL_PROVIDERS.join(", ")}
      </p>

      <Dialog
        open={pendingDeleteId !== null}
        onOpenChange={(open) => {
          if (!open && !deleting) setPendingDeleteId(null);
        }}
      >
        <DialogContent className="sm:max-w-sm">
          <DialogHeader>
            <DialogTitle className="font-mono text-sm uppercase tracking-wider">
              Delete Credential
            </DialogTitle>
            <DialogDescription className="font-mono text-[12px] text-neutral-600">
              {pendingDeleteId
                ? `Remove stored credential for ${pendingDeleteId}? This action cannot be undone.`
                : "Remove this stored credential? This action cannot be undone."}
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <button
              type="button"
              onClick={() => setPendingDeleteId(null)}
              disabled={deleting}
              className="font-mono text-[11px] uppercase tracking-wider border px-3 py-2 text-neutral-600 transition-colors hover:bg-neutral-100 disabled:cursor-not-allowed disabled:opacity-50"
            >
              Cancel
            </button>
            <button
              type="button"
              onClick={() => pendingDeleteId && handleDelete(pendingDeleteId)}
              disabled={!pendingDeleteId || deleting}
              className="font-mono text-[11px] uppercase tracking-wider border border-red-600 bg-red-600 px-3 py-2 text-white transition-colors hover:bg-red-700 disabled:cursor-not-allowed disabled:opacity-50"
            >
              {deleting ? "Deleting..." : "Delete"}
            </button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </main>
  );
}
