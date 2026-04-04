package app

import (
	"fmt"

	"github.com/vindyang/cs464-project/backend/services/shardmap/internal/infra/repository"
	"github.com/vindyang/cs464-project/backend/services/shared/types"
)

// LifecycleService handles lifecycle event business logic.
type LifecycleService interface {
	RecordEvent(event *types.LifecycleEvent) error
	GetFileHistory(fileID string) (*types.FileHistoryResp, error)
	GetAllHistory() (*types.GlobalHistoryResp, error)
}

type lifecycleService struct {
	repo repository.LifecycleRepository
}

// NewLifecycleService creates a new LifecycleService.
func NewLifecycleService(repo repository.LifecycleRepository) LifecycleService {
	return &lifecycleService{repo: repo}
}

// RecordEvent validates and persists a lifecycle event.
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

// GetFileHistory returns all lifecycle events for a file.
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
	return &types.FileHistoryResp{
		FileID: fileID,
		Events: events,
	}, nil
}

// GetAllHistory returns lifecycle events across all files.
func (s *lifecycleService) GetAllHistory() (*types.GlobalHistoryResp, error) {
	events, err := s.repo.GetAll()
	if err != nil {
		return nil, fmt.Errorf("failed to get lifecycle events: %w", err)
	}
	if events == nil {
		events = []types.LifecycleEvent{}
	}
	return &types.GlobalHistoryResp{
		Events: events,
	}, nil
}
