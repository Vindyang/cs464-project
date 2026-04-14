package gdrive_test

import (
	"context"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/joho/godotenv"
	"github.com/vindyang/cs464-project/backend/services/shared/adapter/gdrive"
	"golang.org/x/oauth2/google"
)

const driveScope = "https://www.googleapis.com/auth/drive.file"

// TestMain loads backend/microservice/.env before any tests run so that GDRIVE_OAUTH_CREDENTIALS_FILE
// and GDRIVE_FOLDER_ID are available without manually setting env vars.
// Errors are silently ignored so tests still run when vars are set directly (e.g. CI).
func TestMain(m *testing.M) {
	_ = godotenv.Load("../../../../.env")
	os.Exit(m.Run())
}

// TestGDriveIntegration runs only when real OAuth2 credentials are configured.
// Set the following environment variables (or populate backend/microservice/.env) to run:
//
//	GDRIVE_OAUTH_CREDENTIALS_FILE=/path/to/oauth-credentials.json  (Web app type)
//	GDRIVE_OAUTH_REDIRECT_URI=http://localhost:8080/api/oauth/gdrive/callback
//
// A valid token must already exist in the local SQLite store (connect via the UI first).
func TestGDriveIntegration(t *testing.T) {
	oauthCredsFile := os.Getenv("GDRIVE_OAUTH_CREDENTIALS_FILE")
	redirectURI := os.Getenv("GDRIVE_OAUTH_REDIRECT_URI")
	if oauthCredsFile == "" || redirectURI == "" {
		t.Skip("GDRIVE_OAUTH_CREDENTIALS_FILE and GDRIVE_OAUTH_REDIRECT_URI not set; skipping integration test")
	}

	raw, err := os.ReadFile(oauthCredsFile)
	if err != nil {
		t.Fatalf("read creds file: %v", err)
	}
	config, err := google.ConfigFromJSON(raw, driveScope)
	if err != nil {
		t.Fatalf("parse creds: %v", err)
	}
	config.RedirectURL = redirectURI

	// For integration tests, use a minimal valid-looking token.
	// Real token exchange requires DB — skip if no token available.
	t.Skip("integration test requires a live token from the local SQLite store; connect via UI first")

	a, err := gdrive.NewGDriveAdapter(config, nil, nil)
	if err != nil {
		t.Fatalf("NewGDriveAdapter: %v", err)
	}

	ctx := context.Background()

	t.Run("HealthCheck", func(t *testing.T) {
		if err := a.HealthCheck(ctx); err != nil {
			t.Errorf("HealthCheck: %v", err)
		}
	})

	t.Run("GetMetadata", func(t *testing.T) {
		meta, err := a.GetMetadata(ctx)
		if err != nil {
			t.Fatalf("GetMetadata: %v", err)
		}
		if meta.ProviderID != "googleDrive" {
			t.Errorf("unexpected ProviderID: %s", meta.ProviderID)
		}
		if meta.LatencyMs <= 0 {
			t.Errorf("expected positive latency, got %d ms", meta.LatencyMs)
		}
		t.Logf("quota total=%d used=%d latency=%d ms", meta.QuotaTotal, meta.QuotaUsed, meta.LatencyMs)
	})

	t.Run("UploadDownloadDelete", func(t *testing.T) {
		const payload = "Omnishard-shard-test-payload"

		remoteID, err := a.UploadShard(ctx, "test-file-001", 0, strings.NewReader(payload))
		if err != nil {
			t.Fatalf("UploadShard: %v", err)
		}
		if remoteID == "" {
			t.Fatal("UploadShard returned empty remoteID")
		}
		t.Logf("uploaded shard: remoteID=%s", remoteID)

		t.Cleanup(func() {
			if err := a.DeleteShard(ctx, remoteID); err != nil {
				t.Errorf("DeleteShard cleanup: %v", err)
			}
		})

		rc, err := a.DownloadShard(ctx, remoteID)
		if err != nil {
			t.Fatalf("DownloadShard: %v", err)
		}
		defer rc.Close()

		got, err := io.ReadAll(rc)
		if err != nil {
			t.Fatalf("read downloaded body: %v", err)
		}
		if string(got) != payload {
			t.Errorf("content mismatch: got %q, want %q", got, payload)
		}
	})
}
