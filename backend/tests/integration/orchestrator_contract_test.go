package integration_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/vindyang/cs464-project/backend/services/shared/types"
)

func TestOrchestratorServiceContracts(t *testing.T) {
	t.Parallel()

	// Arrange: mock Adapter service with strict workflow-facing contracts.
	// This verifies orchestrator/adapter compatibility without depending on cloud SDKs.

	adapterStorage := struct {
		sync.Mutex
		nextID int
		data   map[string][]byte
	}{data: map[string][]byte{}}

	adapterServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/providers":
			_ = json.NewEncoder(w).Encode([]map[string]any{
				{"providerId": "provider-a", "displayName": "Provider A", "status": "connected", "latencyMs": 10, "quotaTotalBytes": 1_000_000, "quotaUsedBytes": 1000},
				{"providerId": "provider-b", "displayName": "Provider B", "status": "connected", "latencyMs": 12, "quotaTotalBytes": 1_000_000, "quotaUsedBytes": 2000},
			})
		case r.Method == http.MethodPost && r.URL.Path == "/shards/upload":
			if err := r.ParseMultipartForm(8 << 20); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			provider := r.FormValue("provider")
			file, _, err := r.FormFile("file_data")
			if err != nil {
				http.Error(w, "missing file_data", http.StatusBadRequest)
				return
			}
			defer file.Close()
			payload, _ := io.ReadAll(file)

			adapterStorage.Lock()
			adapterStorage.nextID++
			remoteID := fmt.Sprintf("remote-%d", adapterStorage.nextID)
			adapterStorage.data[provider+"|"+remoteID] = payload
			adapterStorage.Unlock()

			_ = json.NewEncoder(w).Encode(types.UploadShardResp{RemoteID: remoteID, ChecksumSha: "ok"})
		case strings.HasPrefix(r.URL.Path, "/shards/"):
			remoteID := strings.TrimPrefix(r.URL.Path, "/shards/")
			provider := r.URL.Query().Get("provider")
			key := provider + "|" + remoteID
			switch r.Method {
			case http.MethodGet:
				adapterStorage.Lock()
				payload := adapterStorage.data[key]
				adapterStorage.Unlock()
				if payload == nil {
					http.Error(w, "not found", http.StatusNotFound)
					return
				}
				_, _ = w.Write(payload)
			case http.MethodDelete:
				adapterStorage.Lock()
				delete(adapterStorage.data, key)
				adapterStorage.Unlock()
				w.WriteHeader(http.StatusNoContent)
			default:
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			}
		default:
			http.NotFound(w, r)
		}
	}))
	defer adapterServer.Close()

	// Arrange: mock ShardMap service with strict validation used in production handlers.
	// - register requires positive original_size
	// - record expects shard type values DATA/PARITY
	shardMap := struct {
		sync.Mutex
		registerReq types.RegisterFileReq
		recordReq   types.RecordShardReq
		mapResp     types.GetShardMapResp
	}{}

	shardMapServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/shards/register":
			var req types.RegisterFileReq
			_ = json.NewDecoder(r.Body).Decode(&req)
			if req.OriginalSize <= 0 {
				http.Error(w, "original_size must be positive", http.StatusBadRequest)
				return
			}
			shardMap.Lock()
			shardMap.registerReq = req
			shardMap.Unlock()
			_ = json.NewEncoder(w).Encode(types.RegisterFileResp{FileID: "file-12345", Status: "PENDING"})
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/shards/record":
			var req types.RecordShardReq
			_ = json.NewDecoder(r.Body).Decode(&req)
			for _, shard := range req.Shards {
				if shard.Type != "DATA" && shard.Type != "PARITY" {
					http.Error(w, "invalid shard type", http.StatusBadRequest)
					return
				}
			}
			entries := make([]types.ShardMapEntry, 0, len(req.Shards))
			for i, shard := range req.Shards {
				entries = append(entries, types.ShardMapEntry{ShardID: fmt.Sprintf("s-%d", i), ShardIndex: shard.ShardIndex, RemoteID: shard.RemoteID, Provider: shard.Provider, Status: "HEALTHY"})
			}
			shardMap.Lock()
			shardMap.recordReq = req
			shardMap.mapResp = types.GetShardMapResp{FileID: req.FileID, OriginalName: "contract.txt", N: 6, K: 4, Status: "UPLOADED", Shards: entries}
			shardMap.Unlock()
			w.WriteHeader(http.StatusCreated)
		case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/api/v1/shards/file/"):
			shardMap.Lock()
			resp := shardMap.mapResp
			shardMap.Unlock()
			_ = json.NewEncoder(w).Encode(resp)
		default:
			http.NotFound(w, r)
		}
	}))
	defer shardMapServer.Close()

	// Arrange: mock Sharding service using current public API contracts.
	// Endpoint and payload contracts enforced here:
	// - POST /api/sharding/shard with fileId/fileData/n/k
	// - POST /api/sharding/reconstruct
	shardingServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/sharding/shard":
			var req struct {
				FileID   string `json:"fileId"`
				FileData []byte `json:"fileData"`
				N        int    `json:"n"`
				K        int    `json:"k"`
			}
			_ = json.NewDecoder(r.Body).Decode(&req)
			if req.FileID == "" || len(req.FileData) == 0 || req.K <= 0 || req.N <= 0 || req.K > req.N {
				http.Error(w, "invalid shard request", http.StatusBadRequest)
				return
			}
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
			_ = json.NewEncoder(w).Encode(map[string]any{"fileId": req.FileID, "shards": out})
		case r.Method == http.MethodPost && r.URL.Path == "/api/sharding/reconstruct":
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
			_ = json.NewEncoder(w).Encode(map[string]any{"reconstructed_file": req.AvailableShards[0].ShardData})
		default:
			http.NotFound(w, r)
		}
	}))
	defer shardingServer.Close()

	// Arrange: start real orchestrator process wired to mock services.
	orchestratorURL, shutdown := startOrchestrator(t, adapterServer.URL, shardMapServer.URL, shardingServer.URL)
	defer shutdown()

	// Act: upload a file through orchestrator.
	payload := []byte("contract-test-payload")
	resp := uploadFile(t, orchestratorURL, payload)
	if resp.Status != "committed" || resp.FileID == "" {
		t.Fatalf("unexpected upload response: %+v", resp)
	}

	// Assert: cross-service contract invariants were honored by orchestrator.
	shardMap.Lock()
	if shardMap.registerReq.OriginalSize <= 0 {
		t.Fatalf("expected shardmap register original_size > 0")
	}
	for _, shard := range shardMap.recordReq.Shards {
		if shard.Type != "DATA" && shard.Type != "PARITY" {
			t.Fatalf("invalid shard type recorded: %s", shard.Type)
		}
	}
	shardMap.Unlock()

	// Act + Assert: download path still works end-to-end with the same contracts.
	downloadRes, err := http.Get(orchestratorURL + "/api/orchestrator/files/" + resp.FileID + "/download")
	if err != nil {
		t.Fatalf("download request failed: %v", err)
	}
	defer downloadRes.Body.Close()
	body, _ := io.ReadAll(downloadRes.Body)
	if downloadRes.StatusCode != http.StatusOK {
		t.Fatalf("download failed: status=%d body=%s", downloadRes.StatusCode, string(body))
	}
	if !bytes.Equal(body, payload) {
		t.Fatalf("download payload mismatch")
	}
}

