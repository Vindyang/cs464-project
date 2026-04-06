package handlers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/vindyang/cs464-project/backend/services/shared/adapter"
	"github.com/vindyang/cs464-project/backend/services/shared/db"
	"github.com/vindyang/cs464-project/backend/services/shared/s3handler"
	"github.com/vindyang/cs464-project/backend/services/shared/transport/httpx"
)

// CredentialsHandler handles CRUD operations for provider OAuth client credentials.
type CredentialsHandler struct {
	store    *db.Store
	registry *adapter.Registry
}

type CredentialResetSummary struct {
	DeletedCredentials    int `json:"deleted_credentials"`
	DeletedTokens         int `json:"deleted_tokens"`
	DisconnectedProviders int `json:"disconnected_providers"`
}

func NewCredentialsHandler(store *db.Store, registry *adapter.Registry) *CredentialsHandler {
	return &CredentialsHandler{store: store, registry: registry}
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
	id, action := credentialPathParts(r)
	if action == "secret" {
		if r.Method != http.MethodGet {
			httpx.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
			return
		}
		h.getSecret(w, id)
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.get(w, id)
	case http.MethodPut:
		h.upsert(w, r, id)
	case http.MethodDelete:
		h.delete(w, id)
	default:
		httpx.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
	}
}

func credentialPathParts(r *http.Request) (string, string) {
	// r.PathValue works with Go 1.22+ pattern routing.
	// Fall back to trimming the prefix for mux.HandleFunc-style routing.
	if id := r.PathValue("provider"); id != "" {
		return id, ""
	}
	path := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/credentials/"), "/")
	if path == "" {
		return "", ""
	}
	parts := strings.Split(path, "/")
	id := parts[0]
	if len(parts) > 1 {
		return id, parts[1]
	}
	return id, ""
}

// GET /api/credentials/{provider}
// Returns client_id and redirect_uri only — never exposes client_secret.
func (h *CredentialsHandler) get(w http.ResponseWriter, id string) {
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

// GET /api/credentials/{provider}/secret
// Returns the full stored credential for explicit reveal-on-demand flows.
func (h *CredentialsHandler) getSecret(w http.ResponseWriter, id string) {
	if id == "" {
		httpx.WriteError(w, http.StatusBadRequest, "provider is required", nil)
		return
	}

	clientID, clientSecret, redirectURI, err := h.store.LoadCredentials(id)
	if errors.Is(err, db.ErrNotFound) {
		httpx.WriteError(w, http.StatusNotFound, "no credentials configured for provider", nil)
		return
	}
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to load credentials", err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]string{
		"provider_id":   id,
		"client_id":     clientID,
		"client_secret": clientSecret,
		"redirect_uri":  redirectURI,
	})
}

// PUT /api/credentials/{provider}
// Body: { "client_id": "...", "client_secret": "...", "redirect_uri": "..." }
func (h *CredentialsHandler) upsert(w http.ResponseWriter, r *http.Request, id string) {
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
func (h *CredentialsHandler) delete(w http.ResponseWriter, id string) {
	if id == "" {
		httpx.WriteError(w, http.StatusBadRequest, "provider is required", nil)
		return
	}

	if err := h.deleteProviderData(id); err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to delete credentials", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *CredentialsHandler) DeleteAllCredentials() (CredentialResetSummary, error) {
	records, err := h.store.ListCredentials()
	if err != nil {
		return CredentialResetSummary{}, err
	}

	deletedCredentials, err := h.store.DeleteAllCredentials()
	if err != nil {
		return CredentialResetSummary{}, err
	}

	deletedTokens, err := h.store.DeleteAllTokens()
	if err != nil {
		return CredentialResetSummary{}, err
	}

	disconnected := 0
	if h.registry != nil {
		for _, record := range records {
			h.registry.Unregister(record.ProviderID)
			disconnected++
		}
	}

	return CredentialResetSummary{
		DeletedCredentials:    deletedCredentials,
		DeletedTokens:         deletedTokens,
		DisconnectedProviders: disconnected,
	}, nil
}

func (h *CredentialsHandler) deleteProviderData(id string) error {
	if err := h.store.DeleteCredentials(id); err != nil {
		return err
	}
	if err := h.store.DeleteToken(id); err != nil {
		return err
	}
	if h.registry != nil {
		h.registry.Unregister(id)
	}
	return nil
}
