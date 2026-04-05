package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/vindyang/cs464-project/backend/services/orchestrator/internal/app"
	adapterclient "github.com/vindyang/cs464-project/backend/services/shared/clients/adapter"
	"github.com/vindyang/cs464-project/backend/services/shared/clients/sharding"
	shardmapworkflow "github.com/vindyang/cs464-project/backend/services/shared/clients/shardmapworkflow"
	"github.com/vindyang/cs464-project/backend/services/shared/transport/httpx"
)
// CD testing again
// CI testing

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

	// GET /health
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if !httpx.RequireMethod(w, r, http.MethodGet) {
			return
		}

		type serviceHealth struct {
			Service    string `json:"service"`
			URL        string `json:"url"`
			Status     string `json:"status"`
			HTTPStatus int    `json:"httpStatus"`
			Error      string `json:"error,omitempty"`
		}

		checks := []struct {
			service string
			url     string
		}{
			{service: "orchestrator", url: "self"},
			{service: "adapter", url: adapterURL + "/health"},
			{service: "shardmap", url: shardMapURL + "/health"},
			{service: "sharding", url: shardingURL + "/api/sharding/health"},
		}

		results := make([]serviceHealth, 0, len(checks))
		results = append(results, serviceHealth{
			Service:    "orchestrator",
			URL:        "self",
			Status:     "healthy",
			HTTPStatus: http.StatusOK,
		})

		client := &http.Client{Timeout: 3 * time.Second}
		var mu sync.Mutex
		var wg sync.WaitGroup

		for _, check := range checks[1:] {
			wg.Add(1)
			go func(serviceName, url string) {
				defer wg.Done()

				req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, url, nil)
				if err != nil {
					mu.Lock()
					results = append(results, serviceHealth{Service: serviceName, URL: url, Status: "unhealthy", HTTPStatus: 0, Error: err.Error()})
					mu.Unlock()
					return
				}

				resp, err := client.Do(req)
				if err != nil {
					mu.Lock()
					results = append(results, serviceHealth{Service: serviceName, URL: url, Status: "unhealthy", HTTPStatus: 0, Error: err.Error()})
					mu.Unlock()
					return
				}
				defer resp.Body.Close()

				status := "healthy"
				if resp.StatusCode < 200 || resp.StatusCode >= 300 {
					status = "unhealthy"
				}

				mu.Lock()
				results = append(results, serviceHealth{Service: serviceName, URL: url, Status: status, HTTPStatus: resp.StatusCode})
				mu.Unlock()
			}(check.service, check.url)
		}
		wg.Wait()

		overallStatus := "healthy"
		httpStatus := http.StatusOK
		for _, item := range results {
			if item.Status != "healthy" {
				overallStatus = "degraded"
				httpStatus = http.StatusServiceUnavailable
				break
			}
		}

		httpx.WriteJSON(w, httpStatus, map[string]any{
			"status":   overallStatus,
			"service":  "orchestrator",
			"services": results,
		})
	})

	// POST /api/orchestrator/upload
	mux.HandleFunc("/api/orchestrator/upload", func(w http.ResponseWriter, r *http.Request) {
		if !httpx.RequireMethod(w, r, http.MethodPost) {
			return
		}

		fileName, fileData, k, n, err := parseUploadMultipart(r)
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, "Invalid upload payload", err)
			return
		}

		resp, err := service.UploadRawFile(r.Context(), fileName, fileData, k, n)
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, "Failed to upload file", err)
			return
		}

		httpx.WriteJSON(w, http.StatusCreated, resp)
	})

	// GET /api/orchestrator/history
	mux.HandleFunc("/api/orchestrator/history", func(w http.ResponseWriter, r *http.Request) {
		if !httpx.RequireMethod(w, r, http.MethodGet) {
			return
		}
		history, err := service.GetAllHistory(r.Context())
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, "Failed to get lifecycle history", err)
			return
		}
		httpx.WriteJSON(w, http.StatusOK, history)
	})

	// POST /api/orchestrator/files/health/refresh
	mux.HandleFunc("/api/orchestrator/files/health/refresh", func(w http.ResponseWriter, r *http.Request) {
		if !httpx.RequireMethod(w, r, http.MethodPost) {
			return
		}

		summary, err := service.RefreshAllFileHealth(r.Context())
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, "Failed to refresh file health", err)
			return
		}

		httpx.WriteJSON(w, http.StatusOK, summary)
	})

	// GET  /api/orchestrator/files/{fileId}/download
	// GET  /api/orchestrator/files/{fileId}/history
	// DELETE /api/orchestrator/files/{fileId}
	mux.HandleFunc("/api/orchestrator/files/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/api/orchestrator/files/")

		// Compatibility: handle refresh route here too, in case a proxy or trailing-slash
		// variant bypasses the exact /api/orchestrator/files/health/refresh handler.
		if (path == "health/refresh" || path == "health/refresh/") && r.Method == http.MethodPost {
			summary, err := service.RefreshAllFileHealth(r.Context())
			if err != nil {
				httpx.WriteError(w, http.StatusInternalServerError, "Failed to refresh file health", err)
				return
			}
			httpx.WriteJSON(w, http.StatusOK, summary)
			return
		}

		parts := strings.Split(path, "/")

		if len(parts) == 0 || parts[0] == "" {
			httpx.WriteJSON(w, http.StatusNotFound, map[string]string{"error": "Not found"})
			return
		}

		fileID := parts[0]
		if len(fileID) < 5 {
			httpx.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid file ID"})
			return
		}

		// DELETE /api/orchestrator/files/{fileId}
		if r.Method == http.MethodDelete {
			deleteShards := r.URL.Query().Get("delete_shards") == "true"
			if err := service.DeleteFile(r.Context(), fileID, deleteShards); err != nil {
				httpx.WriteError(w, http.StatusInternalServerError, "Failed to delete file", err)
				return
			}
			w.WriteHeader(http.StatusNoContent)
			return
		}

		if r.Method == http.MethodPost && len(parts) == 3 && parts[1] == "health" && parts[2] == "refresh" {
			summary, err := service.RefreshFileHealth(r.Context(), fileID)
			if err != nil {
				httpx.WriteError(w, http.StatusInternalServerError, "Failed to refresh file health", err)
				return
			}
			httpx.WriteJSON(w, http.StatusOK, summary)
			return
		}

		if r.Method != http.MethodGet {
			httpx.WriteJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed"})
			return
		}

		if len(parts) != 2 || parts[1] == "" {
			httpx.WriteJSON(w, http.StatusNotFound, map[string]string{"error": "Not found"})
			return
		}

		switch parts[1] {
		case "download":
			resp, err := service.DownloadFile(r.Context(), fileID)
			if err != nil {
				var recoverabilityErr *app.RecoverabilityError
				if errors.As(err, &recoverabilityErr) {
					httpx.WriteError(w, http.StatusConflict, "File cannot be reconstructed", err)
				} else if strings.Contains(err.Error(), "404") {
					httpx.WriteError(w, http.StatusNotFound, "File not found", err)
				} else {
					httpx.WriteError(w, http.StatusInternalServerError, "Failed to download file", err)
				}
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

		case "history":
			history, err := service.GetFileHistory(r.Context(), fileID)
			if err != nil {
				httpx.WriteError(w, http.StatusInternalServerError, "Failed to get file history", err)
				return
			}
			httpx.WriteJSON(w, http.StatusOK, history)

		default:
			httpx.WriteJSON(w, http.StatusNotFound, map[string]string{"error": "Not found"})
		}
	})

	// Start server
	log.Printf("Orchestrator service starting on :%s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("failed to start server: %v\n", err)
	}
}

func parseUploadMultipart(r *http.Request) (fileName string, fileData []byte, k int, n int, err error) {
	reader, err := r.MultipartReader()
	if err != nil {
		return "", nil, 0, 0, fmt.Errorf("failed to open multipart reader: %w", err)
	}

	var kRaw, nRaw string
	for {
		part, partErr := reader.NextPart()
		if errors.Is(partErr, io.EOF) {
			break
		}
		if partErr != nil {
			return "", nil, 0, 0, fmt.Errorf("failed to read multipart part: %w", partErr)
		}

		name := part.FormName()
		switch name {
		case "file":
			fileName = part.FileName()
			fileData, err = io.ReadAll(part)
			if err != nil {
				return "", nil, 0, 0, fmt.Errorf("failed to read uploaded file: %w", err)
			}
		case "k":
			b, readErr := io.ReadAll(part)
			if readErr != nil {
				return "", nil, 0, 0, fmt.Errorf("failed to read k value: %w", readErr)
			}
			kRaw = strings.TrimSpace(string(b))
		case "n":
			b, readErr := io.ReadAll(part)
			if readErr != nil {
				return "", nil, 0, 0, fmt.Errorf("failed to read n value: %w", readErr)
			}
			nRaw = strings.TrimSpace(string(b))
		default:
			_, _ = io.Copy(io.Discard, part)
		}
		_ = part.Close()
	}

	if fileName == "" || len(fileData) == 0 {
		return "", nil, 0, 0, fmt.Errorf("missing file field")
	}
	parsedK, err := strconv.Atoi(kRaw)
	if err != nil {
		return "", nil, 0, 0, fmt.Errorf("invalid k value: %w", err)
	}
	parsedN, err := strconv.Atoi(nRaw)
	if err != nil {
		return "", nil, 0, 0, fmt.Errorf("invalid n value: %w", err)
	}

	return fileName, fileData, parsedK, parsedN, nil
}
