package service_test

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/vindyang/cs464-project/backend/internal/api/dto"
	"github.com/vindyang/cs464-project/backend/internal/models"
	"github.com/vindyang/cs464-project/backend/internal/service"
)

type mockFileRepo struct {
	createFn       func(*models.File) error
	getByIDFn      func(uuid.UUID) (*models.File, error)
	getAllFn       func() ([]*models.File, error)
	updateStatusFn func(uuid.UUID, models.FileStatus) error
	deleteFn       func(uuid.UUID) error
}

func (m *mockFileRepo) Create(file *models.File) error {
	if m.createFn != nil {
		return m.createFn(file)
	}
	return nil
}
func (m *mockFileRepo) GetByID(id uuid.UUID) (*models.File, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(id)
	}
	return nil, errors.New("not implemented")
}
func (m *mockFileRepo) GetAll() ([]*models.File, error) {
	if m.getAllFn != nil {
		return m.getAllFn()
	}
	return nil, nil
}
func (m *mockFileRepo) UpdateStatus(id uuid.UUID, status models.FileStatus) error {
	if m.updateStatusFn != nil {
		return m.updateStatusFn(id, status)
	}
	return nil
}
func (m *mockFileRepo) Delete(id uuid.UUID) error {
	if m.deleteFn != nil {
		return m.deleteFn(id)
	}
	return nil
}

type mockShardRepo struct {
	createFn         func(*models.Shard) error
	createBatchFn    func([]*models.Shard) error
	getByIDFn        func(uuid.UUID) (*models.Shard, error)
	getByFileIDFn    func(uuid.UUID) ([]*models.Shard, error)
	getByFileChunkFn func(uuid.UUID, int) ([]*models.Shard, error)
	updateStatusFn   func(uuid.UUID, models.ShardStatus) error
	deleteFn         func(uuid.UUID) error
}

func (m *mockShardRepo) Create(shard *models.Shard) error {
	if m.createFn != nil {
		return m.createFn(shard)
	}
	return nil
}
func (m *mockShardRepo) CreateBatch(shards []*models.Shard) error {
	if m.createBatchFn != nil {
		return m.createBatchFn(shards)
	}
	return nil
}
func (m *mockShardRepo) GetByID(id uuid.UUID) (*models.Shard, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(id)
	}
	return nil, errors.New("not implemented")
}
func (m *mockShardRepo) GetByFileID(fileID uuid.UUID) ([]*models.Shard, error) {
	if m.getByFileIDFn != nil {
		return m.getByFileIDFn(fileID)
	}
	return nil, nil
}
func (m *mockShardRepo) GetByFileAndChunk(fileID uuid.UUID, chunkIndex int) ([]*models.Shard, error) {
	if m.getByFileChunkFn != nil {
		return m.getByFileChunkFn(fileID, chunkIndex)
	}
	return nil, nil
}
func (m *mockShardRepo) UpdateStatus(id uuid.UUID, status models.ShardStatus) error {
	if m.updateStatusFn != nil {
		return m.updateStatusFn(id, status)
	}
	return nil
}
func (m *mockShardRepo) Delete(id uuid.UUID) error {
	if m.deleteFn != nil {
		return m.deleteFn(id)
	}
	return nil
}

func TestShardMapService_RegisterFileValidation(t *testing.T) {
	svc := service.NewShardMapService(&mockFileRepo{}, &mockShardRepo{})

	_, err := svc.RegisterFile(&dto.RegisterFileRequest{K: 3, N: 2})
	if err == nil {
		t.Fatalf("expected validation error when K > N")
	}
}

func TestShardMapService_RegisterFileSuccess(t *testing.T) {
	called := false
	svc := service.NewShardMapService(&mockFileRepo{
		createFn: func(file *models.File) error {
			called = true
			if file.Status != models.FileStatusPending {
				t.Fatalf("expected PENDING status, got %s", file.Status)
			}
			return nil
		},
	}, &mockShardRepo{})

	resp, err := svc.RegisterFile(&dto.RegisterFileRequest{
		OriginalName: "a.txt",
		OriginalSize: 100,
		TotalChunks:  1,
		N:            3,
		K:            2,
		ShardSize:    64,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatalf("expected file repo Create to be called")
	}
	if resp.FileID == "" || resp.Status != string(models.FileStatusPending) {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestShardMapService_RecordShardsValidation(t *testing.T) {
	svc := service.NewShardMapService(&mockFileRepo{}, &mockShardRepo{})

	_, err := svc.RecordShards(&dto.RecordShardsRequest{FileID: "not-a-uuid"})
	if err == nil {
		t.Fatalf("expected invalid file id error")
	}
}

func TestShardMapService_RecordShardsSuccess(t *testing.T) {
	fileID := uuid.New()
	updated := false
	batchCalled := false

	fileRepo := &mockFileRepo{
		getByIDFn: func(id uuid.UUID) (*models.File, error) {
			if id != fileID {
				t.Fatalf("unexpected fileID")
			}
			name := "file.bin"
			return &models.File{ID: fileID, OriginalName: &name, N: 2, K: 1}, nil
		},
		updateStatusFn: func(id uuid.UUID, status models.FileStatus) error {
			updated = true
			if status != models.FileStatusUploaded {
				t.Fatalf("expected uploaded status")
			}
			return nil
		},
	}
	shardRepo := &mockShardRepo{
		createBatchFn: func(shards []*models.Shard) error {
			batchCalled = true
			if len(shards) != 2 {
				t.Fatalf("expected 2 shards, got %d", len(shards))
			}
			return nil
		},
	}

	svc := service.NewShardMapService(fileRepo, shardRepo)
	resp, err := svc.RecordShards(&dto.RecordShardsRequest{
		FileID: fileID.String(),
		Shards: []dto.ShardInfo{
			{ChunkIndex: 0, ShardIndex: 0, Type: "DATA", RemoteID: "r1", Provider: "p1", ChecksumSHA256: "abc"},
			{ChunkIndex: 0, ShardIndex: 1, Type: "PARITY", RemoteID: "r2", Provider: "p2", ChecksumSHA256: "def"},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !batchCalled || !updated {
		t.Fatalf("expected batch create and status update to be called")
	}
	if len(resp.Shards) != 2 {
		t.Fatalf("unexpected response shard count: %d", len(resp.Shards))
	}
}

func TestShardMapService_GetAndUpdate(t *testing.T) {
	fileID := uuid.New()
	shardID := uuid.New()
	name := "n.txt"

	fileRepo := &mockFileRepo{
		getByIDFn: func(id uuid.UUID) (*models.File, error) {
			return &models.File{ID: fileID, OriginalName: &name, OriginalSize: 5, TotalChunks: 1, N: 2, K: 1, ShardSize: 3, Status: models.FileStatusUploaded}, nil
		},
	}
	shardRepo := &mockShardRepo{
		getByIDFn: func(id uuid.UUID) (*models.Shard, error) {
			return &models.Shard{ID: shardID, FileID: fileID, ChunkIndex: 0, ShardIndex: 0, ShardType: models.ShardTypeData, RemoteID: "r", Provider: "p", ChecksumSHA256: "x", Status: models.ShardStatusHealthy}, nil
		},
		getByFileIDFn: func(id uuid.UUID) ([]*models.Shard, error) {
			return []*models.Shard{{ID: shardID, FileID: fileID, ChunkIndex: 0, ShardIndex: 0, ShardType: models.ShardTypeData, RemoteID: "r", Provider: "p", ChecksumSHA256: "x", Status: models.ShardStatusHealthy}}, nil
		},
		updateStatusFn: func(id uuid.UUID, status models.ShardStatus) error {
			if status != models.ShardStatusCorrupted {
				t.Fatalf("unexpected status update %s", status)
			}
			return nil
		},
	}

	svc := service.NewShardMapService(fileRepo, shardRepo)

	mapResp, err := svc.GetShardMap(fileID)
	if err != nil || mapResp.FileID == "" || len(mapResp.Shards) != 1 {
		t.Fatalf("unexpected shard map response: %+v err=%v", mapResp, err)
	}

	shardResp, err := svc.GetShardByID(shardID)
	if err != nil || shardResp.ShardID == "" {
		t.Fatalf("unexpected shard response: %+v err=%v", shardResp, err)
	}

	if err := svc.MarkShardStatus(shardID, &dto.MarkShardStatusRequest{Status: "CORRUPTED"}); err != nil {
		t.Fatalf("unexpected mark status error: %v", err)
	}

	if err := svc.MarkShardStatus(shardID, &dto.MarkShardStatusRequest{Status: "BAD"}); err == nil {
		t.Fatalf("expected error for invalid status")
	}
}
