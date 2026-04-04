package app

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/vindyang/cs464-project/backend/services/shardmap/internal/infra/repository"
	"github.com/vindyang/cs464-project/backend/services/shared/api/dto"
	"github.com/vindyang/cs464-project/backend/services/shared/models"
)

type ShardMapService interface {
	RegisterFile(req *dto.RegisterFileRequest) (*dto.RegisterFileResponse, error)
	RecordShards(req *dto.RecordShardsRequest) (*dto.RecordShardsResponse, error)
	GetShardMap(fileID uuid.UUID) (*dto.GetShardMapResponse, error)
	GetFileMetadata(fileID uuid.UUID) (*dto.FileMetadataResponse, error)
	GetShardByID(shardID uuid.UUID) (*dto.ShardInfo, error)
	MarkShardStatus(shardID uuid.UUID, req *dto.MarkShardStatusRequest) error
	ListFiles() ([]dto.FileMetadataResponse, error)
	DeleteFile(fileID uuid.UUID) error
}

type shardMapService struct {
	fileRepo      repository.FileRepository
	shardRepo     repository.ShardRepository
	lifecycleRepo repository.LifecycleRepository
}

func NewShardMapService(
	fileRepo repository.FileRepository,
	shardRepo repository.ShardRepository,
	lifecycleRepo repository.LifecycleRepository,
) ShardMapService {
	return &shardMapService{
		fileRepo:      fileRepo,
		shardRepo:     shardRepo,
		lifecycleRepo: lifecycleRepo,
	}
}

func (s *shardMapService) RegisterFile(req *dto.RegisterFileRequest) (*dto.RegisterFileResponse, error) {
	if req.K <= 0 || req.N <= 0 || req.K > req.N {
		return nil, fmt.Errorf("invalid erasure coding parameters: K must be > 0, N must be >= K")
	}

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

	if err := s.fileRepo.Create(file); err != nil {
		return nil, fmt.Errorf("failed to register file: %w", err)
	}

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

func (s *shardMapService) RecordShards(req *dto.RecordShardsRequest) (*dto.RecordShardsResponse, error) {
	fileID, err := uuid.Parse(req.FileID)
	if err != nil {
		return nil, fmt.Errorf("invalid file ID: %w", err)
	}

	file, err := s.fileRepo.GetByID(fileID)
	if err != nil {
		return nil, fmt.Errorf("file not found: %w", err)
	}

	if len(req.Shards) != file.N {
		return nil, fmt.Errorf("expected %d shards but received %d", file.N, len(req.Shards))
	}

	shards := make([]*models.Shard, 0, len(req.Shards))
	now := time.Now()

	for _, shardDTO := range req.Shards {
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

	if err := s.shardRepo.CreateBatch(shards); err != nil {
		return nil, fmt.Errorf("failed to record shards: %w", err)
	}

	if err := s.fileRepo.UpdateStatus(fileID, models.FileStatusUploaded); err != nil {
		return nil, fmt.Errorf("failed to update file status: %w", err)
	}

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

func (s *shardMapService) GetShardMap(fileID uuid.UUID) (*dto.GetShardMapResponse, error) {
	file, err := s.fileRepo.GetByID(fileID)
	if err != nil {
		return nil, fmt.Errorf("file not found: %w", err)
	}

	shards, err := s.shardRepo.GetByFileID(fileID)
	if err != nil {
		return nil, fmt.Errorf("failed to get shards: %w", err)
	}

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

	summary, err := s.lifecycleRepo.GetLifecycleSummary(fileID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get lifecycle summary: %w", err)
	}

	var firstCreatedAt *string
	var lastDownloadedAt *string
	if summary != nil && summary.FirstCreatedAt != nil {
		v := summary.FirstCreatedAt.Format(time.RFC3339)
		firstCreatedAt = &v
	}
	if summary != nil && summary.LastDownloadedAt != nil {
		v := summary.LastDownloadedAt.Format(time.RFC3339)
		lastDownloadedAt = &v
	}

	name := ""
	if file.OriginalName != nil {
		name = *file.OriginalName
	}
	return &dto.GetShardMapResponse{
		FileID:           file.ID.String(),
		OriginalName:     name,
		OriginalSize:     file.OriginalSize,
		TotalChunks:      file.TotalChunks,
		N:                file.N,
		K:                file.K,
		ShardSize:        file.ShardSize,
		Status:           string(file.Status),
		FirstCreatedAt:   firstCreatedAt,
		LastDownloadedAt: lastDownloadedAt,
		Shards:           shardMetadata,
	}, nil
}

func (s *shardMapService) GetFileMetadata(fileID uuid.UUID) (*dto.FileMetadataResponse, error) {
	file, err := s.fileRepo.GetByID(fileID)
	if err != nil {
		return nil, fmt.Errorf("file not found: %w", err)
	}

	shards, err := s.shardRepo.GetByFileID(fileID)
	if err != nil {
		return nil, fmt.Errorf("failed to get shards: %w", err)
	}

	var healthyShards, corruptedShards, missingShards int
	for _, shard := range shards {
		switch shard.Status {
		case models.ShardStatusHealthy:
			healthyShards++
		case models.ShardStatusCorrupted:
			corruptedShards++
		case models.ShardStatusMissing:
			missingShards++
		}
	}
	totalShards := len(shards)
	healthPercent := 0.0
	if totalShards > 0 {
		healthPercent = float64(healthyShards) / float64(totalShards) * 100
	}

	summary, err := s.lifecycleRepo.GetLifecycleSummary(fileID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get lifecycle summary: %w", err)
	}

	var firstCreatedAt *string
	var lastDownloadedAt *string
	if summary != nil && summary.FirstCreatedAt != nil {
		v := summary.FirstCreatedAt.Format(time.RFC3339)
		firstCreatedAt = &v
	}
	if summary != nil && summary.LastDownloadedAt != nil {
		v := summary.LastDownloadedAt.Format(time.RFC3339)
		lastDownloadedAt = &v
	}

	name := ""
	if file.OriginalName != nil {
		name = *file.OriginalName
	}

	return &dto.FileMetadataResponse{
		FileID:           file.ID.String(),
		OriginalName:     name,
		OriginalSize:     file.OriginalSize,
		TotalChunks:      file.TotalChunks,
		TotalShards:      totalShards,
		N:                file.N,
		K:                file.K,
		ShardSize:        file.ShardSize,
		Status:           string(file.Status),
		CreatedAt:        file.CreatedAt.Format(time.RFC3339),
		UpdatedAt:        file.UpdatedAt.Format(time.RFC3339),
		FirstCreatedAt:   firstCreatedAt,
		LastDownloadedAt: lastDownloadedAt,
		HealthStatus: &dto.FileHealthStatus{
			HealthyShards:   healthyShards,
			CorruptedShards: corruptedShards,
			MissingShards:   missingShards,
			TotalShards:     totalShards,
			HealthPercent:   healthPercent,
			Recoverable:     healthyShards >= file.K,
		},
	}, nil
}

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

func (s *shardMapService) DeleteFile(fileID uuid.UUID) error {
	if err := s.fileRepo.Delete(fileID); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

func (s *shardMapService) ListFiles() ([]dto.FileMetadataResponse, error) {
	files, err := s.fileRepo.GetAllWithHealth()
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	responses := make([]dto.FileMetadataResponse, len(files))
	for i, f := range files {
		name := ""
		if f.OriginalName != nil {
			name = *f.OriginalName
		}

		totalShards := f.TotalShards
		healthPercent := 0.0
		if totalShards > 0 {
			healthPercent = float64(f.HealthyShards) / float64(totalShards) * 100
		}
		recoverable := f.HealthyShards >= f.K

		responses[i] = dto.FileMetadataResponse{
			FileID:       f.ID.String(),
			OriginalName: name,
			OriginalSize: f.OriginalSize,
			TotalChunks:  f.TotalChunks,
			TotalShards:  totalShards,
			N:            f.N,
			K:            f.K,
			ShardSize:    f.ShardSize,
			Status:       string(f.Status),
			CreatedAt:    f.CreatedAt.Format(time.RFC3339),
			UpdatedAt:    f.UpdatedAt.Format(time.RFC3339),
			HealthStatus: &dto.FileHealthStatus{
				HealthyShards:   f.HealthyShards,
				CorruptedShards: f.CorruptedShards,
				MissingShards:   f.MissingShards,
				TotalShards:     totalShards,
				HealthPercent:   healthPercent,
				Recoverable:     recoverable,
			},
		}

		summary, err := s.lifecycleRepo.GetLifecycleSummary(f.ID.String())
		if err != nil {
			return nil, fmt.Errorf("failed to get lifecycle summary: %w", err)
		}
		if summary != nil && summary.FirstCreatedAt != nil {
			v := summary.FirstCreatedAt.Format(time.RFC3339)
			responses[i].FirstCreatedAt = &v
		}
		if summary != nil && summary.LastDownloadedAt != nil {
			v := summary.LastDownloadedAt.Format(time.RFC3339)
			responses[i].LastDownloadedAt = &v
		}
	}

	return responses, nil
}

func (s *shardMapService) MarkShardStatus(shardID uuid.UUID, req *dto.MarkShardStatusRequest) error {
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

	if err := s.shardRepo.UpdateStatus(shardID, status); err != nil {
		return fmt.Errorf("failed to update shard status: %w", err)
	}

	return nil
}
