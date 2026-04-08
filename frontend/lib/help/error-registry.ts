export type HelpEntry = {
  code: string
  title: string
  why: string
  steps: string[]
  docsAnchor: string
}

type Registry = Record<string, HelpEntry>

export const registry: Registry = {
  PROVIDER_TIMEOUT: {
    code: "PROVIDER_TIMEOUT",
    title: "Provider connection timed out",
    why: "The storage provider (Google Drive, OneDrive, or S3) did not respond in time. This is usually a temporary network or provider-side issue.",
    steps: [
      "Check your internet connection.",
      "Wait 30 seconds and try again — provider hiccups are usually brief.",
      "Go to Providers and check the provider's status indicator.",
      "If the provider shows as disconnected, disconnect and reconnect it.",
    ],
    docsAnchor: "provider-timeout",
  },
  PROVIDER_AUTH_EXPIRED: {
    code: "PROVIDER_AUTH_EXPIRED",
    title: "Provider authentication expired",
    why: "The authorization token for your storage provider has expired or been revoked. This happens periodically with OAuth providers.",
    steps: [
      "Go to Providers.",
      "Disconnect the affected provider.",
      "Reconnect it by clicking Connect and completing the OAuth flow again.",
      "Re-try your original action.",
    ],
    docsAnchor: "provider-auth-expired",
  },
  PROVIDER_QUOTA_EXCEEDED: {
    code: "PROVIDER_QUOTA_EXCEEDED",
    title: "Storage quota exceeded",
    why: "The provider's storage is full. Omnishard cannot write new shards to it.",
    steps: [
      "Go to Providers and check the quota bar for the affected provider.",
      "Free up space in your provider account (e.g. delete old files from Google Drive).",
      "Alternatively, add a new provider with available space.",
      "Re-try the upload once space is available.",
    ],
    docsAnchor: "provider-quota-exceeded",
  },
  PROVIDER_NOT_FOUND: {
    code: "PROVIDER_NOT_FOUND",
    title: "Provider not found",
    why: "Omnishard tried to use a provider that no longer exists in your account. It may have been removed.",
    steps: [
      "Go to Providers and check your connected providers.",
      "If the provider is missing, reconnect it.",
      "If you recently removed a provider that stored shards, some files may no longer be recoverable.",
    ],
    docsAnchor: "provider-not-found",
  },
  PROVIDER_UNAVAILABLE: {
    code: "PROVIDER_UNAVAILABLE",
    title: "No connected providers",
    why: "You have no connected storage providers. Omnishard needs at least one to store shards.",
    steps: [
      "Go to Providers.",
      "Connect at least one provider (Google Drive, OneDrive, or S3).",
      "Re-try the upload.",
    ],
    docsAnchor: "provider-unavailable",
  },
  SHARD_UPLOAD_PARTIAL: {
    code: "SHARD_UPLOAD_PARTIAL",
    title: "Upload failed — not all shards were written",
    why: "Omnishard splits your file into 6 shards and needs all 6 to succeed. One or more providers failed during the upload. Any shards that were written have been rolled back automatically.",
    steps: [
      "Go to Providers and confirm all providers are connected and have available quota.",
      "If a provider is showing as degraded or disconnected, reconnect it.",
      "Re-try the upload.",
    ],
    docsAnchor: "shard-upload-partial",
  },
  SHARD_NOT_RECOVERABLE: {
    code: "SHARD_NOT_RECOVERABLE",
    title: "File cannot be recovered",
    why: "Omnishard needs at least 4 of 6 shards to reconstruct a file. Fewer than 4 healthy shards are currently reachable.",
    steps: [
      "Go to the file's detail page and check the shard health status.",
      "Click Refresh Health to get the latest shard status from providers.",
      "If some providers are offline, reconnect them and try refreshing health again.",
      "If 3 or more providers are permanently lost, the file cannot be recovered.",
    ],
    docsAnchor: "shard-not-recoverable",
  },
  SHARD_DECODE_FAILED: {
    code: "SHARD_DECODE_FAILED",
    title: "File reconstruction failed",
    why: "The Reed-Solomon decoder could not reconstruct the file from the available shards. The shard data may be corrupted.",
    steps: [
      "Refresh file health to re-probe all shards.",
      "If shards are marked as CORRUPTED, the file data at those providers may be damaged.",
      "Try the download again — occasionally transient network errors cause this.",
      "If it keeps failing, the file may be unrecoverable from the current shard state.",
    ],
    docsAnchor: "shard-decode-failed",
  },
  SHARD_ENCODE_FAILED: {
    code: "SHARD_ENCODE_FAILED",
    title: "File sharding failed",
    why: "Omnishard could not split your file into shards before uploading. This is usually an internal error.",
    steps: [
      "Try the upload again — this is usually transient.",
      "If it keeps failing, check that the file is not corrupted and try with a different file.",
    ],
    docsAnchor: "shard-encode-failed",
  },
  FILE_NOT_FOUND: {
    code: "FILE_NOT_FOUND",
    title: "File not found",
    why: "The file you are trying to access no longer exists in Omnishard's database. It may have been deleted.",
    steps: [
      "Refresh the Files page — the file list may be stale.",
      "If the file was recently deleted, it cannot be recovered.",
    ],
    docsAnchor: "file-not-found",
  },
  FILE_REGISTER_FAILED: {
    code: "FILE_REGISTER_FAILED",
    title: "Upload failed — could not register file",
    why: "Omnishard failed to create the file record before distributing shards. This is usually a temporary database issue.",
    steps: [
      "Wait a few seconds and try the upload again.",
      "If it keeps failing, check that the backend services are running (Dashboard → System Health).",
    ],
    docsAnchor: "file-register-failed",
  },
  SHARD_RECORD_FAILED: {
    code: "SHARD_RECORD_FAILED",
    title: "Upload failed — could not save shard map",
    why: "The shards were uploaded to providers but Omnishard could not save the shard location records. The shards have been rolled back automatically.",
    steps: [
      "Try the upload again.",
      "If it keeps failing, check that the backend services are running.",
    ],
    docsAnchor: "shard-record-failed",
  },
  HEALTH_REFRESH_FAILED: {
    code: "HEALTH_REFRESH_FAILED",
    title: "Health refresh failed",
    why: "Omnishard could not reach one or more providers to probe shard health. Provider connectivity issues are the most common cause.",
    steps: [
      "Go to Providers and check the status of each provider.",
      "Reconnect any disconnected providers.",
      "Try the health refresh again.",
    ],
    docsAnchor: "health-refresh-failed",
  },
  INVALID_ERASURE_PARAMS: {
    code: "INVALID_ERASURE_PARAMS",
    title: "Invalid erasure coding parameters",
    why: "The k/n values provided for Reed-Solomon coding are out of valid range (k must be > 0, n must be >= k).",
    steps: [
      "Use the default settings (k=4, n=6) unless you have a specific reason to change them.",
      "If you are using the API directly, ensure k ≤ n and both values are positive integers.",
    ],
    docsAnchor: "invalid-erasure-params",
  },
  UNKNOWN_ERROR: {
    code: "UNKNOWN_ERROR",
    title: "Something went wrong",
    why: "An unexpected error occurred. Check the browser console or service logs for details.",
    steps: [
      "Try the action again.",
      "Refresh the page and retry.",
      "If the problem persists, check that all backend services are running.",
    ],
    docsAnchor: "unknown-error",
  },
}

export function lookupHelp(code: string | undefined): HelpEntry {
  if (code && registry[code]) {
    return registry[code]
  }
  return registry["UNKNOWN_ERROR"]
}
