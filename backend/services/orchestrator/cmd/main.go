package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/vindyang/cs464-project/backend/services/orchestrator/internal/app"
	"github.com/vindyang/cs464-project/backend/services/shared/orchestrator/clients"
)

func main() {
	// Load service URLs from env
	adapterURL := os.Getenv("ADAPTER_URL")
	if adapterURL == "" {
		adapterURL = "http://localhost:8080"
	}

	shardMapURL := os.Getenv("SHARDMAP_URL")
	if shardMapURL == "" {
		shardMapURL = "http://localhost:8081"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}

	// Initialize clients and service
	adapter := clients.NewAdapterClient(adapterURL, nil)
	shardMap := clients.NewShardMapClient(shardMapURL, nil)
	service := app.NewService(adapter, shardMap)

	// HTTP handlers
	mux := http.NewServeMux()

	// POST /api/orchestrator/upload
	mux.HandleFunc("/api/orchestrator/upload", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			FileName    string   `json:"fileName"`
			Shards      []string `json:"shards"` // base64-encoded shards
			IsDataShard []bool   `json:"isDataShard"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}

		// Decode shards from base64 (or handle however frontend sends them)
		shards := make([][]byte, len(req.Shards))
		for i, s := range req.Shards {
			shards[i] = []byte(s) // TODO: base64.StdEncoding.DecodeString(s)
		}

		resp, err := service.UploadFile(r.Context(), req.FileName, shards, req.IsDataShard)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	// GET /api/orchestrator/files/{fileId}/download
	mux.HandleFunc("/api/orchestrator/files/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		fileID := r.URL.Path[len("/api/orchestrator/files/"):]
		if fileID == "" || len(fileID) < 5 {
			http.Error(w, "invalid file ID", http.StatusBadRequest)
			return
		}

		resp, err := service.DownloadFile(r.Context(), fileID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	// Start server
	log.Printf("Orchestrator service starting on :%s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
