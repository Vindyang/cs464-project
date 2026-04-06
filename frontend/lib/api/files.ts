import { getApiBaseUrl } from "./base-url";

export interface FileHealthStatus {
  healthy_shards: number;
  corrupted_shards: number;
  missing_shards: number;
  total_shards: number;
  health_percent: number;
  recoverable: boolean;
}

export interface LifecycleEvent {
  file_id: string;
  event_type: string;
  file_name?: string;
  file_size?: number;
  shard_count?: number;
  providers?: string[];
  started_at: string;
  ended_at: string;
  duration_ms: number;
  status: string;
  error_msg?: string;
}

export interface FileHistoryResponse {
  file_id: string;
  events: LifecycleEvent[];
}

export interface FileMetadata {
  file_id: string;
  original_name: string;
  original_size: number;
  total_chunks: number;
  total_shards: number;
  n: number;
  k: number;
  chunk_size: number;
  shard_size: number;
  status: string;
  created_at: string;
  updated_at: string;
  last_health_refresh_at?: string | null;
  first_created_at?: string | null;
  last_downloaded_at?: string | null;
  health_status?: FileHealthStatus;
}

export interface ShardInfo {
  shard_id: string;
  chunk_index: number;
  shard_index: number;
  type: string;
  remote_id: string;
  provider: string;
  checksum_sha256: string;
  status: string;
}

export interface ShardMap {
  file_id: string;
  original_name: string;
  original_size: number;
  total_chunks: number;
  n: number;
  k: number;
  shard_size: number;
  status: string;
  first_created_at?: string | null;
  last_downloaded_at?: string | null;
  shards: ShardInfo[];
}

export interface UploadFileResult {
  status: string;
  file_id?: string;
  error?: string;
  details?: string;
}

function getServerFetchOptions(revalidate: number) {
  return typeof window === "undefined"
    ? { next: { revalidate } }
    : { cache: "no-store" as const }
}

export async function getFiles(): Promise<FileMetadata[]> {
  const API_URL = getApiBaseUrl();
  const res = await fetch(`${API_URL}/api/v1/files`, getServerFetchOptions(15));
  if (!res.ok) throw new Error("Failed to fetch files");
  return res.json();
}

export async function getFileById(fileId: string): Promise<FileMetadata | null> {
  const API_URL = getApiBaseUrl();
  const res = await fetch(`${API_URL}/api/v1/files/${fileId}`, getServerFetchOptions(15));
  if (res.status === 404) return null;
  if (!res.ok) throw new Error("Failed to fetch file");
  return res.json();
}

export async function getFileShards(fileId: string): Promise<ShardMap | null> {
  const API_URL = getApiBaseUrl();
  const res = await fetch(`${API_URL}/api/v1/shards/file/${fileId}`, getServerFetchOptions(15));
  if (res.status === 404) return null;
  if (!res.ok) throw new Error("Failed to fetch shards");
  return res.json();
}

export async function getFileHistory(fileId: string): Promise<FileHistoryResponse | null> {
  const historyBaseUrl =
    typeof window === "undefined"
      ? process.env.GATEWAY_URL || "http://localhost:8084"
      : "";
  const historyUrl =
    typeof window === "undefined"
      ? `${historyBaseUrl}/api/v1/history/${fileId}`
      : `/api/history/${fileId}`;

  const res = await fetch(historyUrl, {
    ...(typeof window === "undefined" ? { next: { revalidate: 15 } } : { cache: "no-store" }),
  });
  if (res.status === 404) return null;
  if (!res.ok) throw new Error("Failed to fetch file history");
  return res.json();
}

export async function deleteFile(fileId: string, deleteShards = false): Promise<void> {
  const url = `/api/files/${fileId}${deleteShards ? "?delete_shards=true" : ""}`;
  const res = await fetch(url, { method: "DELETE" });
  if (!res.ok) throw new Error("Failed to delete file");
}

export async function uploadFile(file: File, k = 4, n = 6): Promise<UploadFileResult> {
  const formData = new FormData();
  formData.append("file", file);
  formData.append("k", String(k));
  formData.append("n", String(n));

  const res = await fetch("/api/upload", {
    method: "POST",
    body: formData,
  });

  const data = await res.json().catch(() => null);
  if (!res.ok) {
    const message = data?.details || data?.error || "Failed to upload file";
    throw new Error(message);
  }

  const result = (data ?? {}) as UploadFileResult;
  const ok = result.status === "success" || result.status === "committed";
  if (!ok) {
    throw new Error(result.error || result.details || "Upload did not complete successfully");
  }

  return result;
}
