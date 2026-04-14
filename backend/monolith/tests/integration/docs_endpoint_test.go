package integration_test

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	monolithapp "github.com/vindyang/cs464-project/backend/monolith/internal/app"
)

func TestDocsEndpointReturnsHTMLIndex(t *testing.T) {
	t.Parallel()

	app := newTestApp(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/docs", nil)
	res := httptest.NewRecorder()

	app.Handler().ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, res.Code)
	}
	if contentType := res.Header().Get("Content-Type"); !strings.Contains(contentType, "text/html") {
		t.Fatalf("expected html content type, got %q", contentType)
	}
	body := res.Body.String()
	for _, fragment := range []string{
		"Omnishard Monolith API",
		"/api/v1/files",
		"/api/orchestrator/files/{fileId}/download",
		"/api/v1/docs/openapi.yml",
	} {
		if !strings.Contains(body, fragment) {
			t.Fatalf("expected docs page to contain %q", fragment)
		}
	}
}

func TestDocsEndpointReturnsRawSpecOnYAMLPath(t *testing.T) {
	t.Parallel()

	app := newTestApp(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/docs/openapi.yml", nil)
	res := httptest.NewRecorder()

	app.Handler().ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, res.Code)
	}
	if contentType := res.Header().Get("Content-Type"); !strings.Contains(contentType, "yaml") {
		t.Fatalf("expected yaml content type, got %q", contentType)
	}
	if disposition := res.Header().Get("Content-Disposition"); !strings.Contains(disposition, "inline") {
		t.Fatalf("expected inline content disposition, got %q", disposition)
	}
	body := res.Body.String()
	if !strings.Contains(body, "openapi: 3.0.3") {
		t.Fatalf("expected raw OpenAPI spec, got %q", body)
	}
}

func newTestApp(t *testing.T) *monolithapp.App {
	t.Helper()

	tempDir := t.TempDir()
	app, err := monolithapp.New(monolithapp.Config{
		StorePath:    filepath.Join(tempDir, "Omnishard.db"),
		ShardMapPath: filepath.Join(tempDir, "Omnishard-shardmap.db"),
	})
	if err != nil {
		t.Fatalf("failed to create monolith app: %v", err)
	}
	t.Cleanup(func() {
		if err := app.Close(); err != nil {
			t.Fatalf("failed to close monolith app: %v", err)
		}
	})

	return app
}
