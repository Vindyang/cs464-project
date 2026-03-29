// Package oauthhandler implements HTTP handlers for the Google Drive OAuth2 flow.
package oauthhandler

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/vindyang/cs464-project/backend/services/shared/adapter"
	"github.com/vindyang/cs464-project/backend/services/shared/adapter/gdrive"
	"github.com/vindyang/cs464-project/backend/services/shared/db"
)

const driveScope = "https://www.googleapis.com/auth/drive.file"

// stateEntry holds a pending OAuth state token with an expiry.
type stateEntry struct {
	expiry time.Time
}

// GDriveHandler handles the three Google Drive OAuth endpoints.
type GDriveHandler struct {
	db          *db.DB
	registry    *adapter.Registry
	oauthConfig *oauth2.Config
	folderID    string
	frontendURL string

	mu     sync.Mutex
	states map[string]stateEntry
}

// New constructs a GDriveHandler. Returns an error if credentials cannot be loaded.
func New(database *db.DB, registry *adapter.Registry) (*GDriveHandler, error) {
	raw, err := loadGDriveCredentials()
	if err != nil {
		return nil, fmt.Errorf("oauthhandler: %w", err)
	}

	config, err := google.ConfigFromJSON(raw, driveScope)
	if err != nil {
		return nil, fmt.Errorf("oauthhandler: parse credentials: %w", err)
	}

	redirectURI := os.Getenv("GDRIVE_OAUTH_REDIRECT_URI")
	if redirectURI == "" {
		return nil, fmt.Errorf("oauthhandler: GDRIVE_OAUTH_REDIRECT_URI not set")
	}
	config.RedirectURL = redirectURI

	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:3000"
	}

	folderID := os.Getenv("GDRIVE_FOLDER_ID")

	return &GDriveHandler{
		db:          database,
		registry:    registry,
		oauthConfig: config,
		folderID:    folderID,
		frontendURL: frontendURL,
		states:      make(map[string]stateEntry),
	}, nil
}

// Authorize handles GET /api/oauth/gdrive/authorize.
// Returns a JSON body: { "authURL": "https://accounts.google.com/..." }
func (h *GDriveHandler) Authorize(w http.ResponseWriter, r *http.Request) {
	state, err := generateState()
	if err != nil {
		http.Error(w, "failed to generate state", http.StatusInternalServerError)
		return
	}

	h.mu.Lock()
	h.states[state] = stateEntry{expiry: time.Now().Add(10 * time.Minute)}
	h.mu.Unlock()

	authURL := h.oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.ApprovalForce)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"authURL": authURL})
}

// Callback handles GET /api/oauth/gdrive/callback.
// Exchanges the authorization code for a token, stores it in Supabase,
// registers the adapter, then redirects to the frontend.
func (h *GDriveHandler) Callback(w http.ResponseWriter, r *http.Request) {
	state := r.URL.Query().Get("state")
	code := r.URL.Query().Get("code")

	if !h.consumeState(state) {
		http.Error(w, "invalid or expired state", http.StatusBadRequest)
		return
	}

	tok, err := h.oauthConfig.Exchange(r.Context(), code)
	if err != nil {
		log.Printf("oauthhandler: token exchange failed: %v", err)
		http.Redirect(w, r, h.frontendURL+"/providers?error=oauth_failed", http.StatusFound)
		return
	}

	if err := h.db.UpsertProviderToken(r.Context(), "googleDrive", tok); err != nil {
		log.Printf("oauthhandler: save token: %v", err)
		http.Redirect(w, r, h.frontendURL+"/providers?error=save_failed", http.StatusFound)
		return
	}

	gda, err := gdrive.NewGDriveAdapter(h.oauthConfig, tok, h.folderID)
	if err != nil {
		log.Printf("oauthhandler: init adapter: %v", err)
		http.Redirect(w, r, h.frontendURL+"/providers?error=adapter_failed", http.StatusFound)
		return
	}
	h.registry.Register("googleDrive", gda)

	http.Redirect(w, r, h.frontendURL+"/providers?connected=googleDrive", http.StatusFound)
}

// Disconnect handles POST /api/oauth/gdrive/disconnect.
// Removes the token from Supabase and unregisters the adapter.
func (h *GDriveHandler) Disconnect(w http.ResponseWriter, r *http.Request) {
	if err := h.db.DeleteProviderToken(r.Context(), "googleDrive"); err != nil {
		log.Printf("oauthhandler: delete token: %v", err)
		http.Error(w, "failed to disconnect", http.StatusInternalServerError)
		return
	}
	h.registry.Unregister("googleDrive")
	w.WriteHeader(http.StatusNoContent)
}

// consumeState validates and removes a state token (one-time use).
func (h *GDriveHandler) consumeState(state string) bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	entry, ok := h.states[state]
	if !ok || time.Now().After(entry.expiry) {
		delete(h.states, state)
		return false
	}
	delete(h.states, state)
	return true
}

// loadGDriveCredentials returns the raw GCP OAuth2 credentials JSON.
// It prefers GDRIVE_OAUTH_CREDENTIALS_JSON (raw JSON, for cloud/env-only deploys)
// and falls back to GDRIVE_OAUTH_CREDENTIALS_FILE (file path, for local dev).
func loadGDriveCredentials() ([]byte, error) {
	if raw := os.Getenv("GDRIVE_OAUTH_CREDENTIALS_JSON"); raw != "" {
		return []byte(raw), nil
	}
	credsFile := os.Getenv("GDRIVE_OAUTH_CREDENTIALS_FILE")
	if credsFile == "" {
		return nil, fmt.Errorf("neither GDRIVE_OAUTH_CREDENTIALS_JSON nor GDRIVE_OAUTH_CREDENTIALS_FILE is set")
	}
	return os.ReadFile(credsFile)
}

func generateState() (string, error) {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
