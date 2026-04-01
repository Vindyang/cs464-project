import { getApiBaseUrl } from "./base-url";

export interface FileHealthStatus {
  healthy_shards: number;
  corrupted_shards: number;
  missing_shards: number;
  total_shards: number;
  health_percent: number;
  recoverable: boolean;
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
  shards: ShardInfo[];
}

export interface UploadFileResult {
  status: string;
  file_id?: string;
  error?: string;
  details?: string;
}

export async function getFiles(): Promise<FileMetadata[]> {
  const API_URL = getApiBaseUrl();
  const res = await fetch(`${API_URL}/api/v1/files`, {
    cache: "no-store",
  });
  if (!res.ok) throw new Error("Failed to fetch files");
  return res.json();
}

export async function getFileById(fileId: string): Promise<ShardMap | null> {
  const API_URL = getApiBaseUrl();
  const res = await fetch(`${API_URL}/api/v1/files/${fileId}`, {
    cache: "no-store",
  });
  if (res.status === 404) return null;
  if (!res.ok) throw new Error("Failed to fetch file");
  return res.json();
}

export async function getFileShards(fileId: string): Promise<ShardMap | null> {
  const API_URL = getApiBaseUrl();
  const res = await fetch(`${API_URL}/api/v1/shards/file/${fileId}`, {
    cache: "no-store",
  });
  if (res.status === 404) return null;
  if (!res.ok) throw new Error("Failed to fetch shards");
  return res.json();
}

export async function deleteFile(fileId: string, deleteShards = false): Promise<void> {
  const API_URL = getApiBaseUrl();
  const url = `${API_URL}/api/v1/files/${fileId}${deleteShards ? "?delete_shards=true" : ""}`;
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
