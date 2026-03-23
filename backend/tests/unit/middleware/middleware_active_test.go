package middleware_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/vindyang/cs464-project/backend/internal/api/middleware"
)

func TestCORS_OptionsPreflight_Active(t *testing.T) {
	called := false
	h := middleware.CORS(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodOptions, "/any", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if called {
		t.Fatalf("expected next handler not to be called for OPTIONS")
	}
	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rr.Code)
	}
}

func TestLogger_InvokesNext_Active(t *testing.T) {
	h := middleware.Logger(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	}))

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d", rr.Code)
	}
}

func TestRecovery_RecoversPanic_Active(t *testing.T) {
	h := middleware.Recovery(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("boom")
	}))

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rr.Code)
	}

	var body map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("expected JSON body, got error: %v", err)
	}
	if body["error"] != "Internal server error" {
		t.Fatalf("unexpected body: %+v", body)
	}
}
