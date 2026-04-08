"use client";

import { useMemo, useState } from "react";
import { useRouter } from "next/navigation";
import { Eye, EyeOff, Loader2, Trash2 } from "lucide-react";
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
	revealCredential,
	RevealedProviderCredential,
	saveCredential,
} from "@/lib/api/credentials";
import { formatUtcDateTime } from "@/lib/utils";
import { toast } from "sonner";
import { helpToast } from "@/lib/help/help-toast";

const PROVIDERS = [
	{ id: "googleDrive", label: "Google Drive" },
	{ id: "awsS3", label: "AWS S3" },
	{ id: "oneDrive", label: "OneDrive" },
]

const PROVIDER_FIELDS: Record<string, {
	field1: { label: string; placeholder: string };
	field2: { label: string; placeholder: string };
	field3: { label: string; placeholder: string; default: string };
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
}

interface CredentialsClientProps {
	initialCredentials: ProviderCredential[]
}

export function CredentialsClient({ initialCredentials }: CredentialsClientProps) {
	const router = useRouter()
	const [credentials, setCredentials] = useState(initialCredentials)
	const [revealedCredentialIds, setRevealedCredentialIds] = useState<string[]>([])
	const [revealedCredentialDetails, setRevealedCredentialDetails] = useState<Record<string, RevealedProviderCredential>>({})
	const [revealingId, setRevealingId] = useState<string | null>(null)
	const [providerId, setProviderId] = useState("googleDrive")
	const [clientId, setClientId] = useState("")
	const [clientSecret, setClientSecret] = useState("")
	const [redirectUri, setRedirectUri] = useState("http://localhost:8080/api/oauth/gdrive/callback")
	const [pendingDeleteId, setPendingDeleteId] = useState<string | null>(null)
	const [deleting, setDeleting] = useState(false)
	const [saving, setSaving] = useState(false)

	const fields = PROVIDER_FIELDS[providerId] ?? PROVIDER_FIELDS.googleDrive
	const selectedName = useMemo(
		() => PROVIDERS.find((provider) => provider.id === providerId)?.label ?? providerId,
		[providerId],
	)

	async function handleSave() {
		if (!clientId || !clientSecret || !redirectUri) {
			toast.error(`${fields.field1.label}, ${fields.field2.label}, and ${fields.field3.label} are required`)
			return
		}

		setSaving(true)
		try {
			await saveCredential(providerId, {
				clientId,
				clientSecret,
				redirectUri,
			})
			const persisted = await getCredentials()
			setCredentials(persisted)
			setRevealedCredentialIds((current) => current.filter((value) => value !== providerId))
			setRevealedCredentialDetails((current) => {
				const next = { ...current }
				delete next[providerId]
				return next
			})
			toast.success(`${selectedName} credentials saved`)
			router.refresh()
		} catch {
			helpToast({ error: "Failed to save credentials", code: "UNKNOWN_ERROR" })
		} finally {
			setSaving(false)
		}
	}

	async function handleDelete(id: string) {
		setDeleting(true)
		try {
			await deleteCredential(id)
			const persisted = await getCredentials()
			setCredentials(persisted)
			setRevealedCredentialIds((current) => current.filter((value) => value !== id))
			setRevealedCredentialDetails((current) => {
				const next = { ...current }
				delete next[id]
				return next
			})
			toast.success("Credentials deleted")
			router.refresh()
			setPendingDeleteId(null)
		} catch {
			helpToast({ error: "Failed to delete credentials", code: "UNKNOWN_ERROR" })
		} finally {
			setDeleting(false)
		}
	}

	async function handleContinue() {
		try {
			const status = await getCredentialStatus()
			if (!status.configured) {
				toast.error("Please save at least one credential first")
				return
			}
			router.push("/dashboard")
		} catch {
			helpToast({ error: "Unable to verify credentials status", code: "UNKNOWN_ERROR" })
		}
	}

	async function toggleReveal(provider: string) {
		if (revealedCredentialIds.includes(provider)) {
			setRevealedCredentialIds((current) => current.filter((value) => value !== provider))
			return
		}

		if (!revealedCredentialDetails[provider]) {
			setRevealingId(provider)
			try {
				const revealed = await revealCredential(provider)
				setRevealedCredentialDetails((current) => ({
					...current,
					[provider]: revealed,
				}))
			} catch {
				helpToast({ error: "Failed to reveal credential secret", code: "UNKNOWN_ERROR" })
				setRevealingId(null)
				return
			} finally {
				setRevealingId(null)
			}
		}

		setRevealedCredentialIds((current) => [...current, provider])
	}

	function maskValue(value: string, provider: string, hiddenFallback = "••••••••••") {
		if (revealedCredentialIds.includes(provider)) {
			return value
		}
		if (!value) {
			return hiddenFallback
		}
		if (value.length <= 8) {
			return "•".repeat(Math.max(value.length, 4))
		}
		return `${value.slice(0, 4)}••••${value.slice(-4)}`
	}

	return (
		<main className="max-w-5xl">
			<div className="mb-6 border-b border-neutral-200 pb-4 dark:border-neutral-800">
				<p className="font-mono text-[12px] uppercase tracking-widest text-neutral-500 dark:text-neutral-400">
					Setup
				</p>
				<h1 className="text-2xl font-semibold tracking-tight text-neutral-950 dark:text-neutral-100">Provider Credentials</h1>
				<p className="mt-1 text-sm text-neutral-600 dark:text-neutral-400">
					Add at least one provider credential to unlock dashboard, upload, and download operations.
				</p>
			</div>

			<section className="grid gap-4 lg:grid-cols-[1.2fr_1fr]">
				<div className="border border-neutral-200 p-4 dark:border-neutral-800 dark:bg-neutral-950">
					<h2 className="font-mono text-sm uppercase tracking-wider text-neutral-500 dark:text-neutral-400">
						Add or Update
					</h2>
					<div className="mt-3 space-y-3">
						<label className="block">
							<span className="mb-1 block font-mono text-[11px] uppercase tracking-wider text-neutral-500 dark:text-neutral-400">
								Provider
							</span>
							<select
								value={providerId}
								onChange={(event) => {
									const id = event.target.value
									setProviderId(id)
									setClientId("")
									setClientSecret("")
									setRedirectUri(PROVIDER_FIELDS[id]?.field3.default ?? "")
								}}
								className="w-full border border-neutral-200 bg-white px-3 py-2 font-mono text-sm text-neutral-900 outline-none focus:border-sky-500 focus:ring-1 focus:ring-sky-500 dark:border-neutral-800 dark:bg-neutral-950 dark:text-neutral-100"
							>
								{PROVIDERS.map((provider) => (
									<option key={provider.id} value={provider.id}>
										{provider.label}
									</option>
								))}
							</select>
						</label>

						<label className="block">
							<span className="mb-1 block font-mono text-[11px] uppercase tracking-wider text-neutral-500 dark:text-neutral-400">
								{fields.field1.label}
							</span>
							<input
								value={clientId}
								onChange={(event) => setClientId(event.target.value)}
								className="w-full border border-neutral-200 bg-white px-3 py-2 font-mono text-sm text-neutral-900 outline-none focus:border-sky-500 focus:ring-1 focus:ring-sky-500 dark:border-neutral-800 dark:bg-neutral-950 dark:text-neutral-100"
								placeholder={fields.field1.placeholder}
							/>
						</label>

						<label className="block">
							<span className="mb-1 block font-mono text-[11px] uppercase tracking-wider text-neutral-500 dark:text-neutral-400">
								{fields.field2.label}
							</span>
							<input
								type="password"
								value={clientSecret}
								onChange={(event) => setClientSecret(event.target.value)}
								className="w-full border border-neutral-200 bg-white px-3 py-2 font-mono text-sm text-neutral-900 outline-none focus:border-sky-500 focus:ring-1 focus:ring-sky-500 dark:border-neutral-800 dark:bg-neutral-950 dark:text-neutral-100"
								placeholder={fields.field2.placeholder}
							/>
						</label>

						<label className="block">
							<span className="mb-1 block font-mono text-[11px] uppercase tracking-wider text-neutral-500 dark:text-neutral-400">
								{fields.field3.label}
							</span>
							<input
								value={redirectUri}
								onChange={(event) => setRedirectUri(event.target.value)}
								className="w-full border border-neutral-200 bg-white px-3 py-2 font-mono text-sm text-neutral-900 outline-none focus:border-sky-500 focus:ring-1 focus:ring-sky-500 dark:border-neutral-800 dark:bg-neutral-950 dark:text-neutral-100"
								placeholder={fields.field3.placeholder}
							/>
						</label>

						<button
							onClick={handleSave}
							disabled={saving}
							className="border border-sky-600 bg-sky-600 px-4 py-2 font-mono text-[11px] uppercase tracking-wider text-white transition-colors hover:bg-sky-700 disabled:cursor-not-allowed disabled:opacity-50"
						>
							{saving ? "Saving..." : "Save Credentials"}
						</button>
					</div>
				</div>

				<div className="border border-neutral-200 p-4 dark:border-neutral-800 dark:bg-neutral-950">
					<div className="mb-3 flex items-center justify-between">
						<h2 className="font-mono text-sm uppercase tracking-wider text-neutral-500 dark:text-neutral-400">
							Stored Credentials
						</h2>
						<span className="font-mono text-[11px] text-neutral-400 dark:text-neutral-500">
							{credentials.length} total
						</span>
					</div>

					{credentials.length === 0 ? (
						<p className="font-mono text-sm text-neutral-400 dark:text-neutral-500">No credentials saved yet.</p>
					) : (
						<ul className="space-y-2">
							{credentials.map((credential) => {
								const revealed = revealedCredentialDetails[credential.provider_id]
								const isRevealed = revealedCredentialIds.includes(credential.provider_id)
								const isRevealing = revealingId === credential.provider_id

								return (
									<li key={credential.provider_id} className="border border-neutral-200 px-3 py-3 dark:border-neutral-800 dark:bg-neutral-950">
										<div className="flex items-start justify-between gap-3">
											<div>
												<span className="font-mono text-sm text-neutral-900 dark:text-neutral-100">{credential.provider_id}</span>
												<p className="mt-1 font-mono text-[11px] text-neutral-500 dark:text-neutral-400">
													Updated {formatUtcDateTime(credential.updated_at)}
												</p>
											</div>
											<div className="flex items-center gap-2">
												<button
													type="button"
													onClick={() => toggleReveal(credential.provider_id)}
													disabled={isRevealing}
													aria-label={isRevealed ? "Hide credential" : "Reveal credential"}
													className="text-neutral-500 transition-colors hover:text-black disabled:cursor-wait disabled:opacity-60 dark:text-neutral-400 dark:hover:text-white"
												>
													{isRevealing ? (
														<Loader2 className="h-4 w-4 animate-spin" />
													) : isRevealed ? (
														<EyeOff className="h-4 w-4" />
													) : (
														<Eye className="h-4 w-4" />
													)}
												</button>
												<button
													onClick={() => setPendingDeleteId(credential.provider_id)}
													aria-label={`Delete ${credential.provider_id} credential`}
													className="text-red-500 transition-colors hover:text-red-700"
												>
													<Trash2 className="h-4 w-4" />
												</button>
											</div>
										</div>
										<dl className="mt-3 space-y-2">
											<div className="grid grid-cols-[108px_1fr] gap-2">
												<dt className="font-mono text-[11px] uppercase tracking-wider text-neutral-400 dark:text-neutral-500">
													Client ID
												</dt>
												<dd className="font-mono text-sm break-all text-neutral-700 dark:text-neutral-200">
													{maskValue(revealed?.client_id ?? credential.client_id, credential.provider_id)}
												</dd>
											</div>
											<div className="grid grid-cols-[108px_1fr] gap-2">
												<dt className="font-mono text-[11px] uppercase tracking-wider text-neutral-400 dark:text-neutral-500">
													Client Secret
												</dt>
												<dd className="font-mono text-sm break-all text-neutral-700 dark:text-neutral-200">
													{isRevealed ? revealed?.client_secret ?? "Unavailable" : maskValue("", credential.provider_id)}
												</dd>
											</div>
											<div className="grid grid-cols-[108px_1fr] gap-2">
												<dt className="font-mono text-[11px] uppercase tracking-wider text-neutral-400 dark:text-neutral-500">
													{credential.provider_id === "awsS3" ? "Region" : "Redirect URI"}
												</dt>
												<dd className="font-mono text-sm break-all text-neutral-700 dark:text-neutral-200">
													{revealed?.redirect_uri ?? credential.redirect_uri}
												</dd>
											</div>
										</dl>
									</li>
								)
							})}
						</ul>
					)}

					<button
						onClick={handleContinue}
						disabled={credentials.length === 0}
						className="mt-4 border border-neutral-300 px-4 py-2 font-mono text-[11px] uppercase tracking-wider text-neutral-700 transition-colors hover:bg-black hover:text-white disabled:cursor-not-allowed disabled:border-neutral-200 disabled:text-neutral-300 disabled:hover:bg-transparent dark:border-neutral-700 dark:text-neutral-200 dark:hover:bg-neutral-100 dark:hover:text-neutral-950 dark:disabled:border-neutral-800 dark:disabled:text-neutral-600"
					>
						Continue to Dashboard
					</button>
				</div>
			</section>
			<p className="mt-4 font-mono text-[11px] text-neutral-400 dark:text-neutral-500">
				Supported providers in backend right now: {CREDENTIAL_PROVIDERS.join(", ")}
			</p>

			<Dialog
				open={pendingDeleteId !== null}
				onOpenChange={(open) => {
					if (!open && !deleting) setPendingDeleteId(null)
				}}
			>
				<DialogContent className="sm:max-w-sm dark:border-neutral-800 dark:bg-neutral-950">
					<DialogHeader>
						<DialogTitle className="font-mono text-sm uppercase tracking-wider text-neutral-950 dark:text-neutral-100">
							Delete Credential
						</DialogTitle>
						<DialogDescription className="font-mono text-[12px] text-neutral-600 dark:text-neutral-400">
							{pendingDeleteId
								? `Remove stored credential for ${pendingDeleteId}? This also disconnects the active provider session.`
								: "Remove this stored credential? This action cannot be undone."}
						</DialogDescription>
					</DialogHeader>
					<DialogFooter>
						<button
							type="button"
							onClick={() => setPendingDeleteId(null)}
							disabled={deleting}
							className="border border-neutral-300 px-3 py-2 font-mono text-[11px] uppercase tracking-wider text-neutral-600 transition-colors hover:bg-neutral-100 disabled:cursor-not-allowed disabled:opacity-50 dark:border-neutral-700 dark:text-neutral-300 dark:hover:bg-neutral-900"
						>
							Cancel
						</button>
						<button
							type="button"
							onClick={() => pendingDeleteId && handleDelete(pendingDeleteId)}
							disabled={!pendingDeleteId || deleting}
							className="border border-red-600 bg-red-600 px-3 py-2 font-mono text-[11px] uppercase tracking-wider text-white transition-colors hover:bg-red-700 disabled:cursor-not-allowed disabled:opacity-50"
						>
							{deleting ? "Deleting..." : "Delete"}
						</button>
					</DialogFooter>
				</DialogContent>
			</Dialog>
		</main>
	)
}
