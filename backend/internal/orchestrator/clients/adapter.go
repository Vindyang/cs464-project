package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/vindyang/cs464-project/backend/internal/types"
)

// AdapterClient calls Adapter Service at ADAPTER_URL (:8080)
type AdapterClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewAdapterClient(baseURL string, httpClient *http.Client) *AdapterClient {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &AdapterClient{
		baseURL:    baseURL,
		httpClient: httpClient,
	}
}

// GetProviders calls GET /api/providers
func (c *AdapterClient) GetProviders(ctx context.Context) ([]types.ProviderInfo, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/api/providers", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get providers: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("adapter returned %d: %s", resp.StatusCode, body)
	}

	var providers []types.ProviderInfo
	if err := json.NewDecoder(resp.Body).Decode(&providers); err != nil {
		return nil, fmt.Errorf("failed to decode providers: %w", err)
	}
	return providers, nil
}

// UploadShard calls POST /shards/upload with multipart form
func (c *AdapterClient) UploadShard(ctx context.Context, shardID string, provider string, data []byte) (*types.UploadShardResp, error) {
	buf := &bytes.Buffer{}
	w := multipart.NewWriter(buf)

	// Add form fields
	w.WriteField("shard_id", shardID)
	w.WriteField("provider", provider)
	w.WriteField("size", fmt.Sprintf("%d", len(data)))

	// Add file data
	part, err := w.CreateFormFile("file_data", "shard")
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}
	if _, err := part.Write(data); err != nil {
		return nil, fmt.Errorf("failed to write to form: %w", err)
	}
	w.Close()

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/shards/upload", buf)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to upload shard: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("adapter returned %d: %s", resp.StatusCode, body)
	}

	var uploadResp types.UploadShardResp
	if err := json.NewDecoder(resp.Body).Decode(&uploadResp); err != nil {
		return nil, fmt.Errorf("failed to decode upload response: %w", err)
	}
	return &uploadResp, nil
}

// DownloadShard calls GET /shards/{remoteId}?provider=xxx
func (c *AdapterClient) DownloadShard(ctx context.Context, remoteID string, provider string) ([]byte, error) {
	url := fmt.Sprintf("%s/shards/%s?provider=%s", c.baseURL, remoteID, provider)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download shard: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("adapter returned %d: %s", resp.StatusCode, body)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read shard data: %w", err)
	}
	return data, nil
}

// DeleteShard calls DELETE /shards/{remoteId}?provider=xxx
func (c *AdapterClient) DeleteShard(ctx context.Context, remoteID string, provider string) error {
	url := fmt.Sprintf("%s/shards/%s?provider=%s", c.baseURL, remoteID, provider)
	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete shard: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("adapter returned %d: %s", resp.StatusCode, body)
	}
	return nil
}
