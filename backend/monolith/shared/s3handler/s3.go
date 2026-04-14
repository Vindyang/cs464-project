package s3handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/vindyang/cs464-project/backend/monolith/shared/adapter"
	s3adapter "github.com/vindyang/cs464-project/backend/monolith/shared/adapter/s3"
	"github.com/vindyang/cs464-project/backend/monolith/shared/db"
)

const bucketConfigKey = "awsS3_bucket"

var awsRegionPattern = regexp.MustCompile(`^[a-z]{2}(?:-gov)?-[a-z]+(?:-[a-z]+)*-[0-9]+$`)

type S3Handler struct {
	store    *db.Store
	registry *adapter.Registry
}

func New(store *db.Store, registry *adapter.Registry) *S3Handler {
	return &S3Handler{store: store, registry: registry}
}

func (h *S3Handler) RestoreAdapter() error {
	adp, err := h.buildAdapter(context.Background())
	if err != nil {
		log.Println("No stored S3 credentials — connect via UI")
		return nil
	}
	if err := adp.HealthCheck(context.Background()); err != nil {
		return fmt.Errorf("s3handler: restore health check failed: %w", err)
	}
	h.registry.Register("awsS3", adp)
	log.Println("S3 adapter restored from stored credentials")
	return nil
}

func (h *S3Handler) Connect(w http.ResponseWriter, r *http.Request) {
	adp, err := h.buildAdapter(r.Context())
	if err != nil {
		http.Error(w, "S3 credentials not configured: "+err.Error(), http.StatusBadRequest)
		return
	}
	if err := adp.HealthCheck(r.Context()); err != nil {
		http.Error(w, "S3 connection failed: "+err.Error(), http.StatusBadGateway)
		return
	}
	h.registry.Register("awsS3", adp)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "connected"})
}

func (h *S3Handler) Disconnect(w http.ResponseWriter, r *http.Request) {
	h.registry.Unregister("awsS3")
	w.WriteHeader(http.StatusNoContent)
}

func (h *S3Handler) buildAdapter(ctx context.Context) (*s3adapter.S3Adapter, error) {
	accessKeyID, secretAccessKey, region, err := h.store.LoadCredentials("awsS3")
	if err != nil {
		return nil, fmt.Errorf("load credentials: %w", err)
	}
	if err := ValidateRegion(region); err != nil {
		return nil, fmt.Errorf("invalid aws region: %w", err)
	}

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region), config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, "")))
	if err != nil {
		return nil, fmt.Errorf("aws config: %w", err)
	}

	bucket, err := h.ensureBucket(ctx, cfg, region)
	if err != nil {
		return nil, fmt.Errorf("ensure bucket: %w", err)
	}

	return s3adapter.NewS3Adapter(accessKeyID, secretAccessKey, region, bucket)
}

func (h *S3Handler) ensureBucket(ctx context.Context, cfg aws.Config, region string) (string, error) {
	if cached, err := h.store.GetConfig(bucketConfigKey); err == nil && cached != "" {
		if isValidBucketName(cached) {
			return cached, nil
		}
		log.Printf("s3handler: ignoring invalid cached bucket name %q", cached)
	}

	stsClient := sts.NewFromConfig(cfg)
	identity, err := stsClient.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return "", fmt.Errorf("get caller identity: %w", err)
	}
	bucket := "omnishard-" + aws.ToString(identity.Account)

	client := s3.NewFromConfig(cfg)
	if err := createBucketIfNotExists(ctx, client, bucket, region); err != nil {
		return "", err
	}

	if err := h.store.SetConfig(bucketConfigKey, bucket); err != nil {
		log.Printf("s3handler: failed to cache bucket name: %v", err)
	}

	return bucket, nil
}

func createBucketIfNotExists(ctx context.Context, client *s3.Client, bucket, region string) error {
	_, err := client.HeadBucket(ctx, &s3.HeadBucketInput{Bucket: aws.String(bucket)})
	if err == nil {
		return nil
	}

	input := createBucketInput(bucket, region)
	_, createErr := client.CreateBucket(ctx, input)
	if createErr != nil {
		var alreadyOwned *s3types.BucketAlreadyOwnedByYou
		if errors.As(createErr, &alreadyOwned) {
			return nil
		}
		return fmt.Errorf("create bucket %q: %w", bucket, createErr)
	}

	log.Printf("s3handler: created bucket %q in %s", bucket, region)
	return nil
}

func ValidateRegion(region string) error {
	region = strings.TrimSpace(region)
	if region == "" {
		return errors.New("region is required")
	}
	if !awsRegionPattern.MatchString(region) {
		return fmt.Errorf("%q is not a valid AWS region", region)
	}
	return nil
}

func bucketNameForAccount(accountID string) (string, error) {
	bucket := strings.ToLower("omnishard-" + strings.TrimSpace(accountID))
	if !isValidBucketName(bucket) {
		return "", fmt.Errorf("generated invalid bucket name %q", bucket)
	}
	return bucket, nil
}

func isValidBucketName(bucket string) bool {
	if len(bucket) < 3 || len(bucket) > 63 {
		return false
	}
	for _, char := range bucket {
		switch {
		case char >= 'a' && char <= 'z':
		case char >= '0' && char <= '9':
		case char == '-':
		default:
			return false
		}
	}
	if bucket[0] == '-' || bucket[len(bucket)-1] == '-' {
		return false
	}
	return true
}

func createBucketInput(bucket, region string) *s3.CreateBucketInput {
	input := &s3.CreateBucketInput{Bucket: aws.String(bucket)}
	if region != "us-east-1" {
		input.CreateBucketConfiguration = &s3types.CreateBucketConfiguration{LocationConstraint: s3types.BucketLocationConstraint(region)}
	}
	return input
}
