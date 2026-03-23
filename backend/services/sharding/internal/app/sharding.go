package app

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/klauspost/reedsolomon"
)

// ShardingService defines sharding and reconstruction operations for this service.
type ShardingService interface {
	ChunkFile(data []byte, chunkSize int64) ([][]byte, error)
	EncodeChunk(chunkData []byte, k, n int) ([][]byte, error)
	DecodeChunk(shards [][]byte, k, n int) ([]byte, error)
	CalculateChecksum(data []byte) string
}

type shardingService struct{}

// NewShardingService constructs the service-local sharding implementation.
func NewShardingService() ShardingService {
	return &shardingService{}
}

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

func (s *shardingService) EncodeChunk(chunkData []byte, k, n int) ([][]byte, error) {
	if k <= 0 || n <= 0 || k > n {
		return nil, fmt.Errorf("invalid erasure coding parameters: K=%d, N=%d (K must be > 0 and K <= N)", k, n)
	}

	if len(chunkData) == 0 {
		return nil, fmt.Errorf("chunk data cannot be empty")
	}

	enc, err := reedsolomon.New(k, n-k)
	if err != nil {
		return nil, fmt.Errorf("failed to create Reed-Solomon encoder: %w", err)
	}

	shardSize := (len(chunkData) + k - 1) / k

	shards := make([][]byte, n)
	for i := 0; i < k; i++ {
		shards[i] = make([]byte, shardSize)

		start := i * shardSize
		end := start + shardSize
		if end > len(chunkData) {
			end = len(chunkData)
		}

		if start < len(chunkData) {
			copy(shards[i], chunkData[start:end])
		}
	}

	for i := k; i < n; i++ {
		shards[i] = make([]byte, shardSize)
	}

	if err := enc.Encode(shards); err != nil {
		return nil, fmt.Errorf("failed to encode shards: %w", err)
	}

	return shards, nil
}

func (s *shardingService) DecodeChunk(shards [][]byte, k, n int) ([]byte, error) {
	if k <= 0 || n <= 0 || k > n {
		return nil, fmt.Errorf("invalid erasure coding parameters: K=%d, N=%d", k, n)
	}

	if len(shards) != n {
		return nil, fmt.Errorf("expected %d shards but got %d", n, len(shards))
	}

	availableCount := 0
	for _, shard := range shards {
		if shard != nil {
			availableCount++
		}
	}

	if availableCount < k {
		return nil, fmt.Errorf("insufficient shards for reconstruction: have %d, need %d", availableCount, k)
	}

	enc, err := reedsolomon.New(k, n-k)
	if err != nil {
		return nil, fmt.Errorf("failed to create Reed-Solomon decoder: %w", err)
	}

	if err := enc.Reconstruct(shards); err != nil {
		return nil, fmt.Errorf("failed to reconstruct shards: %w", err)
	}

	ok, err := enc.Verify(shards)
	if err != nil {
		return nil, fmt.Errorf("failed to verify shards: %w", err)
	}
	if !ok {
		return nil, fmt.Errorf("shard verification failed")
	}

	var buf bytes.Buffer
	if err := enc.Join(&buf, shards, len(shards[0])*k); err != nil {
		return nil, fmt.Errorf("failed to join shards: %w", err)
	}

	return buf.Bytes(), nil
}

func (s *shardingService) CalculateChecksum(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}
