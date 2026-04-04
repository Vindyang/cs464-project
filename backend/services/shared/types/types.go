package types

import "time"

// RegisterFileReq from Shard Map Service
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

// RecordShardReq to Shard Map Service
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
	Type        string `json:"type"` // "data" or "parity"
	RemoteID    string `json:"remote_id"`
	Provider    string `json:"provider"`
	ChecksumSha string `json:"checksum_sha256"`
}

// GetShardMapResp from Shard Map Service
type GetShardMapResp struct {
	FileID       string          `json:"file_id"`
	OriginalName string          `json:"original_name"`
	N            int             `json:"n"`
	K            int             `json:"k"`
	Status       string          `json:"status"`
	Shards       []ShardMapEntry `json:"shards"`
}

type ShardMapEntry struct {
	ShardID    string `json:"shard_id"`
	ShardIndex int    `json:"shard_index"`
	RemoteID   string `json:"remote_id"`
	Provider   string `json:"provider"`
	Status     string `json:"status"`
	Checksum   string `json:"checksum_sha256"`
}

// ProviderInfo from Adapter Service
type ProviderInfo struct {
	ProviderID      string `json:"providerId"`
	DisplayName     string `json:"displayName"`
	Status          string `json:"status"`
	LatencyMs       int64  `json:"latencyMs"`
	QuotaTotalBytes int64  `json:"quotaTotalBytes"`
	QuotaUsedBytes  int64  `json:"quotaUsedBytes"`
}

// UploadShardResp from Adapter Service
type UploadShardResp struct {
	RemoteID    string `json:"remote_id"`
	ChecksumSha string `json:"checksum_sha256"`
}

// UploadReq to Orchestrator
type UploadReq struct {
	FileName    string   `json:"file_name"`
	FileSize    int64    `json:"file_size"`
	ShardCount  int      `json:"shard_count"`
	ShardData   [][]byte `json:"shard_data"`
	IsDataShard []bool   `json:"is_data_shard"`
}

// UploadResp from Orchestrator
type UploadResp struct {
	Status string      `json:"status"`
	FileID string      `json:"file_id"`
	Error  string      `json:"error,omitempty"`
	Shards []ShardInfo `json:"shards,omitempty"`
}

// DownloadResp from Orchestrator
type DownloadResp struct {
	FileID   string   `json:"file_id"`
	FileName string   `json:"file_name"`
	ShardIDs []string `json:"shard_ids"`
	Shards   [][]byte `json:"shards"`
}

// LifecycleEvent is sent from the orchestrator to the shardmap service
// to record file operation history (upload or download).
type LifecycleEvent struct {
	FileID     string    `json:"file_id"`
	EventType  string    `json:"event_type"`  // "upload" or "download"
	FileName   string    `json:"file_name,omitempty"`
	FileSize   int64     `json:"file_size,omitempty"`
	ShardCount int       `json:"shard_count,omitempty"`
	Providers  []string  `json:"providers,omitempty"`
	StartedAt  time.Time `json:"started_at"`
	EndedAt    time.Time `json:"ended_at"`
	DurationMs int64     `json:"duration_ms"`
	Status     string    `json:"status"`             // "success" or "failed"
	ErrorMsg   string    `json:"error_msg,omitempty"`
}

// FileHistoryResp is returned by the shardmap lifecycle history endpoint.
type FileHistoryResp struct {
	FileID string           `json:"file_id"`
	Events []LifecycleEvent `json:"events"`
}
