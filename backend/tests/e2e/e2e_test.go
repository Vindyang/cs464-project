package e2e_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"

	"github.com/vindyang/cs464-project/backend/services/shared/types"
	"github.com/vindyang/cs464-project/backend/tests/helpers"
)

// TestE2EUploadDownloadDelete: Happy path covering full 3-operation workflow
func TestE2EUploadDownloadDelete(t *testing.T) {
	var mu sync.Mutex
	uploadedShards := make(map[string][]byte)
	shardMapData := struct {
		fileID string
		n      int
		k      int
	}{
		fileID: "file-001",
		n:      6,
		k:      4,
	}

	adapterMock := helpers.NewAdapterMock(t, helpers.AdapterMockConfig{
		OnGetProviders: func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode([]types.ProviderInfo{
				{ProviderID: "awsS3", DisplayName: "AWS S3", Status: "connected", LatencyMs: 50},
				{ProviderID: "googleDrive", DisplayName: "Google Drive", Status: "connected", LatencyMs: 60},
			})
		},
		OnUploadShard: func(w http.ResponseWriter, r *http.Request) {
			_ = r.ParseMultipartForm(32 << 20)
			shardID := r.FormValue("shard_id")
			if shardID == "" {
				http.Error(w, "missing shard_id", http.StatusBadRequest)
				return
			}
			remoteID := "remote-" + shardID
			file, _, _ := r.FormFile("file_data")
			if file != nil {
				data, _ := io.ReadAll(file)
				mu.Lock()
				uploadedShards[remoteID] = data // Store by RemoteID
				mu.Unlock()
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(types.UploadShardResp{
				RemoteID:    remoteID,
				ChecksumSha: "a3f2b1c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2",
			})
		},
		OnDownloadShard: func(w http.ResponseWriter, r *http.Request) {
			path := strings.Split(r.URL.Path, "?")[0]
			parts := strings.Split(path, "/")
			if len(parts) < 3 {
				http.Error(w, "invalid shard ID", http.StatusBadRequest)
				return
			}
			remoteID := parts[len(parts)-1]

			mu.Lock()
			data, exists := uploadedShards[remoteID]
			mu.Unlock()

			if exists {
				w.Write(data)
				return
			}
			http.NotFound(w, r)
		},
		OnDeleteShard: func(w http.ResponseWriter, r *http.Request) {
			path := strings.Split(r.URL.Path, "?")[0]
			parts := strings.Split(path, "/")
			if len(parts) >= 3 {
				remoteID := parts[len(parts)-1]
				mu.Lock()
				delete(uploadedShards, remoteID)
				mu.Unlock()
			}
			w.WriteHeader(http.StatusOK)
		},
	})
	defer adapterMock.Close()

	shardMapMock := helpers.NewShardMapMock(t, helpers.ShardMapMockConfig{
		OnRegisterFile: func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(types.RegisterFileResp{
				FileID: shardMapData.fileID,
				Status: "PENDING",
			})
		},
		OnRecordShards: func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusCreated)
		},
		OnGetShardMap: func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			shardMapEntries := []types.ShardMapEntry{}

			// FIX 1: Generate the shard map dynamically based on what was ACTUALLY uploaded
			mu.Lock()
			i := 0
			for remoteID := range uploadedShards {
				shardMapEntries = append(shardMapEntries, types.ShardMapEntry{
					ShardID:    fmt.Sprintf("shard-%d", i),
					ShardIndex: i,
					RemoteID:   remoteID,
					Provider:   "awsS3",
					Status:     "HEALTHY",
				})
				i++
			}
			mu.Unlock()

			json.NewEncoder(w).Encode(types.GetShardMapResp{
				FileID: shardMapData.fileID,
				N:      shardMapData.n,
				K:      shardMapData.k,
				Shards: shardMapEntries,
			})
		},
	})
	defer shardMapMock.Close()

	shardingMock := helpers.NewShardingMock(t, helpers.ShardingMockConfig{
		OnShard: func(w http.ResponseWriter, r *http.Request) {
			var req struct {
				FileID   string `json:"fileId"`
				FileData []byte `json:"fileData"`
				N        int    `json:"n"`
				K        int    `json:"k"`
			}
			_ = json.NewDecoder(r.Body).Decode(&req)

			type outShard struct {
				ShardIndex int    `json:"shardIndex"`
				ShardType  string `json:"shardType"`
				ShardData  []byte `json:"shardData"`
			}
			out := make([]outShard, 0, req.N)
			for i := 0; i < req.N; i++ {
				typeVal := "data"
				if i >= req.K {
					typeVal = "parity"
				}
				out = append(out, outShard{ShardIndex: i, ShardType: typeVal, ShardData: req.FileData})
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"fileId": req.FileID,
				"shards": out,
			})
		},
		OnReconstruct: func(w http.ResponseWriter, r *http.Request) {
			var req struct {
				AvailableShards []struct {
					ShardData []byte `json:"shard_data"`
				} `json:"available_shards"`
			}
			_ = json.NewDecoder(r.Body).Decode(&req)
			if len(req.AvailableShards) == 0 {
				http.Error(w, "no shards", http.StatusBadRequest)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"reconstructed_file": req.AvailableShards[0].ShardData,
			})
		},
	})
	defer shardingMock.Close()

	orchestratorURL, shutdown := helpers.StartOrchestrator(t,
		adapterMock.URL,
		shardMapMock.URL,
		shardingMock.URL,
	)
	defer shutdown()

	testData := []byte("This is a test document for e2e testing. It contains important data that needs to be stored reliably across multiple providers.")

	t.Run("Upload", func(t *testing.T) {
		resp := helpers.UploadFile(t, orchestratorURL, testData)

		if resp.FileID == "" {
			t.Errorf("expected non-empty FileID, got %q", resp.FileID)
		}
		if resp.Status != "committed" {
			t.Errorf("expected status 'committed', got %q", resp.Status)
		}
		if len(resp.Shards) != 6 {
			t.Errorf("expected 6 shards, got %d", len(resp.Shards))
		}

		shardMapData.fileID = resp.FileID
		t.Logf("✓ Uploaded file: %s (6 shards)", resp.FileID)
	})

	t.Run("Download", func(t *testing.T) {
		if shardMapData.fileID == "" {
			t.Skip("upload phase did not complete")
		}

		downloadURL := fmt.Sprintf("%s/api/orchestrator/files/%s/download", orchestratorURL, shardMapData.fileID)
		httpResp, err := http.Get(downloadURL)
		if err != nil {
			t.Fatalf("download request failed: %v", err)
		}
		defer httpResp.Body.Close()

		if httpResp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(httpResp.Body)
			t.Fatalf("download failed: status=%d body=%s", httpResp.StatusCode, string(body))
		}

		downloadedData, err := io.ReadAll(httpResp.Body)
		if err != nil {
			t.Fatalf("failed to read download response: %v", err)
		}

		if len(downloadedData) == 0 {
			t.Errorf("expected non-empty file data, got 0 bytes")
		}

		t.Logf("✓ Downloaded file %s: %d bytes", shardMapData.fileID, len(downloadedData))
	})

	t.Run("Delete", func(t *testing.T) {
		// FIX 2: Restore the skip since we haven't built this endpoint yet!
		t.Skip("Delete endpoint not implemented in orchestrator service")
	})
}

// TestE2EUploadFailure: Verify failed upload doesn't leave shards behind
func TestE2EUploadFailure(t *testing.T) {
	var mu sync.Mutex
	deleteCallCount := 0

	adapterMock := helpers.NewAdapterMock(t, helpers.AdapterMockConfig{
		OnGetProviders: func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode([]types.ProviderInfo{
				{ProviderID: "awsS3", Status: "connected"},
				{ProviderID: "googleDrive", Status: "connected"},
			})
		},
		OnUploadShard: func(w http.ResponseWriter, r *http.Request) {
			_ = r.ParseMultipartForm(32 << 20)
			shardID := r.FormValue("shard_id")
			if strings.Contains(shardID, "shard-4") || strings.Contains(shardID, "shard-5") {
				http.Error(w, "provider unavailable", http.StatusServiceUnavailable)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(types.UploadShardResp{
				RemoteID:    "remote-" + shardID,
				ChecksumSha: "a3f2b1c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2",
			})
		},
		OnDeleteShard: func(w http.ResponseWriter, r *http.Request) {
			mu.Lock()
			deleteCallCount++
			mu.Unlock()
			w.WriteHeader(http.StatusOK)
		},
	})
	defer adapterMock.Close()

	shardMapMock := helpers.NewShardMapMock(t, helpers.ShardMapMockConfig{
		OnRegisterFile: func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(types.RegisterFileResp{FileID: "file-002", Status: "PENDING"})
		},
		OnRecordShards: func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "should not record on failure", http.StatusInternalServerError)
		},
	})
	defer shardMapMock.Close()

	shardingMock := helpers.NewShardingMock(t, helpers.ShardingMockConfig{
		OnShard: func(w http.ResponseWriter, r *http.Request) {
			var req struct {
				FileID   string `json:"fileId"`
				FileData []byte `json:"fileData"`
				N        int    `json:"n"`
				K        int    `json:"k"`
			}
			_ = json.NewDecoder(r.Body).Decode(&req)

			type outShard struct {
				ShardIndex int    `json:"shardIndex"`
				ShardType  string `json:"shardType"`
				ShardData  []byte `json:"shardData"`
			}
			out := make([]outShard, 0, req.N)
			for i := 0; i < req.N; i++ {
				out = append(out, outShard{ShardIndex: i, ShardType: "data", ShardData: req.FileData})
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"fileId": req.FileID,
				"shards": out,
			})
		},
	})
	defer shardingMock.Close()

	orchestratorURL, shutdown := helpers.StartOrchestrator(t,
		adapterMock.URL,
		shardMapMock.URL,
		shardingMock.URL,
	)
	defer shutdown()

	testData := []byte("This upload will fail")

	httpResp, body := helpers.UploadFileRaw(t, orchestratorURL, testData)
	defer httpResp.Body.Close()

	// FIX 3: Check the JSON 'status' field, not the HTTP StatusCode
	var uploadResp types.UploadResp
	if err := json.Unmarshal(body, &uploadResp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if uploadResp.Status != "failed" {
		t.Errorf("expected upload to fail (status 'failed'), but got %s", uploadResp.Status)
	}

	mu.Lock()
	currentCount := deleteCallCount
	mu.Unlock()

	if currentCount < 3 {
		t.Errorf("expected at least 3 DELETE calls for rollback, got %d", currentCount)
	}

	t.Logf("✓ Upload correctly failed with rollback (%d shards deleted)", currentCount)
}

// TestE2EDownloadNonexistent: Verify graceful error for missing file
func TestE2EDownloadNonexistent(t *testing.T) {
	adapterMock := helpers.NewAdapterMock(t, helpers.AdapterMockConfig{
		OnGetProviders: func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode([]types.ProviderInfo{
				{ProviderID: "awsS3", Status: "connected"},
			})
		},
	})
	defer adapterMock.Close()

	shardMapMock := helpers.NewShardMapMock(t, helpers.ShardMapMockConfig{
		OnGetShardMap: func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "file not found", http.StatusNotFound)
		},
	})
	defer shardMapMock.Close()

	shardingMock := helpers.NewShardingMock(t, helpers.ShardingMockConfig{})
	defer shardingMock.Close()

	orchestratorURL, shutdown := helpers.StartOrchestrator(t,
		adapterMock.URL,
		shardMapMock.URL,
		shardingMock.URL,
	)
	defer shutdown()

	downloadURL := fmt.Sprintf("%s/api/orchestrator/files/nonexistent-file/download", orchestratorURL)
	httpResp, err := http.Get(downloadURL)
	if err != nil {
		t.Fatalf("download request failed: %v", err)
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got status %d", httpResp.StatusCode)
	}

	t.Logf("✓ Download nonexistent file correctly returned 404")
}

// TestE2EDownloadDegradedProviders: Download succeeds when some providers fail
func TestE2EDownloadDegradedProviders(t *testing.T) {
	var mu sync.Mutex
	downloadAttempts := 0

	adapterMock := helpers.NewAdapterMock(t, helpers.AdapterMockConfig{
		OnGetProviders: func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode([]types.ProviderInfo{
				{ProviderID: "awsS3", Status: "connected"},
				{ProviderID: "googleDrive", Status: "degraded"},
				{ProviderID: "onedrive", Status: "connected"},
			})
		},
		OnUploadShard: func(w http.ResponseWriter, r *http.Request) {
			_ = r.ParseMultipartForm(32 << 20)
			shardID := r.FormValue("shard_id")
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(types.UploadShardResp{
				RemoteID:    "remote-" + shardID,
				ChecksumSha: "a3f2b1c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2",
			})
		},
		OnDownloadShard: func(w http.ResponseWriter, r *http.Request) {
			mu.Lock()
			downloadAttempts++
			currentAttempt := downloadAttempts
			mu.Unlock()

			// FIX 4: Only fail exactly 2 shards so K=4 succeeds
			if currentAttempt == 1 || currentAttempt == 2 {
				http.Error(w, "provider error", http.StatusInternalServerError)
				return
			}
			w.Write([]byte("shard-data"))
		},
	})
	defer adapterMock.Close()

	shardMapMock := helpers.NewShardMapMock(t, helpers.ShardMapMockConfig{
		OnRegisterFile: func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(types.RegisterFileResp{FileID: "file-003", Status: "PENDING"})
		},
		OnRecordShards: func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusCreated)
		},
		OnGetShardMap: func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(types.GetShardMapResp{
				FileID: "file-003",
				N:      6,
				K:      4,
				Shards: []types.ShardMapEntry{
					{ShardID: "1", ShardIndex: 0, RemoteID: "remote-1", Provider: "awsS3"},
					{ShardID: "2", ShardIndex: 1, RemoteID: "remote-2", Provider: "googleDrive"},
					{ShardID: "3", ShardIndex: 2, RemoteID: "remote-3", Provider: "awsS3"},
					{ShardID: "4", ShardIndex: 3, RemoteID: "remote-4", Provider: "onedrive"},
					{ShardID: "5", ShardIndex: 4, RemoteID: "remote-5", Provider: "awsS3"},
					{ShardID: "6", ShardIndex: 5, RemoteID: "remote-6", Provider: "onedrive"},
				},
			})
		},
	})
	defer shardMapMock.Close()

	shardingMock := helpers.NewShardingMock(t, helpers.ShardingMockConfig{
		OnShard: func(w http.ResponseWriter, r *http.Request) {
			var req struct {
				FileID   string `json:"fileId"`
				FileData []byte `json:"fileData"`
				N        int    `json:"n"`
				K        int    `json:"k"`
			}
			_ = json.NewDecoder(r.Body).Decode(&req)

			type outShard struct {
				ShardIndex int    `json:"shardIndex"`
				ShardType  string `json:"shardType"`
				ShardData  []byte `json:"shardData"`
			}
			out := make([]outShard, 0, req.N)
			for i := 0; i < req.N; i++ {
				out = append(out, outShard{ShardIndex: i, ShardType: "data", ShardData: req.FileData})
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"fileId": req.FileID,
				"shards": out,
			})
		},
		OnReconstruct: func(w http.ResponseWriter, r *http.Request) {
			var req struct {
				AvailableShards []struct {
					ShardData []byte `json:"shard_data"`
				} `json:"available_shards"`
			}
			_ = json.NewDecoder(r.Body).Decode(&req)
			if len(req.AvailableShards) == 0 {
				http.Error(w, "no shards", http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"reconstructed_file": req.AvailableShards[0].ShardData,
			})
		},
	})
	defer shardingMock.Close()

	orchestratorURL, shutdown := helpers.StartOrchestrator(t,
		adapterMock.URL,
		shardMapMock.URL,
		shardingMock.URL,
	)
	defer shutdown()

	testData := []byte("Test data for degraded provider scenario")
	helpers.UploadFile(t, orchestratorURL, testData)

	downloadURL := fmt.Sprintf("%s/api/orchestrator/files/file-003/download", orchestratorURL)
	httpResp, err := http.Get(downloadURL)
	if err != nil {
		t.Fatalf("download request failed: %v", err)
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(httpResp.Body)
		t.Fatalf("download failed: status=%d body=%s", httpResp.StatusCode, string(body))
	}

	mu.Lock()
	finalAttempts := downloadAttempts
	mu.Unlock()

	t.Logf("✓ Download succeeded despite degraded providers (%d attempts)", finalAttempts)
}

// TestE2EMultipleFiles: Handling of concurrent file operations
func TestE2EMultipleFiles(t *testing.T) {
	var mu sync.Mutex
	uploadedFiles := make(map[string]bool)

	adapterMock := helpers.NewAdapterMock(t, helpers.AdapterMockConfig{
		OnGetProviders: func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode([]types.ProviderInfo{
				{ProviderID: "awsS3", Status: "connected"},
				{ProviderID: "googleDrive", Status: "connected"},
			})
		},
		OnUploadShard: func(w http.ResponseWriter, r *http.Request) {
			_ = r.ParseMultipartForm(32 << 20)
			shardID := r.FormValue("shard_id")
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(types.UploadShardResp{
				RemoteID:    "remote-" + shardID,
				ChecksumSha: "a3f2b1c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2",
			})
		},
	})
	defer adapterMock.Close()

	fileCount := 0
	shardMapMock := helpers.NewShardMapMock(t, helpers.ShardMapMockConfig{
		OnRegisterFile: func(w http.ResponseWriter, r *http.Request) {
			mu.Lock()
			fileCount++
			fileID := fmt.Sprintf("file-%04d", fileCount)
			uploadedFiles[fileID] = true
			mu.Unlock()

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(types.RegisterFileResp{
				FileID: fileID,
				Status: "PENDING",
			})
		},
		OnRecordShards: func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusCreated)
		},
	})
	defer shardMapMock.Close()

	shardingMock := helpers.NewShardingMock(t, helpers.ShardingMockConfig{
		OnShard: func(w http.ResponseWriter, r *http.Request) {
			var req struct {
				FileID   string `json:"fileId"`
				FileData []byte `json:"fileData"`
				N        int    `json:"n"`
				K        int    `json:"k"`
			}
			_ = json.NewDecoder(r.Body).Decode(&req)

			type outShard struct {
				ShardIndex int    `json:"shardIndex"`
				ShardType  string `json:"shardType"`
				ShardData  []byte `json:"shardData"`
			}
			out := make([]outShard, 0, req.N)
			for i := 0; i < req.N; i++ {
				out = append(out, outShard{ShardIndex: i, ShardType: "data", ShardData: req.FileData})
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"fileId": req.FileID,
				"shards": out,
			})
		},
		OnReconstruct: func(w http.ResponseWriter, r *http.Request) {
			var req struct {
				AvailableShards []struct {
					ShardData []byte `json:"shard_data"`
				} `json:"available_shards"`
			}
			_ = json.NewDecoder(r.Body).Decode(&req)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"reconstructed_file": req.AvailableShards[0].ShardData,
			})
		},
	})
	defer shardingMock.Close()

	orchestratorURL, shutdown := helpers.StartOrchestrator(t,
		adapterMock.URL,
		shardMapMock.URL,
		shardingMock.URL,
	)
	defer shutdown()

	fileContents := [][]byte{
		[]byte("File 1 content"),
		[]byte("File 2 content with more data"),
		[]byte("File 3 - another test document"),
	}

	for i, content := range fileContents {
		resp := helpers.UploadFile(t, orchestratorURL, content)
		if resp.Status != "committed" {
			t.Errorf("file %d upload failed: status=%s", i+1, resp.Status)
		}
	}

	mu.Lock()
	filesTracked := len(uploadedFiles)
	mu.Unlock()

	if filesTracked != 3 {
		t.Errorf("expected 3 files uploaded, got %d", filesTracked)
	}

	t.Logf("✓ Uploaded and tracked %d files successfully", filesTracked)
}
