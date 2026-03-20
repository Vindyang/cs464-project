package service

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"

	"github.com/klauspost/reedsolomon"
)

// ShardingService defines the interface for file sharding and reconstruction operations
type ShardingService interface {
	// ChunkFile splits a file into chunks of specified size
	ChunkFile(data []byte, chunkSize int64) ([][]byte, error)

	// EncodeChunk applies Reed-Solomon encoding to a chunk, producing N shards (K data + N-K parity)
	EncodeChunk(chunkData []byte, k, n int) ([][]byte, error)

	// DecodeChunk reconstructs original data from K or more shards
	DecodeChunk(shards [][]byte, k, n int) ([]byte, error)

	// CalculateChecksum computes SHA256 checksum of data
	CalculateChecksum(data []byte) string
}

// shardingService implements ShardingService
type shardingService struct{}

// NewShardingService creates a new ShardingService instance
func NewShardingService() ShardingService {
	return &shardingService{}
}

// ChunkFile splits file data into chunks of specified size
func (s *shardingService) ChunkFile(data []byte, chunkSize int64) ([][]byte, error) {
	if chunkSize <= 0 {
		return nil, fmt.Errorf("chunk size must be positive")
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("data cannot be empty")
	}

	var chunks [][]byte
	dataLen := int64(len(data))

	for offset := int64(0); offset < dataLen; offset += chunkSize {
		end := offset + chunkSize
		if end > dataLen {
			end = dataLen
		}

		chunk := make([]byte, end-offset)
		copy(chunk, data[offset:end])
		chunks = append(chunks, chunk)
	}

	return chunks, nil
}

// EncodeChunk applies Reed-Solomon encoding to a chunk
// Returns N shards where first K are data shards and last N-K are parity shards
func (s *shardingService) EncodeChunk(chunkData []byte, k, n int) ([][]byte, error) {
	// Validate parameters
	if k <= 0 || n <= 0 || k > n {
		return nil, fmt.Errorf("invalid erasure coding parameters: K=%d, N=%d (K must be > 0 and K <= N)", k, n)
	}

	if len(chunkData) == 0 {
		return nil, fmt.Errorf("chunk data cannot be empty")
	}

	// Create Reed-Solomon encoder
	// k = data shards, n-k = parity shards
	enc, err := reedsolomon.New(k, n-k)
	if err != nil {
		return nil, fmt.Errorf("failed to create Reed-Solomon encoder: %w", err)
	}

	// Calculate shard size (each data shard will be this size)
	shardSize := (len(chunkData) + k - 1) / k // Ceiling division

	// Split data into K data shards
	shards := make([][]byte, n)
	for i := 0; i < k; i++ {
		shards[i] = make([]byte, shardSize)

		// Copy data into shard
		start := i * shardSize
		end := start + shardSize
		if end > len(chunkData) {
			end = len(chunkData)
		}

		if start < len(chunkData) {
			copy(shards[i], chunkData[start:end])
		}
	}

	// Create empty parity shards
	for i := k; i < n; i++ {
		shards[i] = make([]byte, shardSize)
	}

	// Encode to generate parity shards
	if err := enc.Encode(shards); err != nil {
		return nil, fmt.Errorf("failed to encode shards: %w", err)
	}

	return shards, nil
}

// DecodeChunk reconstructs original data from available shards
// Requires at least K shards to be present (can be any combination of data and parity shards)
func (s *shardingService) DecodeChunk(shards [][]byte, k, n int) ([]byte, error) {
	// Validate parameters
	if k <= 0 || n <= 0 || k > n {
		return nil, fmt.Errorf("invalid erasure coding parameters: K=%d, N=%d", k, n)
	}

	if len(shards) != n {
		return nil, fmt.Errorf("expected %d shards but got %d", n, len(shards))
	}

	// Count available shards
	availableCount := 0
	for _, shard := range shards {
		if shard != nil {
			availableCount++
		}
	}

	if availableCount < k {
		return nil, fmt.Errorf("insufficient shards for reconstruction: have %d, need %d", availableCount, k)
	}

	// Create Reed-Solomon decoder
	enc, err := reedsolomon.New(k, n-k)
	if err != nil {
		return nil, fmt.Errorf("failed to create Reed-Solomon decoder: %w", err)
	}

	// Reconstruct missing shards
	if err := enc.Reconstruct(shards); err != nil {
		return nil, fmt.Errorf("failed to reconstruct shards: %w", err)
	}

	// Verify reconstruction
	ok, err := enc.Verify(shards)
	if err != nil {
		return nil, fmt.Errorf("failed to verify shards: %w", err)
	}
	if !ok {
		return nil, fmt.Errorf("shard verification failed")
	}

	// Reconstruct original data from data shards only
	var buf bytes.Buffer
	if err := enc.Join(&buf, shards, len(shards[0])*k); err != nil {
		return nil, fmt.Errorf("failed to join shards: %w", err)
	}

	return buf.Bytes(), nil
}

// CalculateChecksum computes SHA256 checksum of data
func (s *shardingService) CalculateChecksum(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// CalculateChecksumFromReader computes SHA256 checksum from an io.Reader
func CalculateChecksumFromReader(r io.Reader) (string, error) {
	hasher := sha256.New()
	if _, err := io.Copy(hasher, r); err != nil {
		return "", fmt.Errorf("failed to calculate checksum: %w", err)
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}
