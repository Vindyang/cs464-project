package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
)

// In-memory mock data stores
var (
	files  = make(map[string]*mockFile)
	shards = make(map[string]*mockShard)
	mu     sync.RWMutex
)

type mockFile struct {
	FileID       string
	OriginalName string
	Status       string
	N            int
	K            int
}

type mockShard struct {
	ShardID string
	FileID  string
	Status  string
	Data    []byte
}

func main() {
	shardMapPort := os.Getenv("SHARDMAP_PORT")
	if shardMapPort == "" {
		shardMapPort = "8081"
	}

	adapterPort := os.Getenv("ADAPTER_PORT")
	if adapterPort == "" {
		adapterPort = "8080"
	}

	// ---- Shard Map Server ----
	shardMapMux := http.NewServeMux()

	shardMapMux.HandleFunc("/api/v1/shards/register", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			OriginalName string `json:"original_name"`
			OriginalSize int64  `json:"original_size"`
			TotalChunks  int    `json:"total_chunks"`
			N            int    `json:"n"`
			K            int    `json:"k"`
			ShardSize    int64  `json:"shard_size"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}

		fileID := fmt.Sprintf("file-%d", len(files)+1)
		mu.Lock()
		files[fileID] = &mockFile{
			FileID:       fileID,
			OriginalName: req.OriginalName,
			Status:       "PENDING",
			N:            req.N,
			K:            req.K,
		}
		mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{
			"file_id": fileID,
			"status":  "PENDING",
		})
	})

	shardMapMux.HandleFunc("/api/v1/shards/record", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			FileID string `json:"file_id"`
			Shards []struct {
				ChunkIndex  int    `json:"chunk_index"`
				ShardIndex  int    `json:"shard_index"`
				Type        string `json:"type"`
				RemoteID    string `json:"remote_id"`
				Provider    string `json:"provider"`
				ChecksumSha string `json:"checksum_sha256"`
			} `json:"shards"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}

		if len(req.Shards) == 0 {
			http.Error(w, "must provide at least 1 shard", http.StatusBadRequest)
			return
		}

		// In a real system, validate len(req.Shards) == N
		mu.Lock()
		if f, ok := files[req.FileID]; ok {
			f.Status = "COMMITTED"
		}
		mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	shardMapMux.HandleFunc("/api/v1/shards/file/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		fileID := r.URL.Path[len("/api/v1/shards/file/"):]
		mu.RLock()
		f, ok := files[fileID]
		mu.RUnlock()

		if !ok {
			http.Error(w, "file not found", http.StatusNotFound)
			return
		}

		// Return mock shard map with 64-char checksums
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"file_id":       f.FileID,
			"original_name": f.OriginalName,
			"n":             f.N,
			"k":             f.K,
			"status":        f.Status,
			"shards": []map[string]interface{}{
				{"shard_id": "1", "shard_index": 0, "remote_id": "remote-1", "provider": "awsS3", "checksum_sha256": "a3f2b1c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2", "status": "HEALTHY"},
				{"shard_id": "2", "shard_index": 1, "remote_id": "remote-2", "provider": "googleDrive", "checksum_sha256": "b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1", "status": "HEALTHY"},
				{"shard_id": "3", "shard_index": 2, "remote_id": "remote-3", "provider": "awsS3", "checksum_sha256": "c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2", "status": "HEALTHY"},
				{"shard_id": "4", "shard_index": 3, "remote_id": "remote-4", "provider": "googleDrive", "checksum_sha256": "d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3", "status": "HEALTHY"},
				{"shard_id": "5", "shard_index": 4, "remote_id": "remote-5", "provider": "onedrive", "checksum_sha256": "e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3d4", "status": "HEALTHY"},
				{"shard_id": "6", "shard_index": 5, "remote_id": "remote-6", "provider": "onedrive", "checksum_sha256": "f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3d4e5", "status": "HEALTHY"},
			},
		})
	})

	// ---- Adapter Server ----
	adapterMux := http.NewServeMux()

	adapterMux.HandleFunc("/api/providers", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]interface{}{
			{
				"providerId":      "awsS3",
				"displayName":     "AWS S3",
				"status":          "connected",
				"latencyMs":       int64(50),
				"quotaTotalBytes": int64(1000000000),
				"quotaUsedBytes":  int64(100000000),
			},
			{
				"providerId":      "googleDrive",
				"displayName":     "Google Drive",
				"status":          "connected",
				"latencyMs":       int64(60),
				"quotaTotalBytes": int64(1000000000),
				"quotaUsedBytes":  int64(150000000),
			},
			{
				"providerId":      "onedrive",
				"displayName":     "OneDrive",
				"status":          "connected",
				"latencyMs":       int64(70),
				"quotaTotalBytes": int64(1000000000),
				"quotaUsedBytes":  int64(200000000),
			},
		})
	})

	adapterMux.HandleFunc("/shards/upload", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		shardID := fmt.Sprintf("remote-shard-%d", len(shards)+1)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{
			"remote_id":       shardID,
			"checksum_sha256": "a3f2b1c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2",
		})
	})

	adapterMux.HandleFunc("/shards/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Write([]byte("mock-shard-data"))
		} else if r.Method == http.MethodDelete {
			w.WriteHeader(http.StatusNoContent)
		} else {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Start Shard Map Server
	go func() {
		log.Printf("Mock Shard Map Service starting on :%s", shardMapPort)
		if err := http.ListenAndServe(":"+shardMapPort, shardMapMux); err != nil {
			log.Fatalf("Shard Map server failed: %v", err)
		}
	}()

	// Start Adapter Server
	log.Printf("Mock Adapter Service starting on :%s", adapterPort)
	if err := http.ListenAndServe(":"+adapterPort, adapterMux); err != nil {
		log.Fatalf("Adapter server failed: %v", err)
	}
}
