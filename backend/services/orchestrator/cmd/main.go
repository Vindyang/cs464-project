package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/vindyang/cs464-project/backend/services/orchestrator/internal/app"
	adapterclient "github.com/vindyang/cs464-project/backend/services/shared/clients/adapter"
	"github.com/vindyang/cs464-project/backend/services/shared/clients/sharding"
	shardmapworkflow "github.com/vindyang/cs464-project/backend/services/shared/clients/shardmapworkflow"
	"github.com/vindyang/cs464-project/backend/services/shared/transport/httpx"
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

	shardingURL := os.Getenv("SHARDING_URL")
	if shardingURL == "" {
		shardingURL = "http://localhost:8083"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}

	// Initialize clients and service
	adapter := adapterclient.NewClient(adapterURL, nil)
	shardMap := shardmapworkflow.NewClient(shardMapURL, nil)
	shardingClient := sharding.NewClient(shardingURL, nil)
	service := app.NewServiceWithSharding(adapter, shardMap, shardingClient)

	// HTTP handlers
	mux := http.NewServeMux()

	// POST /api/orchestrator/upload
	mux.HandleFunc("/api/orchestrator/upload", func(w http.ResponseWriter, r *http.Request) {
		if !httpx.RequireMethod(w, r, http.MethodPost) {
			return
		}

		if err := r.ParseMultipartForm(100 << 20); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, "Failed to parse multipart form", err)
			return
		}

		file, fileHeader, err := r.FormFile("file")
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, "Missing file field", err)
			return
		}
		defer file.Close()

		fileData, err := io.ReadAll(file)
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, "Failed to read uploaded file", err)
			return
		}

		k, err := strconv.Atoi(r.FormValue("k"))
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, "Invalid k value", err)
			return
		}

		n, err := strconv.Atoi(r.FormValue("n"))
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, "Invalid n value", err)
			return
		}

		resp, err := service.UploadRawFile(r.Context(), fileHeader.Filename, fileData, k, n)
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, "Failed to upload file", err)
			return
		}

		httpx.WriteJSON(w, http.StatusCreated, resp)
	})

	// GET /api/orchestrator/files/{fileId}/download
	mux.HandleFunc("/api/orchestrator/files/", func(w http.ResponseWriter, r *http.Request) {
		if !httpx.RequireMethod(w, r, http.MethodGet) {
			return
		}

		path := strings.TrimPrefix(r.URL.Path, "/api/orchestrator/files/")
		parts := strings.Split(path, "/")
		if len(parts) != 2 || parts[0] == "" || parts[1] != "download" {
			httpx.WriteJSON(w, http.StatusNotFound, map[string]string{"error": "Not found"})
			return
		}

		fileID := parts[0]
		if len(fileID) < 5 {
			httpx.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid file ID"})
			return
		}

		resp, err := service.DownloadFile(r.Context(), fileID)
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, "Failed to download file", err)
			return
		}

		if len(resp.Shards) == 0 {
			httpx.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "No file data reconstructed"})
			return
		}

		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Disposition", "attachment; filename=\""+resp.FileName+"\"")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(resp.Shards[0])
	})

	// Start server
	log.Printf("Orchestrator service starting on :%s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
