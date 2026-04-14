package types

import "time"

type RegisterFileReq struct {
	OriginalName string `json:"original_name"`
	OriginalSize int64  `json:"original_size"`
	TotalChunks  int    `json:"total_chunks"`
	N            int    `json:"n"`
	K            int    `json:"k"`
	ShardSize    int64  `json:"shard_size"`
}

type RegisterFileResp struct {
	FileID string `json:"file_id"`
	Status string `json:"status"`
}

type RecordShardReq struct {
	FileID string      `json:"file_id"`
	Shards []ShardInfo `json:"shards"`
}

type RecordShardResp struct {
	FileID string      `json:"file_id"`
	Shards []ShardInfo `json:"shards"`
}

type ShardInfo struct {
	ShardID     string `json:"shard_id,omitempty"`
	ChunkIndex  int    `json:"chunk_index"`
	ShardIndex  int    `json:"shard_index"`
	Type        string `json:"type"`
	RemoteID    string `json:"remote_id"`
	Provider    string `json:"provider"`
	ChecksumSha string `json:"checksum_sha256"`
}

type GetShardMapResp struct {
	FileID           string          `json:"file_id"`
	OriginalName     string          `json:"original_name"`
	N                int             `json:"n"`
	K                int             `json:"k"`
	Status           string          `json:"status"`
	FirstCreatedAt   *string         `json:"first_created_at,omitempty"`
	LastDownloadedAt *string         `json:"last_downloaded_at,omitempty"`
	Shards           []ShardMapEntry `json:"shards"`
}

type FileMetadata struct {
	FileID              string  `json:"file_id"`
	OriginalName        string  `json:"original_name"`
	OriginalSize        int64   `json:"original_size"`
	TotalChunks         int     `json:"total_chunks"`
	TotalShards         int     `json:"total_shards"`
	N                   int     `json:"n"`
	K                   int     `json:"k"`
	ChunkSize           int64   `json:"chunk_size"`
	ShardSize           int64   `json:"shard_size"`
	Status              string  `json:"status"`
	CreatedAt           string  `json:"created_at"`
	UpdatedAt           string  `json:"updated_at"`
	LastHealthRefreshAt *string `json:"last_health_refresh_at,omitempty"`
	FirstCreatedAt      *string `json:"first_created_at,omitempty"`
	LastDownloadedAt    *string `json:"last_downloaded_at,omitempty"`
}

type HealthRefreshSummary struct {
	FilesScanned  int      `json:"files_scanned"`
	ShardsChecked int      `json:"shards_checked"`
	MarkedHealthy int      `json:"marked_healthy"`
	MarkedMissing int      `json:"marked_missing"`
	SkippedErrors int      `json:"skipped_errors"`
	FailedFiles   int      `json:"failed_files"`
	RefreshedAt   string   `json:"refreshed_at,omitempty"`
	ErrorMessages []string `json:"error_messages,omitempty"`
}

type ShardMapEntry struct {
	ShardID    string `json:"shard_id"`
	ShardIndex int    `json:"shard_index"`
	RemoteID   string `json:"remote_id"`
	Provider   string `json:"provider"`
	Status     string `json:"status"`
	Checksum   string `json:"checksum_sha256"`
}

type ProviderInfo struct {
	ProviderID      string `json:"providerId"`
	DisplayName     string `json:"displayName"`
	Status          string `json:"status"`
	LatencyMs       int64  `json:"latencyMs"`
	QuotaTotalBytes int64  `json:"quotaTotalBytes"`
	QuotaUsedBytes  int64  `json:"quotaUsedBytes"`
}

type UploadShardResp struct {
	RemoteID    string `json:"remote_id"`
	ChecksumSha string `json:"checksum_sha256"`
}

type UploadReq struct {
	FileName    string   `json:"file_name"`
	FileSize    int64    `json:"file_size"`
	ShardCount  int      `json:"shard_count"`
	ShardData   [][]byte `json:"shard_data"`
	IsDataShard []bool   `json:"is_data_shard"`
}

type UploadResp struct {
	Status string      `json:"status"`
	FileID string      `json:"file_id"`
	Error  string      `json:"error,omitempty"`
	Shards []ShardInfo `json:"shards,omitempty"`
}

type DownloadResp struct {
	FileID   string   `json:"file_id"`
	FileName string   `json:"file_name"`
	ShardIDs []string `json:"shard_ids"`
	Shards   [][]byte `json:"shards"`
}

type LifecycleEvent struct {
	FileID     string    `json:"file_id"`
	EventType  string    `json:"event_type"`
	FileName   string    `json:"file_name,omitempty"`
	FileSize   int64     `json:"file_size,omitempty"`
	ShardCount int       `json:"shard_count,omitempty"`
	Providers  []string  `json:"providers,omitempty"`
	StartedAt  time.Time `json:"started_at"`
	EndedAt    time.Time `json:"ended_at"`
	DurationMs int64     `json:"duration_ms"`
	Status     string    `json:"status"`
	ErrorMsg   string    `json:"error_msg,omitempty"`
}

type FileHistoryResp struct {
	FileID string           `json:"file_id"`
	Events []LifecycleEvent `json:"events"`
}

type GlobalHistoryResp struct {
	Events []LifecycleEvent `json:"events"`
}
