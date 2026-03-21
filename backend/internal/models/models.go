package models

import (
	"time"

	"github.com/google/uuid"
)

// FileStatus represents the overall health status of a file
type FileStatus string

const (
	FileStatusPending  FileStatus = "PENDING"
	FileStatusUploaded FileStatus = "UPLOADED"
	FileStatusDegraded FileStatus = "DEGRADED"
	FileStatusDeleted  FileStatus = "DELETED"
)

// ShardStatus represents the health status of an individual shard
type ShardStatus string

const (
	ShardStatusPending   ShardStatus = "PENDING"
	ShardStatusHealthy   ShardStatus = "HEALTHY"
	ShardStatusCorrupted ShardStatus = "CORRUPTED"
	ShardStatusMissing   ShardStatus = "MISSING"
)

// ShardType indicates whether a shard is data or parity
type ShardType string

const (
	ShardTypeData   ShardType = "DATA"
	ShardTypeParity ShardType = "PARITY"
)

// File represents metadata for an uploaded file
type File struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	OriginalName *string    `json:"originalName,omitempty" db:"original_name"`
	OriginalSize int64      `json:"originalSize" db:"original_size"`
	TotalChunks  int        `json:"totalChunks" db:"total_chunks"`
	N            int        `json:"n" db:"n"` // Total shards per chunk
	K            int        `json:"k" db:"k"` // Data shards per chunk
	ShardSize    int64      `json:"shardSize" db:"shard_size"`
	Status       FileStatus `json:"status" db:"status"`
	CreatedAt    time.Time  `json:"createdAt" db:"created_at"`
	UpdatedAt    time.Time  `json:"updatedAt" db:"updated_at"`
}

// Shard represents an individual shard stored on a cloud provider
type Shard struct {
	ID             uuid.UUID   `json:"id" db:"id"`
	FileID         uuid.UUID   `json:"fileId" db:"file_id"`
	ChunkIndex     int         `json:"chunkIndex" db:"chunk_index"`
	ShardIndex     int         `json:"shardIndex" db:"shard_index"`
	ShardType      ShardType   `json:"shardType" db:"shard_type"`
	RemoteID       string      `json:"remoteId" db:"remote_id"`
	Provider       string      `json:"provider" db:"provider"`
	ChecksumSHA256 string      `json:"checksumSha256" db:"checksum_sha256"`
	Status         ShardStatus `json:"status" db:"status"`
	CreatedAt      time.Time   `json:"createdAt" db:"created_at"`
	UpdatedAt      time.Time   `json:"updatedAt" db:"updated_at"`
}

// FileWithShards represents a complete file with all its shards
type FileWithShards struct {
	File   File    `json:"file"`
	Shards []Shard `json:"shards"`
}
