package service

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"
	"github.com/vindyang/cs464-project/backend/internal/adapter"
	"github.com/vindyang/cs464-project/backend/internal/api/dto"
)

// FileOperationsService handles file upload and download operations
type FileOperationsService interface {
	// UploadFile handles the entire file upload process: chunking, encoding, uploading shards, and registration
	UploadFile(ctx context.Context, filename string, fileData []byte, k, n int, chunkSizeMB int, providers []string) (*dto.UploadFileResponse, error)

	// DownloadFile handles file reconstruction and download
	DownloadFile(ctx context.Context, fileID uuid.UUID) ([]byte, string, error)

	// GetFileMetadata retrieves metadata for a file
	GetFileMetadata(ctx context.Context, fileID uuid.UUID) (*dto.FileMetadataResponse, error)

	// DeleteFile deletes a file and optionally its shards from cloud storage
	DeleteFile(ctx context.Context, fileID uuid.UUID, deleteShards bool) error
}

// fileOperationsService implements FileOperationsService
type fileOperationsService struct {
	shardingService ShardingService
	shardMapService ShardMapService
	adapterRegistry *adapter.Registry
}

// NewFileOperationsService creates a new FileOperationsService instance
func NewFileOperationsService(
	shardingService ShardingService,
	shardMapService ShardMapService,
	adapterRegistry *adapter.Registry,
) FileOperationsService {
	return &fileOperationsService{
		shardingService: shardingService,
		shardMapService: shardMapService,
		adapterRegistry: adapterRegistry,
	}
}

// UploadFile orchestrates the complete file upload process
func (s *fileOperationsService) UploadFile(
	ctx context.Context,
	filename string,
	fileData []byte,
	k, n int,
	chunkSizeMB int,
	providers []string,
) (*dto.UploadFileResponse, error) {
	startTime := time.Now()

	// Validate parameters
	if len(fileData) == 0 {
		return nil, fmt.Errorf("file data cannot be empty")
	}
	if k <= 0 || n <= 0 || k > n {
		return nil, fmt.Errorf("invalid erasure coding parameters: k=%d, n=%d", k, n)
	}
	if chunkSizeMB <= 0 {
		return nil, fmt.Errorf("chunk size must be positive")
	}
	if len(providers) == 0 {
		return nil, fmt.Errorf("at least one provider must be specified")
	}

	// Validate that all providers exist
	for _, providerID := range providers {
		if _, err := s.adapterRegistry.Get(providerID); err != nil {
			return nil, fmt.Errorf("provider %s not available: %w", providerID, err)
		}
	}

	// Calculate chunk size in bytes
	chunkSizeBytes := int64(chunkSizeMB) * 1024 * 1024

	// Step 1: Chunk the file
	chunks, err := s.shardingService.ChunkFile(fileData, chunkSizeBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to chunk file: %w", err)
	}

	totalChunks := len(chunks)
	totalShards := totalChunks * n

	// Step 2: Register file in shard map first
	registerReq := &dto.RegisterFileRequest{
		OriginalName: filename,
		OriginalSize: int64(len(fileData)),
		TotalChunks:  totalChunks,
		N:            n,
		K:            k,
		ShardSize:    0, // Will be calculated from first chunk
	}

	// Calculate shard size from first chunk
	if len(chunks) > 0 {
		shards, err := s.shardingService.EncodeChunk(chunks[0], k, n)
		if err != nil {
			return nil, fmt.Errorf("failed to encode first chunk: %w", err)
		}
		if len(shards) > 0 {
			registerReq.ShardSize = int64(len(shards[0]))
		}
	}

	registerResp, err := s.shardMapService.RegisterFile(registerReq)
	if err != nil {
		return nil, fmt.Errorf("failed to register file: %w", err)
	}

	// Parse file ID from response
	parsedFileID, err := uuid.Parse(registerResp.FileID)
	if err != nil {
		return nil, fmt.Errorf("invalid file ID returned: %w", err)
	}

	// Step 3: Process each chunk: encode and upload shards
	stats := &dto.UploadStatistics{
		TotalShards:   totalShards,
		ProviderStats: make(map[string]int),
	}

	allShardInfos := make([]dto.ShardInfo, 0, totalShards)
	providerIndex := 0

	for chunkIdx, chunkData := range chunks {
		// Encode chunk into shards
		shards, err := s.shardingService.EncodeChunk(chunkData, k, n)
		if err != nil {
			return nil, fmt.Errorf("failed to encode chunk %d: %w", chunkIdx, err)
		}

		// Upload each shard
		for shardIdx, shardData := range shards {
			// Determine shard type
			shardType := "DATA"
			if shardIdx >= k {
				shardType = "PARITY"
			}

			// Calculate checksum
			checksum := s.shardingService.CalculateChecksum(shardData)

			// Select provider (round-robin)
			providerID := providers[providerIndex%len(providers)]
			providerIndex++

			// Get provider from registry
			provider, err := s.adapterRegistry.Get(providerID)
			if err != nil {
				stats.FailedShards++
				continue
			}

			// Upload shard
			shardReader := bytes.NewReader(shardData)
			remoteID, err := provider.UploadShard(ctx, parsedFileID.String(), shardIdx, shardReader)
			if err != nil {
				stats.FailedShards++
				continue
			}

			// Record successful upload
			stats.SuccessfulShards++
			stats.ProviderStats[providerID]++

			// Add shard info for registration
			allShardInfos = append(allShardInfos, dto.ShardInfo{
				ChunkIndex:     chunkIdx,
				ShardIndex:     shardIdx,
				Type:           shardType,
				RemoteID:       remoteID,
				Provider:       providerID,
				ChecksumSHA256: checksum,
				Status:         "HEALTHY",
			})
		}
	}

	// Step 4: Record all shards in shard map
	recordReq := &dto.RecordShardsRequest{
		FileID: parsedFileID.String(),
		Shards: allShardInfos,
	}

	_, err = s.shardMapService.RecordShards(recordReq)
	if err != nil {
		return nil, fmt.Errorf("failed to record shards: %w", err)
	}

	// Calculate duration
	stats.DurationMS = time.Since(startTime).Milliseconds()

	// Build response
	response := &dto.UploadFileResponse{
		FileID:       parsedFileID.String(),
		OriginalName: filename,
		OriginalSize: int64(len(fileData)),
		TotalChunks:  totalChunks,
		TotalShards:  totalShards,
		N:            n,
		K:            k,
		ChunkSize:    chunkSizeBytes,
		ShardSize:    registerReq.ShardSize,
		Status:       "UPLOADED",
		Message:      fmt.Sprintf("Successfully uploaded %d shards across %d chunks", stats.SuccessfulShards, totalChunks),
		UploadStats:  stats,
	}

	return response, nil
}

// DownloadFile retrieves and reconstructs a file from shards
func (s *fileOperationsService) DownloadFile(ctx context.Context, fileID uuid.UUID) ([]byte, string, error) {
	// Step 1: Get shard map
	shardMap, err := s.shardMapService.GetShardMap(fileID)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get shard map: %w", err)
	}

	// Group shards by chunk
	chunkShards := make(map[int][]dto.ShardInfo)
	for _, shard := range shardMap.Shards {
		chunkShards[shard.ChunkIndex] = append(chunkShards[shard.ChunkIndex], shard)
	}

	// Step 2: Process each chunk
	var reconstructedChunks [][]byte

	for chunkIdx := 0; chunkIdx < shardMap.TotalChunks; chunkIdx++ {
		shards := chunkShards[chunkIdx]

		// Need at least K shards to reconstruct
		if len(shards) < shardMap.K {
			return nil, "", fmt.Errorf("insufficient shards for chunk %d: have %d, need %d", chunkIdx, len(shards), shardMap.K)
		}

		// Download shards
		shardData := make([][]byte, shardMap.N)
		downloadedCount := 0

		for _, shardInfo := range shards {
			// Skip if we already have enough shards
			if downloadedCount >= shardMap.K {
				break
			}

			// Get provider
			provider, err := s.adapterRegistry.Get(shardInfo.Provider)
			if err != nil {
				continue // Skip failed providers
			}

			// Download shard
			reader, err := provider.DownloadShard(ctx, shardInfo.RemoteID)
			if err != nil {
				continue // Skip failed downloads
			}

			data, err := io.ReadAll(reader)
			reader.Close()
			if err != nil {
				continue
			}

			// Verify checksum
			checksum := s.shardingService.CalculateChecksum(data)
			if checksum != shardInfo.ChecksumSHA256 {
				continue // Skip corrupted shards
			}

			shardData[shardInfo.ShardIndex] = data
			downloadedCount++
		}

		// Check if we have enough shards
		if downloadedCount < shardMap.K {
			return nil, "", fmt.Errorf("failed to download enough shards for chunk %d: got %d, need %d", chunkIdx, downloadedCount, shardMap.K)
		}

		// Reconstruct chunk
		chunkData, err := s.shardingService.DecodeChunk(shardData, shardMap.K, shardMap.N)
		if err != nil {
			return nil, "", fmt.Errorf("failed to reconstruct chunk %d: %w", chunkIdx, err)
		}

		reconstructedChunks = append(reconstructedChunks, chunkData)
	}

	// Step 3: Reassemble file from chunks
	var fileBuffer bytes.Buffer
	for _, chunk := range reconstructedChunks {
		fileBuffer.Write(chunk)
	}

	fileData := fileBuffer.Bytes()

	// Trim to original file size if needed
	if int64(len(fileData)) > shardMap.OriginalSize {
		fileData = fileData[:shardMap.OriginalSize]
	}

	return fileData, shardMap.OriginalName, nil
}

// GetFileMetadata retrieves metadata for a file
func (s *fileOperationsService) GetFileMetadata(ctx context.Context, fileID uuid.UUID) (*dto.FileMetadataResponse, error) {
	shardMap, err := s.shardMapService.GetShardMap(fileID)
	if err != nil {
		return nil, fmt.Errorf("failed to get shard map: %w", err)
	}

	// Calculate health status
	healthyCount := 0
	corruptedCount := 0
	missingCount := 0

	for _, shard := range shardMap.Shards {
		switch shard.Status {
		case "HEALTHY":
			healthyCount++
		case "CORRUPTED":
			corruptedCount++
		case "MISSING":
			missingCount++
		}
	}

	totalShards := len(shardMap.Shards)
	healthPercent := float64(healthyCount) / float64(totalShards) * 100

	// Check if file is recoverable (need K shards per chunk)
	shardsPerChunk := shardMap.N
	chunksRecoverable := healthyCount >= (shardMap.K * shardMap.TotalChunks)

	response := &dto.FileMetadataResponse{
		FileID:       shardMap.FileID,
		OriginalName: shardMap.OriginalName,
		OriginalSize: shardMap.OriginalSize,
		TotalChunks:  shardMap.TotalChunks,
		TotalShards:  totalShards,
		N:            shardMap.N,
		K:            shardMap.K,
		ChunkSize:    int64(shardsPerChunk),
		ShardSize:    shardMap.ShardSize,
		Status:       shardMap.Status,
		HealthStatus: &dto.FileHealthStatus{
			HealthyShards:   healthyCount,
			CorruptedShards: corruptedCount,
			MissingShards:   missingCount,
			TotalShards:     totalShards,
			HealthPercent:   healthPercent,
			Recoverable:     chunksRecoverable,
		},
	}

	return response, nil
}

// DeleteFile deletes a file and optionally its shards
func (s *fileOperationsService) DeleteFile(ctx context.Context, fileID uuid.UUID, deleteShards bool) error {
	if deleteShards {
		// Get shard map to find all shards
		shardMap, err := s.shardMapService.GetShardMap(fileID)
		if err != nil {
			return fmt.Errorf("failed to get shard map: %w", err)
		}

		// Delete each shard from cloud storage
		for _, shard := range shardMap.Shards {
			provider, err := s.adapterRegistry.Get(shard.Provider)
			if err != nil {
				continue // Skip if provider not available
			}

			// Best effort deletion
			_ = provider.DeleteShard(ctx, shard.RemoteID)
		}
	}

	// Note: Actual file/shard deletion from database would be implemented here
	// For now, we would just mark the status as deleted
	// This would require adding a Delete method to the FileRepository

	return nil
}
