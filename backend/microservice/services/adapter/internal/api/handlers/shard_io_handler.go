package handlers

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
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

	shardID, providerID, payload, err := parseShardUploadMultipart(r)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "Invalid shard upload payload", err)
		return
	}

	provider, err := h.registry.Get(providerID)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "Invalid provider", err)
		return
	}

	if shardID == "" || providerID == "" {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "shard_id and provider are required"})
		return
	}

	if len(payload) == 0 {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "file_data is required"})
		return
	}

	payloadReader := bytes.NewReader(payload)
	remoteID, err := provider.UploadShard(r.Context(), parseFileID(shardID), parseShardIndex(shardID), payloadReader)
	if err != nil {
		httpx.WriteErrorWithCode(w, http.StatusInternalServerError, "Failed to upload shard", httpx.ClassifyProviderError(err.Error()), err)
		return
	}

	hash := sha256.Sum256(payload)
	httpx.WriteJSON(w, http.StatusCreated, map[string]string{
		"remote_id":       remoteID,
		"checksum_sha256": hex.EncodeToString(hash[:]),
	})
}

func parseShardUploadMultipart(r *http.Request) (shardID string, providerID string, payload []byte, err error) {
	reader, err := r.MultipartReader()
	if err != nil {
		return "", "", nil, fmt.Errorf("failed to open multipart reader: %w", err)
	}

	for {
		part, partErr := reader.NextPart()
		if errors.Is(partErr, io.EOF) {
			break
		}
		if partErr != nil {
			return "", "", nil, fmt.Errorf("failed to read multipart part: %w", partErr)
		}

		switch part.FormName() {
		case "shard_id":
			b, readErr := io.ReadAll(part)
			if readErr != nil {
				return "", "", nil, fmt.Errorf("failed to read shard_id: %w", readErr)
			}
			shardID = strings.TrimSpace(string(b))
		case "provider":
			b, readErr := io.ReadAll(part)
			if readErr != nil {
				return "", "", nil, fmt.Errorf("failed to read provider: %w", readErr)
			}
			providerID = strings.TrimSpace(string(b))
		case "file_data":
			b, readErr := io.ReadAll(part)
			if readErr != nil {
				return "", "", nil, fmt.Errorf("failed to read file_data: %w", readErr)
			}
			payload = b
		default:
			_, _ = io.Copy(io.Discard, part)
		}
		_ = part.Close()
	}

	return shardID, providerID, payload, nil
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

	log.Printf("adapter shard download requested: provider=%q remote_id=%q", providerID, remoteID)

	provider, err := h.registry.Get(providerID)
	if err != nil {
		log.Printf("adapter shard download provider lookup failed: provider=%q remote_id=%q err=%v", providerID, remoteID, err)
		httpx.WriteError(w, http.StatusBadRequest, "Invalid provider", err)
		return
	}

	reader, err := provider.DownloadShard(r.Context(), remoteID)
	if err != nil {
		log.Printf("adapter shard download failed: provider=%q remote_id=%q err=%v", providerID, remoteID, err)
		lowerErr := strings.ToLower(err.Error())
		if errors.Is(err, adapter.ErrShardNotFound) ||
			strings.Contains(lowerErr, "not found") ||
			strings.Contains(lowerErr, "no such key") ||
			strings.Contains(lowerErr, "404") {
			httpx.WriteErrorWithCode(w, http.StatusNotFound, "Shard not found", "FILE_NOT_FOUND", err)
			return
		}
		httpx.WriteErrorWithCode(w, http.StatusInternalServerError, "Failed to download shard", httpx.ClassifyProviderError(err.Error()), err)
		return
	}
	defer reader.Close()

	shardData, err := io.ReadAll(reader)
	if err != nil {
		log.Printf("adapter shard download read failed: provider=%q remote_id=%q err=%v", providerID, remoteID, err)
		httpx.WriteError(w, http.StatusInternalServerError, "Failed to read shard data", err)
		return
	}

	log.Printf("adapter shard download completed: provider=%q remote_id=%q bytes=%d", providerID, remoteID, len(shardData))

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

// parseFileID extracts the file UUID from a shard ID of the form "{fileID}-shard-{index}".
func parseFileID(shardID string) string {
	if idx := strings.LastIndex(shardID, "-shard-"); idx != -1 {
		return shardID[:idx]
	}
	return shardID
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
