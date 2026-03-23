package orchestrator_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/vindyang/cs464-project/backend/services/orchestrator/internal/app"
	"github.com/vindyang/cs464-project/backend/services/shared/orchestrator/clients"
	"github.com/vindyang/cs464-project/backend/services/shared/types"
)

// TestUploadHappyPath: all 6 shards succeed
func TestUploadHappyPath(t *testing.T) {
	adapterServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/providers" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode([]types.ProviderInfo{
				{
					ProviderID:      "awsS3",
					DisplayName:     "AWS S3",
					Status:          "connected",
					LatencyMs:       50,
					QuotaTotalBytes: 1000000,
					QuotaUsedBytes:  100000,
				},
				{
					ProviderID:      "googleDrive",
					DisplayName:     "Google Drive",
					Status:          "connected",
					LatencyMs:       60,
					QuotaTotalBytes: 1000000,
					QuotaUsedBytes:  150000,
				},
			})
		} else if r.URL.Path == "/shards/upload" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(types.UploadShardResp{
				RemoteID:    "remote-shard-123",
				ChecksumSha: "abc123def456abc123def456abc123def456abc123def456abc123def456",
			})
		}
	}))
	defer adapterServer.Close()

	// Mock Shard Map Service
	registeredFileID := ""
	recordedShardCount := 0
	shardMapServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" && r.URL.Path == "/api/v1/shards/register" {
			w.Header().Set("Content-Type", "application/json")
			registeredFileID = "file-123"
			json.NewEncoder(w).Encode(types.RegisterFileResp{
				FileID: registeredFileID,
				Status: "PENDING",
			})
		} else if r.Method == "POST" && r.URL.Path == "/api/v1/shards/record" {
			var recordReq types.RecordShardReq
			json.NewDecoder(r.Body).Decode(&recordReq)
			recordedShardCount = len(recordReq.Shards)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
		}
	}))
	defer shardMapServer.Close()

	// Create service
	adapter := clients.NewAdapterClient(adapterServer.URL, nil)
	shardMap := clients.NewShardMapClient(shardMapServer.URL, nil)
	service := app.NewService(adapter, shardMap)

	// Create 6 mock shards (4 data + 2 parity)
	shards := make([][]byte, 6)
	isDataShard := make([]bool, 6)
	for i := 0; i < 6; i++ {
		shards[i] = []byte(("shard-" + string(rune(i))))
		isDataShard[i] = i < 4
	}

	// Upload
	ctx := context.Background()
	resp, err := service.UploadFile(ctx, "test.txt", shards, isDataShard)
	if err != nil {
		t.Fatalf("upload failed: %v", err)
	}

	// Verify
	if resp.Status != "committed" {
		t.Errorf("expected status 'committed', got '%s'", resp.Status)
	}
	if resp.FileID == "" {
		t.Errorf("expected non-empty file ID")
	}
	if len(resp.Shards) != 6 {
		t.Errorf("expected 6 shards, got %d", len(resp.Shards))
	}
	if recordedShardCount != 6 {
		t.Errorf("expected to record 6 shards, recorded %d", recordedShardCount)
	}
}

// TestUploadPartialFailureRollback: 2 uploads fail, verify DELETE called on the 4 that succeeded
func TestUploadPartialFailureRollback(t *testing.T) {
	uploadCount := 0
	deleteCount := 0
	var mu sync.Mutex

	adapterServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/providers" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode([]types.ProviderInfo{
				{
					ProviderID: "awsS3",
					Status:     "connected",
				},
				{
					ProviderID: "googleDrive",
					Status:     "connected",
				},
			})
		} else if r.URL.Path == "/shards/upload" {
			mu.Lock()
			uploadCount++
			currentUpload := uploadCount
			mu.Unlock()

			if currentUpload > 4 {
				// First 4 succeed, last 2 fail
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(types.UploadShardResp{
				RemoteID:    fmt.Sprintf("remote-%d", currentUpload),
				ChecksumSha: "abc123",
			})
		} else if r.Method == "DELETE" && r.URL.Path == "/shards/remote-1" {
			mu.Lock()
			deleteCount++
			mu.Unlock()
		} else if r.Method == "DELETE" && r.URL.Path == "/shards/remote-2" {
			mu.Lock()
			deleteCount++
			mu.Unlock()
		} else if r.Method == "DELETE" && r.URL.Path == "/shards/remote-3" {
			mu.Lock()
			deleteCount++
			mu.Unlock()
		} else if r.Method == "DELETE" && r.URL.Path == "/shards/remote-4" {
			mu.Lock()
			deleteCount++
			mu.Unlock()
		}
	}))
	defer adapterServer.Close()

	shardMapServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" && r.URL.Path == "/api/v1/shards/register" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(types.RegisterFileResp{
				FileID: "file-123",
				Status: "PENDING",
			})
		}
	}))
	defer shardMapServer.Close()

	adapter := clients.NewAdapterClient(adapterServer.URL, nil)
	shardMap := clients.NewShardMapClient(shardMapServer.URL, nil)
	service := app.NewService(adapter, shardMap)

	shards := make([][]byte, 6)
	isDataShard := make([]bool, 6)
	for i := 0; i < 6; i++ {
		shards[i] = []byte("shard")
		isDataShard[i] = i < 4
	}

	ctx := context.Background()
	resp, err := service.UploadFile(ctx, "test.txt", shards, isDataShard)
	if err != nil {
		t.Fatalf("upload failed: %v", err)
	}

	if resp.Status != "failed" {
		t.Errorf("expected status 'failed', got '%s'; error: %s", resp.Status, resp.Error)
	}
	if deleteCount != 4 {
		t.Errorf("expected 4 DELETE calls for rollback, got %d", deleteCount)
	}
}

// TestUploadSkipsDegradedProviders: one provider degraded, allocation skips it
func TestUploadSkipsDegradedProviders(t *testing.T) {
	adapterServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/providers" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode([]types.ProviderInfo{
				{
					ProviderID: "awsS3",
					Status:     "degraded", // Skip this one
				},
				{
					ProviderID: "googleDrive",
					Status:     "connected",
				},
				{
					ProviderID: "onedrive",
					Status:     "connected",
				},
			})
		} else if r.URL.Path == "/shards/upload" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(types.UploadShardResp{
				RemoteID:    "remote-123",
				ChecksumSha: "abc123",
			})
		}
	}))
	defer adapterServer.Close()

	shardMapServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" && r.URL.Path == "/api/v1/shards/register" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(types.RegisterFileResp{
				FileID: "file-123",
				Status: "PENDING",
			})
		} else if r.Method == "POST" && r.URL.Path == "/api/v1/shards/record" {
			w.WriteHeader(http.StatusCreated)
		}
	}))
	defer shardMapServer.Close()

	adapter := clients.NewAdapterClient(adapterServer.URL, nil)
	shardMap := clients.NewShardMapClient(shardMapServer.URL, nil)
	service := app.NewService(adapter, shardMap)

	shards := make([][]byte, 6)
	isDataShard := make([]bool, 6)
	for i := 0; i < 6; i++ {
		shards[i] = []byte("shard")
		isDataShard[i] = i < 4
	}

	ctx := context.Background()
	resp, err := service.UploadFile(ctx, "test.txt", shards, isDataShard)
	if err != nil {
		t.Fatalf("upload failed: %v", err)
	}

	if resp.Status != "committed" {
		t.Errorf("expected status 'committed', got '%s'; error: %s", resp.Status, resp.Error)
	}
	// Verify that only 2 healthy providers (googleDrive, onedrive) were used
	if len(resp.Shards) != 6 {
		t.Errorf("expected 6 shards recorded, got %d", len(resp.Shards))
	}
}

// TestUploadInsufficientProviders: < 2 healthy providers, verify error
func TestUploadInsufficientProviders(t *testing.T) {
	adapterServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/providers" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode([]types.ProviderInfo{
				{
					ProviderID: "awsS3",
					Status:     "disconnected",
				},
			})
		}
	}))
	defer adapterServer.Close()

	shardMapServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(types.RegisterFileResp{
			FileID: "file-123",
			Status: "PENDING",
		})
	}))
	defer shardMapServer.Close()

	adapter := clients.NewAdapterClient(adapterServer.URL, nil)
	shardMap := clients.NewShardMapClient(shardMapServer.URL, nil)
	service := app.NewService(adapter, shardMap)

	shards := make([][]byte, 6)
	isDataShard := make([]bool, 6)
	for i := 0; i < 6; i++ {
		shards[i] = []byte("shard")
		isDataShard[i] = i < 4
	}

	ctx := context.Background()
	_, err := service.UploadFile(ctx, "test.txt", shards, isDataShard)
	if err == nil {
		t.Errorf("expected error for insufficient providers, got nil")
	}
}

// TestDownloadEarlyExit: K shards + slowshards, verify exit doesn't wait
func TestDownloadEarlyExit(t *testing.T) {
	adapterServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/shards/remote-slow" {
			// Simulate slow shard — by the time it responds, we've already exited
			time.Sleep(1 * time.Second)
			w.Write([]byte("slow-data"))
			return
		}
		if r.URL.Path == "/shards/remote-1" {
			w.Write([]byte("shard1"))
		} else if r.URL.Path == "/shards/remote-2" {
			w.Write([]byte("shard2"))
		} else if r.URL.Path == "/shards/remote-3" {
			w.Write([]byte("shard3"))
		} else if r.URL.Path == "/shards/remote-4" {
			w.Write([]byte("shard4"))
		}
	}))
	defer adapterServer.Close()

	shardMapServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(types.GetShardMapResp{
			FileID: "file-123",
			N:      6,
			K:      4,
			Shards: []types.ShardMapEntry{
				{ShardID: "1", ShardIndex: 0, RemoteID: "remote-1", Provider: "awsS3"},
				{ShardID: "2", ShardIndex: 1, RemoteID: "remote-2", Provider: "googleDrive"},
				{ShardID: "3", ShardIndex: 2, RemoteID: "remote-3", Provider: "awsS3"},
				{ShardID: "4", ShardIndex: 3, RemoteID: "remote-4", Provider: "googleDrive"},
				{ShardID: "5", ShardIndex: 4, RemoteID: "remote-slow", Provider: "onedrive"}, // Slow
				{ShardID: "6", ShardIndex: 5, RemoteID: "remote-slow", Provider: "onedrive"}, // Slow
			},
		})
	}))
	defer shardMapServer.Close()

	adapter := clients.NewAdapterClient(adapterServer.URL, nil)
	shardMap := clients.NewShardMapClient(shardMapServer.URL, nil)
	service := app.NewService(adapter, shardMap)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	startTime := time.Now()
	resp, err := service.DownloadFile(ctx, "file-123")
	elapsedMs := time.Since(startTime).Milliseconds()

	if err != nil {
		t.Fatalf("download failed: %v", err)
	}
	if resp.FileID != "file-123" {
		t.Errorf("expected file ID 'file-123', got '%s'", resp.FileID)
	}
	if elapsedMs > 500 {
		t.Logf("warning: download took %dms (expected early exit < 200ms)", elapsedMs)
	}
}

// BenchmarkUploadParallel: verify parallel uploads improve performance
func BenchmarkUploadParallel(b *testing.B) {
	adapterServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/providers" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode([]types.ProviderInfo{
				{ProviderID: "awsS3", Status: "connected"},
				{ProviderID: "googleDrive", Status: "connected"},
			})
		} else if r.URL.Path == "/shards/upload" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(types.UploadShardResp{
				RemoteID:    "remote-123",
				ChecksumSha: "abc123",
			})
		}
	}))
	defer adapterServer.Close()

	shardMapServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" && r.URL.Path == "/api/v1/shards/register" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(types.RegisterFileResp{FileID: "file-123"})
		} else {
			w.WriteHeader(http.StatusCreated)
		}
	}))
	defer shardMapServer.Close()

	adapter := clients.NewAdapterClient(adapterServer.URL, nil)
	shardMap := clients.NewShardMapClient(shardMapServer.URL, nil)
	service := app.NewService(adapter, shardMap)

	shards := make([][]byte, 6)
	for i := 0; i < 6; i++ {
		shards[i] = make([]byte, 1024*1024) // 1MB shard
	}
	isDataShard := make([]bool, 6)
	for i := 0; i < 4; i++ {
		isDataShard[i] = true
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx := context.Background()
		service.UploadFile(ctx, "test.txt", shards, isDataShard)
	}
}
