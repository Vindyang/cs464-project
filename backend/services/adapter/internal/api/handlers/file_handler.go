package handlers

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/vindyang/cs464-project/backend/services/adapter/internal/service"
)

// FileHandler handles HTTP requests for file operations
type FileHandler struct {
	service service.FileOperationsService
}

// NewFileHandler creates a new FileHandler instance
func NewFileHandler(service service.FileOperationsService) *FileHandler {
	return &FileHandler{
		service: service,
	}
}

// RegisterRoutes registers all file operation routes
func (h *FileHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/files/upload", h.UploadFile)
	mux.HandleFunc("/api/v1/files/", h.handleFileRoutes)
}

// handleFileRoutes routes file-specific endpoints
func (h *FileHandler) handleFileRoutes(w http.ResponseWriter, r *http.Request) {
	// Extract file ID from path
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/files/")
	parts := strings.Split(path, "/")

	if len(parts) == 0 || parts[0] == "" {
		// GET /api/v1/files - List files (not implemented yet)
		writeJSON(w, http.StatusNotImplemented, map[string]string{
			"error": "List files endpoint not yet implemented",
		})
		return
	}

	fileIDStr := parts[0]

	if len(parts) == 1 {
		// GET /api/v1/files/:fileId - Get file metadata
		if r.Method == http.MethodGet {
			h.GetFileMetadata(w, r, fileIDStr)
		} else if r.Method == http.MethodDelete {
			h.DeleteFile(w, r, fileIDStr)
		} else {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed"})
		}
		return
	}

	if len(parts) == 2 && parts[1] == "download" && r.Method == http.MethodGet {
		// GET /api/v1/files/:fileId/download
		h.DownloadFile(w, r, fileIDStr)
		return
	}

	writeJSON(w, http.StatusNotFound, map[string]string{"error": "Not found"})
}

// UploadFile handles POST /api/v1/files/upload
func (h *FileHandler) UploadFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed"})
		return
	}

	// Parse multipart form (max 100MB in memory)
	if err := r.ParseMultipartForm(100 << 20); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error":   "Failed to parse multipart form",
			"details": err.Error(),
		})
		return
	}

	// Get file from form
	file, header, err := r.FormFile("file")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error":   "No file provided",
			"details": err.Error(),
		})
		return
	}
	defer file.Close()

	// Read file data
	fileData, err := io.ReadAll(file)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error":   "Failed to read file",
			"details": err.Error(),
		})
		return
	}

	// Parse parameters
	kStr := r.FormValue("k")
	nStr := r.FormValue("n")
	chunkSizeMBStr := r.FormValue("chunk_size_mb")
	providersStr := r.FormValue("providers")

	// Validate required parameters
	if kStr == "" || nStr == "" || chunkSizeMBStr == "" || providersStr == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Missing required parameters: k, n, chunk_size_mb, providers",
		})
		return
	}

	// Parse integers
	k, err := strconv.Atoi(kStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error":   "Invalid k parameter",
			"details": err.Error(),
		})
		return
	}

	n, err := strconv.Atoi(nStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error":   "Invalid n parameter",
			"details": err.Error(),
		})
		return
	}

	chunkSizeMB, err := strconv.Atoi(chunkSizeMBStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error":   "Invalid chunk_size_mb parameter",
			"details": err.Error(),
		})
		return
	}

	// Parse providers (comma-separated)
	providers := strings.Split(providersStr, ",")
	for i := range providers {
		providers[i] = strings.TrimSpace(providers[i])
	}

	// Validate parameters
	if k <= 0 || n <= 0 || k > n {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Invalid erasure coding parameters: k must be > 0 and k <= n",
		})
		return
	}

	if chunkSizeMB <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "chunk_size_mb must be positive",
		})
		return
	}

	if len(providers) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "At least one provider must be specified",
		})
		return
	}

	// Call service
	resp, err := h.service.UploadFile(r.Context(), header.Filename, fileData, k, n, chunkSizeMB, providers)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error":   "Failed to upload file",
			"details": err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusCreated, resp)
}

// DownloadFile handles GET /api/v1/files/:fileId/download
func (h *FileHandler) DownloadFile(w http.ResponseWriter, r *http.Request, fileIDStr string) {
	// Parse file ID
	fileID, err := uuid.Parse(fileIDStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error":   "Invalid file ID format",
			"details": err.Error(),
		})
		return
	}

	// Download and reconstruct file
	fileData, filename, err := h.service.DownloadFile(r.Context(), fileID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error":   "Failed to download file",
			"details": err.Error(),
		})
		return
	}

	// Set headers for file download
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	w.Header().Set("Content-Length", strconv.Itoa(len(fileData)))

	// Write file data
	w.WriteHeader(http.StatusOK)
	w.Write(fileData)
}

// GetFileMetadata handles GET /api/v1/files/:fileId
func (h *FileHandler) GetFileMetadata(w http.ResponseWriter, r *http.Request, fileIDStr string) {
	// Parse file ID
	fileID, err := uuid.Parse(fileIDStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error":   "Invalid file ID format",
			"details": err.Error(),
		})
		return
	}

	// Get metadata
	metadata, err := h.service.GetFileMetadata(r.Context(), fileID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{
			"error":   "File not found",
			"details": err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, metadata)
}

// DeleteFile handles DELETE /api/v1/files/:fileId
func (h *FileHandler) DeleteFile(w http.ResponseWriter, r *http.Request, fileIDStr string) {
	// Parse file ID
	fileID, err := uuid.Parse(fileIDStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error":   "Invalid file ID format",
			"details": err.Error(),
		})
		return
	}

	// Check if we should delete shards too
	deleteShards := r.URL.Query().Get("delete_shards") == "true"

	// Delete file
	err = h.service.DeleteFile(r.Context(), fileID, deleteShards)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error":   "Failed to delete file",
			"details": err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success":        true,
		"message":        "File deleted successfully",
		"file_id":        fileIDStr,
		"shards_deleted": deleteShards,
	})
}
