package shardmap

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/vindyang/cs464-project/backend/monolith/shared/api/dto"
	"github.com/vindyang/cs464-project/backend/monolith/shared/models"
	"github.com/vindyang/cs464-project/backend/monolith/shared/types"
)

type ShardMapService interface {
	RegisterFile(req *dto.RegisterFileRequest) (*dto.RegisterFileResponse, error)
	RecordShards(req *dto.RecordShardsRequest) (*dto.RecordShardsResponse, error)
	GetShardMap(fileID uuid.UUID) (*dto.GetShardMapResponse, error)
	GetFileMetadata(fileID uuid.UUID) (*dto.FileMetadataResponse, error)
	GetShardByID(shardID uuid.UUID) (*dto.ShardInfo, error)
	MarkShardStatus(shardID uuid.UUID, req *dto.MarkShardStatusRequest) error
	UpdateFileHealthRefresh(fileID uuid.UUID, refreshedAt time.Time) error
	ListFiles() ([]dto.FileMetadataResponse, error)
	DeleteFile(fileID uuid.UUID) error
}

type LifecycleService interface {
	RecordEvent(event *types.LifecycleEvent) error
	GetFileHistory(fileID string) (*types.FileHistoryResp, error)
	GetAllHistory() (*types.GlobalHistoryResp, error)
	DeleteAllHistory() (int, error)
}

type shardMapService struct {
	fileRepo      FileRepository
	shardRepo     ShardRepository
	lifecycleRepo LifecycleRepository
}

type lifecycleService struct {
	repo LifecycleRepository
}

func NewShardMapService(fileRepo FileRepository, shardRepo ShardRepository, lifecycleRepo LifecycleRepository) ShardMapService {
	return &shardMapService{fileRepo: fileRepo, shardRepo: shardRepo, lifecycleRepo: lifecycleRepo}
}

func NewLifecycleService(repo LifecycleRepository) LifecycleService {
	return &lifecycleService{repo: repo}
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

	now := time.Now()
	shards := make([]*models.Shard, 0, len(req.Shards))
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
	for index, shard := range shards {
		shardMetadata[index] = dto.ShardInfo{
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

	return &dto.RecordShardsResponse{FileID: fileID.String(), Shards: shardMetadata}, nil
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
	for index, shard := range shards {
		shardMetadata[index] = dto.ShardInfo{
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
		value := summary.FirstCreatedAt.Format(time.RFC3339)
		firstCreatedAt = &value
	}
	if summary != nil && summary.LastDownloadedAt != nil {
		value := summary.LastDownloadedAt.Format(time.RFC3339)
		lastDownloadedAt = &value
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

	var healthyShards int
	var corruptedShards int
	var missingShards int
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
	var lastHealthRefreshAt *string
	if summary != nil && summary.FirstCreatedAt != nil {
		value := summary.FirstCreatedAt.Format(time.RFC3339)
		firstCreatedAt = &value
	}
	if summary != nil && summary.LastDownloadedAt != nil {
		value := summary.LastDownloadedAt.Format(time.RFC3339)
		lastDownloadedAt = &value
	}
	if file.LastHealthRefreshAt != nil {
		value := file.LastHealthRefreshAt.Format(time.RFC3339)
		lastHealthRefreshAt = &value
	}

	name := ""
	if file.OriginalName != nil {
		name = *file.OriginalName
	}

	return &dto.FileMetadataResponse{
		FileID:              file.ID.String(),
		OriginalName:        name,
		OriginalSize:        file.OriginalSize,
		TotalChunks:         file.TotalChunks,
		TotalShards:         totalShards,
		N:                   file.N,
		K:                   file.K,
		ShardSize:           file.ShardSize,
		Status:              string(file.Status),
		CreatedAt:           file.CreatedAt.Format(time.RFC3339),
		UpdatedAt:           file.UpdatedAt.Format(time.RFC3339),
		LastHealthRefreshAt: lastHealthRefreshAt,
		FirstCreatedAt:      firstCreatedAt,
		LastDownloadedAt:    lastDownloadedAt,
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
	for index, file := range files {
		name := ""
		if file.OriginalName != nil {
			name = *file.OriginalName
		}

		healthPercent := 0.0
		if file.TotalShards > 0 {
			healthPercent = float64(file.HealthyShards) / float64(file.TotalShards) * 100
		}

		responses[index] = dto.FileMetadataResponse{
			FileID:       file.ID.String(),
			OriginalName: name,
			OriginalSize: file.OriginalSize,
			TotalChunks:  file.TotalChunks,
			TotalShards:  file.TotalShards,
			N:            file.N,
			K:            file.K,
			ShardSize:    file.ShardSize,
			Status:       string(file.Status),
			CreatedAt:    file.CreatedAt.Format(time.RFC3339),
			UpdatedAt:    file.UpdatedAt.Format(time.RFC3339),
			HealthStatus: &dto.FileHealthStatus{
				HealthyShards:   file.HealthyShards,
				CorruptedShards: file.CorruptedShards,
				MissingShards:   file.MissingShards,
				TotalShards:     file.TotalShards,
				HealthPercent:   healthPercent,
				Recoverable:     file.HealthyShards >= file.K,
			},
		}

		if file.LastHealthRefreshAt != nil {
			value := file.LastHealthRefreshAt.Format(time.RFC3339)
			responses[index].LastHealthRefreshAt = &value
		}
		summary, err := s.lifecycleRepo.GetLifecycleSummary(file.ID.String())
		if err != nil {
			return nil, fmt.Errorf("failed to get lifecycle summary: %w", err)
		}
		if summary != nil && summary.FirstCreatedAt != nil {
			value := summary.FirstCreatedAt.Format(time.RFC3339)
			responses[index].FirstCreatedAt = &value
		}
		if summary != nil && summary.LastDownloadedAt != nil {
			value := summary.LastDownloadedAt.Format(time.RFC3339)
			responses[index].LastDownloadedAt = &value
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

	shard, err := s.shardRepo.GetByID(shardID)
	if err != nil {
		return fmt.Errorf("failed to load shard after status update: %w", err)
	}
	fileShards, err := s.shardRepo.GetByFileID(shard.FileID)
	if err != nil {
		return fmt.Errorf("failed to load file shards for status recompute: %w", err)
	}

	healthyCount := 0
	for _, existingShard := range fileShards {
		if existingShard.Status == models.ShardStatusHealthy {
			healthyCount++
		}
	}

	var fileStatus models.FileStatus
	if healthyCount == len(fileShards) {
		fileStatus = models.FileStatusUploaded
	} else {
		file, err := s.fileRepo.GetByID(shard.FileID)
		if err != nil {
			return fmt.Errorf("failed to load file for erasure coding threshold: %w", err)
		}
		if healthyCount >= file.K {
			fileStatus = models.FileStatusDegraded
		} else {
			fileStatus = models.FileStatusCorrupted
		}
	}

	if err := s.fileRepo.UpdateStatus(shard.FileID, fileStatus); err != nil {
		return fmt.Errorf("failed to update file status after shard status change: %w", err)
	}
	return nil
}

func (s *shardMapService) UpdateFileHealthRefresh(fileID uuid.UUID, refreshedAt time.Time) error {
	if err := s.fileRepo.UpdateLastHealthRefresh(fileID, refreshedAt.UTC()); err != nil {
		return fmt.Errorf("failed to update health refresh time: %w", err)
	}
	return nil
}

func (s *lifecycleService) RecordEvent(event *types.LifecycleEvent) error {
	if event.FileID == "" {
		return fmt.Errorf("file_id is required")
	}
	if event.EventType != "upload" && event.EventType != "download" && event.EventType != "delete" {
		return fmt.Errorf("event_type must be 'upload', 'download', or 'delete', got: %q", event.EventType)
	}
	if event.Status != "success" && event.Status != "failed" {
		return fmt.Errorf("status must be 'success' or 'failed', got: %q", event.Status)
	}
	if err := s.repo.Insert(event); err != nil {
		return fmt.Errorf("failed to record lifecycle event: %w", err)
	}
	return nil
}

func (s *lifecycleService) GetFileHistory(fileID string) (*types.FileHistoryResp, error) {
	if fileID == "" {
		return nil, fmt.Errorf("file_id is required")
	}
	events, err := s.repo.GetByFileID(fileID)
	if err != nil {
		return nil, fmt.Errorf("failed to get file history: %w", err)
	}
	if events == nil {
		events = []types.LifecycleEvent{}
	}
	return &types.FileHistoryResp{FileID: fileID, Events: events}, nil
}

func (s *lifecycleService) GetAllHistory() (*types.GlobalHistoryResp, error) {
	events, err := s.repo.GetAll()
	if err != nil {
		return nil, fmt.Errorf("failed to get lifecycle events: %w", err)
	}
	if events == nil {
		events = []types.LifecycleEvent{}
	}
	return &types.GlobalHistoryResp{Events: events}, nil
}

func (s *lifecycleService) DeleteAllHistory() (int, error) {
	deleted, err := s.repo.DeleteAll()
	if err != nil {
		return 0, fmt.Errorf("failed to delete lifecycle events: %w", err)
	}
	return deleted, nil
}
