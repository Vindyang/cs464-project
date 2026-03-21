package dto

import "github.com/google/uuid"

// ShardChunkRequest - POST /api/sharding/shard
// Shards a single chunk of data into N total shards (K data + T parity)
type ShardChunkRequest struct {
	FileID     uuid.UUID `json:"fileId" binding:"required"`
	ChunkIndex int       `json:"chunkIndex" binding:"required,gte=0"`
	ChunkData  []byte    `json:"chunkData" binding:"required"` // Raw binary data
	N          int       `json:"n" binding:"required,gt=0"`     // Total shards
	K          int       `json:"k" binding:"required,gt=0"`     // Data shards
}

// ShardChunkResponse - Response for POST /api/sharding/shard
type ShardChunkResponse struct {
	FileID     uuid.UUID         `json:"fileId"`
	ChunkIndex int               `json:"chunkIndex"`
	Shards     []ShardDataOutput `json:"shards"`
	Metadata   ShardMetadata     `json:"metadata"`
}

// ShardDataOutput represents a single generated shard
type ShardDataOutput struct {
	ShardIndex     int    `json:"shardIndex"`
	ShardType      string `json:"shardType"` // "DATA" or "PARITY"
	ShardData      []byte `json:"shardData"` // Raw binary shard data
	ChecksumSHA256 string `json:"checksumSha256"`
}

// ShardMetadata contains metadata about the sharding operation
type ShardMetadata struct {
	N                int   `json:"n"`
	K                int   `json:"k"`
	OriginalSize     int64 `json:"originalSize"`
	ShardSize        int64 `json:"shardSize"`
	TotalDataShards  int   `json:"totalDataShards"`
	TotalParityShards int  `json:"totalParityShards"`
}

// ReconstructChunkRequest - POST /api/sharding/reconstruct
// Reconstructs original chunk from available shards (minimum K required)
type ReconstructChunkRequest struct {
	FileID          uuid.UUID           `json:"fileId" binding:"required"`
	ChunkIndex      int                 `json:"chunkIndex" binding:"required,gte=0"`
	N               int                 `json:"n" binding:"required,gt=0"`
	K               int                 `json:"k" binding:"required,gt=0"`
	AvailableShards []AvailableShardInput `json:"availableShards" binding:"required"`
}

// AvailableShardInput represents a shard available for reconstruction
type AvailableShardInput struct {
	ShardIndex     int    `json:"shardIndex" binding:"required,gte=0"`
	ShardType      string `json:"shardType" binding:"required,oneof=DATA PARITY"`
	ShardData      []byte `json:"shardData" binding:"required"`
	ChecksumSHA256 string `json:"checksumSha256" binding:"required,len=64"`
}

// ReconstructChunkResponse - Response for POST /api/sharding/reconstruct
type ReconstructChunkResponse struct {
	FileID            uuid.UUID            `json:"fileId"`
	ChunkIndex        int                  `json:"chunkIndex"`
	ReconstructedData []byte               `json:"reconstructedData"` // Original chunk data
	Metadata          ReconstructionMetadata `json:"metadata"`
}

// ReconstructionMetadata contains metadata about the reconstruction operation
type ReconstructionMetadata struct {
	OriginalSize         int64  `json:"originalSize"`
	ShardsUsed           int    `json:"shardsUsed"`
	ReconstructionMethod string `json:"reconstructionMethod"` // "data_only", "with_parity", "parity_only"
	Success              bool   `json:"success"`
}

// HealthCheckResponse - GET /api/sharding/health
type HealthCheckResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
	Version string `json:"version"`
}
