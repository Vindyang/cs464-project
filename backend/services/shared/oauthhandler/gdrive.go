// Package oauthhandler implements HTTP handlers for the Google Drive OAuth2 flow.
package oauthhandler

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
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
	store       *db.Store
	registry    *adapter.Registry
	frontendURL string

	mu     sync.Mutex
	states map[string]stateEntry
}

// New constructs a GDriveHandler. The store is used for both credential loading
// and token persistence. No credentials are required at construction time —
// they are loaded lazily when the authorize endpoint is called.
func New(store *db.Store, registry *adapter.Registry) *GDriveHandler {
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:3000"
	}
	return &GDriveHandler{
		store:       store,
		registry:    registry,
		frontendURL: frontendURL,
		states:      make(map[string]stateEntry),
	}
}

// RestoreAdapter rebuilds and registers the GDrive adapter from a previously stored token.
// Called on server startup to reconnect without requiring the user to re-authenticate.
func (h *GDriveHandler) RestoreAdapter(registry *adapter.Registry, tok *oauth2.Token) error {
	oauthConfig, err := h.loadOAuthConfig()
	if err != nil {
		return fmt.Errorf("oauthhandler: restore adapter: %w", err)
	}
	gda, err := gdrive.NewGDriveAdapter(oauthConfig, tok, h.store)
	if err != nil {
		return fmt.Errorf("oauthhandler: restore adapter: %w", err)
	}
	registry.Register("googleDrive", gda)
	return nil
}

// Authorize handles GET /api/oauth/gdrive/authorize.
// Returns a JSON body: { "authURL": "https://accounts.google.com/..." }
func (h *GDriveHandler) Authorize(w http.ResponseWriter, r *http.Request) {
	oauthConfig, err := h.loadOAuthConfig()
	if err != nil {
		http.Error(w, "Google OAuth credentials not configured: "+err.Error(), http.StatusBadRequest)
		return
	}

	state, err := generateState()
	if err != nil {
		http.Error(w, "failed to generate state", http.StatusInternalServerError)
		return
	}

	h.mu.Lock()
	h.states[state] = stateEntry{expiry: time.Now().Add(10 * time.Minute)}
	h.mu.Unlock()

	authURL := oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.ApprovalForce)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"authURL": authURL})
}

// Callback handles GET /api/oauth/gdrive/callback.
// Exchanges the authorization code for a token, stores it in SQLite,
// registers the adapter, then redirects to the frontend.
func (h *GDriveHandler) Callback(w http.ResponseWriter, r *http.Request) {
	state := r.URL.Query().Get("state")
	code := r.URL.Query().Get("code")

	if !h.consumeState(state) {
		http.Error(w, "invalid or expired state", http.StatusBadRequest)
		return
	}

	oauthConfig, err := h.loadOAuthConfig()
	if err != nil {
		log.Printf("oauthhandler: load credentials: %v", err)
		http.Redirect(w, r, h.frontendURL+"/providers?error=credentials_missing", http.StatusFound)
		return
	}

	tok, err := oauthConfig.Exchange(r.Context(), code)
	if err != nil {
		log.Printf("oauthhandler: token exchange failed: %v", err)
		http.Redirect(w, r, h.frontendURL+"/providers?error=oauth_failed", http.StatusFound)
		return
	}

	if err := h.store.UpsertToken("googleDrive", tok); err != nil {
		log.Printf("oauthhandler: save token: %v", err)
		http.Redirect(w, r, h.frontendURL+"/providers?error=save_failed", http.StatusFound)
		return
	}

	gda, err := gdrive.NewGDriveAdapter(oauthConfig, tok, h.store)
	if err != nil {
		log.Printf("oauthhandler: init adapter: %v", err)
		http.Redirect(w, r, h.frontendURL+"/providers?error=adapter_failed", http.StatusFound)
		return
	}
	h.registry.Register("googleDrive", gda)

	http.Redirect(w, r, h.frontendURL+"/providers?connected=googleDrive", http.StatusFound)
}

// Disconnect handles POST /api/oauth/gdrive/disconnect.
// Removes the token from SQLite and unregisters the adapter.
func (h *GDriveHandler) Disconnect(w http.ResponseWriter, r *http.Request) {
	if err := h.store.DeleteToken("googleDrive"); err != nil {
		log.Printf("oauthhandler: delete token: %v", err)
		http.Error(w, "failed to disconnect", http.StatusInternalServerError)
		return
	}
	h.registry.Unregister("googleDrive")
	w.WriteHeader(http.StatusNoContent)
}

// loadOAuthConfig builds an oauth2.Config from stored credentials, falling back
// to environment variables if no runtime credentials have been configured.
func (h *GDriveHandler) loadOAuthConfig() (*oauth2.Config, error) {
	raw, redirectURI, err := h.loadCredentialsJSON()
	if err != nil {
		return nil, err
	}

	config, err := google.ConfigFromJSON(raw, driveScope)
	if err != nil {
		return nil, fmt.Errorf("parse credentials: %w", err)
	}
	config.RedirectURL = redirectURI
	return config, nil
}

// loadCredentialsJSON returns the raw GCP OAuth2 credentials JSON and redirect URI.
// Priority: SQLite store → GDRIVE_OAUTH_CREDENTIALS_JSON env
func (h *GDriveHandler) loadCredentialsJSON() (raw []byte, redirectURI string, err error) {
	// 1. SQLite store (runtime-configured via PUT /api/credentials/googleDrive)
	clientID, clientSecret, redirectURI, storeErr := h.store.LoadCredentials("googleDrive")
	if storeErr == nil {
		// Build the credentials JSON that google.ConfigFromJSON expects.
		creds := map[string]interface{}{
			"installed": map[string]interface{}{
				"client_id":                   clientID,
				"client_secret":               clientSecret,
				"redirect_uris":               []string{redirectURI},
				"auth_uri":                    "https://accounts.google.com/o/oauth2/auth",
				"token_uri":                   "https://oauth2.googleapis.com/token",
				"auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
			},
		}
		raw, err = json.Marshal(creds)
		if err != nil {
			return nil, "", fmt.Errorf("marshal credentials: %w", err)
		}
		return raw, redirectURI, nil
	}
	if !errors.Is(storeErr, db.ErrNotFound) {
		return nil, "", fmt.Errorf("load credentials from store: %w", storeErr)
	}

	// 2. Env fallback: raw JSON string
	if envJSON := os.Getenv("GDRIVE_OAUTH_CREDENTIALS_JSON"); envJSON != "" {
		redirectURI = os.Getenv("GDRIVE_OAUTH_REDIRECT_URI")
		return []byte(envJSON), redirectURI, nil
	}

	return nil, "", fmt.Errorf("no Google OAuth credentials configured — use PUT /api/credentials/googleDrive or set env vars")
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

func generateState() (string, error) {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
