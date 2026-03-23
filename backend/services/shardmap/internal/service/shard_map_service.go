package service

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/vindyang/cs464-project/backend/services/shardmap/internal/api/dto"
	"github.com/vindyang/cs464-project/backend/services/shardmap/internal/models"
	"github.com/vindyang/cs464-project/backend/services/shardmap/internal/repository"
)

// ShardMapService defines the interface for shard map operations
type ShardMapService interface {
	RegisterFile(req *dto.RegisterFileRequest) (*dto.RegisterFileResponse, error)
	RecordShards(req *dto.RecordShardsRequest) (*dto.RecordShardsResponse, error)
	GetShardMap(fileID uuid.UUID) (*dto.GetShardMapResponse, error)
	GetShardByID(shardID uuid.UUID) (*dto.ShardInfo, error)
	MarkShardStatus(shardID uuid.UUID, req *dto.MarkShardStatusRequest) error
}

// shardMapService implements ShardMapService
type shardMapService struct {
	fileRepo  repository.FileRepository
	shardRepo repository.ShardRepository
}

// NewShardMapService creates a new ShardMapService instance
func NewShardMapService(fileRepo repository.FileRepository, shardRepo repository.ShardRepository) ShardMapService {
	return &shardMapService{
		fileRepo:  fileRepo,
		shardRepo: shardRepo,
	}
}

// RegisterFile creates a new file entry in the shard map
func (s *shardMapService) RegisterFile(req *dto.RegisterFileRequest) (*dto.RegisterFileResponse, error) {
	// Validate erasure coding parameters
	if req.K <= 0 || req.N <= 0 || req.K > req.N {
		return nil, fmt.Errorf("invalid erasure coding parameters: K must be > 0, N must be >= K")
	}

	// Create file model
	originalName := req.OriginalName
	file := &models.File{
		ID:           uuid.New(),
		OriginalName: &originalName,
		OriginalSize: req.OriginalSize,
		TotalChunks:  req.TotalChunks,
		N:            req.N,
		K:            req.K,
		ShardSize:    req.ShardSize,
		Status:       models.FileStatusPending,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Save to database
	if err := s.fileRepo.Create(file); err != nil {
		return nil, fmt.Errorf("failed to register file: %w", err)
	}

	// Build response
	name := ""
	if file.OriginalName != nil {
		name = *file.OriginalName
	}
	return &dto.RegisterFileResponse{
		FileID:       file.ID.String(),
		OriginalName: name,
		OriginalSize: file.OriginalSize,
		TotalChunks:  file.TotalChunks,
		N:            file.N,
		K:            file.K,
		ShardSize:    file.ShardSize,
		Status:       string(file.Status),
	}, nil
}

// RecordShards saves shard metadata to the shard map
func (s *shardMapService) RecordShards(req *dto.RecordShardsRequest) (*dto.RecordShardsResponse, error) {
	// Validate file exists
	fileID, err := uuid.Parse(req.FileID)
	if err != nil {
		return nil, fmt.Errorf("invalid file ID: %w", err)
	}

	file, err := s.fileRepo.GetByID(fileID)
	if err != nil {
		return nil, fmt.Errorf("file not found: %w", err)
	}

	// Validate shard count matches erasure coding parameters
	if len(req.Shards) != file.N {
		return nil, fmt.Errorf("expected %d shards but received %d", file.N, len(req.Shards))
	}

	// Create shard models
	shards := make([]*models.Shard, 0, len(req.Shards))
	now := time.Now()

	for _, shardDTO := range req.Shards {
		// Parse shard type
		var shardType models.ShardType
		switch shardDTO.Type {
		case "DATA":
			shardType = models.ShardTypeData
		case "PARITY":
			shardType = models.ShardTypeParity
		default:
			return nil, fmt.Errorf("invalid shard type: %s", shardDTO.Type)
		}

		shard := &models.Shard{
			ID:             uuid.New(),
			FileID:         fileID,
			ChunkIndex:     shardDTO.ChunkIndex,
			ShardIndex:     shardDTO.ShardIndex,
			ShardType:      shardType,
			RemoteID:       shardDTO.RemoteID,
			Provider:       shardDTO.Provider,
			ChecksumSHA256: shardDTO.ChecksumSHA256,
			Status:         models.ShardStatusPending,
			CreatedAt:      now,
			UpdatedAt:      now,
		}

		shards = append(shards, shard)
	}

	// Save shards in batch
	if err := s.shardRepo.CreateBatch(shards); err != nil {
		return nil, fmt.Errorf("failed to record shards: %w", err)
	}

	// Update file status if all chunks are recorded
	// For now, just set to UPLOADED - more sophisticated logic can be added later
	if err := s.fileRepo.UpdateStatus(fileID, models.FileStatusUploaded); err != nil {
		return nil, fmt.Errorf("failed to update file status: %w", err)
	}

	// Build response
	shardMetadata := make([]dto.ShardInfo, len(shards))
	for i, shard := range shards {
		shardMetadata[i] = dto.ShardInfo{
			ShardID:        shard.ID.String(),
			ChunkIndex:     shard.ChunkIndex,
			ShardIndex:     shard.ShardIndex,
			Type:           string(shard.ShardType),
			RemoteID:       shard.RemoteID,
			Provider:       shard.Provider,
			ChecksumSHA256: shard.ChecksumSHA256,
			Status:         string(shard.Status),
		}
	}

	return &dto.RecordShardsResponse{
		FileID: fileID.String(),
		Shards: shardMetadata,
	}, nil
}

// GetShardMap retrieves the complete shard map for a file
func (s *shardMapService) GetShardMap(fileID uuid.UUID) (*dto.GetShardMapResponse, error) {
	// Get file
	file, err := s.fileRepo.GetByID(fileID)
	if err != nil {
		return nil, fmt.Errorf("file not found: %w", err)
	}

	// Get shards
	shards, err := s.shardRepo.GetByFileID(fileID)
	if err != nil {
		return nil, fmt.Errorf("failed to get shards: %w", err)
	}

	// Build response
	shardMetadata := make([]dto.ShardInfo, len(shards))
	for i, shard := range shards {
		shardMetadata[i] = dto.ShardInfo{
			ShardID:        shard.ID.String(),
			ChunkIndex:     shard.ChunkIndex,
			ShardIndex:     shard.ShardIndex,
			Type:           string(shard.ShardType),
			RemoteID:       shard.RemoteID,
			Provider:       shard.Provider,
			ChecksumSHA256: shard.ChecksumSHA256,
			Status:         string(shard.Status),
		}
	}

	name := ""
	if file.OriginalName != nil {
		name = *file.OriginalName
	}
	return &dto.GetShardMapResponse{
		FileID:       file.ID.String(),
		OriginalName: name,
		OriginalSize: file.OriginalSize,
		TotalChunks:  file.TotalChunks,
		N:            file.N,
		K:            file.K,
		ShardSize:    file.ShardSize,
		Status:       string(file.Status),
		Shards:       shardMetadata,
	}, nil
}

// GetShardByID retrieves metadata for a specific shard
func (s *shardMapService) GetShardByID(shardID uuid.UUID) (*dto.ShardInfo, error) {
	shard, err := s.shardRepo.GetByID(shardID)
	if err != nil {
		return nil, fmt.Errorf("shard not found: %w", err)
	}

	return &dto.ShardInfo{
		ShardID:        shard.ID.String(),
		ChunkIndex:     shard.ChunkIndex,
		ShardIndex:     shard.ShardIndex,
		Type:           string(shard.ShardType),
		RemoteID:       shard.RemoteID,
		Provider:       shard.Provider,
		ChecksumSHA256: shard.ChecksumSHA256,
		Status:         string(shard.Status),
	}, nil
}

// MarkShardStatus updates the status of a shard
func (s *shardMapService) MarkShardStatus(shardID uuid.UUID, req *dto.MarkShardStatusRequest) error {
	// Parse status
	var status models.ShardStatus
	switch req.Status {
	case "PENDING":
		status = models.ShardStatusPending
	case "HEALTHY":
		status = models.ShardStatusHealthy
	case "CORRUPTED":
		status = models.ShardStatusCorrupted
	case "MISSING":
		status = models.ShardStatusMissing
	default:
		return fmt.Errorf("invalid shard status: %s", req.Status)
	}

	// Update status
	if err := s.shardRepo.UpdateStatus(shardID, status); err != nil {
		return fmt.Errorf("failed to update shard status: %w", err)
	}

	return nil
}
