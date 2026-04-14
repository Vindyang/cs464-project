package onedrivehandler

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
	"golang.org/x/oauth2/microsoft"

	"github.com/vindyang/cs464-project/backend/monolith/shared/adapter"
	"github.com/vindyang/cs464-project/backend/monolith/shared/adapter/onedrive"
	"github.com/vindyang/cs464-project/backend/monolith/shared/db"
)

var oneDriveScopes = []string{"Files.ReadWrite.All", "User.Read", "offline_access"}

type stateEntry struct {
	expiry time.Time
}

type OneDriveHandler struct {
	store       *db.Store
	registry    *adapter.Registry
	frontendURL string

	mu     sync.Mutex
	states map[string]stateEntry
}

func New(store *db.Store, registry *adapter.Registry) *OneDriveHandler {
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:3000"
	}
	return &OneDriveHandler{store: store, registry: registry, frontendURL: frontendURL, states: make(map[string]stateEntry)}
}

func (h *OneDriveHandler) RestoreAdapter(registry *adapter.Registry, tok *oauth2.Token) error {
	oauthConfig, err := h.loadOAuthConfig()
	if err != nil {
		return fmt.Errorf("onedrivehandler: restore adapter: %w", err)
	}
	oda, err := onedrive.NewOneDriveAdapter(oauthConfig, tok, h.store)
	if err != nil {
		return fmt.Errorf("onedrivehandler: restore adapter: %w", err)
	}
	registry.Register("oneDrive", oda)
	return nil
}

func (h *OneDriveHandler) Authorize(w http.ResponseWriter, r *http.Request) {
	oauthConfig, err := h.loadOAuthConfig()
	if err != nil {
		http.Error(w, "OneDrive OAuth credentials not configured: "+err.Error(), http.StatusBadRequest)
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

	authURL := oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.SetAuthURLParam("prompt", "select_account"))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"authURL": authURL})
}

func (h *OneDriveHandler) Callback(w http.ResponseWriter, r *http.Request) {
	state := r.URL.Query().Get("state")
	code := r.URL.Query().Get("code")

	if !h.consumeState(state) {
		http.Error(w, "invalid or expired state", http.StatusBadRequest)
		return
	}

	oauthConfig, err := h.loadOAuthConfig()
	if err != nil {
		log.Printf("onedrivehandler: load credentials: %v", err)
		http.Redirect(w, r, h.frontendURL+"/providers?error=credentials_missing", http.StatusFound)
		return
	}

	tok, err := oauthConfig.Exchange(r.Context(), code)
	if err != nil {
		log.Printf("onedrivehandler: token exchange failed: %v", err)
		http.Redirect(w, r, h.frontendURL+"/providers?error=oauth_failed", http.StatusFound)
		return
	}

	if err := h.store.UpsertToken("oneDrive", tok); err != nil {
		log.Printf("onedrivehandler: save token: %v", err)
		http.Redirect(w, r, h.frontendURL+"/providers?error=save_failed", http.StatusFound)
		return
	}

	oda, err := onedrive.NewOneDriveAdapter(oauthConfig, tok, h.store)
	if err != nil {
		log.Printf("onedrivehandler: init adapter: %v", err)
		http.Redirect(w, r, h.frontendURL+"/providers?error=adapter_failed", http.StatusFound)
		return
	}
	h.registry.Register("oneDrive", oda)

	http.Redirect(w, r, h.frontendURL+"/providers?connected=oneDrive", http.StatusFound)
}

func (h *OneDriveHandler) Disconnect(w http.ResponseWriter, r *http.Request) {
	if err := h.store.DeleteToken("oneDrive"); err != nil {
		log.Printf("onedrivehandler: delete token: %v", err)
		http.Error(w, "failed to disconnect", http.StatusInternalServerError)
		return
	}
	h.registry.Unregister("oneDrive")
	w.WriteHeader(http.StatusNoContent)
}

func (h *OneDriveHandler) loadOAuthConfig() (*oauth2.Config, error) {
	clientID, clientSecret, redirectURI, err := h.loadCredentials()
	if err != nil {
		return nil, err
	}
	return &oauth2.Config{ClientID: clientID, ClientSecret: clientSecret, RedirectURL: redirectURI, Scopes: oneDriveScopes, Endpoint: microsoft.AzureADEndpoint("common")}, nil
}

func (h *OneDriveHandler) loadCredentials() (clientID, clientSecret, redirectURI string, err error) {
	clientID, clientSecret, redirectURI, storeErr := h.store.LoadCredentials("oneDrive")
	if storeErr == nil {
		return clientID, clientSecret, redirectURI, nil
	}
	if !errors.Is(storeErr, db.ErrNotFound) {
		return "", "", "", fmt.Errorf("load credentials from store: %w", storeErr)
	}

	clientID = os.Getenv("ONEDRIVE_CLIENT_ID")
	clientSecret = os.Getenv("ONEDRIVE_CLIENT_SECRET")
	redirectURI = os.Getenv("ONEDRIVE_REDIRECT_URI")
	if clientID != "" && clientSecret != "" {
		return clientID, clientSecret, redirectURI, nil
	}

	return "", "", "", fmt.Errorf("no OneDrive OAuth credentials configured — use PUT /api/credentials/oneDrive or set env vars")
}

func (h *OneDriveHandler) consumeState(state string) bool {
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
