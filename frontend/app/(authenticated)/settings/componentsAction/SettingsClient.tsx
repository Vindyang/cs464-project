"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import {
	Dialog,
	DialogContent,
	DialogDescription,
	DialogFooter,
	DialogHeader,
	DialogTitle,
} from "@/components/ui/dialog";
import { Switch } from "@/components/ui/switch";
import { cn } from "@/lib/utils";
import {
	AppSettings,
	REDUNDANCY_PRESETS,
	ResetScope,
	resetStoredData,
	saveSettings,
} from "@/lib/api/settings";
import { toast } from "sonner";

const RESET_ACTIONS: Record<ResetScope, { title: string; description: string; confirmLabel: string }> = {
	files: {
		title: "Delete All File Data",
		description: "Delete all file metadata and attempt to remove every provider-side shard as part of the purge.",
		confirmLabel: "Delete File Data",
	},
	credentials: {
		title: "Delete All Credentials",
		description: "Delete all stored client credentials, provider tokens, and active provider connections.",
		confirmLabel: "Delete Credentials",
	},
	all_data: {
		title: "Delete All Data",
		description: "Delete all file data, provider-side shards, stored credentials, and active provider sessions.",
		confirmLabel: "Delete All Data",
	},
}

interface SettingsClientProps {
	initialSettings: AppSettings
}

export function SettingsClient({ initialSettings }: SettingsClientProps) {
	const router = useRouter()
	const [redundancy, setRedundancy] = useState(initialSettings.redundancy)
	const [encryptDefault, setEncryptDefault] = useState(initialSettings.encrypt_default)
	const [autoDelete, setAutoDelete] = useState(initialSettings.auto_delete)
	const [saving, setSaving] = useState(false)
	const [pendingResetScope, setPendingResetScope] = useState<ResetScope | null>(null)
	const [resetting, setResetting] = useState(false)

	const handleSave = async () => {
		setSaving(true)
		try {
			await saveSettings({
				redundancy,
				encrypt_default: encryptDefault,
				auto_delete: autoDelete,
			})
			toast.success("Preferences saved")
			router.refresh()
		} catch {
			toast.error("Failed to save")
		} finally {
			setSaving(false)
		}
	}

	const handleReset = async () => {
		if (!pendingResetScope) return

		setResetting(true)
		try {
			const result = await resetStoredData(pendingResetScope, true)

			if (pendingResetScope === "files") {
				toast.success(`Deleted ${result.file_summary?.deleted_files ?? 0} file records`)
				router.refresh()
			} else if (pendingResetScope === "credentials") {
				toast.success(`Deleted ${result.credential_summary?.deleted_credentials ?? 0} credentials and disconnected providers`)
				router.push("/dashboard")
				router.refresh()
			} else {
				toast.success("Deleted file data and credentials")
				router.push("/dashboard")
				router.refresh()
			}

			setPendingResetScope(null)
		} catch {
			toast.error("Failed to delete stored data")
		} finally {
			setResetting(false)
		}
	}

	return (
		<div className="space-y-6">
			<div>
				<h2 className="font-mono text-sm font-bold uppercase tracking-widest text-neutral-950 dark:text-neutral-100">System Settings</h2>
				<p className="mt-1 font-mono text-[12px] text-neutral-500 dark:text-neutral-400">
					Vault configuration, redundancy strategy, and storage behaviors.
				</p>
			</div>

            <SettingsCard>
                <SettingsCardHeader
                    title="Reed-Solomon Configuration"
                    subtitle="Erasure coding preset for future uploads. Does not affect existing files."
                    tag="Preference only"
                />
                <div className="space-y-5 p-5">
                    <div className="space-y-2">
                        {REDUNDANCY_PRESETS.map((preset) => {
                            const active = redundancy === preset.val
                            return (
                                <button
                                    key={preset.val}
                                    onClick={() => setRedundancy(preset.val)}
                                    className={cn(
                                        "relative w-full border px-4 py-3 text-left transition-colors",
                                        active
                                            ? "border-sky-600 bg-sky-50 dark:border-sky-500 dark:bg-sky-950/30"
                                            : "border-neutral-200 bg-neutral-50 hover:border-neutral-400 dark:border-neutral-800 dark:bg-neutral-900/60 dark:hover:border-neutral-700",
                                    )}
                                >
                                    <div className="mb-0.5 flex items-baseline justify-between gap-3">
                                        <span className="font-mono text-sm font-bold text-neutral-950 dark:text-neutral-100">{preset.label}</span>
                                        <span className="font-mono text-[11px] uppercase tracking-widest text-neutral-400 dark:text-neutral-500">{preset.name}</span>
                                    </div>
                                    <div className="font-mono text-[11px] text-neutral-500 dark:text-neutral-400">{preset.desc}</div>
                                    <div className="mt-0.5 font-mono text-[11px] text-neutral-400 dark:text-neutral-500">{preset.overhead}</div>
                                </button>
                            )
                        })}
                    </div>
                    <div className="flex justify-end pt-2">
                        <button
                            onClick={handleSave}
                            disabled={saving}
                            className="border border-sky-600 bg-sky-600 px-4 py-2 font-mono text-[11px] uppercase tracking-wider text-white transition-colors hover:bg-sky-700 disabled:cursor-not-allowed disabled:opacity-50"
                        >
                            {saving ? "Saving..." : "Save Configuration"}
                        </button>
                    </div>
                </div>
            </SettingsCard>

            <SettingsCard>
                <SettingsCardHeader
                    title="Storage Behaviors"
                    subtitle="Default behaviors for uploads and retention. Does not affect existing files."
                    tag="Preference only"
                />
                <div className="divide-y divide-neutral-200 p-5 dark:divide-neutral-800">
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
                <div className="flex justify-end px-5 pb-5">
                    <button
                        onClick={handleSave}
                        disabled={saving}
                        className="border border-sky-600 bg-sky-600 px-4 py-2 font-mono text-[11px] uppercase tracking-wider text-white transition-colors hover:bg-sky-700 disabled:cursor-not-allowed disabled:opacity-50"
                    >
                        {saving ? "Saving..." : "Save Preferences"}
                    </button>
                </div>
            </SettingsCard>

			<SettingsCard>
				<SettingsCardHeader
					title="Danger Zone"
					subtitle="Destructive operations for data removal. File purges also attempt provider-side shard deletion."
					tag="Requires confirmation"
				/>
				<div className="space-y-3 p-5">
					{(Object.keys(RESET_ACTIONS) as ResetScope[]).map((scope) => (
						<div key={scope} className="flex flex-col gap-3 border border-red-200 bg-red-50/60 p-4 dark:border-red-950 dark:bg-red-950/20 lg:flex-row lg:items-center lg:justify-between">
							<div>
								<p className="font-mono text-sm font-medium text-red-900 dark:text-red-200">{RESET_ACTIONS[scope].title}</p>
								<p className="mt-1 font-mono text-[11px] text-red-700 dark:text-red-300">{RESET_ACTIONS[scope].description}</p>
							</div>
							<button
								type="button"
								onClick={() => setPendingResetScope(scope)}
								className="border border-red-600 bg-red-600 px-4 py-2 font-mono text-[11px] uppercase tracking-wider text-white transition-colors hover:bg-red-700"
							>
								{RESET_ACTIONS[scope].confirmLabel}
							</button>
						</div>
					))}
				</div>
			</SettingsCard>

			<Dialog
				open={pendingResetScope !== null}
				onOpenChange={(open) => {
					if (!open && !resetting) setPendingResetScope(null)
				}}
			>
				<DialogContent className="sm:max-w-md dark:border-neutral-800 dark:bg-neutral-950">
					<DialogHeader>
						<DialogTitle className="font-mono text-sm uppercase tracking-wider text-neutral-950 dark:text-neutral-100">
							{pendingResetScope ? RESET_ACTIONS[pendingResetScope].title : "Confirm Reset"}
						</DialogTitle>
						<DialogDescription className="font-mono text-[12px] text-neutral-600 dark:text-neutral-400">
							{pendingResetScope ? RESET_ACTIONS[pendingResetScope].description : "This action cannot be undone."}
						</DialogDescription>
					</DialogHeader>
					<DialogFooter>
						<button
							type="button"
							onClick={() => setPendingResetScope(null)}
							disabled={resetting}
							className="border border-neutral-300 px-3 py-2 font-mono text-[11px] uppercase tracking-wider text-neutral-600 transition-colors hover:bg-neutral-100 disabled:cursor-not-allowed disabled:opacity-50 dark:border-neutral-700 dark:text-neutral-300 dark:hover:bg-neutral-900"
						>
							Cancel
						</button>
						<button
							type="button"
							onClick={handleReset}
							disabled={!pendingResetScope || resetting}
							className="border border-red-600 bg-red-600 px-3 py-2 font-mono text-[11px] uppercase tracking-wider text-white transition-colors hover:bg-red-700 disabled:cursor-not-allowed disabled:opacity-50"
						>
							{resetting ? "Deleting..." : pendingResetScope ? RESET_ACTIONS[pendingResetScope].confirmLabel : "Confirm"}
						</button>
					</DialogFooter>
				</DialogContent>
			</Dialog>
		</div>
	)
}

function SettingsCard({ children }: { children: React.ReactNode }) {
	return (
		<div className="relative border border-neutral-200 dark:border-neutral-800 dark:bg-neutral-950">
			<span className="pointer-events-none absolute -left-px -top-px h-1.5 w-1.5 border-l border-t border-neutral-400 opacity-50 dark:border-neutral-700" />
			<span className="pointer-events-none absolute -right-px -top-px h-1.5 w-1.5 border-r border-t border-neutral-400 opacity-50 dark:border-neutral-700" />
			<span className="pointer-events-none absolute -bottom-px -left-px h-1.5 w-1.5 border-b border-l border-neutral-400 opacity-50 dark:border-neutral-700" />
			<span className="pointer-events-none absolute -bottom-px -right-px h-1.5 w-1.5 border-b border-r border-neutral-400 opacity-50 dark:border-neutral-700" />
			{children}
		</div>
	)
}

function SettingsCardHeader({
	title,
	subtitle,
	tag,
}: {
	title: string
	subtitle: string
	tag?: string
}) {
	return (
		<div className="flex items-start justify-between gap-4 border-b border-neutral-200 bg-neutral-50 px-5 py-4 dark:border-neutral-800 dark:bg-neutral-900/70">
			<div>
				<h3 className="font-mono text-sm font-bold uppercase tracking-wider text-neutral-950 dark:text-neutral-100">{title}</h3>
				<p className="mt-0.5 font-mono text-[11px] text-neutral-500 dark:text-neutral-400">{subtitle}</p>
			</div>
			{tag && (
				<span className="shrink-0 border border-neutral-200 px-2 py-0.5 font-mono text-[11px] uppercase tracking-widest text-neutral-400 dark:border-neutral-700 dark:text-neutral-500">
					{tag}
				</span>
			)}
		</div>
	)
}

function ToggleRow({
	label,
	description,
	checked,
	onCheckedChange,
}: {
	label: string
	description: string
	checked: boolean
	onCheckedChange: (value: boolean) => void
}) {
	return (
		<div className="flex items-center justify-between gap-4 py-4">
			<div>
				<div className="font-mono text-sm font-medium text-neutral-950 dark:text-neutral-100">{label}</div>
				<div className="mt-0.5 font-mono text-[11px] text-neutral-500 dark:text-neutral-400">{description}</div>
			</div>
			<Switch checked={checked} onCheckedChange={onCheckedChange} />
		</div>
	)
}
