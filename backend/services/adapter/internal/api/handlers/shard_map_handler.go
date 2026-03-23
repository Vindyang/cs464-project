package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/vindyang/cs464-project/backend/services/shared/api/dto"
	"github.com/vindyang/cs464-project/backend/services/shared/service"
)

// ShardMapHandler handles HTTP requests for shard map operations
type ShardMapHandler struct {
	service service.ShardMapService
}

// NewShardMapHandler creates a new ShardMapHandler instance
func NewShardMapHandler(service service.ShardMapService) *ShardMapHandler {
	return &ShardMapHandler{
		service: service,
	}
}

// RegisterRoutes registers all shard map routes
func (h *ShardMapHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/shards/register", h.RegisterFile)
	mux.HandleFunc("/api/v1/shards/record", h.RecordShards)
	mux.HandleFunc("/api/v1/shards/file/", h.GetShardMap)
	mux.HandleFunc("/api/v1/shards/", h.handleShardRoutes)
}

// handleShardRoutes routes between GetShardByID and MarkShardStatus
func (h *ShardMapHandler) handleShardRoutes(w http.ResponseWriter, r *http.Request) {
	// Extract shard ID from path
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/shards/")
	parts := strings.Split(path, "/")

	if len(parts) == 1 && r.Method == http.MethodGet {
		h.GetShardByID(w, r)
	} else if len(parts) == 2 && parts[1] == "status" && r.Method == http.MethodPut {
		h.MarkShardStatus(w, r)
	} else {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Not found"})
	}
}

// RegisterFile handles POST /api/v1/shards/register
func (h *ShardMapHandler) RegisterFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed"})
		return
	}

	var req dto.RegisterFileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Validate request
	if req.OriginalName == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "original_name is required"})
		return
	}

	if req.OriginalSize <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "original_size must be positive"})
		return
	}

	if req.TotalChunks <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "total_chunks must be positive"})
		return
	}

	if req.K <= 0 || req.N <= 0 || req.K > req.N {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid erasure coding parameters: K must be > 0, N must be >= K"})
		return
	}

	if req.ShardSize <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "shard_size must be positive"})
		return
	}

	// Call service
	resp, err := h.service.RegisterFile(&req)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error":   "Failed to register file",
			"details": err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusCreated, resp)
}

// RecordShards handles POST /api/v1/shards/record
func (h *ShardMapHandler) RecordShards(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed"})
		return
	}

	var req dto.RecordShardsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Validate request
	if req.FileID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "file_id is required"})
		return
	}

	if len(req.Shards) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "shards array cannot be empty"})
		return
	}

	// Validate each shard
	for i, shard := range req.Shards {
		if shard.ChunkIndex < 0 {
			writeJSON(w, http.StatusBadRequest, map[string]interface{}{
				"error":       "invalid chunk_index in shard",
				"shard_index": i,
			})
			return
		}

		if shard.ShardIndex < 0 {
			writeJSON(w, http.StatusBadRequest, map[string]interface{}{
				"error":       "invalid shard_index in shard",
				"shard_index": i,
			})
			return
		}

		if shard.Type != "DATA" && shard.Type != "PARITY" {
			writeJSON(w, http.StatusBadRequest, map[string]interface{}{
				"error":       "shard type must be DATA or PARITY",
				"shard_index": i,
			})
			return
		}

		if shard.RemoteID == "" {
			writeJSON(w, http.StatusBadRequest, map[string]interface{}{
				"error":       "remote_id is required in shard",
				"shard_index": i,
			})
			return
		}

		if shard.Provider == "" {
			writeJSON(w, http.StatusBadRequest, map[string]interface{}{
				"error":       "provider is required in shard",
				"shard_index": i,
			})
			return
		}

		if shard.ChecksumSHA256 == "" {
			writeJSON(w, http.StatusBadRequest, map[string]interface{}{
				"error":       "checksum_sha256 is required in shard",
				"shard_index": i,
			})
			return
		}
	}

	// Call service
	resp, err := h.service.RecordShards(&req)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error":   "Failed to record shards",
			"details": err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusCreated, resp)
}

// GetShardMap handles GET /api/v1/shards/file/:fileId
func (h *ShardMapHandler) GetShardMap(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed"})
		return
	}

	// Extract file ID from path
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/shards/file/")
	fileIDStr := strings.TrimSuffix(path, "/")

	fileID, err := uuid.Parse(fileIDStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error":   "Invalid file ID format",
			"details": err.Error(),
		})
		return
	}

	// Call service
	resp, err := h.service.GetShardMap(fileID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{
			"error":   "File not found",
			"details": err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

// GetShardByID handles GET /api/v1/shards/:shardId
func (h *ShardMapHandler) GetShardByID(w http.ResponseWriter, r *http.Request) {
	// Extract shard ID from path
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/shards/")
	shardIDStr := strings.TrimSuffix(path, "/")

	shardID, err := uuid.Parse(shardIDStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error":   "Invalid shard ID format",
			"details": err.Error(),
		})
		return
	}

	// Call service
	resp, err := h.service.GetShardByID(shardID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{
			"error":   "Shard not found",
			"details": err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

// MarkShardStatus handles PUT /api/v1/shards/:shardId/status
func (h *ShardMapHandler) MarkShardStatus(w http.ResponseWriter, r *http.Request) {
	// Extract shard ID from path
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/shards/")
	parts := strings.Split(path, "/")
	shardIDStr := parts[0]

	shardID, err := uuid.Parse(shardIDStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error":   "Invalid shard ID format",
			"details": err.Error(),
		})
		return
	}

	var req dto.MarkShardStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Validate status
	validStatuses := map[string]bool{
		"PENDING":   true,
		"HEALTHY":   true,
		"CORRUPTED": true,
		"MISSING":   true,
	}

	if !validStatuses[req.Status] {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error":   "Invalid status value",
			"details": "status must be one of: PENDING, HEALTHY, CORRUPTED, MISSING",
		})
		return
	}

	// Call service
	if err := h.service.MarkShardStatus(shardID, &req); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error":   "Failed to update shard status",
			"details": err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message":  "Shard status updated successfully",
		"shard_id": shardIDStr,
		"status":   req.Status,
	})
}

// writeJSON is a helper to write JSON responses
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
