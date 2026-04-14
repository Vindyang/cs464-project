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

	"github.com/vindyang/cs464-project/backend/monolith/shared/adapter"
	"github.com/vindyang/cs464-project/backend/monolith/shared/db"
)

const driveScope = "https://www.googleapis.com/auth/drive.file"

const OmnishardFolderConfigKey = "gdrive_Omnishard_folder_id"

type GDriveAdapter struct {
	folderID     string
	store        *db.Store
	service      *drive.Service
	mu           sync.Mutex
	fileFolderMu sync.Mutex
	fileFolders  map[string]string
}

func NewGDriveAdapter(config *oauth2.Config, token *oauth2.Token, store *db.Store) (*GDriveAdapter, error) {
	tokenSource := config.TokenSource(context.Background(), token)

	svc, err := drive.NewService(context.Background(), option.WithTokenSource(tokenSource))
	if err != nil {
		return nil, fmt.Errorf("gdrive: create drive service: %w", err)
	}

	return &GDriveAdapter{store: store, service: svc, fileFolders: make(map[string]string)}, nil
}

func (g *GDriveAdapter) GetMetadata(ctx context.Context) (*adapter.ProviderMetadata, error) {
	start := time.Now()

	about, err := g.service.About.Get().Fields("storageQuota").Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("gdrive: about.get: %w", err)
	}

	latency := time.Since(start).Milliseconds()

	var quotaTotal int64
	var quotaUsed int64
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

func (g *GDriveAdapter) ensureOmnishardFolder(ctx context.Context) (string, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.folderID != "" {
		return g.folderID, nil
	}

	if id, err := g.store.GetConfig(OmnishardFolderConfigKey); err == nil {
		g.folderID = id
		return id, nil
	}

	query := "name='Omnishard' and mimeType='application/vnd.google-apps.folder' and trashed=false"
	list, err := g.service.Files.List().Q(query).Fields("files(id,name)").PageSize(1).Context(ctx).Do()
	if err != nil {
		return "", fmt.Errorf("gdrive: search for Omnishard folder: %w", err)
	}

	var folderID string
	if len(list.Files) > 0 {
		folderID = list.Files[0].Id
	} else {
		folder, err := g.service.Files.Create(&drive.File{Name: "Omnishard", MimeType: "application/vnd.google-apps.folder"}).Fields("id").Context(ctx).Do()
		if err != nil {
			return "", fmt.Errorf("gdrive: create Omnishard folder: %w", err)
		}
		folderID = folder.Id
	}

	if err := g.store.SetConfig(OmnishardFolderConfigKey, folderID); err != nil {
		return "", fmt.Errorf("gdrive: persist Omnishard folder id: %w", err)
	}
	g.folderID = folderID
	return folderID, nil
}

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

	query := fmt.Sprintf("name='%s' and mimeType='application/vnd.google-apps.folder' and '%s' in parents and trashed=false", fileID, rootFolderID)
	list, err := g.service.Files.List().Q(query).Fields("files(id)").Context(ctx).Do()
	if err != nil {
		return "", fmt.Errorf("gdrive: search file folder %s: %w", fileID, err)
	}
	if len(list.Files) > 0 {
		g.fileFolders[fileID] = list.Files[0].Id
		return list.Files[0].Id, nil
	}

	folder, err := g.service.Files.Create(&drive.File{Name: fileID, MimeType: "application/vnd.google-apps.folder", Parents: []string{rootFolderID}}).Fields("id").Context(ctx).Do()
	if err != nil {
		return "", fmt.Errorf("gdrive: create file folder %s: %w", fileID, err)
	}
	g.fileFolders[fileID] = folder.Id
	return folder.Id, nil
}

func (g *GDriveAdapter) UploadShard(ctx context.Context, fileID string, index int, data io.Reader) (string, error) {
	folderID, err := g.ensureFileFolder(ctx, fileID)
	if err != nil {
		return "", err
	}

	name := fmt.Sprintf("shard_%03d", index)
	meta := &drive.File{Name: name, Parents: []string{folderID}}

	file, err := g.service.Files.Create(meta).Media(data).Fields("id").Context(ctx).Do()
	if err != nil {
		return "", fmt.Errorf("gdrive: upload shard %s[%d]: %w", fileID, index, err)
	}

	return file.Id, nil
}

func (g *GDriveAdapter) DownloadShard(ctx context.Context, remoteID string) (io.ReadCloser, error) {
	meta, err := g.service.Files.Get(remoteID).Fields("id,trashed").SupportsAllDrives(true).Context(ctx).Do()
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

	resp, err := g.service.Files.Get(remoteID).SupportsAllDrives(true).Context(ctx).Download()
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

func (g *GDriveAdapter) DeleteShard(ctx context.Context, remoteID string) error {
	file, err := g.service.Files.Get(remoteID).Fields("parents").Context(ctx).Do()
	var parentID string
	if err == nil && len(file.Parents) > 0 {
		parentID = file.Parents[0]
	}

	if err := g.service.Files.Delete(remoteID).Context(ctx).Do(); err != nil {
		var apiErr *googleapi.Error
		if errors.As(err, &apiErr) && apiErr.Code == 404 {
			return nil
		}
		return fmt.Errorf("gdrive: delete shard %q: %w", remoteID, err)
	}

	if parentID != "" {
		g.fileFolderMu.Lock()
		defer g.fileFolderMu.Unlock()

		list, err := g.service.Files.List().Q(fmt.Sprintf("'%s' in parents and trashed=false", parentID)).Fields("files(id)").PageSize(1).Context(ctx).Do()
		if err == nil && len(list.Files) == 0 {
			_ = g.service.Files.Delete(parentID).Context(ctx).Do()
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

func (g *GDriveAdapter) HealthCheck(ctx context.Context) error {
	_, err := g.service.Files.List().PageSize(1).Fields("files(id)").Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("gdrive: health check: %w", err)
	}
	return nil
}
