package integration_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/vindyang/cs464-project/backend/services/shared/types"
)

// TestOrchestratorUploadFailsWhenShardMapRegisterErrors verifies that orchestrator
// returns a stable 500 JSON error response when shard-map file registration fails.
// This protects the upload contract for dependency failure propagation.
func TestOrchestratorUploadFailsWhenShardMapRegisterErrors(t *testing.T) {
	t.Parallel()

	adapterServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/providers":
			_ = json.NewEncoder(w).Encode([]map[string]any{
				{"providerId": "provider-a", "displayName": "Provider A", "status": "connected"},
				{"providerId": "provider-b", "displayName": "Provider B", "status": "connected"},
			})
		case r.Method == http.MethodPost && r.URL.Path == "/shards/upload":
			_ = json.NewEncoder(w).Encode(types.UploadShardResp{RemoteID: "unused", ChecksumSha: "unused"})
		default:
			http.NotFound(w, r)
		}
	}))
	defer adapterServer.Close()

	shardMapServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == "/api/v1/shards/register" {
			http.Error(w, "register unavailable", http.StatusInternalServerError)
			return
		}
		http.NotFound(w, r)
	}))
	defer shardMapServer.Close()

	shardingServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == "/api/sharding/shard" {
			shards := make([]map[string]any, 0, 6)
			for i := 0; i < 6; i++ {
				shards = append(shards, map[string]any{"shardIndex": i, "shardType": "data", "shardData": []byte("ok")})
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"shards": shards})
			return
		}
		http.NotFound(w, r)
	}))
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

	adapterServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/api/providers" {
			_ = json.NewEncoder(w).Encode([]map[string]any{})
			return
		}
		http.NotFound(w, r)
	}))
	defer adapterServer.Close()

	shardMapServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer shardMapServer.Close()

	shardingServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == "/api/sharding/shard" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"shards":[{"shardData":`))
			return
		}
		http.NotFound(w, r)
	}))
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