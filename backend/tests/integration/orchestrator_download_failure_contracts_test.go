package integration_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestOrchestratorDownloadFailsWhenShardMapLookupErrors verifies that orchestrator
// returns a stable 500 JSON error response when shard-map lookup fails during download.
// This protects the download contract for upstream dependency failures.
func TestOrchestratorDownloadFailsWhenShardMapLookupErrors(t *testing.T) {
	t.Parallel()

	adapterServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer adapterServer.Close()

	shardMapServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/api/v1/shards/file/") {
			http.Error(w, "shard map unavailable", http.StatusInternalServerError)
			return
		}
		http.NotFound(w, r)
	}))
	defer shardMapServer.Close()

	shardingServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer shardingServer.Close()

	orchestratorURL, shutdown := startOrchestrator(t, adapterServer.URL, shardMapServer.URL, shardingServer.URL)
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