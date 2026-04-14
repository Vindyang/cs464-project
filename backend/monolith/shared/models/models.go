package models

import (
	"time"

	"github.com/google/uuid"
)

type FileStatus string

const (
	FileStatusPending   FileStatus = "PENDING"
	FileStatusUploaded  FileStatus = "UPLOADED"
	FileStatusDegraded  FileStatus = "DEGRADED"
	FileStatusCorrupted FileStatus = "CORRUPTED"
	FileStatusDeleted   FileStatus = "DELETED"
)

type ShardStatus string

const (
	ShardStatusPending   ShardStatus = "PENDING"
	ShardStatusHealthy   ShardStatus = "HEALTHY"
	ShardStatusCorrupted ShardStatus = "CORRUPTED"
	ShardStatusMissing   ShardStatus = "MISSING"
)

type ShardType string

const (
	ShardTypeData   ShardType = "DATA"
	ShardTypeParity ShardType = "PARITY"
)

type File struct {
	ID                  uuid.UUID  `json:"id" db:"id"`
	OriginalName        *string    `json:"originalName,omitempty" db:"original_name"`
	OriginalSize        int64      `json:"originalSize" db:"original_size"`
	TotalChunks         int        `json:"totalChunks" db:"total_chunks"`
	N                   int        `json:"n" db:"n"`
	K                   int        `json:"k" db:"k"`
	ShardSize           int64      `json:"shardSize" db:"shard_size"`
	Status              FileStatus `json:"status" db:"status"`
	CreatedAt           time.Time  `json:"createdAt" db:"created_at"`
	UpdatedAt           time.Time  `json:"updatedAt" db:"updated_at"`
	LastHealthRefreshAt *time.Time `json:"lastHealthRefreshAt,omitempty" db:"last_health_refresh_at"`
}

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

type FileWithShards struct {
	File   File    `json:"file"`
	Shards []Shard `json:"shards"`
}

type FileWithHealth struct {
	File
	HealthyShards   int `db:"healthy_shards"`
	CorruptedShards int `db:"corrupted_shards"`
	MissingShards   int `db:"missing_shards"`
	TotalShards     int `db:"total_shards"`
}
