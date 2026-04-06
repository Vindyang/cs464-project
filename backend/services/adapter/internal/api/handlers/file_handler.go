package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/vindyang/cs464-project/backend/services/shared/adapter"
	"github.com/vindyang/cs464-project/backend/services/shared/api/dto"
	"github.com/vindyang/cs464-project/backend/services/shared/transport/httpx"
)

// FileHandler proxies file and shard-map requests to the shardmap service.
type FileHandler struct {
	shardmapURL string
	registry    *adapter.Registry
}

type FileDeleteSummary struct {
	DeletedFiles       int  `json:"deleted_files"`
	DeletedShards      int  `json:"deleted_shards"`
	FailedShardDeletes int  `json:"failed_shard_deletes"`
	DeleteShards       bool `json:"delete_shards"`
}

// NewFileHandler creates a new FileHandler instance.
func NewFileHandler(shardmapURL string, registry *adapter.Registry) *FileHandler {
	return &FileHandler{shardmapURL: shardmapURL, registry: registry}
}

// RegisterRoutes registers all file and shard-map proxy routes.
func (h *FileHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/files", h.handleFileRoutes)
	mux.HandleFunc("/api/v1/files/", h.handleFileRoutes)
	mux.HandleFunc("/api/v1/shards/file/", h.proxyShardMap)
}

// handleFileRoutes dispatches based on path shape and HTTP method.
func (h *FileHandler) handleFileRoutes(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/files/")
	parts := strings.Split(path, "/")

	if len(parts) == 0 || parts[0] == "" {
		h.ListFiles(w, r)
		return
	}

	fileIDStr := parts[0]

	if len(parts) == 1 {
		if r.Method == http.MethodGet {
			h.GetFileMetadata(w, r, fileIDStr)
		} else if r.Method == http.MethodDelete {
			h.DeleteFile(w, r, fileIDStr)
		} else {
			httpx.WriteJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed"})
		}
		return
	}

	if len(parts) == 2 && parts[1] == "download" && r.Method == http.MethodGet {
		httpx.WriteJSON(w, http.StatusNotImplemented, map[string]string{"error": "Download not yet supported via this endpoint; use the orchestrator"})
		return
	}

	httpx.WriteJSON(w, http.StatusNotFound, map[string]string{"error": "Not found"})
}

// ListFiles handles GET /api/v1/files by proxying to the shardmap service.
func (h *FileHandler) ListFiles(w http.ResponseWriter, r *http.Request) {
	h.proxyGET(w, r, h.shardmapURL+"/api/v1/files")
}

// GetFileMetadata handles GET /api/v1/files/:fileId by proxying to the shardmap service.
func (h *FileHandler) GetFileMetadata(w http.ResponseWriter, r *http.Request, fileIDStr string) {
	h.proxyGET(w, r, h.shardmapURL+"/api/v1/files/"+fileIDStr)
}

// DeleteFile handles DELETE /api/v1/files/:fileId.
// If ?delete_shards=true, it first deletes each shard from its provider, then removes the DB record.
func (h *FileHandler) DeleteFile(w http.ResponseWriter, r *http.Request, fileIDStr string) {
	summary, err := h.deleteFileData(r.Context(), fileIDStr, r.URL.Query().Get("delete_shards") == "true")
	if err != nil {
		httpx.WriteJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"success":              true,
		"file_id":              fileIDStr,
		"message":              "File deleted successfully",
		"shards_deleted":       summary.DeletedShards,
		"failed_shard_deletes": summary.FailedShardDeletes,
	})
}

// proxyShardMap handles GET /api/v1/shards/file/:fileId by proxying to the shardmap service.
func (h *FileHandler) proxyShardMap(w http.ResponseWriter, r *http.Request) {
	fileIDStr := strings.TrimPrefix(r.URL.Path, "/api/v1/shards/file/")
	h.proxyGET(w, r, h.shardmapURL+"/api/v1/shards/file/"+fileIDStr)
}

// proxyGET is a helper for proxying GET requests to an upstream URL.
func (h *FileHandler) proxyGET(w http.ResponseWriter, r *http.Request, url string) {
	upstream, err := http.NewRequestWithContext(r.Context(), http.MethodGet, url, nil)
	if err != nil {
		httpx.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to build upstream request"})
		return
	}
	h.doProxy(w, upstream)
}

// doProxy executes an upstream request and streams the response back.
func (h *FileHandler) doProxy(w http.ResponseWriter, req *http.Request) {
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		httpx.WriteJSON(w, http.StatusBadGateway, map[string]string{"error": "Failed to reach shardmap service"})
		return
	}
	defer resp.Body.Close()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func (h *FileHandler) DeleteAllFiles(ctx context.Context, deleteShards bool) (FileDeleteSummary, error) {
	upstream, err := http.NewRequestWithContext(ctx, http.MethodGet, h.shardmapURL+"/api/v1/files", nil)
	if err != nil {
		return FileDeleteSummary{}, err
	}

	resp, err := http.DefaultClient.Do(upstream)
	if err != nil {
		return FileDeleteSummary{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return FileDeleteSummary{}, errorsFromBody(resp, "failed to list files for reset")
	}

	var files []dto.FileMetadataResponse
	if err := json.NewDecoder(resp.Body).Decode(&files); err != nil {
		return FileDeleteSummary{}, err
	}

	summary := FileDeleteSummary{DeleteShards: deleteShards}
	for _, file := range files {
		result, err := h.deleteFileData(ctx, file.FileID, deleteShards)
		if err != nil {
			return summary, err
		}
		summary.DeletedFiles += result.DeletedFiles
		summary.DeletedShards += result.DeletedShards
		summary.FailedShardDeletes += result.FailedShardDeletes
	}

	return summary, nil
}

func (h *FileHandler) deleteFileData(ctx context.Context, fileIDStr string, deleteShards bool) (FileDeleteSummary, error) {
	type shardEntry struct {
		RemoteID string `json:"remote_id"`
		Provider string `json:"provider"`
	}
	type shardMapResp struct {
		Shards []shardEntry `json:"shards"`
	}

	summary := FileDeleteSummary{DeleteShards: deleteShards}

	if deleteShards {
		shardMapReq, err := http.NewRequestWithContext(ctx, http.MethodGet, h.shardmapURL+"/api/v1/shards/file/"+fileIDStr, nil)
		if err != nil {
			return summary, err
		}
		smResp, err := http.DefaultClient.Do(shardMapReq)
		if err != nil {
			return summary, err
		}
		defer smResp.Body.Close()

		if smResp.StatusCode < http.StatusOK || smResp.StatusCode >= http.StatusMultipleChoices {
			return summary, errorsFromBody(smResp, "failed to load shard map")
		}

		var fileShardMap shardMapResp
		if err := json.NewDecoder(smResp.Body).Decode(&fileShardMap); err != nil {
			return summary, err
		}

		for _, shard := range fileShardMap.Shards {
			p, err := h.registry.Get(shard.Provider)
			if err != nil {
				summary.FailedShardDeletes++
				continue
			}
			if err := p.DeleteShard(ctx, shard.RemoteID); err != nil {
				log.Printf("DeleteShard(%s, %s): %v", shard.Provider, shard.RemoteID, err)
				summary.FailedShardDeletes++
				continue
			}
			summary.DeletedShards++
		}
	}

	upstream, err := http.NewRequestWithContext(ctx, http.MethodDelete, h.shardmapURL+"/api/v1/files/"+fileIDStr, nil)
	if err != nil {
		return summary, err
	}

	resp, err := http.DefaultClient.Do(upstream)
	if err != nil {
		return summary, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return summary, errorsFromBody(resp, "failed to delete file metadata")
	}

	summary.DeletedFiles = 1
	return summary, nil
}

func errorsFromBody(resp *http.Response, fallback string) error {
	var payload struct {
		Error string `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err == nil && payload.Error != "" {
		return fmt.Errorf("%s: %s", fallback, payload.Error)
	}
	return fmt.Errorf("%s: upstream status %d", fallback, resp.StatusCode)
}
