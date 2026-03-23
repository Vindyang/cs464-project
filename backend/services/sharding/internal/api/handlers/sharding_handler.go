package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/vindyang/cs464-project/backend/services/sharding/internal/service"
)

// ShardingHandler exposes sharding and reconstruction endpoints.
type ShardingHandler struct {
	service service.ShardingService
}

// NewShardingHandler creates a new ShardingHandler.
func NewShardingHandler(service service.ShardingService) *ShardingHandler {
	return &ShardingHandler{service: service}
}

// RegisterRoutes registers all sharding routes.
func (h *ShardingHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/health", h.Health)
	mux.HandleFunc("/shard", h.Shard)
	mux.HandleFunc("/reconstruct", h.Reconstruct)
}

type shardRequest struct {
	FileID   string `json:"file_id"`
	FileData []byte `json:"file_data"`
	N        int    `json:"N"`
	K        int    `json:"K"`
}

type shardResponse struct {
	FileID   string            `json:"file_id"`
	Shards   []shardOutput     `json:"shards"`
	Metadata shardResponseMeta `json:"metadata"`
}

type shardOutput struct {
	ShardIndex int    `json:"shard_index"`
	ShardType  string `json:"shard_type"`
	ShardData  []byte `json:"shard_data"`
}

type shardResponseMeta struct {
	N            int `json:"N"`
	K            int `json:"K"`
	OriginalSize int `json:"original_size"`
	ShardSize    int `json:"shard_size"`
}

type reconstructRequest struct {
	FileID          string                 `json:"file_id"`
	N               int                    `json:"N"`
	K               int                    `json:"K"`
	AvailableShards []reconstructShardData `json:"available_shards"`
}

type reconstructShardData struct {
	ShardIndex int    `json:"shard_index"`
	ShardType  string `json:"shard_type"`
	ShardData  []byte `json:"shard_data"`
}

type reconstructResponse struct {
	FileID            string                  `json:"file_id"`
	ReconstructedFile []byte                  `json:"reconstructed_file"`
	Metadata          reconstructResponseMeta `json:"metadata"`
}

type reconstructResponseMeta struct {
	OriginalSize         int    `json:"original_size"`
	ShardsUsed           int    `json:"shards_used"`
	ReconstructionMethod string `json:"reconstruction_method"`
}

// Health handles GET /health.
func (h *ShardingHandler) Health(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"status":  "healthy",
		"service": "sharding",
	})
}

// Shard handles POST /shard.
func (h *ShardingHandler) Shard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed"})
		return
	}

	var req shardRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body", "details": err.Error()})
		return
	}

	if _, err := uuid.Parse(req.FileID); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "file_id must be a valid UUID", "details": err.Error()})
		return
	}

	if req.K <= 0 || req.N <= 0 || req.K > req.N {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid erasure coding parameters: K must be > 0 and K <= N"})
		return
	}

	if len(req.FileData) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "file_data is required"})
		return
	}

	shards, err := h.service.EncodeChunk(req.FileData, req.K, req.N)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to shard data", "details": err.Error()})
		return
	}

	outputs := make([]shardOutput, 0, len(shards))
	for i, shard := range shards {
		shardType := "data"
		if i >= req.K {
			shardType = "parity"
		}

		outputs = append(outputs, shardOutput{
			ShardIndex: i,
			ShardType:  shardType,
			ShardData:  shard,
		})
	}

	resp := shardResponse{
		FileID: req.FileID,
		Shards: outputs,
		Metadata: shardResponseMeta{
			N:            req.N,
			K:            req.K,
			OriginalSize: len(req.FileData),
			ShardSize:    len(shards[0]),
		},
	}

	writeJSON(w, http.StatusOK, resp)
}

// Reconstruct handles POST /reconstruct.
func (h *ShardingHandler) Reconstruct(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed"})
		return
	}

	var req reconstructRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body", "details": err.Error()})
		return
	}

	if _, err := uuid.Parse(req.FileID); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "file_id must be a valid UUID", "details": err.Error()})
		return
	}

	if req.K <= 0 || req.N <= 0 || req.K > req.N {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid erasure coding parameters: K must be > 0 and K <= N"})
		return
	}

	if len(req.AvailableShards) < req.K {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("insufficient available shards: have %d, need at least %d", len(req.AvailableShards), req.K)})
		return
	}

	shards := make([][]byte, req.N)
	hasParity := false
	seen := map[int]bool{}

	for _, shard := range req.AvailableShards {
		if shard.ShardIndex < 0 || shard.ShardIndex >= req.N {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("shard_index %d is out of range [0, %d)", shard.ShardIndex, req.N)})
			return
		}

		if seen[shard.ShardIndex] {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("duplicate shard_index %d", shard.ShardIndex)})
			return
		}
		seen[shard.ShardIndex] = true

		st := strings.ToLower(shard.ShardType)
		if st != "data" && st != "parity" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "shard_type must be either data or parity"})
			return
		}
		if st == "parity" {
			hasParity = true
		}

		if len(shard.ShardData) == 0 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "shard_data cannot be empty"})
			return
		}

		shards[shard.ShardIndex] = shard.ShardData
	}

	reconstructed, err := h.service.DecodeChunk(shards, req.K, req.N)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to reconstruct data", "details": err.Error()})
		return
	}

	method := "data"
	if hasParity {
		method = "parity"
	}

	resp := reconstructResponse{
		FileID:            req.FileID,
		ReconstructedFile: reconstructed,
		Metadata: reconstructResponseMeta{
			OriginalSize:         len(reconstructed),
			ShardsUsed:           len(req.AvailableShards),
			ReconstructionMethod: method,
		},
	}

	writeJSON(w, http.StatusOK, resp)
}
