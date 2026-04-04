import { getFiles } from "@/lib/api/files";
import { FilesPageClient } from "./componentsAction/FilesPageClient";

export default async function FilesPage() {
  const files = await getFiles().catch(() => []);
  return <FilesPageClient initialFiles={files} />;
}
