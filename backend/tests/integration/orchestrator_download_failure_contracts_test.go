package integration_test

import (
	"github.com/vindyang/cs464-project/backend/tests/helpers"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

// TestOrchestratorDownloadFailsWhenShardMapLookupErrors verifies that orchestrator
// returns a stable 500 JSON error response when shard-map lookup fails during download.
// This protects the download contract for upstream dependency failures.
func TestOrchestratorDownloadFailsWhenShardMapLookupErrors(t *testing.T) {
	t.Parallel()

	adapterServer := helpers.NewAdapterMock(t, helpers.AdapterMockConfig{})
	defer adapterServer.Close()

	shardMapServer := helpers.NewShardMapMock(t, helpers.ShardMapMockConfig{
		OnGetShardMap: func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "shard map unavailable", http.StatusInternalServerError)
		},
	})
	defer shardMapServer.Close()

	shardingServer := helpers.NewShardingMock(t, helpers.ShardingMockConfig{})
	defer shardingServer.Close()

	orchestratorURL, shutdown := helpers.StartOrchestrator(t, adapterServer.URL, shardMapServer.URL, shardingServer.URL)
	defer shutdown()

	res, err := http.Get(orchestratorURL + "/api/orchestrator/files/file-12345/download")
	if err != nil {
		t.Fatalf("download request failed: %v", err)
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	if res.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusInternalServerError, res.StatusCode, string(body))
	}

	payload := map[string]string{}
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("failed to parse error response: %v body=%s", err, string(body))
	}
	if payload["error"] != "Failed to download file" {
		t.Fatalf("unexpected error message: %q", payload["error"])
	}
	if !strings.Contains(payload["details"], "file not found") {
		t.Fatalf("expected details to contain shard map lookup failure, got: %q", payload["details"])
	}
}

// TestOrchestratorDownloadFailsWhenShardingReconstructReturnsError verifies that
// orchestrator surfaces a stable 500 error when sharding reconstruct returns 500.
func TestOrchestratorDownloadFailsWhenShardingReconstructReturnsError(t *testing.T) {
	t.Parallel()

	adapterServer := helpers.NewAdapterMock(t, helpers.AdapterMockConfig{
		OnDownloadShard: func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("ok-shard-bytes"))
		},
	})
	defer adapterServer.Close()

	shardMapServer := helpers.NewShardMapMock(t, helpers.ShardMapMockConfig{
		OnGetShardMap: func(w http.ResponseWriter, r *http.Request) {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"file_id":       "file-12345",
				"original_name": "contract.txt",
				"n":             6,
				"k":             4,
				"status":        "UPLOADED",
				"shards": []map[string]any{
					{"shard_id": "s-0", "shard_index": 0, "remote_id": "remote-0", "provider": "provider-a", "status": "HEALTHY"},
					{"shard_id": "s-1", "shard_index": 1, "remote_id": "remote-1", "provider": "provider-b", "status": "HEALTHY"},
					{"shard_id": "s-2", "shard_index": 2, "remote_id": "remote-2", "provider": "provider-a", "status": "HEALTHY"},
					{"shard_id": "s-3", "shard_index": 3, "remote_id": "remote-3", "provider": "provider-b", "status": "HEALTHY"},
				},
			})
		},
	})
	defer shardMapServer.Close()

	shardingServer := helpers.NewShardingMock(t, helpers.ShardingMockConfig{
		OnReconstruct: func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "reconstruct failed", http.StatusInternalServerError)
		},
	})
	defer shardingServer.Close()

	orchestratorURL, shutdown := helpers.StartOrchestrator(t, adapterServer.URL, shardMapServer.URL, shardingServer.URL)
	defer shutdown()

	res, err := http.Get(orchestratorURL + "/api/orchestrator/files/file-12345/download")
	if err != nil {
		t.Fatalf("download request failed: %v", err)
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	if res.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusInternalServerError, res.StatusCode, string(body))
	}

	payload := map[string]string{}
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("failed to parse error response: %v body=%s", err, string(body))
	}
	if payload["error"] != "Failed to download file" {
		t.Fatalf("unexpected error message: %q", payload["error"])
	}
	if !strings.Contains(payload["details"], "failed to reconstruct file") {
		t.Fatalf("expected details to contain reconstruct failure, got: %q", payload["details"])
	}
}

// TestOrchestratorDownloadFailsWhenShardingReconstructPayloadIsMalformed verifies
// that orchestrator returns a stable 500 error when sharding reconstruct returns
// HTTP 200 with malformed JSON.
func TestOrchestratorDownloadFailsWhenShardingReconstructPayloadIsMalformed(t *testing.T) {
	t.Parallel()

	adapterServer := helpers.NewAdapterMock(t, helpers.AdapterMockConfig{
		OnDownloadShard: func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("ok-shard-bytes"))
		},
	})
	defer adapterServer.Close()

	shardMapServer := helpers.NewShardMapMock(t, helpers.ShardMapMockConfig{
		OnGetShardMap: func(w http.ResponseWriter, r *http.Request) {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"file_id":       "file-12345",
				"original_name": "contract.txt",
				"n":             6,
				"k":             4,
				"status":        "UPLOADED",
				"shards": []map[string]any{
					{"shard_id": "s-0", "shard_index": 0, "remote_id": "remote-0", "provider": "provider-a", "status": "HEALTHY"},
					{"shard_id": "s-1", "shard_index": 1, "remote_id": "remote-1", "provider": "provider-b", "status": "HEALTHY"},
					{"shard_id": "s-2", "shard_index": 2, "remote_id": "remote-2", "provider": "provider-a", "status": "HEALTHY"},
					{"shard_id": "s-3", "shard_index": 3, "remote_id": "remote-3", "provider": "provider-b", "status": "HEALTHY"},
				},
			})
		},
	})
	defer shardMapServer.Close()

	shardingServer := helpers.NewShardingMock(t, helpers.ShardingMockConfig{
		OnReconstruct: func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"reconstructed_file":`))
		},
	})
	defer shardingServer.Close()

	orchestratorURL, shutdown := helpers.StartOrchestrator(t, adapterServer.URL, shardMapServer.URL, shardingServer.URL)
	defer shutdown()

	res, err := http.Get(orchestratorURL + "/api/orchestrator/files/file-12345/download")
	if err != nil {
		t.Fatalf("download request failed: %v", err)
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	if res.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusInternalServerError, res.StatusCode, string(body))
	}

	payload := map[string]string{}
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("failed to parse error response: %v body=%s", err, string(body))
	}
	if payload["error"] != "Failed to download file" {
		t.Fatalf("unexpected error message: %q", payload["error"])
	}
	if !strings.Contains(payload["details"], "failed to reconstruct file") {
		t.Fatalf("expected details to contain reconstruct decode failure, got: %q", payload["details"])
	}
}

// TestOrchestratorDownloadFailsWhenAvailableShardsAreBelowK verifies that
// orchestrator returns a stable 500 error when shard-map returns fewer than K shards.
func TestOrchestratorDownloadFailsWhenAvailableShardsAreBelowK(t *testing.T) {
	t.Parallel()

	adapterServer := helpers.NewAdapterMock(t, helpers.AdapterMockConfig{})
	defer adapterServer.Close()

	shardMapServer := helpers.NewShardMapMock(t, helpers.ShardMapMockConfig{
		OnGetShardMap: func(w http.ResponseWriter, r *http.Request) {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"file_id":       "file-12345",
				"original_name": "contract.txt",
				"n":             6,
				"k":             4,
				"status":        "UPLOADED",
				"shards": []map[string]any{
					{"shard_id": "s-0", "shard_index": 0, "remote_id": "remote-0", "provider": "provider-a", "status": "HEALTHY"},
					{"shard_id": "s-1", "shard_index": 1, "remote_id": "remote-1", "provider": "provider-b", "status": "HEALTHY"},
					{"shard_id": "s-2", "shard_index": 2, "remote_id": "remote-2", "provider": "provider-a", "status": "HEALTHY"},
				},
			})
		},
	})
	defer shardMapServer.Close()

	shardingServer := helpers.NewShardingMock(t, helpers.ShardingMockConfig{})
	defer shardingServer.Close()

	orchestratorURL, shutdown := helpers.StartOrchestrator(t, adapterServer.URL, shardMapServer.URL, shardingServer.URL)
	defer shutdown()

	res, err := http.Get(orchestratorURL + "/api/orchestrator/files/file-12345/download")
	if err != nil {
		t.Fatalf("download request failed: %v", err)
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	if res.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusInternalServerError, res.StatusCode, string(body))
	}

	payload := map[string]string{}
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("failed to parse error response: %v body=%s", err, string(body))
	}
	if payload["error"] != "Failed to download file" {
		t.Fatalf("unexpected error message: %q", payload["error"])
	}
	if !strings.Contains(payload["details"], "insufficient shards available") {
		t.Fatalf("expected details to contain insufficient shards failure, got: %q", payload["details"])
	}
}
