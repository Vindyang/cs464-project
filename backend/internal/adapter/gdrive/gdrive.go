package gdrive

import (
	"context"
	"io"
	"github.com/vindyang/cs464-project/backend/internal/adapter"
)

type GDriveAdapter struct {
	FolderID string
}

func NewGDriveAdapter(folderID string) *GDriveAdapter {
	return &GDriveAdapter{
		FolderID: folderID,
	}
}

func (a *GDriveAdapter) GetMetadata(ctx context.Context) (*adapter.ProviderMetadata, error) {
	return &adapter.ProviderMetadata{
		ProviderID:   "googleDrive",
		DisplayName:  "Google Drive",
		Status:       "connected",
		Region:       "global",
		Capabilities: map[string]interface{}{"supportsVersioning": true},
	}, nil
}

func (a *GDriveAdapter) UploadShard(ctx context.Context, fileID string, index int, data io.Reader) (string, error) {
	// TODO: Implement Google Drive API upload
	return "gdrive-mock-remote-id", nil
}

func (a *GDriveAdapter) DownloadShard(ctx context.Context, remoteID string) (io.ReadCloser, error) {
	// TODO: Implement Google Drive API download
	return nil, nil
}

func (a *GDriveAdapter) DeleteShard(ctx context.Context, remoteID string) error {
	// TODO: Implement Google Drive API delete
	return nil
}

func (a *GDriveAdapter) HealthCheck(ctx context.Context) error {
	return nil
}
