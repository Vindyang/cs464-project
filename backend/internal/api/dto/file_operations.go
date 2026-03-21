package dto

// UploadFileRequest - POST /api/v1/files/upload (multipart/form-data)
// File is sent as multipart form data with field name "file"
type UploadFileRequest struct {
	K              int      `form:"k" binding:"required,gt=0"`               // Number of data shards
	N              int      `form:"n" binding:"required,gt=0"`               // Total shards (data + parity)
	ChunkSizeMB    int      `form:"chunk_size_mb" binding:"required,gt=0"`   // Chunk size in MB
	Providers      []string `form:"providers" binding:"required,min=1"`      // List of cloud providers to use
	DistributeMode string   `form:"distribute_mode" binding:"required,oneof=round-robin random"` // How to distribute shards
}

// UploadFileResponse - Response for POST /api/v1/files/upload
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

// UploadStatistics contains statistics about the upload process
type UploadStatistics struct {
	TotalShards      int               `json:"total_shards"`
	SuccessfulShards int               `json:"successful_shards"`
	FailedShards     int               `json:"failed_shards"`
	ProviderStats    map[string]int    `json:"provider_stats"` // provider -> shard count
	DurationMS       int64             `json:"duration_ms"`
}

// DownloadFileRequest - GET /api/v1/files/:fileId/download
// Parameters are sent as query parameters
type DownloadFileRequest struct {
	FileID string `uri:"fileId" binding:"required,uuid"`
}

// DownloadFileResponse - Response for GET /api/v1/files/:fileId/download
// The file is returned as binary data with appropriate headers
type DownloadFileResponse struct {
	// This is returned as file download, so no JSON response
	// Content-Type: application/octet-stream or the original file's MIME type
	// Content-Disposition: attachment; filename="original_filename"
}

// ChunkUploadInfo represents information about a chunk and its shards during upload
type ChunkUploadInfo struct {
	ChunkIndex int                `json:"chunk_index"`
	ChunkSize  int64              `json:"chunk_size"`
	Shards     []ShardUploadInfo  `json:"shards"`
}

// ShardUploadInfo represents information about a shard upload
type ShardUploadInfo struct {
	ShardIndex     int    `json:"shard_index"`
	ShardType      string `json:"shard_type"` // "DATA" or "PARITY"
	Provider       string `json:"provider"`
	RemoteID       string `json:"remote_id"`
	ChecksumSHA256 string `json:"checksum_sha256"`
	SizeBytes      int64  `json:"size_bytes"`
	Status         string `json:"status"` // "success" or "failed"
	Error          string `json:"error,omitempty"`
}

// FileMetadataResponse - GET /api/v1/files/:fileId
type FileMetadataResponse struct {
	FileID        string              `json:"file_id"`
	OriginalName  string              `json:"original_name"`
	OriginalSize  int64               `json:"original_size"`
	TotalChunks   int                 `json:"total_chunks"`
	TotalShards   int                 `json:"total_shards"`
	N             int                 `json:"n"`
	K             int                 `json:"k"`
	ChunkSize     int64               `json:"chunk_size"`
	ShardSize     int64               `json:"shard_size"`
	Status        string              `json:"status"`
	CreatedAt     string              `json:"created_at"`
	UpdatedAt     string              `json:"updated_at"`
	HealthStatus  *FileHealthStatus   `json:"health_status,omitempty"`
}

// FileHealthStatus represents the health status of a file's shards
type FileHealthStatus struct {
	HealthyShards   int     `json:"healthy_shards"`
	CorruptedShards int     `json:"corrupted_shards"`
	MissingShards   int     `json:"missing_shards"`
	TotalShards     int     `json:"total_shards"`
	HealthPercent   float64 `json:"health_percent"`
	Recoverable     bool    `json:"recoverable"` // true if we have at least K healthy shards per chunk
}

// ListFilesRequest - GET /api/v1/files
type ListFilesRequest struct {
	Page     int    `form:"page" binding:"omitempty,gt=0"`
	PageSize int    `form:"page_size" binding:"omitempty,gt=0,lte=100"`
	Status   string `form:"status" binding:"omitempty,oneof=PENDING UPLOADED HEALTHY DEGRADED CORRUPTED"`
}

// ListFilesResponse - Response for GET /api/v1/files
type ListFilesResponse struct {
	Files      []FileMetadataResponse `json:"files"`
	Page       int                    `json:"page"`
	PageSize   int                    `json:"page_size"`
	TotalFiles int                    `json:"total_files"`
	TotalPages int                    `json:"total_pages"`
}

// DeleteFileRequest - DELETE /api/v1/files/:fileId
type DeleteFileRequest struct {
	FileID       string `uri:"fileId" binding:"required,uuid"`
	DeleteShards bool   `form:"delete_shards"` // If true, also delete shards from cloud storage
}

// DeleteFileResponse - Response for DELETE /api/v1/files/:fileId
type DeleteFileResponse struct {
	FileID         string `json:"file_id"`
	Success        bool   `json:"success"`
	Message        string `json:"message"`
	ShardsDeleted  int    `json:"shards_deleted,omitempty"`
}
