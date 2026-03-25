package handlers

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/vindyang/cs464-project/backend/services/shared/adapter"
	"github.com/vindyang/cs464-project/backend/services/shared/transport/httpx"
)

// ShardIOHandler exposes shard-level storage operations for orchestrator workflow calls.
type ShardIOHandler struct {
	registry *adapter.Registry
}

func NewShardIOHandler(registry *adapter.Registry) *ShardIOHandler {
	return &ShardIOHandler{registry: registry}
}

func (h *ShardIOHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/shards/upload", h.UploadShard)
	mux.HandleFunc("/shards/", h.handleShardByRemoteID)
}

func (h *ShardIOHandler) UploadShard(w http.ResponseWriter, r *http.Request) {
	if !httpx.RequireMethod(w, r, http.MethodPost) {
		return
	}

	if err := r.ParseMultipartForm(25 << 20); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "Failed to parse multipart form", err)
		return
	}

	shardID := r.FormValue("shard_id")
	providerID := r.FormValue("provider")
	if shardID == "" || providerID == "" {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "shard_id and provider are required"})
		return
	}

	provider, err := h.registry.Get(providerID)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "Invalid provider", err)
		return
	}

	file, _, err := r.FormFile("file_data")
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "Missing file_data form field", err)
		return
	}
	defer file.Close()

	payload, err := io.ReadAll(file)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "Failed to read shard payload", err)
		return
	}

	remoteID, err := provider.UploadShard(r.Context(), shardID, parseShardIndex(shardID), bytes.NewReader(payload))
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "Failed to upload shard", err)
		return
	}

	hash := sha256.Sum256(payload)
	httpx.WriteJSON(w, http.StatusCreated, map[string]string{
		"remote_id":       remoteID,
		"checksum_sha256": hex.EncodeToString(hash[:]),
	})
}

func (h *ShardIOHandler) handleShardByRemoteID(w http.ResponseWriter, r *http.Request) {
	remoteID := strings.TrimPrefix(r.URL.Path, "/shards/")
	if remoteID == "" {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "remote shard ID is required"})
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.downloadShard(w, r, remoteID)
	case http.MethodDelete:
		h.deleteShard(w, r, remoteID)
	default:
		httpx.WriteJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed"})
	}
}

func (h *ShardIOHandler) downloadShard(w http.ResponseWriter, r *http.Request, remoteID string) {
	providerID := r.URL.Query().Get("provider")
	if providerID == "" {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "provider query parameter is required"})
		return
	}

	provider, err := h.registry.Get(providerID)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "Invalid provider", err)
		return
	}

	reader, err := provider.DownloadShard(r.Context(), remoteID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "Failed to download shard", err)
		return
	}
	defer reader.Close()

	shardData, err := io.ReadAll(reader)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "Failed to read shard data", err)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(shardData)
}

func (h *ShardIOHandler) deleteShard(w http.ResponseWriter, r *http.Request, remoteID string) {
	providerID := r.URL.Query().Get("provider")
	if providerID == "" {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "provider query parameter is required"})
		return
	}

	provider, err := h.registry.Get(providerID)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "Invalid provider", err)
		return
	}

	if err := provider.DeleteShard(r.Context(), remoteID); err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "Failed to delete shard", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func parseShardIndex(shardID string) int {
	if shardID == "" {
		// Return a clearly invalid index for empty shard IDs.
		return -1
	}

	parts := strings.Split(shardID, "-")
	if len(parts) == 0 {
		// No parts to parse; treat as invalid.
		return -1
	}

	idx, err := strconv.Atoi(parts[len(parts)-1])
	if err != nil {
		// Unparseable shard index; treat as invalid.
		return -1
	}
	return idx
}
