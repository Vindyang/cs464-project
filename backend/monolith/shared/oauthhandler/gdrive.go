package oauthhandler

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/vindyang/cs464-project/backend/monolith/shared/adapter"
	"github.com/vindyang/cs464-project/backend/monolith/shared/adapter/gdrive"
	"github.com/vindyang/cs464-project/backend/monolith/shared/db"
)

const driveScope = "https://www.googleapis.com/auth/drive.file"
const frontendURL = "http://localhost:3000"

type stateEntry struct {
	expiry time.Time
}

type GDriveHandler struct {
	store       *db.Store
	registry    *adapter.Registry
	frontendURL string

	mu     sync.Mutex
	states map[string]stateEntry
}

func New(store *db.Store, registry *adapter.Registry) *GDriveHandler {
	return &GDriveHandler{store: store, registry: registry, frontendURL: frontendURL, states: make(map[string]stateEntry)}
}

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

func (h *GDriveHandler) Disconnect(w http.ResponseWriter, r *http.Request) {
	if err := h.store.DeleteToken("googleDrive"); err != nil {
		log.Printf("oauthhandler: delete token: %v", err)
		http.Error(w, "failed to disconnect", http.StatusInternalServerError)
		return
	}
	h.registry.Unregister("googleDrive")
	w.WriteHeader(http.StatusNoContent)
}

func (h *GDriveHandler) loadOAuthConfig() (*oauth2.Config, error) {
	clientID, clientSecret, redirectURI, err := h.loadCredentials()
	if err != nil {
		return nil, err
	}
	return &oauth2.Config{ClientID: clientID, ClientSecret: clientSecret, RedirectURL: redirectURI, Scopes: []string{driveScope}, Endpoint: google.Endpoint}, nil
}

func (h *GDriveHandler) loadCredentials() (clientID, clientSecret, redirectURI string, err error) {
	clientID, clientSecret, redirectURI, err = h.store.LoadCredentials("googleDrive")
	if err != nil {
		return "", "", "", fmt.Errorf("load credentials from store: %w", err)
	}
	return clientID, clientSecret, redirectURI, nil
}

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
