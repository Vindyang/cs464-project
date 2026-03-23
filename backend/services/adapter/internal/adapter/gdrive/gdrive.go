package gdrive

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"golang.org/x/oauth2"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"

	"github.com/vindyang/cs464-project/backend/services/adapter/internal/adapter"
)

// driveScope grants access only to files created by this app.
// With user OAuth2, drive.file tokens are bound to the OAuth2 client ID (not a session),
// so files remain accessible across server restarts.
const driveScope = "https://www.googleapis.com/auth/drive.file"

// GDriveAdapter implements StorageProvider using the Google Drive API v3.
// Authenticates via OAuth2 user credentials (Web app flow).
type GDriveAdapter struct {
	FolderID string
	service  *drive.Service
}

// NewGDriveAdapter constructs a GDriveAdapter from an OAuth2 config and token.
// The token source auto-refreshes using the refresh token.
//
//   - config: OAuth2 config with client credentials and redirect URI.
//   - token: stored OAuth2 token (loaded from DB by the caller).
//   - folderID: Google Drive folder ID where shards will be stored.
func NewGDriveAdapter(config *oauth2.Config, token *oauth2.Token, folderID string) (*GDriveAdapter, error) {
	tokenSource := config.TokenSource(context.Background(), token)

	svc, err := drive.NewService(context.Background(), option.WithTokenSource(tokenSource))
	if err != nil {
		return nil, fmt.Errorf("gdrive: create drive service: %w", err)
	}

	return &GDriveAdapter{
		FolderID: folderID,
		service:  svc,
	}, nil
}

// GetMetadata fetches real quota and latency from the Drive API.
func (g *GDriveAdapter) GetMetadata(ctx context.Context) (*adapter.ProviderMetadata, error) {
	start := time.Now()

	about, err := g.service.About.Get().
		Fields("storageQuota").
		Context(ctx).
		Do()
	if err != nil {
		return nil, fmt.Errorf("gdrive: about.get: %w", err)
	}

	latency := time.Since(start).Milliseconds()

	var quotaTotal, quotaUsed int64
	if about.StorageQuota != nil {
		quotaTotal = about.StorageQuota.Limit
		quotaUsed = about.StorageQuota.Usage
	}

	return &adapter.ProviderMetadata{
		ProviderID:   "googleDrive",
		DisplayName:  "Google Drive",
		Status:       "connected",
		LatencyMs:    latency,
		Region:       "global",
		Capabilities: map[string]any{"supportsVersioning": true},
		QuotaTotal:   quotaTotal,
		QuotaUsed:    quotaUsed,
		LastCheck:    time.Now().UTC().Format(time.RFC3339),
	}, nil
}

// UploadShard uploads binary shard data as a file in the configured folder.
// Returns the Drive file ID as the remoteID for future retrieval or deletion.
func (g *GDriveAdapter) UploadShard(ctx context.Context, fileID string, index int, data io.Reader) (string, error) {
	name := fmt.Sprintf("shard_%s_%03d", fileID, index)

	meta := &drive.File{
		Name:    name,
		Parents: []string{g.FolderID},
	}

	file, err := g.service.Files.Create(meta).
		Media(data).
		Fields("id").
		Context(ctx).
		Do()
	if err != nil {
		return "", fmt.Errorf("gdrive: upload shard %s[%d]: %w", fileID, index, err)
	}

	return file.Id, nil
}

// DownloadShard downloads a shard by its Drive file ID.
// The returned ReadCloser must be closed by the caller.
func (g *GDriveAdapter) DownloadShard(ctx context.Context, remoteID string) (io.ReadCloser, error) {
	resp, err := g.service.Files.Get(remoteID).
		Context(ctx).
		Download()
	if err != nil {
		return nil, fmt.Errorf("gdrive: download shard %q: %w", remoteID, err)
	}
	return resp.Body, nil
}

// DeleteShard permanently deletes a shard by its Drive file ID (bypasses trash).
// Treats 404 as success for idempotent rollback behavior.
func (g *GDriveAdapter) DeleteShard(ctx context.Context, remoteID string) error {
	err := g.service.Files.Delete(remoteID).
		Context(ctx).
		Do()
	if err != nil {
		var apiErr *googleapi.Error
		if errors.As(err, &apiErr) && apiErr.Code == 404 {
			return nil // already deleted; treat as success
		}
		return fmt.Errorf("gdrive: delete shard %q: %w", remoteID, err)
	}
	return nil
}

// HealthCheck performs a lightweight liveness ping against the Drive API.
func (g *GDriveAdapter) HealthCheck(ctx context.Context) error {
	_, err := g.service.Files.List().
		PageSize(1).
		Fields("files(id)").
		Context(ctx).
		Do()
	if err != nil {
		return fmt.Errorf("gdrive: health check: %w", err)
	}
	return nil
}
