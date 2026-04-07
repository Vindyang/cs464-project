import Link from "next/link";

export const dynamic = 'force-dynamic';

type LifecycleEvent = {
  event_type: string;
  status: string;
  started_at: string;
  ended_at: string;
  duration_ms: number;
  error_msg?: string;
};

type FileHistoryResp = {
  events: (LifecycleEvent & { file_id: string; file_name?: string })[];
};

type HistoryRow = {
  key: string;
  file_id: string;
  file_name?: string;
  event: LifecycleEvent;
};

async function getGlobalHistory(): Promise<FileHistoryResp> {
  const gatewayURL = process.env.GATEWAY_URL || "http://localhost:8084";
  try {
    const res = await fetch(`${gatewayURL}/api/v1/history`, {
      next: { revalidate: 15 },
    });
    if (!res.ok) {
      return { events: [] };
    }
    return res.json();
  } catch {
    return { events: [] };
  }
}

function toRow(event: LifecycleEvent & { file_id: string; file_name?: string }, idx: number): HistoryRow {
  return {
    key: `${event.file_id}-${event.event_type}-${event.started_at}-${idx}`,
    file_id: event.file_id,
    file_name: event.file_name,
    event,
  };
}

export default async function HistoryPage() {
  const history = await getGlobalHistory();
  const rows = history.events
    .map((event, idx) => toRow(event, idx))
    .sort((a, b) => new Date(b.event.ended_at).getTime() - new Date(a.event.ended_at).getTime());

  return (
    <div className="space-y-5">
      <div className="flex items-end justify-between border-b border-neutral-200 pb-4 dark:border-neutral-800">
        <div>
          <p className="mb-0.5 font-mono text-[11px] uppercase tracking-[0.15em] text-neutral-400 dark:text-neutral-500">
            Activity
          </p>
          <h1 className="text-2xl font-semibold tracking-tight text-neutral-950 dark:text-neutral-100">History</h1>
        </div>
        <Link
          href="/files"
          className="border border-neutral-300 px-4 py-2 font-mono text-[11px] uppercase tracking-wider text-neutral-700 transition-colors hover:bg-black hover:text-white dark:border-neutral-700 dark:text-neutral-200 dark:hover:bg-neutral-100 dark:hover:text-neutral-950"
        >
          Back to Files
        </Link>
      </div>

      {rows.length === 0 ? (
        <section className="border border-neutral-200 bg-white dark:border-neutral-800 dark:bg-neutral-950">
          <div className="px-5 py-10 text-center font-mono text-xs text-neutral-400 dark:text-neutral-500">
            No lifecycle events yet.
            <br />
            Upload and download a file to populate history.
          </div>
        </section>
      ) : (
        <section className="border border-neutral-200 bg-white dark:border-neutral-800 dark:bg-neutral-950">
          <div className="grid grid-cols-[1.6fr_0.8fr_1fr_0.8fr_1.4fr] gap-4 border-b border-neutral-200 bg-neutral-50 px-5 py-3 dark:border-neutral-800 dark:bg-neutral-900/70">
            {["File", "Event", "Status", "Duration", "Completed"].map((h) => (
              <span key={h} className="font-mono text-[11px] uppercase tracking-[0.08em] text-neutral-400 dark:text-neutral-500">
                {h}
              </span>
            ))}
          </div>

          <div className="divide-y divide-neutral-200 dark:divide-neutral-800">
            {rows.map(({ file_id, file_name, event, key }) => (
              <div
                key={key}
                className="grid grid-cols-[1.6fr_0.8fr_1fr_0.8fr_1.4fr] items-center gap-4 px-5 py-3 transition-colors hover:bg-neutral-50 dark:hover:bg-neutral-900/60"
              >
                <div className="min-w-0">
                  <p className="truncate font-mono text-sm text-neutral-800 dark:text-neutral-100">{file_name || "Unknown file"}</p>
                  <p className="truncate font-mono text-[11px] text-neutral-400 dark:text-neutral-500">{file_id}</p>
                </div>

                <span className="font-mono text-[11px] uppercase tracking-wider text-neutral-500 dark:text-neutral-400">
                  {event.event_type}
                </span>

                <span
                  className={`font-mono text-[11px] uppercase tracking-wider ${
                    event.status === "success" ? "text-neutral-600" : "text-red-600"
                  }`}
                >
                  {event.status}
                </span>

                <span className="font-mono text-[11px] text-neutral-500 dark:text-neutral-400">{event.duration_ms}ms</span>

                <span className="font-mono text-[11px] text-neutral-500 dark:text-neutral-400">
                  {new Date(event.ended_at).toLocaleString()}
                </span>
              </div>
            ))}
          </div>
        </section>
      )}
    </div>
  );
}
