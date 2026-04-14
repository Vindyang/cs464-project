package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/vindyang/cs464-project/backend/services/shardmap/internal/app"
	"github.com/vindyang/cs464-project/backend/services/shared/types"
)

// LifecycleHandler handles HTTP requests for lifecycle event operations.
type LifecycleHandler struct {
	service app.LifecycleService
}

// NewLifecycleHandler creates a new LifecycleHandler.
func NewLifecycleHandler(service app.LifecycleService) *LifecycleHandler {
	return &LifecycleHandler{service: service}
}

// RegisterRoutes registers lifecycle routes on the given mux.
//
//	POST /api/v1/lifecycle           — record a lifecycle event
//	GET  /api/v1/lifecycle           — get global lifecycle events
//	DELETE /api/v1/lifecycle         — delete all lifecycle events
//	GET  /api/v1/lifecycle/{fileId}  — get history for a file
func (h *LifecycleHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/lifecycle", h.handleRoot)
	mux.HandleFunc("/api/v1/lifecycle/", h.handleHistory)
}

// handleRoot dispatches POST/GET/DELETE /api/v1/lifecycle.
func (h *LifecycleHandler) handleRoot(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		resp, err := h.service.GetAllHistory()
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to get lifecycle history", "UNKNOWN_ERROR", err.Error())
			return
		}
		writeJSON(w, http.StatusOK, resp)
		return
	}

	if r.Method == http.MethodDelete {
		deleted, err := h.service.DeleteAllHistory()
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to delete lifecycle history", "UNKNOWN_ERROR", err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"deleted_events": deleted})
		return
	}

	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	var event types.LifecycleEvent
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error":   "invalid request body",
			"details": err.Error(),
		})
		return
	}

	if err := h.service.RecordEvent(&event); err != nil {
		writeError(w, http.StatusBadRequest, "failed to record lifecycle event", "UNKNOWN_ERROR", err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{"status": "recorded"})
}

// handleHistory handles GET /api/v1/lifecycle/{fileId}
func (h *LifecycleHandler) handleHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	fileID := strings.TrimPrefix(r.URL.Path, "/api/v1/lifecycle/")
	fileID = strings.TrimSuffix(fileID, "/")
	if fileID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "file_id is required in path"})
		return
	}

	resp, err := h.service.GetFileHistory(fileID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get file history", "UNKNOWN_ERROR", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, resp)
}
