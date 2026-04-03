import Link from "next/link";
import { getFiles } from "@/lib/api/files";
import { FilesTableClient } from "./FilesTableClient";

export default async function FilesPage() {
  const files = await getFiles().catch(() => []);

  return (
    <div className="space-y-5">
      <div className="flex items-end justify-between border-b pb-4">
        <div>
          <p className="font-mono text-[11px] uppercase tracking-[0.15em] text-neutral-400 mb-0.5">
            Storage
          </p>
          <h1 className="text-2xl font-semibold tracking-tight">Files</h1>
        </div>
        <Link
          href="/providers"
          className="font-mono text-[11px] uppercase tracking-wider border px-4 py-2 hover:bg-black hover:text-white transition-colors"
        >
          Upload More
        </Link>
      </div>

      <FilesTableClient initialFiles={files} />
    </div>
  );
}
