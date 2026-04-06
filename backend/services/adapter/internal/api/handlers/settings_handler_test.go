package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

type stubFileResetter struct {
	deleteAllFilesFn func(deleteShards bool) (FileDeleteSummary, error)
}

func (s stubFileResetter) DeleteAllFiles(_ context.Context, deleteShards bool) (FileDeleteSummary, error) {
	if s.deleteAllFilesFn != nil {
		return s.deleteAllFilesFn(deleteShards)
	}
	return FileDeleteSummary{}, nil
}

type stubCredentialResetter struct {
	deleteAllCredentialsFn func() (CredentialResetSummary, error)
}

func (s stubCredentialResetter) DeleteAllCredentials() (CredentialResetSummary, error) {
	if s.deleteAllCredentialsFn != nil {
		return s.deleteAllCredentialsFn()
	}
	return CredentialResetSummary{}, nil
}

func TestSettingsResetScope_AllData(t *testing.T) {
	store := newTestCredentialStore(t)
	h := NewSettingsHandler(
		store,
		stubFileResetter{deleteAllFilesFn: func(deleteShards bool) (FileDeleteSummary, error) {
			if !deleteShards {
				t.Fatalf("expected delete_shards true")
			}
			return FileDeleteSummary{DeletedFiles: 3, DeletedShards: 12, DeleteShards: true}, nil
		}},
		stubCredentialResetter{deleteAllCredentialsFn: func() (CredentialResetSummary, error) {
			return CredentialResetSummary{DeletedCredentials: 2, DeletedTokens: 2, DisconnectedProviders: 2}, nil
		}},
	)

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	body := []byte(`{"scope":"all_data"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/settings/reset", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status code: got %d want %d body=%s", rr.Code, http.StatusOK, rr.Body.String())
	}

	var got map[string]json.RawMessage
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if _, ok := got["file_summary"]; !ok {
		t.Fatalf("missing file_summary in response")
	}
	if _, ok := got["credential_summary"]; !ok {
		t.Fatalf("missing credential_summary in response")
	}
}
