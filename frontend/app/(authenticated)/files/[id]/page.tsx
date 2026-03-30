import { getFileById } from "@/lib/api/files";
import { notFound } from "next/navigation";
import { FileDetailsClient } from "./FileDetailsClient";

export default async function FileDetailsPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;
  const file = await getFileById(id);

  if (!file) notFound();

  return <FileDetailsClient file={file} />;
}
