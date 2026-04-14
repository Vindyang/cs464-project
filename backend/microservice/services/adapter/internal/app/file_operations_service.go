package app

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"
	"github.com/vindyang/cs464-project/backend/services/shared/adapter"
	"github.com/vindyang/cs464-project/backend/services/shared/api/dto"
)

type ShardingService interface {
	ChunkFile(data []byte, chunkSize int64) ([][]byte, error)
	EncodeChunk(chunkData []byte, k, n int) ([][]byte, error)
	DecodeChunk(shards [][]byte, k, n int) ([]byte, error)
	CalculateChecksum(data []byte) string
}

type ShardMapService interface {
	RegisterFile(req *dto.RegisterFileRequest) (*dto.RegisterFileResponse, error)
	RecordShards(req *dto.RecordShardsRequest) (*dto.RecordShardsResponse, error)
	GetShardMap(fileID uuid.UUID) (*dto.GetShardMapResponse, error)
	GetShardByID(shardID uuid.UUID) (*dto.ShardInfo, error)
	MarkShardStatus(shardID uuid.UUID, req *dto.MarkShardStatusRequest) error
}

// FileOperationsService handles file upload and download operations.
type FileOperationsService interface {
	UploadFile(ctx context.Context, filename string, fileData []byte, k, n int, chunkSizeMB int, providers []string) (*dto.UploadFileResponse, error)
	DownloadFile(ctx context.Context, fileID uuid.UUID) ([]byte, string, error)
	GetFileMetadata(ctx context.Context, fileID uuid.UUID) (*dto.FileMetadataResponse, error)
	DeleteFile(ctx context.Context, fileID uuid.UUID, deleteShards bool) error
}

type fileOperationsService struct {
	shardingService ShardingService
	shardMapService ShardMapService
	adapterRegistry *adapter.Registry
}

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

func (s *fileOperationsService) UploadFile(
	ctx context.Context,
	filename string,
	fileData []byte,
	k, n int,
	chunkSizeMB int,
	providers []string,
) (*dto.UploadFileResponse, error) {
	startTime := time.Now()

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

	for _, providerID := range providers {
		if _, err := s.adapterRegistry.Get(providerID); err != nil {
			return nil, fmt.Errorf("provider %s not available: %w", providerID, err)
		}
	}

	chunkSizeBytes := int64(chunkSizeMB) * 1024 * 1024

	chunks, err := s.shardingService.ChunkFile(fileData, chunkSizeBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to chunk file: %w", err)
	}

	totalChunks := len(chunks)
	totalShards := totalChunks * n

	registerReq := &dto.RegisterFileRequest{
		OriginalName: filename,
		OriginalSize: int64(len(fileData)),
		TotalChunks:  totalChunks,
		N:            n,
		K:            k,
		ShardSize:    0,
	}

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

	parsedFileID, err := uuid.Parse(registerResp.FileID)
	if err != nil {
		return nil, fmt.Errorf("invalid file ID returned: %w", err)
	}

	stats := &dto.UploadStatistics{
		TotalShards:   totalShards,
		ProviderStats: make(map[string]int),
	}

	allShardInfos := make([]dto.ShardInfo, 0, totalShards)
	providerIndex := 0

	for chunkIdx, chunkData := range chunks {
		shards, err := s.shardingService.EncodeChunk(chunkData, k, n)
		if err != nil {
			return nil, fmt.Errorf("failed to encode chunk %d: %w", chunkIdx, err)
		}

		for shardIdx, shardData := range shards {
			shardType := "DATA"
			if shardIdx >= k {
				shardType = "PARITY"
			}

			checksum := s.shardingService.CalculateChecksum(shardData)

			providerID := providers[providerIndex%len(providers)]
			providerIndex++

			provider, err := s.adapterRegistry.Get(providerID)
			if err != nil {
				stats.FailedShards++
				continue
			}

			shardReader := bytes.NewReader(shardData)
			remoteID, err := provider.UploadShard(ctx, parsedFileID.String(), shardIdx, shardReader)
			if err != nil {
				stats.FailedShards++
				continue
			}

			stats.SuccessfulShards++
			stats.ProviderStats[providerID]++

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

	recordReq := &dto.RecordShardsRequest{
		FileID: parsedFileID.String(),
		Shards: allShardInfos,
	}

	_, err = s.shardMapService.RecordShards(recordReq)
	if err != nil {
		return nil, fmt.Errorf("failed to record shards: %w", err)
	}

	stats.DurationMS = time.Since(startTime).Milliseconds()

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

func (s *fileOperationsService) DownloadFile(ctx context.Context, fileID uuid.UUID) ([]byte, string, error) {
	shardMap, err := s.shardMapService.GetShardMap(fileID)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get shard map: %w", err)
	}

	chunkShards := make(map[int][]dto.ShardInfo)
	for _, shard := range shardMap.Shards {
		chunkShards[shard.ChunkIndex] = append(chunkShards[shard.ChunkIndex], shard)
	}

	var reconstructedChunks [][]byte

	for chunkIdx := 0; chunkIdx < shardMap.TotalChunks; chunkIdx++ {
		shards := chunkShards[chunkIdx]

		if len(shards) < shardMap.K {
			return nil, "", fmt.Errorf("insufficient shards for chunk %d: have %d, need %d", chunkIdx, len(shards), shardMap.K)
		}

		shardData := make([][]byte, shardMap.N)
		downloadedCount := 0

		for _, shardInfo := range shards {
			if downloadedCount >= shardMap.K {
				break
			}

			provider, err := s.adapterRegistry.Get(shardInfo.Provider)
			if err != nil {
				continue
			}

			reader, err := provider.DownloadShard(ctx, shardInfo.RemoteID)
			if err != nil {
				continue
			}

			data, err := io.ReadAll(reader)
			reader.Close()
			if err != nil {
				continue
			}

			checksum := s.shardingService.CalculateChecksum(data)
			if checksum != shardInfo.ChecksumSHA256 {
				continue
			}

			shardData[shardInfo.ShardIndex] = data
			downloadedCount++
		}

		if downloadedCount < shardMap.K {
			return nil, "", fmt.Errorf("failed to download enough shards for chunk %d: got %d, need %d", chunkIdx, downloadedCount, shardMap.K)
		}

		chunkData, err := s.shardingService.DecodeChunk(shardData, shardMap.K, shardMap.N)
		if err != nil {
			return nil, "", fmt.Errorf("failed to reconstruct chunk %d: %w", chunkIdx, err)
		}

		reconstructedChunks = append(reconstructedChunks, chunkData)
	}

	var fileBuffer bytes.Buffer
	for _, chunk := range reconstructedChunks {
		fileBuffer.Write(chunk)
	}

	fileData := fileBuffer.Bytes()
	if int64(len(fileData)) > shardMap.OriginalSize {
		fileData = fileData[:shardMap.OriginalSize]
	}

	return fileData, shardMap.OriginalName, nil
}

func (s *fileOperationsService) GetFileMetadata(ctx context.Context, fileID uuid.UUID) (*dto.FileMetadataResponse, error) {
	shardMap, err := s.shardMapService.GetShardMap(fileID)
	if err != nil {
		return nil, fmt.Errorf("failed to get shard map: %w", err)
	}

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

func (s *fileOperationsService) DeleteFile(ctx context.Context, fileID uuid.UUID, deleteShards bool) error {
	if deleteShards {
		shardMap, err := s.shardMapService.GetShardMap(fileID)
		if err != nil {
			return fmt.Errorf("failed to get shard map: %w", err)
		}

		for _, shard := range shardMap.Shards {
			provider, err := s.adapterRegistry.Get(shard.Provider)
			if err != nil {
				continue
			}

			_ = provider.DeleteShard(ctx, shard.RemoteID)
		}
	}

	return nil
}
