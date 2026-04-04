import Link from "next/link";

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
  const res = await fetch(`${gatewayURL}/api/v1/history`, {
    cache: "no-store",
  });
  if (!res.ok) {
    return { events: [] };
  }
  return res.json();
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
      <div className="flex items-end justify-between border-b pb-4">
        <div>
          <p className="mb-0.5 font-mono text-[11px] uppercase tracking-[0.15em] text-neutral-400">
            Activity
          </p>
          <h1 className="text-2xl font-semibold tracking-tight">History</h1>
        </div>
        <Link
          href="/files"
          className="font-mono text-[11px] uppercase tracking-wider border px-4 py-2 transition-colors hover:bg-black hover:text-white"
        >
          Back to Files
        </Link>
      </div>

      {rows.length === 0 ? (
        <section className="border bg-white">
          <div className="px-5 py-10 text-center font-mono text-xs text-neutral-400">
            No lifecycle events yet.
            <br />
            Upload and download a file to populate history.
          </div>
        </section>
      ) : (
        <section className="border bg-white">
          <div className="grid grid-cols-[1.6fr_0.8fr_1fr_0.8fr_1.4fr] gap-4 border-b bg-neutral-50 px-5 py-3">
            {["File", "Event", "Status", "Duration", "Completed"].map((h) => (
              <span key={h} className="font-mono text-[11px] uppercase tracking-[0.08em] text-neutral-400">
                {h}
              </span>
            ))}
          </div>

          <div className="divide-y">
            {rows.map(({ file_id, file_name, event, key }) => (
              <div
                key={key}
                className="grid grid-cols-[1.6fr_0.8fr_1fr_0.8fr_1.4fr] items-center gap-4 px-5 py-3"
              >
                <div className="min-w-0">
                  <p className="truncate font-mono text-sm text-neutral-800">{file_name || "Unknown file"}</p>
                  <p className="truncate font-mono text-[11px] text-neutral-400">{file_id}</p>
                </div>

                <span className="font-mono text-[11px] uppercase tracking-wider text-neutral-500">
                  {event.event_type}
                </span>

                <span
                  className={`font-mono text-[11px] uppercase tracking-wider ${
                    event.status === "success" ? "text-neutral-600" : "text-red-600"
                  }`}
                >
                  {event.status}
                </span>

                <span className="font-mono text-[11px] text-neutral-500">{event.duration_ms}ms</span>

                <span className="font-mono text-[11px] text-neutral-500">
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
