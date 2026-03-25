package integration_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"testing"

	"github.com/vindyang/cs464-project/backend/services/shared/types"
)

// TestOrchestratorUploadFailsWhenShardMapRegisterErrors verifies that orchestrator
// returns a stable 500 JSON error response when shard-map file registration fails.
// This protects the upload contract for dependency failure propagation.
func TestOrchestratorUploadFailsWhenShardMapRegisterErrors(t *testing.T) {
	t.Parallel()

	adapterServer := newAdapterMock(t, adapterMockConfig{
		OnGetProviders: func(w http.ResponseWriter, r *http.Request) {
			_ = json.NewEncoder(w).Encode([]map[string]any{
				{"providerId": "provider-a", "displayName": "Provider A", "status": "connected"},
				{"providerId": "provider-b", "displayName": "Provider B", "status": "connected"},
			})
		},
		OnUploadShard: func(w http.ResponseWriter, r *http.Request) {
			_ = json.NewEncoder(w).Encode(types.UploadShardResp{RemoteID: "unused", ChecksumSha: "unused"})
		},
	})
	defer adapterServer.Close()

	shardMapServer := newShardMapMock(t, shardMapMockConfig{
		OnRegisterFile: func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "register unavailable", http.StatusInternalServerError)
		},
	})
	defer shardMapServer.Close()

	shardingServer := newShardingMock(t, shardingMockConfig{
		OnShard: func(w http.ResponseWriter, r *http.Request) {
			shards := make([]map[string]any, 0, 6)
			for i := 0; i < 6; i++ {
				shards = append(shards, map[string]any{"shardIndex": i, "shardType": "data", "shardData": []byte("ok")})
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"shards": shards})
		},
	})
	defer shardingServer.Close()

	orchestratorURL, shutdown := startOrchestrator(t, adapterServer.URL, shardMapServer.URL, shardingServer.URL)
	defer shutdown()

	httpResp, body := uploadFileRaw(t, orchestratorURL, []byte("should-fail"))
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusInternalServerError, httpResp.StatusCode, string(body))
	}

	payload := map[string]string{}
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("failed to parse error response: %v body=%s", err, string(body))
	}
	if payload["error"] != "Failed to upload file" {
		t.Fatalf("unexpected error message: %q", payload["error"])
	}
	if !strings.Contains(payload["details"], "failed to register file") {
		t.Fatalf("expected details to contain register failure, got: %q", payload["details"])
	}
}

// TestOrchestratorUploadFailsWhenShardingPayloadIsMalformed verifies that orchestrator
// fails upload with a 500 JSON error when sharding returns malformed JSON.
// This guards decoder/contract regressions on /api/sharding/shard.
func TestOrchestratorUploadFailsWhenShardingPayloadIsMalformed(t *testing.T) {
	t.Parallel()

	adapterServer := newAdapterMock(t, adapterMockConfig{
		OnGetProviders: func(w http.ResponseWriter, r *http.Request) {
			_ = json.NewEncoder(w).Encode([]map[string]any{})
		},
	})
	defer adapterServer.Close()

	shardMapServer := newShardMapMock(t, shardMapMockConfig{})
	defer shardMapServer.Close()

	shardingServer := newShardingMock(t, shardingMockConfig{
		OnShard: func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"shards":[{"shardData":`))
		},
	})
	defer shardingServer.Close()

	orchestratorURL, shutdown := startOrchestrator(t, adapterServer.URL, shardMapServer.URL, shardingServer.URL)
	defer shutdown()

	httpResp, body := uploadFileRaw(t, orchestratorURL, []byte("malformed-sharding-response"))
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusInternalServerError, httpResp.StatusCode, string(body))
	}

	payload := map[string]string{}
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("failed to parse error response: %v body=%s", err, string(body))
	}
	if payload["error"] != "Failed to upload file" {
		t.Fatalf("unexpected error message: %q", payload["error"])
	}
	if !strings.Contains(payload["details"], "failed to shard file data") {
		t.Fatalf("expected details to contain shard failure, got: %q", payload["details"])
	}
}

// TestOrchestratorUploadRollsBackSuccessfulShardsOnPartialUploadFailure verifies
// that orchestrator attempts cleanup for all successfully uploaded shards when any
// shard upload fails in the same request.
func TestOrchestratorUploadRollsBackSuccessfulShardsOnPartialUploadFailure(t *testing.T) {
	t.Parallel()

	adapterState := struct {
		sync.Mutex
		nextID      int
		uploaded    map[string]struct{}
		deleted     map[string]struct{}
		uploadCalls int
		deleteCalls int
	}{
		uploaded: map[string]struct{}{},
		deleted:  map[string]struct{}{},
	}

	adapterServer := newAdapterMock(t, adapterMockConfig{
		OnGetProviders: func(w http.ResponseWriter, r *http.Request) {
			_ = json.NewEncoder(w).Encode([]map[string]any{
				{"providerId": "provider-a", "displayName": "Provider A", "status": "connected"},
				{"providerId": "provider-b", "displayName": "Provider B", "status": "connected"},
			})
		},
		OnUploadShard: func(w http.ResponseWriter, r *http.Request) {
			if err := r.ParseMultipartForm(8 << 20); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			shardID := r.FormValue("shard_id")
			provider := r.FormValue("provider")

			adapterState.Lock()
			adapterState.uploadCalls++
			adapterState.Unlock()

			if strings.Contains(shardID, "-shard-2") {
				http.Error(w, "forced upload failure", http.StatusInternalServerError)
				return
			}

			adapterState.Lock()
			adapterState.nextID++
			remoteID := fmt.Sprintf("remote-%d", adapterState.nextID)
			adapterState.uploaded[provider+"|"+remoteID] = struct{}{}
			adapterState.Unlock()

			_ = json.NewEncoder(w).Encode(types.UploadShardResp{RemoteID: remoteID, ChecksumSha: "ok"})
		},
		OnDeleteShard: func(w http.ResponseWriter, r *http.Request) {
			remoteID := strings.TrimPrefix(r.URL.Path, "/shards/")
			provider := r.URL.Query().Get("provider")
			key := provider + "|" + remoteID

			adapterState.Lock()
			adapterState.deleteCalls++
			adapterState.deleted[key] = struct{}{}
			adapterState.Unlock()

			w.WriteHeader(http.StatusNoContent)
		},
	})
	defer adapterServer.Close()

	shardMapState := struct {
		sync.Mutex
		recordCalled bool
	}{}

	shardMapServer := newShardMapMock(t, shardMapMockConfig{
		OnRegisterFile: func(w http.ResponseWriter, r *http.Request) {
			_ = json.NewEncoder(w).Encode(types.RegisterFileResp{FileID: "file-rollback", Status: "PENDING"})
		},
		OnRecordShards: func(w http.ResponseWriter, r *http.Request) {
			shardMapState.Lock()
			shardMapState.recordCalled = true
			shardMapState.Unlock()
			w.WriteHeader(http.StatusCreated)
		},
	})
	defer shardMapServer.Close()

	shardingServer := newShardingMock(t, shardingMockConfig{
		OnShard: func(w http.ResponseWriter, r *http.Request) {
			shards := make([]map[string]any, 0, 6)
			for i := 0; i < 6; i++ {
				shards = append(shards, map[string]any{"shardIndex": i, "shardType": "data", "shardData": []byte("ok")})
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"shards": shards})
		},
	})
	defer shardingServer.Close()

	orchestratorURL, shutdown := startOrchestrator(t, adapterServer.URL, shardMapServer.URL, shardingServer.URL)
	defer shutdown()

	resp := uploadFile(t, orchestratorURL, []byte("rollback-check"))
	if resp.Status != "failed" {
		t.Fatalf("expected failed upload status, got: %+v", resp)
	}
	if !strings.Contains(resp.Error, "only 5/6 shards succeeded") {
		t.Fatalf("expected partial-upload failure message, got: %q", resp.Error)
	}

	adapterState.Lock()
	uploadedCount := len(adapterState.uploaded)
	deletedCount := len(adapterState.deleted)
	uploadCalls := adapterState.uploadCalls
	deleteCalls := adapterState.deleteCalls
	adapterState.Unlock()

	if uploadCalls != 6 {
		t.Fatalf("expected 6 upload attempts, got %d", uploadCalls)
	}
	if uploadedCount != 5 {
		t.Fatalf("expected 5 successful uploaded shards, got %d", uploadedCount)
	}
	if deleteCalls != 5 || deletedCount != 5 {
		t.Fatalf("expected rollback delete for all successful shards (5), got deleteCalls=%d uniqueDeleted=%d", deleteCalls, deletedCount)
	}

	shardMapState.Lock()
	recordCalled := shardMapState.recordCalled
	shardMapState.Unlock()
	if recordCalled {
		t.Fatalf("did not expect shard-map record call when upload is partial failure")
	}
}

// TestOrchestratorUploadRollsBackWhenShardRecordFails verifies that orchestrator
// rolls back all uploaded shards when shard-map record fails after successful uploads.
func TestOrchestratorUploadRollsBackWhenShardRecordFails(t *testing.T) {
	t.Parallel()

	adapterState := struct {
		sync.Mutex
		nextID      int
		uploaded    map[string]struct{}
		deleted     map[string]struct{}
		uploadCalls int
		deleteCalls int
	}{
		uploaded: map[string]struct{}{},
		deleted:  map[string]struct{}{},
	}

	adapterServer := newAdapterMock(t, adapterMockConfig{
		OnGetProviders: func(w http.ResponseWriter, r *http.Request) {
			_ = json.NewEncoder(w).Encode([]map[string]any{
				{"providerId": "provider-a", "displayName": "Provider A", "status": "connected"},
				{"providerId": "provider-b", "displayName": "Provider B", "status": "connected"},
			})
		},
		OnUploadShard: func(w http.ResponseWriter, r *http.Request) {
			if err := r.ParseMultipartForm(8 << 20); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			provider := r.FormValue("provider")

			adapterState.Lock()
			adapterState.uploadCalls++
			adapterState.nextID++
			remoteID := fmt.Sprintf("remote-%d", adapterState.nextID)
			adapterState.uploaded[provider+"|"+remoteID] = struct{}{}
			adapterState.Unlock()

			_ = json.NewEncoder(w).Encode(types.UploadShardResp{RemoteID: remoteID, ChecksumSha: "ok"})
		},
		OnDeleteShard: func(w http.ResponseWriter, r *http.Request) {
			remoteID := strings.TrimPrefix(r.URL.Path, "/shards/")
			provider := r.URL.Query().Get("provider")
			key := provider + "|" + remoteID

			adapterState.Lock()
			adapterState.deleteCalls++
			adapterState.deleted[key] = struct{}{}
			adapterState.Unlock()

			w.WriteHeader(http.StatusNoContent)
		},
	})
	defer adapterServer.Close()

	shardMapState := struct {
		sync.Mutex
		recordCalls int
	}{ }

	shardMapServer := newShardMapMock(t, shardMapMockConfig{
		OnRegisterFile: func(w http.ResponseWriter, r *http.Request) {
			_ = json.NewEncoder(w).Encode(types.RegisterFileResp{FileID: "file-record-fail", Status: "PENDING"})
		},
		OnRecordShards: func(w http.ResponseWriter, r *http.Request) {
			shardMapState.Lock()
			shardMapState.recordCalls++
			shardMapState.Unlock()
			http.Error(w, "record unavailable", http.StatusInternalServerError)
		},
	})
	defer shardMapServer.Close()

	shardingServer := newShardingMock(t, shardingMockConfig{
		OnShard: func(w http.ResponseWriter, r *http.Request) {
			shards := make([]map[string]any, 0, 6)
			for i := 0; i < 6; i++ {
				shards = append(shards, map[string]any{"shardIndex": i, "shardType": "data", "shardData": []byte("ok")})
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"shards": shards})
		},
	})
	defer shardingServer.Close()

	orchestratorURL, shutdown := startOrchestrator(t, adapterServer.URL, shardMapServer.URL, shardingServer.URL)
	defer shutdown()

	resp := uploadFile(t, orchestratorURL, []byte("record-failure-rollback"))
	if resp.Status != "failed" {
		t.Fatalf("expected failed upload status, got: %+v", resp)
	}
	if !strings.Contains(resp.Error, "failed to record shards") {
		t.Fatalf("expected record failure in error message, got: %q", resp.Error)
	}

	adapterState.Lock()
	uploadedCount := len(adapterState.uploaded)
	deletedCount := len(adapterState.deleted)
	uploadCalls := adapterState.uploadCalls
	deleteCalls := adapterState.deleteCalls
	adapterState.Unlock()

	if uploadCalls != 6 || uploadedCount != 6 {
		t.Fatalf("expected 6 successful uploads before record failure, got uploadCalls=%d uploaded=%d", uploadCalls, uploadedCount)
	}
	if deleteCalls != 6 || deletedCount != 6 {
		t.Fatalf("expected rollback delete for all successful shards (6), got deleteCalls=%d uniqueDeleted=%d", deleteCalls, deletedCount)
	}

	shardMapState.Lock()
	recordCalls := shardMapState.recordCalls
	shardMapState.Unlock()
	if recordCalls != 1 {
		t.Fatalf("expected exactly one record call, got %d", recordCalls)
	}
}
