package service_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"testing"

	"github.com/google/uuid"
	"github.com/vindyang/cs464-project/backend/services/adapter/internal/app"
	"github.com/vindyang/cs464-project/backend/services/shared/adapter"
	"github.com/vindyang/cs464-project/backend/services/shared/api/dto"
)

type fileOpsMockSharding struct {
	chunkFileFn         func(data []byte, chunkSize int64) ([][]byte, error)
	encodeChunkFn       func(chunkData []byte, k, n int) ([][]byte, error)
	decodeChunkFn       func(shards [][]byte, k, n int) ([]byte, error)
	calculateChecksumFn func(data []byte) string
}

func (m *fileOpsMockSharding) ChunkFile(data []byte, chunkSize int64) ([][]byte, error) {
	if m.chunkFileFn != nil {
		return m.chunkFileFn(data, chunkSize)
	}
	return nil, fmt.Errorf("not implemented")
}

func (m *fileOpsMockSharding) EncodeChunk(chunkData []byte, k, n int) ([][]byte, error) {
	if m.encodeChunkFn != nil {
		return m.encodeChunkFn(chunkData, k, n)
	}
	return nil, fmt.Errorf("not implemented")
}

func (m *fileOpsMockSharding) DecodeChunk(shards [][]byte, k, n int) ([]byte, error) {
	if m.decodeChunkFn != nil {
		return m.decodeChunkFn(shards, k, n)
	}
	return nil, fmt.Errorf("not implemented")
}

func (m *fileOpsMockSharding) CalculateChecksum(data []byte) string {
	if m.calculateChecksumFn != nil {
		return m.calculateChecksumFn(data)
	}
	return ""
}

type fileOpsMockShardMap struct {
	registerFileFn    func(req *dto.RegisterFileRequest) (*dto.RegisterFileResponse, error)
	recordShardsFn    func(req *dto.RecordShardsRequest) (*dto.RecordShardsResponse, error)
	getShardMapFn     func(fileID uuid.UUID) (*dto.GetShardMapResponse, error)
	getShardByIDFn    func(shardID uuid.UUID) (*dto.ShardInfo, error)
	markShardStatusFn func(shardID uuid.UUID, req *dto.MarkShardStatusRequest) error
}

func (m *fileOpsMockShardMap) RegisterFile(req *dto.RegisterFileRequest) (*dto.RegisterFileResponse, error) {
	if m.registerFileFn != nil {
		return m.registerFileFn(req)
	}
	return nil, fmt.Errorf("not implemented")
}

func (m *fileOpsMockShardMap) RecordShards(req *dto.RecordShardsRequest) (*dto.RecordShardsResponse, error) {
	if m.recordShardsFn != nil {
		return m.recordShardsFn(req)
	}
	return nil, fmt.Errorf("not implemented")
}

func (m *fileOpsMockShardMap) GetShardMap(fileID uuid.UUID) (*dto.GetShardMapResponse, error) {
	if m.getShardMapFn != nil {
		return m.getShardMapFn(fileID)
	}
	return nil, fmt.Errorf("not implemented")
}

func (m *fileOpsMockShardMap) GetShardByID(shardID uuid.UUID) (*dto.ShardInfo, error) {
	if m.getShardByIDFn != nil {
		return m.getShardByIDFn(shardID)
	}
	return nil, fmt.Errorf("not implemented")
}

func (m *fileOpsMockShardMap) MarkShardStatus(shardID uuid.UUID, req *dto.MarkShardStatusRequest) error {
	if m.markShardStatusFn != nil {
		return m.markShardStatusFn(shardID, req)
	}
	return nil
}

type fileOpsMockProvider struct {
	uploadShardFn   func(ctx context.Context, fileID string, index int, data io.Reader) (string, error)
	downloadShardFn func(ctx context.Context, remoteID string) (io.ReadCloser, error)
	deleteShardFn   func(ctx context.Context, remoteID string) error
}

func (m *fileOpsMockProvider) GetMetadata(ctx context.Context) (*adapter.ProviderMetadata, error) {
	return &adapter.ProviderMetadata{ProviderID: "mock"}, nil
}

func (m *fileOpsMockProvider) UploadShard(ctx context.Context, fileID string, index int, data io.Reader) (string, error) {
	if m.uploadShardFn != nil {
		return m.uploadShardFn(ctx, fileID, index, data)
	}
	return "", fmt.Errorf("not implemented")
}

func (m *fileOpsMockProvider) DownloadShard(ctx context.Context, remoteID string) (io.ReadCloser, error) {
	if m.downloadShardFn != nil {
		return m.downloadShardFn(ctx, remoteID)
	}
	return nil, fmt.Errorf("not implemented")
}

func (m *fileOpsMockProvider) DeleteShard(ctx context.Context, remoteID string) error {
	if m.deleteShardFn != nil {
		return m.deleteShardFn(ctx, remoteID)
	}
	return nil
}

func (m *fileOpsMockProvider) HealthCheck(ctx context.Context) error { return nil }

func TestFileOperationsService_UploadFile_Success(t *testing.T) {
	ctx := context.Background()
	fileID := uuid.New().String()

	sharding := &fileOpsMockSharding{
		chunkFileFn: func(data []byte, chunkSize int64) ([][]byte, error) {
			return [][]byte{data}, nil
		},
		encodeChunkFn: func(chunkData []byte, k, n int) ([][]byte, error) {
			return [][]byte{[]byte("d0"), []byte("p0")}, nil
		},
		calculateChecksumFn: func(data []byte) string { return "sum" },
	}

	var recorded *dto.RecordShardsRequest
	shardMap := &fileOpsMockShardMap{
		registerFileFn: func(req *dto.RegisterFileRequest) (*dto.RegisterFileResponse, error) {
			return &dto.RegisterFileResponse{FileID: fileID, Status: "PENDING"}, nil
		},
		recordShardsFn: func(req *dto.RecordShardsRequest) (*dto.RecordShardsResponse, error) {
			recorded = req
			return &dto.RecordShardsResponse{FileID: req.FileID, Shards: req.Shards}, nil
		},
	}

	provider := &fileOpsMockProvider{
		uploadShardFn: func(ctx context.Context, fileID string, index int, data io.Reader) (string, error) {
			_, _ = io.ReadAll(data)
			return fmt.Sprintf("remote-%d", index), nil
		},
	}

	reg := adapter.NewRegistry()
	reg.Register("p1", provider)

	svc := app.NewFileOperationsService(sharding, shardMap, reg)
	resp, err := svc.UploadFile(ctx, "a.txt", []byte("abc"), 1, 2, 1, []string{"p1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.FileID != fileID {
		t.Fatalf("unexpected file id: %s", resp.FileID)
	}
	if resp.UploadStats.SuccessfulShards != 2 {
		t.Fatalf("expected 2 successful shards, got %d", resp.UploadStats.SuccessfulShards)
	}
	if recorded == nil || len(recorded.Shards) != 2 {
		t.Fatalf("expected 2 recorded shards")
	}
}

func TestFileOperationsService_UploadFile_Validation(t *testing.T) {
	ctx := context.Background()
	svc := app.NewFileOperationsService(&fileOpsMockSharding{}, &fileOpsMockShardMap{}, adapter.NewRegistry())

	if _, err := svc.UploadFile(ctx, "a.txt", []byte{}, 1, 2, 1, []string{"p1"}); err == nil {
		t.Fatalf("expected error for empty data")
	}
	if _, err := svc.UploadFile(ctx, "a.txt", []byte("x"), 2, 1, 1, []string{"p1"}); err == nil {
		t.Fatalf("expected error for invalid k,n")
	}
	if _, err := svc.UploadFile(ctx, "a.txt", []byte("x"), 1, 2, 0, []string{"p1"}); err == nil {
		t.Fatalf("expected error for chunk size")
	}
	if _, err := svc.UploadFile(ctx, "a.txt", []byte("x"), 1, 2, 1, nil); err == nil {
		t.Fatalf("expected error for providers")
	}
	if _, err := svc.UploadFile(ctx, "a.txt", []byte("x"), 1, 2, 1, []string{"missing"}); err == nil {
		t.Fatalf("expected error for missing provider")
	}
}

func TestFileOperationsService_DownloadFile_Success(t *testing.T) {
	ctx := context.Background()
	fileID := uuid.New()
	data := []byte("hello")

	sharding := &fileOpsMockSharding{
		calculateChecksumFn: func(d []byte) string { return "ok" },
		decodeChunkFn: func(shards [][]byte, k, n int) ([]byte, error) {
			return data, nil
		},
	}

	shardMap := &fileOpsMockShardMap{
		getShardMapFn: func(id uuid.UUID) (*dto.GetShardMapResponse, error) {
			return &dto.GetShardMapResponse{
				FileID:       id.String(),
				OriginalName: "a.txt",
				OriginalSize: int64(len(data)),
				TotalChunks:  1,
				N:            2,
				K:            1,
				Shards:       []dto.ShardInfo{{ChunkIndex: 0, ShardIndex: 0, Provider: "p1", RemoteID: "r1", ChecksumSHA256: "ok"}},
			}, nil
		},
	}

	provider := &fileOpsMockProvider{
		downloadShardFn: func(ctx context.Context, remoteID string) (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewReader(data)), nil
		},
	}
	reg := adapter.NewRegistry()
	reg.Register("p1", provider)

	svc := app.NewFileOperationsService(sharding, shardMap, reg)
	got, name, err := svc.DownloadFile(ctx, fileID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "a.txt" {
		t.Fatalf("unexpected filename: %s", name)
	}
	if !bytes.Equal(got, data) {
		t.Fatalf("unexpected data: %q", got)
	}
}

func TestFileOperationsService_DownloadFile_InsufficientShards(t *testing.T) {
	ctx := context.Background()
	fileID := uuid.New()

	sharding := &fileOpsMockSharding{calculateChecksumFn: func(d []byte) string { return "ok" }}
	shardMap := &fileOpsMockShardMap{
		getShardMapFn: func(id uuid.UUID) (*dto.GetShardMapResponse, error) {
			return &dto.GetShardMapResponse{FileID: id.String(), TotalChunks: 1, N: 2, K: 2, Shards: []dto.ShardInfo{{ChunkIndex: 0, ShardIndex: 0, Provider: "p1", RemoteID: "r1", ChecksumSHA256: "ok"}}}, nil
		},
	}

	reg := adapter.NewRegistry()
	reg.Register("p1", &fileOpsMockProvider{downloadShardFn: func(ctx context.Context, remoteID string) (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader([]byte("x"))), nil
	}})

	svc := app.NewFileOperationsService(sharding, shardMap, reg)
	if _, _, err := svc.DownloadFile(ctx, fileID); err == nil {
		t.Fatalf("expected insufficient shard error")
	}
}

func TestFileOperationsService_GetFileMetadata_And_Delete(t *testing.T) {
	ctx := context.Background()
	fileID := uuid.New()
	deleted := 0

	shardMap := &fileOpsMockShardMap{
		getShardMapFn: func(id uuid.UUID) (*dto.GetShardMapResponse, error) {
			return &dto.GetShardMapResponse{
				FileID:       id.String(),
				OriginalName: "a.txt",
				OriginalSize: 10,
				TotalChunks:  2,
				N:            2,
				K:            1,
				ShardSize:    5,
				Status:       "UPLOADED",
				Shards: []dto.ShardInfo{
					{ChunkIndex: 0, ShardIndex: 0, Status: "HEALTHY", Provider: "p1", RemoteID: "r1"},
					{ChunkIndex: 0, ShardIndex: 1, Status: "MISSING", Provider: "p1", RemoteID: "r2"},
					{ChunkIndex: 1, ShardIndex: 0, Status: "CORRUPTED", Provider: "p1", RemoteID: "r3"},
					{ChunkIndex: 1, ShardIndex: 1, Status: "HEALTHY", Provider: "p1", RemoteID: "r4"},
				},
			}, nil
		},
	}

	provider := &fileOpsMockProvider{
		deleteShardFn: func(ctx context.Context, remoteID string) error {
			deleted++
			return nil
		},
	}

	reg := adapter.NewRegistry()
	reg.Register("p1", provider)

	svc := app.NewFileOperationsService(&fileOpsMockSharding{}, shardMap, reg)

	meta, err := svc.GetFileMetadata(ctx, fileID)
	if err != nil {
		t.Fatalf("unexpected metadata error: %v", err)
	}
	if meta.HealthStatus.HealthyShards != 2 || meta.HealthStatus.CorruptedShards != 1 || meta.HealthStatus.MissingShards != 1 {
		t.Fatalf("unexpected health counts: %+v", meta.HealthStatus)
	}

	if err := svc.DeleteFile(ctx, fileID, true); err != nil {
		t.Fatalf("unexpected delete error: %v", err)
	}
	if deleted != 4 {
		t.Fatalf("expected 4 shard delete attempts, got %d", deleted)
	}
}
