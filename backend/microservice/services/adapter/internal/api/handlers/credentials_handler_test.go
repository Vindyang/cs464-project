package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/vindyang/cs464-project/backend/services/shared/adapter"
	"github.com/vindyang/cs464-project/backend/services/shared/db"
	"golang.org/x/oauth2"
)

type stubProvider struct{}

func (stubProvider) GetMetadata(_ context.Context) (*adapter.ProviderMetadata, error) {
	return &adapter.ProviderMetadata{ProviderID: "stub"}, nil
}
func (stubProvider) UploadShard(_ context.Context, _ string, _ int, _ io.Reader) (string, error) {
	return "", nil
}
func (stubProvider) DownloadShard(_ context.Context, _ string) (io.ReadCloser, error) {
	return io.NopCloser(strings.NewReader("")), nil
}
func (stubProvider) DeleteShard(_ context.Context, _ string) error { return nil }
func (stubProvider) HealthCheck(_ context.Context) error           { return nil }

func newTestCredentialStore(t *testing.T) *db.Store {
	t.Helper()

	storePath := filepath.Join(t.TempDir(), "test.db")
	store, err := db.NewStore(storePath)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	t.Cleanup(func() {
		_ = store.Close()
	})
	return store
}

func newCredentialHandlerServer(t *testing.T) (*db.Store, http.Handler) {
	t.Helper()

	store := newTestCredentialStore(t)
	registry := adapter.NewRegistry()
	h := NewCredentialsHandler(store, registry)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	return store, mux
}

func TestCredentialsCollectionAndStatus_Empty(t *testing.T) {
	_, server := newCredentialHandlerServer(t)

	t.Run("GET /api/credentials returns empty list", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/credentials", nil)
		rr := httptest.NewRecorder()
		server.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("status code: got %d want %d", rr.Code, http.StatusOK)
		}

		var got []db.CredentialRecord
		if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if len(got) != 0 {
			t.Fatalf("expected empty list, got %d items", len(got))
		}
	})

	t.Run("GET /api/credentials/status returns configured false", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/credentials/status", nil)
		rr := httptest.NewRecorder()
		server.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("status code: got %d want %d", rr.Code, http.StatusOK)
		}

		var got struct {
			Configured bool     `json:"configured"`
			Count      int      `json:"count"`
			Providers  []string `json:"providers"`
		}
		if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if got.Configured {
			t.Fatalf("configured: got true want false")
		}
		if got.Count != 0 {
			t.Fatalf("count: got %d want 0", got.Count)
		}
		if len(got.Providers) != 0 {
			t.Fatalf("providers: got %v want empty", got.Providers)
		}
	})
}

func TestCredentialSecretRevealAndDeleteCleansUpProvider(t *testing.T) {
	store := newTestCredentialStore(t)
	registry := adapter.NewRegistry()
	registry.Register("googleDrive", stubProvider{})
	h := NewCredentialsHandler(store, registry)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	if err := store.UpsertCredentials("googleDrive", "client-id", "client-secret", "http://localhost/callback"); err != nil {
		t.Fatalf("upsert credentials: %v", err)
	}
	if err := store.UpsertToken("googleDrive", &oauth2.Token{AccessToken: "token", TokenType: "Bearer"}); err != nil {
		t.Fatalf("upsert token: %v", err)
	}

	t.Run("GET /api/credentials/googleDrive/secret reveals the stored secret", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/credentials/googleDrive/secret", nil)
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("status code: got %d want %d body=%s", rr.Code, http.StatusOK, rr.Body.String())
		}

		var got map[string]string
		if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if got["client_secret"] != "client-secret" {
			t.Fatalf("client_secret: got %q want %q", got["client_secret"], "client-secret")
		}
	})

	t.Run("DELETE /api/credentials/googleDrive removes credentials, token, and live registration", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/api/credentials/googleDrive", nil)
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)

		if rr.Code != http.StatusNoContent {
			t.Fatalf("status code: got %d want %d", rr.Code, http.StatusNoContent)
		}

		if _, _, _, err := store.LoadCredentials("googleDrive"); !errors.Is(err, db.ErrNotFound) {
			t.Fatalf("expected credentials to be deleted, got err=%v", err)
		}
		if _, err := store.LoadToken("googleDrive"); !errors.Is(err, db.ErrNotFound) {
			t.Fatalf("expected token to be deleted, got err=%v", err)
		}
		if _, err := registry.Get("googleDrive"); err == nil {
			t.Fatalf("expected provider to be unregistered")
		}
	})
}

func TestCredentialsCollectionAndStatus_Configured(t *testing.T) {
	store, server := newCredentialHandlerServer(t)
	if err := store.UpsertCredentials("googleDrive", "g-id", "g-secret", "http://localhost:3000/callback"); err != nil {
		t.Fatalf("upsert googleDrive: %v", err)
	}
	if err := store.UpsertCredentials("awsS3", "a-id", "a-secret", "http://localhost:3000/s3-callback"); err != nil {
		t.Fatalf("upsert awsS3: %v", err)
	}

	t.Run("GET /api/credentials returns safe credential list", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/credentials", nil)
		rr := httptest.NewRecorder()
		server.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("status code: got %d want %d", rr.Code, http.StatusOK)
		}

		var got []db.CredentialRecord
		if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if len(got) != 2 {
			t.Fatalf("expected 2 records, got %d", len(got))
		}
		if got[0].ProviderID != "awsS3" || got[1].ProviderID != "googleDrive" {
			t.Fatalf("provider order/content: got [%s %s]", got[0].ProviderID, got[1].ProviderID)
		}
		if rr.Body.String() == "" || strings.Contains(rr.Body.String(), "client_secret") {
			t.Fatalf("response unexpectedly contains secret field")
		}
	})

	t.Run("GET /api/credentials/status returns configured true with providers", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/credentials/status", nil)
		rr := httptest.NewRecorder()
		server.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("status code: got %d want %d", rr.Code, http.StatusOK)
		}

		var got struct {
			Configured bool     `json:"configured"`
			Count      int      `json:"count"`
			Providers  []string `json:"providers"`
		}
		if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if !got.Configured {
			t.Fatalf("configured: got false want true")
		}
		if got.Count != 2 {
			t.Fatalf("count: got %d want 2", got.Count)
		}
		if len(got.Providers) != 2 || got.Providers[0] != "awsS3" || got.Providers[1] != "googleDrive" {
			t.Fatalf("providers: got %v want [awsS3 googleDrive]", got.Providers)
		}
	})
}

func TestCredentialsCollectionAndStatus_MethodNotAllowed(t *testing.T) {
	_, server := newCredentialHandlerServer(t)

	t.Run("POST /api/credentials", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/credentials", nil)
		rr := httptest.NewRecorder()
		server.ServeHTTP(rr, req)
		if rr.Code != http.StatusMethodNotAllowed {
			t.Fatalf("status code: got %d want %d", rr.Code, http.StatusMethodNotAllowed)
		}
	})

	t.Run("POST /api/credentials/status", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/credentials/status", nil)
		rr := httptest.NewRecorder()
		server.ServeHTTP(rr, req)
		if rr.Code != http.StatusMethodNotAllowed {
			t.Fatalf("status code: got %d want %d", rr.Code, http.StatusMethodNotAllowed)
		}
	})
}

func TestCredentialsUpsert_AWSS3RegionValidation(t *testing.T) {
	store, server := newCredentialHandlerServer(t)

	t.Run("PUT /api/credentials/awsS3 accepts valid region", func(t *testing.T) {
		body := []byte(`{"client_id":"a-id","client_secret":"a-secret","redirect_uri":"ap-southeast-1"}`)
		req := httptest.NewRequest(http.MethodPut, "/api/credentials/awsS3", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		server.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("status code: got %d want %d body=%s", rr.Code, http.StatusOK, rr.Body.String())
		}

		clientID, clientSecret, region, err := store.LoadCredentials("awsS3")
		if err != nil {
			t.Fatalf("load credentials: %v", err)
		}
		if clientID != "a-id" || clientSecret != "a-secret" || region != "ap-southeast-1" {
			t.Fatalf("stored credentials mismatch: got %q %q %q", clientID, clientSecret, region)
		}
	})

	t.Run("PUT /api/credentials/awsS3 rejects invalid region", func(t *testing.T) {
		body := []byte(`{"client_id":"a-id","client_secret":"a-secret","redirect_uri":"Sydney"}`)
		req := httptest.NewRequest(http.MethodPut, "/api/credentials/awsS3", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		server.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("status code: got %d want %d body=%s", rr.Code, http.StatusBadRequest, rr.Body.String())
		}

		var got map[string]string
		if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if got["error"] != "invalid AWS region" {
			t.Fatalf("error message: got %q want %q", got["error"], "invalid AWS region")
		}
		if !strings.Contains(got["details"], "not a valid AWS region") {
			t.Fatalf("details: got %q", got["details"])
		}
	})
}
