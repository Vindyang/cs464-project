package dto

import (
	"github.com/google/uuid"
	"github.com/vindyang/cs464-project/backend/services/shared/models"
)

// CreateFileRequest - POST /api/map/files
type CreateFileRequest struct {
	OriginalName      *string                    `json:"originalName,omitempty"`
	OriginalSize      int64                      `json:"originalSize" binding:"required,gt=0"`
	TotalChunks       int                        `json:"totalChunks" binding:"required,gt=0"`
	N                 int                        `json:"n" binding:"required,gt=0"`
	K                 int                        `json:"k" binding:"required,gt=0"`
	ShardSize         int64                      `json:"shardSize" binding:"required,gt=0"`
	ShardDistribution []ShardDistributionRequest `json:"shardDistribution" binding:"required"`
}

// ShardDistributionRequest represents a single shard's metadata during file creation
type ShardDistributionRequest struct {
	ChunkIndex     int              `json:"chunkIndex" binding:"required,gte=0"`
	ShardIndex     int              `json:"shardIndex" binding:"required,gte=0"`
	ShardType      models.ShardType `json:"shardType" binding:"required,oneof=DATA PARITY"`
	RemoteID       string           `json:"remoteId" binding:"required"`
	Provider       string           `json:"provider" binding:"required"`
	ChecksumSHA256 string           `json:"checksumSha256" binding:"required,len=64"`
}

// CreateFileResponse - Response for POST /api/map/files
type CreateFileResponse struct {
	FileID    uuid.UUID `json:"fileId"`
	Status    string    `json:"status"`
	Message   string    `json:"message"`
	CreatedAt string    `json:"createdAt"`
}

// GetFileResponse - GET /api/map/files/{id}
type GetFileResponse struct {
	File   FileDTO    `json:"file"`
	Shards []ShardDTO `json:"shards"`
}

// FileDTO represents file metadata in API responses
type FileDTO struct {
	ID           uuid.UUID `json:"id"`
	OriginalName *string   `json:"originalName,omitempty"`
	OriginalSize int64     `json:"originalSize"`
	TotalChunks  int       `json:"totalChunks"`
	N            int       `json:"n"`
	K            int       `json:"k"`
	ShardSize    int64     `json:"shardSize"`
	Status       string    `json:"status"`
	CreatedAt    string    `json:"createdAt"`
	UpdatedAt    string    `json:"updatedAt"`
}

// ShardDTO represents shard metadata in API responses
type ShardDTO struct {
	ID             uuid.UUID `json:"id"`
	FileID         uuid.UUID `json:"fileId"`
	ChunkIndex     int       `json:"chunkIndex"`
	ShardIndex     int       `json:"shardIndex"`
	ShardType      string    `json:"shardType"`
	RemoteID       string    `json:"remoteId"`
	Provider       string    `json:"provider"`
	ChecksumSHA256 string    `json:"checksumSha256"`
	Status         string    `json:"status"`
	CreatedAt      string    `json:"createdAt"`
	UpdatedAt      string    `json:"updatedAt"`
}

// UpdateShardRequest - PATCH /api/map/files/{id}/shards/{index}
type UpdateShardRequest struct {
	RemoteID       *string `json:"remoteId,omitempty"`
	Provider       *string `json:"provider,omitempty"`
	Status         *string `json:"status,omitempty" binding:"omitempty,oneof=PENDING HEALTHY CORRUPTED MISSING"`
	ChecksumSHA256 *string `json:"checksumSha256,omitempty" binding:"omitempty,len=64"`
}

// UpdateShardResponse - Response for PATCH /api/map/files/{id}/shards/{index}
type UpdateShardResponse struct {
	Success bool      `json:"success"`
	Message string    `json:"message"`
	Shard   *ShardDTO `json:"shard,omitempty"`
}

// ErrorResponse - Standard error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}

// RegisterFileRequest - POST /api/v1/shards/register
type RegisterFileRequest struct {
	OriginalName string `json:"original_name"`
	OriginalSize int64  `json:"original_size"`
	TotalChunks  int    `json:"total_chunks"`
	N            int    `json:"n"`
	K            int    `json:"k"`
	ShardSize    int64  `json:"shard_size"`
}

// RegisterFileResponse - Response for POST /api/v1/shards/register
type RegisterFileResponse struct {
	FileID       string `json:"file_id"`
	OriginalName string `json:"original_name"`
	OriginalSize int64  `json:"original_size"`
	TotalChunks  int    `json:"total_chunks"`
	N            int    `json:"n"`
	K            int    `json:"k"`
	ShardSize    int64  `json:"shard_size"`
	Status       string `json:"status"`
}

// RecordShardsRequest - POST /api/v1/shards/record
type RecordShardsRequest struct {
	FileID string      `json:"file_id"`
	Shards []ShardInfo `json:"shards"`
}

// RecordShardsResponse - Response for POST /api/v1/shards/record
type RecordShardsResponse struct {
	FileID string      `json:"file_id"`
	Shards []ShardInfo `json:"shards"`
}

// ShardInfo represents individual shard details in requests/responses
type ShardInfo struct {
	ShardID        string `json:"shard_id,omitempty"`
	ChunkIndex     int    `json:"chunk_index"`
	ShardIndex     int    `json:"shard_index"`
	Type           string `json:"type"`
	RemoteID       string `json:"remote_id"`
	Provider       string `json:"provider"`
	ChecksumSHA256 string `json:"checksum_sha256"`
	Status         string `json:"status,omitempty"`
}

// GetShardMapResponse - GET /api/v1/shards/file/:fileId
type GetShardMapResponse struct {
	FileID           string      `json:"file_id"`
	OriginalName     string      `json:"original_name"`
	OriginalSize     int64       `json:"original_size"`
	TotalChunks      int         `json:"total_chunks"`
	N                int         `json:"n"`
	K                int         `json:"k"`
	ShardSize        int64       `json:"shard_size"`
	Status           string      `json:"status"`
	FirstCreatedAt   *string     `json:"first_created_at,omitempty"`
	LastDownloadedAt *string     `json:"last_downloaded_at,omitempty"`
	Shards           []ShardInfo `json:"shards"`
}

// MarkShardStatusRequest - PUT /api/v1/shards/:shardId/status
type MarkShardStatusRequest struct {
	Status string `json:"status"`
}
