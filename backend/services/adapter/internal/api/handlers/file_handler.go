package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/vindyang/cs464-project/backend/services/shared/adapter"
	"github.com/vindyang/cs464-project/backend/services/shared/transport/httpx"
)

// FileHandler proxies file and shard-map requests to the shardmap service.
type FileHandler struct {
	shardmapURL string
	registry    *adapter.Registry
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
	h.proxyGET(w, r, h.shardmapURL+"/api/v1/shards/file/"+fileIDStr)
}

// DeleteFile handles DELETE /api/v1/files/:fileId.
// If ?delete_shards=true, it first deletes each shard from its provider, then removes the DB record.
func (h *FileHandler) DeleteFile(w http.ResponseWriter, r *http.Request, fileIDStr string) {
	if r.URL.Query().Get("delete_shards") == "true" {
		// Fetch shard map so we know which provider holds each shard.
		shardMapReq, err := http.NewRequestWithContext(r.Context(), http.MethodGet,
			h.shardmapURL+"/api/v1/shards/file/"+fileIDStr, nil)
		if err != nil {
			httpx.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to build shard map request"})
			return
		}
		smResp, err := http.DefaultClient.Do(shardMapReq)
		if err != nil {
			httpx.WriteJSON(w, http.StatusBadGateway, map[string]string{"error": "Failed to reach shardmap service"})
			return
		}
		defer smResp.Body.Close()

		var shardMap struct {
			Shards []struct {
				RemoteID string `json:"remote_id"`
				Provider string `json:"provider"`
			} `json:"shards"`
		}
		if err := json.NewDecoder(smResp.Body).Decode(&shardMap); err == nil {
			for _, shard := range shardMap.Shards {
				p, err := h.registry.Get(shard.Provider)
				if err != nil {
					continue
				}
				if err := p.DeleteShard(r.Context(), shard.RemoteID); err != nil {
					log.Printf("DeleteShard(%s, %s): %v", shard.Provider, shard.RemoteID, err)
				}
			}
		}
	}

	// Always delete the DB record.
	upstream, err := http.NewRequestWithContext(r.Context(), http.MethodDelete,
		h.shardmapURL+"/api/v1/files/"+fileIDStr, nil)
	if err != nil {
		httpx.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to build upstream request"})
		return
	}
	h.doProxy(w, upstream)
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
