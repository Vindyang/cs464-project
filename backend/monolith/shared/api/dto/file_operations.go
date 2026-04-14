package dto

type UploadFileRequest struct {
	K              int      `form:"k" binding:"required,gt=0"`
	N              int      `form:"n" binding:"required,gt=0"`
	ChunkSizeMB    int      `form:"chunk_size_mb" binding:"required,gt=0"`
	Providers      []string `form:"providers" binding:"required,min=1"`
	DistributeMode string   `form:"distribute_mode" binding:"required,oneof=round-robin random"`
}

type UploadFileResponse struct {
	FileID       string            `json:"file_id"`
	OriginalName string            `json:"original_name"`
	OriginalSize int64             `json:"original_size"`
	TotalChunks  int               `json:"total_chunks"`
	TotalShards  int               `json:"total_shards"`
	N            int               `json:"n"`
	K            int               `json:"k"`
	ChunkSize    int64             `json:"chunk_size"`
	ShardSize    int64             `json:"shard_size"`
	Status       string            `json:"status"`
	Message      string            `json:"message"`
	UploadStats  *UploadStatistics `json:"upload_stats,omitempty"`
}

type UploadStatistics struct {
	TotalShards      int            `json:"total_shards"`
	SuccessfulShards int            `json:"successful_shards"`
	FailedShards     int            `json:"failed_shards"`
	ProviderStats    map[string]int `json:"provider_stats"`
	DurationMS       int64          `json:"duration_ms"`
}

type DownloadFileRequest struct {
	FileID string `uri:"fileId" binding:"required,uuid"`
}

type DownloadFileResponse struct{}

type ChunkUploadInfo struct {
	ChunkIndex int               `json:"chunk_index"`
	ChunkSize  int64             `json:"chunk_size"`
	Shards     []ShardUploadInfo `json:"shards"`
}

type ShardUploadInfo struct {
	ShardIndex     int    `json:"shard_index"`
	ShardType      string `json:"shard_type"`
	Provider       string `json:"provider"`
	RemoteID       string `json:"remote_id"`
	ChecksumSHA256 string `json:"checksum_sha256"`
	SizeBytes      int64  `json:"size_bytes"`
	Status         string `json:"status"`
	Error          string `json:"error,omitempty"`
}

type FileMetadataResponse struct {
	FileID              string            `json:"file_id"`
	OriginalName        string            `json:"original_name"`
	OriginalSize        int64             `json:"original_size"`
	TotalChunks         int               `json:"total_chunks"`
	TotalShards         int               `json:"total_shards"`
	N                   int               `json:"n"`
	K                   int               `json:"k"`
	ChunkSize           int64             `json:"chunk_size"`
	ShardSize           int64             `json:"shard_size"`
	Status              string            `json:"status"`
	CreatedAt           string            `json:"created_at"`
	UpdatedAt           string            `json:"updated_at"`
	LastHealthRefreshAt *string           `json:"last_health_refresh_at,omitempty"`
	FirstCreatedAt      *string           `json:"first_created_at,omitempty"`
	LastDownloadedAt    *string           `json:"last_downloaded_at,omitempty"`
	HealthStatus        *FileHealthStatus `json:"health_status,omitempty"`
}

type FileHealthStatus struct {
	HealthyShards   int     `json:"healthy_shards"`
	CorruptedShards int     `json:"corrupted_shards"`
	MissingShards   int     `json:"missing_shards"`
	TotalShards     int     `json:"total_shards"`
	HealthPercent   float64 `json:"health_percent"`
	Recoverable     bool    `json:"recoverable"`
}

type ListFilesRequest struct {
	Page     int    `form:"page" binding:"omitempty,gt=0"`
	PageSize int    `form:"page_size" binding:"omitempty,gt=0,lte=100"`
	Status   string `form:"status" binding:"omitempty,oneof=PENDING UPLOADED HEALTHY DEGRADED CORRUPTED"`
}

type ListFilesResponse struct {
	Files      []FileMetadataResponse `json:"files"`
	Page       int                    `json:"page"`
	PageSize   int                    `json:"page_size"`
	TotalFiles int                    `json:"total_files"`
	TotalPages int                    `json:"total_pages"`
}

type DeleteFileRequest struct {
	FileID       string `uri:"fileId" binding:"required,uuid"`
	DeleteShards bool   `form:"delete_shards"`
}

type DeleteFileResponse struct {
	FileID        string `json:"file_id"`
	Success       bool   `json:"success"`
	Message       string `json:"message"`
	ShardsDeleted int    `json:"shards_deleted,omitempty"`
}
