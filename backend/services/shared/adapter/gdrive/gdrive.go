package gdrive

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"golang.org/x/oauth2"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"

	"github.com/vindyang/cs464-project/backend/services/shared/adapter"
	"github.com/vindyang/cs464-project/backend/services/shared/db"
)

// driveScope grants access only to files created by this app.
// With user OAuth2, drive.file tokens are bound to the OAuth2 client ID (not a session),
// so files remain accessible across server restarts.
const driveScope = "https://www.googleapis.com/auth/drive.file"

const OmnishardFolderConfigKey = "gdrive_Omnishard_folder_id"

// GDriveAdapter implements StorageProvider using the Google Drive API v3.
// Authenticates via OAuth2 user credentials (Web app flow).
type GDriveAdapter struct {
	folderID     string    // in-memory cache of the resolved Omnishard folder ID
	store        *db.Store // persists the Omnishard folder ID across restarts
	service      *drive.Service
	mu           sync.Mutex        // guards folderID during lazy resolution
	fileFolderMu sync.Mutex        // serializes per-file folder creation to prevent duplicates
	fileFolders  map[string]string // cache of per-file folder IDs: fileID -> Drive folder ID
}

// NewGDriveAdapter constructs a GDriveAdapter from an OAuth2 config, token, and store.
// The token source auto-refreshes using the refresh token.
// The store is used to persist the resolved Omnishard folder ID across restarts.
func NewGDriveAdapter(config *oauth2.Config, token *oauth2.Token, store *db.Store) (*GDriveAdapter, error) {
	tokenSource := config.TokenSource(context.Background(), token)

	svc, err := drive.NewService(context.Background(), option.WithTokenSource(tokenSource))
	if err != nil {
		return nil, fmt.Errorf("gdrive: create drive service: %w", err)
	}

	return &GDriveAdapter{
		store:       store,
		service:     svc,
		fileFolders: make(map[string]string),
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

// ensureOmnishardFolder returns the Drive folder ID for the "Omnishard" folder,
// creating it if it doesn't exist. The result is cached in memory and
// persisted in the local store so it survives restarts.
func (g *GDriveAdapter) ensureOmnishardFolder(ctx context.Context) (string, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	// 1. In-memory cache (fastest path after first call).
	if g.folderID != "" {
		return g.folderID, nil
	}

	// 2. Persisted in SQLite from a previous run.
	if id, err := g.store.GetConfig(OmnishardFolderConfigKey); err == nil {
		g.folderID = id
		return id, nil
	}

	// 3. Search Drive for an existing "Omnishard" folder created by this app.
	query := "name='Omnishard' and mimeType='application/vnd.google-apps.folder' and trashed=false"
	list, err := g.service.Files.List().
		Q(query).
		Fields("files(id,name)").
		PageSize(1).
		Context(ctx).
		Do()
	if err != nil {
		return "", fmt.Errorf("gdrive: search for Omnishard folder: %w", err)
	}

	var folderID string
	if len(list.Files) > 0 {
		folderID = list.Files[0].Id
	} else {
		// 4. Create the folder.
		folder, err := g.service.Files.Create(&drive.File{
			Name:     "Omnishard",
			MimeType: "application/vnd.google-apps.folder",
		}).Fields("id").Context(ctx).Do()
		if err != nil {
			return "", fmt.Errorf("gdrive: create Omnishard folder: %w", err)
		}
		folderID = folder.Id
	}

	// Persist for future restarts.
	if err := g.store.SetConfig(OmnishardFolderConfigKey, folderID); err != nil {
		return "", fmt.Errorf("gdrive: persist Omnishard folder id: %w", err)
	}
	g.folderID = folderID
	return folderID, nil
}

// ensureFileFolder returns the Drive folder ID for a per-file subfolder inside
// the Omnishard root, named by fileID. Creates it if it doesn't exist.
// fileFolderMu serializes the check-and-create so parallel shard uploads
// for the same file cannot each create their own duplicate folder.
func (g *GDriveAdapter) ensureFileFolder(ctx context.Context, fileID string) (string, error) {
	g.fileFolderMu.Lock()
	defer g.fileFolderMu.Unlock()

	if id, ok := g.fileFolders[fileID]; ok {
		return id, nil
	}

	rootFolderID, err := g.ensureOmnishardFolder(ctx)
	if err != nil {
		return "", err
	}

	query := fmt.Sprintf(
		"name='%s' and mimeType='application/vnd.google-apps.folder' and '%s' in parents and trashed=false",
		fileID, rootFolderID,
	)
	list, err := g.service.Files.List().Q(query).Fields("files(id)").Context(ctx).Do()
	if err != nil {
		return "", fmt.Errorf("gdrive: search file folder %s: %w", fileID, err)
	}
	if len(list.Files) > 0 {
		g.fileFolders[fileID] = list.Files[0].Id
		return list.Files[0].Id, nil
	}

	folder, err := g.service.Files.Create(&drive.File{
		Name:     fileID,
		MimeType: "application/vnd.google-apps.folder",
		Parents:  []string{rootFolderID},
	}).Fields("id").Context(ctx).Do()
	if err != nil {
		return "", fmt.Errorf("gdrive: create file folder %s: %w", fileID, err)
	}
	g.fileFolders[fileID] = folder.Id
	return folder.Id, nil
}

// UploadShard uploads binary shard data into a per-file subfolder inside "Omnishard".
// The subfolder is named by fileID and created automatically if it doesn't exist.
// Returns the Drive file ID as the remoteID for future retrieval or deletion.
func (g *GDriveAdapter) UploadShard(ctx context.Context, fileID string, index int, data io.Reader) (string, error) {
	folderID, err := g.ensureFileFolder(ctx, fileID)
	if err != nil {
		return "", err
	}

	name := fmt.Sprintf("shard_%03d", index)

	meta := &drive.File{
		Name:    name,
		Parents: []string{folderID},
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
	// Drive UI deletes often move files to trash first. Treat trashed shards as missing.
	meta, err := g.service.Files.Get(remoteID).
		Fields("id,trashed").
		SupportsAllDrives(true).
		Context(ctx).
		Do()
	if err != nil {
		var apiErr *googleapi.Error
		if errors.As(err, &apiErr) {
			if apiErr.Code == 404 {
				return nil, fmt.Errorf("%w: gdrive shard %q", adapter.ErrShardNotFound, remoteID)
			}
			if apiErr.Code == 403 {
				for _, e := range apiErr.Errors {
					reason := strings.ToLower(e.Reason)
					message := strings.ToLower(e.Message)
					if reason == "notfound" || strings.Contains(message, "not found") {
						return nil, fmt.Errorf("%w: gdrive shard %q", adapter.ErrShardNotFound, remoteID)
					}
				}
				msg := strings.ToLower(apiErr.Message)
				if strings.Contains(msg, "not found") {
					return nil, fmt.Errorf("%w: gdrive shard %q", adapter.ErrShardNotFound, remoteID)
				}
			}
		}
		return nil, fmt.Errorf("gdrive: download shard %q: %w", remoteID, err)
	}
	if meta != nil && meta.Trashed {
		return nil, fmt.Errorf("%w: gdrive shard %q is trashed", adapter.ErrShardNotFound, remoteID)
	}

	resp, err := g.service.Files.Get(remoteID).
		SupportsAllDrives(true).
		Context(ctx).
		Download()
	if err != nil {
		var apiErr *googleapi.Error
		if errors.As(err, &apiErr) {
			if apiErr.Code == 404 {
				return nil, fmt.Errorf("%w: gdrive shard %q", adapter.ErrShardNotFound, remoteID)
			}
			if apiErr.Code == 403 {
				for _, e := range apiErr.Errors {
					reason := strings.ToLower(e.Reason)
					message := strings.ToLower(e.Message)
					if reason == "notfound" || strings.Contains(message, "not found") {
						return nil, fmt.Errorf("%w: gdrive shard %q", adapter.ErrShardNotFound, remoteID)
					}
				}
				msg := strings.ToLower(apiErr.Message)
				if strings.Contains(msg, "not found") {
					return nil, fmt.Errorf("%w: gdrive shard %q", adapter.ErrShardNotFound, remoteID)
				}
			}
		}
		return nil, fmt.Errorf("gdrive: download shard %q: %w", remoteID, err)
	}
	return resp.Body, nil
}

// DeleteShard permanently deletes a shard by its Drive file ID (bypasses trash).
// After deleting the shard, if its parent folder is now empty it is also deleted.
// Treats 404 as success for idempotent rollback behavior.
func (g *GDriveAdapter) DeleteShard(ctx context.Context, remoteID string) error {
	// Fetch the parent folder ID before deleting the file.
	file, err := g.service.Files.Get(remoteID).Fields("parents").Context(ctx).Do()
	var parentID string
	if err == nil && len(file.Parents) > 0 {
		parentID = file.Parents[0]
	}

	if err := g.service.Files.Delete(remoteID).Context(ctx).Do(); err != nil {
		var apiErr *googleapi.Error
		if errors.As(err, &apiErr) && apiErr.Code == 404 {
			return nil // already deleted; treat as success
		}
		return fmt.Errorf("gdrive: delete shard %q: %w", remoteID, err)
	}

	// If we know the parent folder, check whether it is now empty and delete it.
	if parentID != "" {
		g.fileFolderMu.Lock()
		defer g.fileFolderMu.Unlock()

		list, err := g.service.Files.List().
			Q(fmt.Sprintf("'%s' in parents and trashed=false", parentID)).
			Fields("files(id)").
			PageSize(1).
			Context(ctx).
			Do()
		if err == nil && len(list.Files) == 0 {
			_ = g.service.Files.Delete(parentID).Context(ctx).Do()
			// Evict from cache so future uploads recreate the folder cleanly.
			for k, v := range g.fileFolders {
				if v == parentID {
					delete(g.fileFolders, k)
					break
				}
			}
		}
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
