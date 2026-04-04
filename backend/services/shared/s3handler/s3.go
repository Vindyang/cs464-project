// Package s3handler implements HTTP handlers for connecting and disconnecting the AWS S3 provider.
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
	"github.com/vindyang/cs464-project/backend/services/shared/adapter"
	s3adapter "github.com/vindyang/cs464-project/backend/services/shared/adapter/s3"
	"github.com/vindyang/cs464-project/backend/services/shared/db"
)

const bucketConfigKey = "awsS3_bucket"

var awsRegionPattern = regexp.MustCompile(`^[a-z]{2}(?:-gov)?-[a-z]+(?:-[a-z]+)*-[0-9]+$`)

// S3Handler handles connect and disconnect for the AWS S3 provider.
type S3Handler struct {
	store    *db.Store
	registry *adapter.Registry
}

// New constructs an S3Handler.
func New(store *db.Store, registry *adapter.Registry) *S3Handler {
	return &S3Handler{store: store, registry: registry}
}

// RestoreAdapter is called on server startup. It reads stored credentials and
// re-registers the adapter. Non-fatal: logs and returns nil if no credentials found.
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

// Connect handles POST /api/providers/awsS3/connect.
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

// Disconnect handles POST /api/providers/awsS3/disconnect.
func (h *S3Handler) Disconnect(w http.ResponseWriter, r *http.Request) {
	h.registry.Unregister("awsS3")
	w.WriteHeader(http.StatusNoContent)
}

// buildAdapter loads stored credentials, ensures the Omnishard bucket exists, and
// constructs an S3Adapter. The bucket name (Omnishard-<accountID>) is cached in
// provider_config so subsequent startups don't need an extra STS call.
func (h *S3Handler) buildAdapter(ctx context.Context) (*s3adapter.S3Adapter, error) {
	// region is stored in the redirect_uri column
	accessKeyID, secretAccessKey, region, err := h.store.LoadCredentials("awsS3")
	if err != nil {
		return nil, fmt.Errorf("load credentials: %w", err)
	}
	if err := ValidateRegion(region); err != nil {
		return nil, fmt.Errorf("invalid aws region: %w", err)
	}

	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, "")),
	)
	if err != nil {
		return nil, fmt.Errorf("aws config: %w", err)
	}

	bucket, err := h.ensureBucket(ctx, cfg, region)
	if err != nil {
		return nil, fmt.Errorf("ensure bucket: %w", err)
	}

	return s3adapter.NewS3Adapter(accessKeyID, secretAccessKey, region, bucket)
}

// ensureBucket returns the Omnishard bucket name for this account, creating it if needed.
// The name is cached in provider_config to avoid STS calls on every restart.
func (h *S3Handler) ensureBucket(ctx context.Context, cfg aws.Config, region string) (string, error) {
	// Check cache first
	if cached, err := h.store.GetConfig(bucketConfigKey); err == nil && cached != "" {
		if isValidBucketName(cached) {
			return cached, nil
		}
		log.Printf("s3handler: ignoring invalid cached bucket name %q", cached)
	}

	// Derive bucket name from AWS account ID
	stsClient := sts.NewFromConfig(cfg)
	identity, err := stsClient.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return "", fmt.Errorf("get caller identity: %w", err)
	}
	bucket := "omnishard-" + aws.ToString(identity.Account)

	// Create bucket if it doesn't exist
	s3Client := s3.NewFromConfig(cfg)
	if err := createBucketIfNotExists(ctx, s3Client, bucket, region); err != nil {
		return "", err
	}

	// Persist to config cache
	if err := h.store.SetConfig(bucketConfigKey, bucket); err != nil {
		log.Printf("s3handler: failed to cache bucket name: %v", err)
	}

	return bucket, nil
}

func createBucketIfNotExists(ctx context.Context, client *s3.Client, bucket, region string) error {
	// HeadBucket to check existence
	_, err := client.HeadBucket(ctx, &s3.HeadBucketInput{Bucket: aws.String(bucket)})
	if err == nil {
		return nil // bucket already exists and we have access
	}

	input := createBucketInput(bucket, region)

	_, createErr := client.CreateBucket(ctx, input)
	if createErr != nil {
		// BucketAlreadyOwnedByYou is fine
		var alreadyOwned *s3types.BucketAlreadyOwnedByYou
		if errors.As(createErr, &alreadyOwned) {
			return nil
		}
		return fmt.Errorf("create bucket %q: %w", bucket, createErr)
	}

	log.Printf("s3handler: created bucket %q in %s", bucket, region)
	return nil
}

// ValidateRegion checks whether the configured AWS region looks like a valid region identifier.
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
		input.CreateBucketConfiguration = &s3types.CreateBucketConfiguration{
			LocationConstraint: s3types.BucketLocationConstraint(region),
		}
	}
	return input
}
