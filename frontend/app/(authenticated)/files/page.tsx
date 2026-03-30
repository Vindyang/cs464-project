import { auth } from "@/lib/auth";
import { headers } from "next/headers";
import { getFiles, FileMetadata } from "@/lib/api/files";
import { FilesClient } from "./FilesClient";

export default async function FilesPage() {
  const session = await auth.api.getSession({ headers: await headers() });
  const userId = session?.user?.id;
  const files: FileMetadata[] = userId
    ? await getFiles(userId).catch(() => [])
    : [];

  return <FilesClient initialFiles={files} />;
}
