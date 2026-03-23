package s3

import (
	"context"
	"github.com/vindyang/cs464-project/backend/services/orchestrator/internal/adapter"
	"io"
)

type S3Adapter struct {
	Bucket string
	Region string
}

func NewS3Adapter(bucket, region string) *S3Adapter {
	return &S3Adapter{
		Bucket: bucket,
		Region: region,
	}
}

func (a *S3Adapter) GetMetadata(ctx context.Context) (*adapter.ProviderMetadata, error) {
	return &adapter.ProviderMetadata{
		ProviderID:   "awsS3",
		DisplayName:  "AWS S3",
		Status:       "connected",
		Region:       a.Region,
		Capabilities: map[string]interface{}{"maxPartSize": 5242880},
	}, nil
}

func (a *S3Adapter) UploadShard(ctx context.Context, fileID string, index int, data io.Reader) (string, error) {
	// TODO: Implement AWS SDK PutObject
	return "s3-mock-remote-id", nil
}

func (a *S3Adapter) DownloadShard(ctx context.Context, remoteID string) (io.ReadCloser, error) {
	// TODO: Implement AWS SDK GetObject
	return nil, nil
}

func (a *S3Adapter) DeleteShard(ctx context.Context, remoteID string) error {
	// TODO: Implement AWS SDK DeleteObject
	return nil
}

func (a *S3Adapter) HealthCheck(ctx context.Context) error {
	return nil
}
