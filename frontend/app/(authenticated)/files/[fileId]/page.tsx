import Link from "next/link";
import { notFound } from "next/navigation";
import { getFileById, getFileShards } from "@/lib/api/files";
import { formatBytes } from "@/lib/utils";
import { FileHealthRefreshButton } from "../componentsAction/FileHealthRefreshButton";
import { DownloadFileButton } from "../componentsAction/DownloadFileButton";

export default async function FileDetailPage({
  params,
}: {
  params: Promise<{ fileId: string }>;
}) {
  const { fileId } = await params;
  const [file, shardMap] = await Promise.all([
    getFileById(fileId),
    getFileShards(fileId),
  ]);
  if (!file || !shardMap) notFound();

  return (
    <div className="space-y-5">
      <div className="flex items-end justify-between border-b pb-4">
        <div>
          <p className="font-mono text-[11px] uppercase tracking-[0.15em] text-neutral-400 mb-0.5">
            File Details
          </p>
          <h1 className="text-2xl font-semibold tracking-tight">{file.original_name}</h1>
        </div>
        <div className="flex gap-2">
          <FileHealthRefreshButton fileId={file.file_id} fileName={file.original_name} />
          <Link
            href="/files"
            className="font-mono text-[11px] uppercase tracking-wider border px-3 py-2 hover:bg-black hover:text-white transition-colors"
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

      <section className="grid gap-4 lg:grid-cols-3">
        <div className="border p-4 lg:col-span-2">
          <p className="mb-3 font-mono text-[11px] uppercase tracking-widest text-neutral-400">
            Shards
          </p>
          <div className="grid grid-cols-[0.8fr_0.8fr_0.9fr_1.2fr_1fr_0.9fr] gap-3 border-b pb-2">
            {["Chunk", "Index", "Type", "Provider", "Remote ID", "Status"].map((h) => (
              <span key={h} className="font-mono text-[11px] uppercase tracking-[0.08em] text-neutral-400">
                {h}
              </span>
            ))}
          </div>
          <div className="divide-y">
            {shardMap.shards.map((s) => (
              <div
                key={s.shard_id}
                className="grid grid-cols-[0.8fr_0.8fr_0.9fr_1.2fr_1fr_0.9fr] gap-3 py-2.5"
              >
                <span className="font-mono text-sm text-neutral-700">{s.chunk_index}</span>
                <span className="font-mono text-sm text-neutral-700">{s.shard_index}</span>
                <span className="font-mono text-[11px] uppercase tracking-wider text-neutral-500">
                  {s.type}
                </span>
                <span className="truncate font-mono text-sm text-neutral-700">{s.provider}</span>
                <span className="truncate font-mono text-sm text-neutral-500">{s.remote_id}</span>
                <span className="font-mono text-[11px] uppercase tracking-wider text-neutral-500">
                  {s.status}
                </span>
              </div>
            ))}
          </div>
        </div>

        <div className="border p-4">
          <p className="mb-3 font-mono text-[11px] uppercase tracking-widest text-neutral-400">
            Metadata
          </p>
          <dl className="space-y-2">
            <MetaRow label="File ID" value={file.file_id} />
            <MetaRow label="Size" value={formatBytes(file.original_size ?? 0)} />
            <MetaRow label="Status" value={file.status} />
            <MetaRow label="Chunks" value={String(file.total_chunks)} />
            <MetaRow label="Scheme" value={`(${file.n},${file.k})`} />
            <MetaRow label="Shard Size" value={formatBytes(file.shard_size ?? 0)} />
          </dl>
        </div>
      </section>
    </div>
  );
}

function MetaRow({ label, value }: { label: string; value: string }) {
  return (
    <div className="grid grid-cols-[90px_1fr] gap-2">
      <dt className="font-mono text-[11px] uppercase tracking-wider text-neutral-400">{label}</dt>
      <dd className="truncate font-mono text-sm text-neutral-700">{value}</dd>
    </div>
  );
}
