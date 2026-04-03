package s3

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/vindyang/cs464-project/backend/services/shared/adapter"
)

type S3Adapter struct {
	client *s3.Client
	Bucket string
	Region string
}

func NewS3Adapter(accessKeyID, secretAccessKey, region, bucket string) (*S3Adapter, error) {
	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, "")),
	)
	if err != nil {
		return nil, fmt.Errorf("s3: load config: %w", err)
	}
	return &S3Adapter{
		client: s3.NewFromConfig(cfg),
		Bucket: bucket,
		Region: region,
	}, nil
}

func (a *S3Adapter) GetMetadata(ctx context.Context) (*adapter.ProviderMetadata, error) {
	start := time.Now()
	_, err := a.client.HeadBucket(ctx, &s3.HeadBucketInput{Bucket: aws.String(a.Bucket)})
	latency := time.Since(start).Milliseconds()
	status := "connected"
	if err != nil {
		status = "error"
	}
	return &adapter.ProviderMetadata{
		ProviderID:   "awsS3",
		DisplayName:  "AWS S3",
		Status:       status,
		LatencyMs:    latency,
		Region:       a.Region,
		Capabilities: map[string]interface{}{"maxPartSize": 5242880},
		QuotaTotal:   0,
		QuotaUsed:    0,
		LastCheck:    time.Now().UTC().Format(time.RFC3339),
	}, nil
}

func (a *S3Adapter) UploadShard(ctx context.Context, fileID string, index int, data io.Reader) (string, error) {
	key := fmt.Sprintf("shards/%s_%03d", fileID, index)
	_, err := a.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(a.Bucket),
		Key:    aws.String(key),
		Body:   data,
	})
	if err != nil {
		return "", fmt.Errorf("s3: upload shard: %w", err)
	}
	return key, nil
}

func (a *S3Adapter) DownloadShard(ctx context.Context, remoteID string) (io.ReadCloser, error) {
	result, err := a.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(a.Bucket),
		Key:    aws.String(remoteID),
	})
	if err != nil {
		return nil, fmt.Errorf("s3: download shard: %w", err)
	}
	return result.Body, nil
}

func (a *S3Adapter) DeleteShard(ctx context.Context, remoteID string) error {
	_, err := a.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(a.Bucket),
		Key:    aws.String(remoteID),
	})
	if err != nil {
		var noSuchKey *types.NoSuchKey
		if errors.As(err, &noSuchKey) {
			return nil
		}
		return fmt.Errorf("s3: delete shard: %w", err)
	}
	return nil
}

func (a *S3Adapter) HealthCheck(ctx context.Context) error {
	_, err := a.client.HeadBucket(ctx, &s3.HeadBucketInput{Bucket: aws.String(a.Bucket)})
	if err != nil {
		return fmt.Errorf("s3: health check: %w", err)
	}
	return nil
}
