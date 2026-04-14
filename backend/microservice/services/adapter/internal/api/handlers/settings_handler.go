package handlers

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/vindyang/cs464-project/backend/services/shared/db"
	"github.com/vindyang/cs464-project/backend/services/shared/transport/httpx"
)

type SettingsHandler struct {
	store        *db.Store
	fileResetter interface {
		DeleteAllFiles(context.Context, bool) (FileDeleteSummary, error)
	}
	credentialResetter interface {
		DeleteAllCredentials() (CredentialResetSummary, error)
	}
	lifecycleResetter interface {
		DeleteAllHistory(context.Context) (int, error)
	}
}

func NewSettingsHandler(
	store *db.Store,
	fileResetter interface {
		DeleteAllFiles(context.Context, bool) (FileDeleteSummary, error)
	},
	credentialResetter interface {
		DeleteAllCredentials() (CredentialResetSummary, error)
	},
	lifecycleResetter interface {
		DeleteAllHistory(context.Context) (int, error)
	},
) *SettingsHandler {
	return &SettingsHandler{store: store, fileResetter: fileResetter, credentialResetter: credentialResetter, lifecycleResetter: lifecycleResetter}
}

func (h *SettingsHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/settings", h.route)
	mux.HandleFunc("/api/settings/reset", h.reset)
}

func (h *SettingsHandler) route(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.get(w, r)
	case http.MethodPut:
		h.put(w, r)
	default:
		httpx.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
	}
}

func (h *SettingsHandler) reset(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}

	var body struct {
		Scope        string `json:"scope"`
		DeleteShards *bool  `json:"delete_shards,omitempty"`
	}
	if err := httpx.DecodeJSON(r, &body, 1<<20); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", err)
		return
	}

	deleteShards := true
	if body.DeleteShards != nil {
		deleteShards = *body.DeleteShards
	}

	response := map[string]any{
		"scope":         body.Scope,
		"delete_shards": deleteShards,
	}

	switch body.Scope {
	case "files":
		if h.fileResetter == nil {
			httpx.WriteError(w, http.StatusNotImplemented, "file reset is not configured", nil)
			return
		}
		fileSummary, err := h.fileResetter.DeleteAllFiles(r.Context(), deleteShards)
		if err != nil {
			httpx.WriteErrorWithCode(w, http.StatusInternalServerError, "failed to delete file data", "UNKNOWN_ERROR", err)
			return
		}
		response["file_summary"] = fileSummary
	case "credentials":
		if h.credentialResetter == nil {
			httpx.WriteError(w, http.StatusNotImplemented, "credential reset is not configured", nil)
			return
		}
		credentialSummary, err := h.credentialResetter.DeleteAllCredentials()
		if err != nil {
			httpx.WriteErrorWithCode(w, http.StatusInternalServerError, "failed to delete credentials", "UNKNOWN_ERROR", err)
			return
		}
		response["credential_summary"] = credentialSummary
	case "all_data":
		if h.fileResetter == nil || h.credentialResetter == nil || h.lifecycleResetter == nil {
			httpx.WriteError(w, http.StatusNotImplemented, "full reset is not configured", nil)
			return
		}
		fileSummary, err := h.fileResetter.DeleteAllFiles(r.Context(), deleteShards)
		if err != nil {
			httpx.WriteErrorWithCode(w, http.StatusInternalServerError, "failed to delete file data", "UNKNOWN_ERROR", err)
			return
		}
		credentialSummary, err := h.credentialResetter.DeleteAllCredentials()
		if err != nil {
			httpx.WriteErrorWithCode(w, http.StatusInternalServerError, "failed to delete credentials", "UNKNOWN_ERROR", err)
			return
		}
		deletedEvents, err := h.lifecycleResetter.DeleteAllHistory(r.Context())
		if err != nil {
			httpx.WriteErrorWithCode(w, http.StatusInternalServerError, "failed to delete lifecycle history", "UNKNOWN_ERROR", err)
			return
		}
		response["file_summary"] = fileSummary
		response["credential_summary"] = credentialSummary
		response["lifecycle_summary"] = map[string]int{"deleted_events": deletedEvents}
	default:
		httpx.WriteError(w, http.StatusBadRequest, "invalid reset scope", nil)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, response)
}

func (h *SettingsHandler) get(w http.ResponseWriter, _ *http.Request) {
	redundancy, err := h.store.GetConfig("settings_redundancy")
	if errors.Is(err, db.ErrNotFound) {
		redundancy = "(6,4)"
	} else if err != nil {
		httpx.WriteErrorWithCode(w, http.StatusInternalServerError, "failed to load settings", "UNKNOWN_ERROR", err)
		return
	}

	encryptDefault, err := h.getBoolConfig("settings_encrypt_default", true)
	if err != nil {
		httpx.WriteErrorWithCode(w, http.StatusInternalServerError, "failed to load settings", "UNKNOWN_ERROR", err)
		return
	}

	autoDelete, err := h.getBoolConfig("settings_auto_delete", false)
	if err != nil {
		httpx.WriteErrorWithCode(w, http.StatusInternalServerError, "failed to load settings", "UNKNOWN_ERROR", err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"redundancy":      redundancy,
		"encrypt_default": encryptDefault,
		"auto_delete":     autoDelete,
	})
}

func (h *SettingsHandler) put(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Redundancy     string `json:"redundancy"`
		EncryptDefault *bool  `json:"encrypt_default"`
		AutoDelete     *bool  `json:"auto_delete"`
	}
	if err := httpx.DecodeJSON(r, &body, 1<<20); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", err)
		return
	}

	if body.Redundancy != "(6,4)" && body.Redundancy != "(8,4)" && body.Redundancy != "(10,8)" {
		httpx.WriteError(w, http.StatusBadRequest, "invalid redundancy value", nil)
		return
	}
	if body.EncryptDefault == nil || body.AutoDelete == nil {
		httpx.WriteError(w, http.StatusBadRequest, "encrypt_default and auto_delete are required", nil)
		return
	}

	if err := h.store.SetConfig("settings_redundancy", body.Redundancy); err != nil {
		httpx.WriteErrorWithCode(w, http.StatusInternalServerError, "failed to save settings", "UNKNOWN_ERROR", err)
		return
	}
	if err := h.store.SetConfig("settings_encrypt_default", strconv.FormatBool(*body.EncryptDefault)); err != nil {
		httpx.WriteErrorWithCode(w, http.StatusInternalServerError, "failed to save settings", "UNKNOWN_ERROR", err)
		return
	}
	if err := h.store.SetConfig("settings_auto_delete", strconv.FormatBool(*body.AutoDelete)); err != nil {
		httpx.WriteErrorWithCode(w, http.StatusInternalServerError, "failed to save settings", "UNKNOWN_ERROR", err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "settings saved"})
}

func (h *SettingsHandler) getBoolConfig(key string, fallback bool) (bool, error) {
	raw, err := h.store.GetConfig(key)
	if errors.Is(err, db.ErrNotFound) {
		return fallback, nil
	}
	if err != nil {
		return false, err
	}
	v, err := strconv.ParseBool(raw)
	if err != nil {
		return fallback, nil
	}
	return v, nil
}
