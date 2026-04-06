package handlers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/vindyang/cs464-project/backend/services/shared/db"
	"github.com/vindyang/cs464-project/backend/services/shared/s3handler"
	"github.com/vindyang/cs464-project/backend/services/shared/transport/httpx"
)

// CredentialsHandler handles CRUD operations for provider OAuth client credentials.
type CredentialsHandler struct {
	store *db.Store
}

func NewCredentialsHandler(store *db.Store) *CredentialsHandler {
	return &CredentialsHandler{store: store}
}

func (h *CredentialsHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/credentials/status", h.status)
	mux.HandleFunc("/api/credentials", h.collection)
	mux.HandleFunc("/api/credentials/", h.route)
}

// collection handles GET /api/credentials.
func (h *CredentialsHandler) collection(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}

	records, err := h.store.ListCredentials()
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to list credentials", err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, records)
}

// status handles GET /api/credentials/status.
func (h *CredentialsHandler) status(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}

	records, err := h.store.ListCredentials()
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to list credentials status", err)
		return
	}

	providers := make([]string, 0, len(records))
	for _, rec := range records {
		providers = append(providers, rec.ProviderID)
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"configured": len(records) > 0,
		"count":      len(records),
		"providers":  providers,
	})
}

// route dispatches to the correct method handler based on HTTP verb.
func (h *CredentialsHandler) route(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.get(w, r)
	case http.MethodPut:
		h.upsert(w, r)
	case http.MethodDelete:
		h.delete(w, r)
	default:
		httpx.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
	}
}

// providerID extracts the provider segment from /api/credentials/{provider}.
func providerID(r *http.Request) string {
	// r.PathValue works with Go 1.22+ pattern routing.
	// Fall back to trimming the prefix for mux.HandleFunc-style routing.
	if id := r.PathValue("provider"); id != "" {
		return id
	}
	path := r.URL.Path
	const prefix = "/api/credentials/"
	if len(path) > len(prefix) {
		return path[len(prefix):]
	}
	return ""
}

// GET /api/credentials/{provider}
// Returns client_id and redirect_uri only — never exposes client_secret.
func (h *CredentialsHandler) get(w http.ResponseWriter, r *http.Request) {
	id := providerID(r)
	if id == "" {
		httpx.WriteError(w, http.StatusBadRequest, "provider is required", nil)
		return
	}

	clientID, _, redirectURI, err := h.store.LoadCredentials(id)
	if errors.Is(err, db.ErrNotFound) {
		httpx.WriteError(w, http.StatusNotFound, "no credentials configured for provider", nil)
		return
	}
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to load credentials", err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]string{
		"provider_id":  id,
		"client_id":    clientID,
		"redirect_uri": redirectURI,
	})
}

// PUT /api/credentials/{provider}
// Body: { "client_id": "...", "client_secret": "...", "redirect_uri": "..." }
func (h *CredentialsHandler) upsert(w http.ResponseWriter, r *http.Request) {
	id := providerID(r)
	if id == "" {
		httpx.WriteError(w, http.StatusBadRequest, "provider is required", nil)
		return
	}

	var body struct {
		ClientID     string `json:"client_id"`
		ClientSecret string `json:"client_secret"`
		RedirectURI  string `json:"redirect_uri"`
	}
	if err := httpx.DecodeJSON(r, &body, 1<<20); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", err)
		return
	}
	body.ClientID = strings.TrimSpace(body.ClientID)
	body.ClientSecret = strings.TrimSpace(body.ClientSecret)
	body.RedirectURI = strings.TrimSpace(body.RedirectURI)
	if body.ClientID == "" || body.ClientSecret == "" || body.RedirectURI == "" {
		httpx.WriteError(w, http.StatusBadRequest, "client_id, client_secret, and redirect_uri are required", nil)
		return
	}
	if id == "awsS3" {
		if err := s3handler.ValidateRegion(body.RedirectURI); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, "invalid AWS region", err)
			return
		}
	}

	if err := h.store.UpsertCredentials(id, body.ClientID, body.ClientSecret, body.RedirectURI); err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to save credentials", err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]string{
		"provider_id": id,
		"status":      "credentials saved",
	})
}

// DELETE /api/credentials/{provider}
func (h *CredentialsHandler) delete(w http.ResponseWriter, r *http.Request) {
	id := providerID(r)
	if id == "" {
		httpx.WriteError(w, http.StatusBadRequest, "provider is required", nil)
		return
	}

	if err := h.store.DeleteCredentials(id); err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to delete credentials", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
