import Link from "next/link";
import { notFound } from "next/navigation";
import { getFileById, getFileHistory, getFileShards } from "@/lib/api/files";
import { formatBytes, formatDateTime, formatRelativeTime, roundHealthPercent } from "@/lib/utils";
import { FileHealthRefreshButton } from "../componentsAction/FileHealthRefreshButton";
import { DownloadFileButton } from "../componentsAction/DownloadFileButton";

export default async function FileDetailPage({
	params,
}: {
	params: Promise<{ fileId: string }>;
}) {
	const { fileId } = await params;
	const [file, shardMap, history] = await Promise.all([
		getFileById(fileId),
		getFileShards(fileId),
		getFileHistory(fileId),
	]);

	if (!file || !shardMap) notFound();

	const recentEvents = [...(history?.events ?? [])]
		.sort((left, right) => new Date(right.ended_at).getTime() - new Date(left.ended_at).getTime())
		.slice(0, 5);
	const missingShards = file.health_status?.missing_shards ?? 0;
	const corruptedShards = file.health_status?.corrupted_shards ?? 0;
	const healthPercent = roundHealthPercent(file.health_status?.health_percent, 100);

	return (
		<div className="space-y-5">
			<div className="flex flex-wrap items-end justify-between gap-3 border-b border-neutral-200 pb-4 dark:border-neutral-800">
				<div>
					<p className="mb-0.5 font-mono text-[11px] uppercase tracking-[0.15em] text-neutral-400 dark:text-neutral-500">
						File Details
					</p>
					<h1 className="text-2xl font-semibold tracking-tight text-neutral-950 dark:text-neutral-100">{file.original_name}</h1>
					<div className="mt-2 flex flex-wrap items-center gap-2 text-[11px]">
						<span className="border border-sky-200 bg-sky-50 px-2 py-1 font-mono uppercase tracking-[0.08em] text-sky-700 dark:border-sky-900 dark:bg-sky-950/60 dark:text-sky-300">
							Health Synced {formatRelativeTime(file.last_health_refresh_at)}
						</span>
						<span className="font-mono text-neutral-500 dark:text-neutral-400">
							{formatDateTime(file.last_health_refresh_at)}
						</span>
					</div>
				</div>
				<div className="flex gap-2">
					<FileHealthRefreshButton fileId={file.file_id} fileName={file.original_name} />
					<Link
						href="/files"
						className="border border-neutral-300 px-3 py-2 font-mono text-[11px] uppercase tracking-wider text-neutral-700 transition-colors hover:bg-black hover:text-white dark:border-neutral-700 dark:text-neutral-200 dark:hover:bg-neutral-100 dark:hover:text-neutral-950"
					>
						Back
					</Link>
					<DownloadFileButton
						fileId={file.file_id}
						fileName={file.original_name}
						requiredShards={file.k}
						healthStatus={file.health_status}
						variant="primary"
					/>
				</div>
			</div>

			{(missingShards > 0 || corruptedShards > 0) && (
				<section className="border border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-900 dark:border-amber-900 dark:bg-amber-950/40 dark:text-amber-200">
					<p className="font-mono text-[11px] uppercase tracking-[0.1em] text-amber-700 dark:text-amber-300">
						File Health Warning
					</p>
					<p className="mt-1">
						{missingShards > 0 ? `${missingShards} shard${missingShards === 1 ? " is" : "s are"} missing.` : ""}
						{missingShards > 0 && corruptedShards > 0 ? " " : ""}
						{corruptedShards > 0 ? `${corruptedShards} shard${corruptedShards === 1 ? " is" : "s are"} corrupted.` : ""}
						{" "}Recovery health is currently {healthPercent}%.
					</p>
				</section>
			)}

			<section className="grid gap-4 xl:grid-cols-[minmax(300px,0.95fr)_minmax(0,1.25fr)]">
				<div className="border border-neutral-200 p-4 dark:border-neutral-800 dark:bg-neutral-950">
					<p className="mb-3 font-mono text-[11px] uppercase tracking-widest text-neutral-400 dark:text-neutral-500">
						Metadata
					</p>
					<dl className="space-y-2">
						<MetaRow label="File ID" value={file.file_id} />
						<MetaRow label="Size" value={formatBytes(file.original_size ?? 0)} />
						<MetaRow label="Status" value={file.status} />
						<MetaRow label="Chunks" value={String(file.total_chunks)} />
						<MetaRow label="Scheme" value={`(${file.n},${file.k})`} />
						<MetaRow label="Shard Size" value={formatBytes(file.shard_size ?? 0)} />
						<MetaRow label="Last Sync" value={formatDateTime(file.last_health_refresh_at)} />
					</dl>
				</div>

				<div className="border border-neutral-200 p-4 dark:border-neutral-800 dark:bg-neutral-950">
					<div className="mb-3 flex items-center justify-between">
						<p className="font-mono text-[11px] uppercase tracking-widest text-neutral-400 dark:text-neutral-500">
							Recent Activity
						</p>
						<span className="font-mono text-[11px] text-neutral-400 dark:text-neutral-500">
							Top {recentEvents.length}
						</span>
					</div>
					{recentEvents.length === 0 ? (
						<p className="font-mono text-sm text-neutral-400 dark:text-neutral-500">No activity has been recorded for this file yet.</p>
					) : (
						<div className="max-h-[360px] space-y-3 overflow-y-auto pr-1">
							{recentEvents.map((event) => (
								<div key={`${event.event_type}-${event.started_at}`} className="border border-neutral-200 px-3 py-2 dark:border-neutral-800">
									<div className="flex items-center justify-between gap-3">
										<span className="font-mono text-[11px] uppercase tracking-[0.08em] text-neutral-500 dark:text-neutral-400">
											{event.event_type}
										</span>
										<span className={`font-mono text-[11px] uppercase tracking-[0.08em] ${event.status === "success" ? "text-emerald-600" : "text-red-600"}`}>
											{event.status}
										</span>
									</div>
									<p className="mt-2 font-mono text-[12px] text-neutral-600 dark:text-neutral-300">
										{formatDateTime(event.ended_at)} · {event.duration_ms}ms
									</p>
									{event.error_msg && (
										<p className="mt-2 text-sm text-red-600">{event.error_msg}</p>
									)}
								</div>
							))}
						</div>
					)}
				</div>
			</section>

			<section className="border border-neutral-200 p-4 dark:border-neutral-800 dark:bg-neutral-950">
				<p className="mb-3 font-mono text-[11px] uppercase tracking-widest text-neutral-400 dark:text-neutral-500">
					Shards
				</p>
				<div className="grid grid-cols-[0.7fr_0.8fr_0.9fr_1fr_1.6fr_0.9fr] gap-3 border-b border-neutral-200 pb-2 dark:border-neutral-800">
					{["Chunk", "Index", "Type", "Provider", "Remote ID", "Status"].map((header) => (
						<span key={header} className="font-mono text-[11px] uppercase tracking-[0.08em] text-neutral-400 dark:text-neutral-500">
							{header}
						</span>
					))}
				</div>
				<div className="divide-y divide-neutral-200 dark:divide-neutral-800">
					{shardMap.shards.map((shard) => (
						<div
							key={shard.shard_id}
							className="grid grid-cols-[0.7fr_0.8fr_0.9fr_1fr_1.6fr_0.9fr] gap-3 py-2.5"
						>
							<span className="font-mono text-sm text-neutral-700 dark:text-neutral-200">{shard.chunk_index}</span>
							<span className="font-mono text-sm text-neutral-700 dark:text-neutral-200">{shard.shard_index}</span>
							<span className="font-mono text-[11px] uppercase tracking-wider text-neutral-500 dark:text-neutral-400">
								{shard.type}
							</span>
							<span className="truncate font-mono text-sm text-neutral-700 dark:text-neutral-200">{shard.provider}</span>
							<span className="break-all whitespace-normal font-mono text-sm text-neutral-500 dark:text-neutral-400" title={shard.remote_id}>
								{shard.remote_id}
							</span>
							<span className={`font-mono text-[11px] uppercase tracking-wider ${shard.status === "HEALTHY" ? "text-emerald-600" : shard.status === "MISSING" ? "text-amber-700" : "text-red-600"}`}>
								{shard.status}
							</span>
						</div>
					))}
				</div>
			</section>
		</div>
	);
}

function MetaRow({ label, value }: { label: string; value: string }) {
	return (
		<div className="grid grid-cols-[90px_1fr] gap-2">
			<dt className="font-mono text-[11px] uppercase tracking-wider text-neutral-400 dark:text-neutral-500">{label}</dt>
			<dd className="break-all font-mono text-sm text-neutral-700 dark:text-neutral-200">{value}</dd>
		</div>
	);
}
