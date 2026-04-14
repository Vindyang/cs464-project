package adapterclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/vindyang/cs464-project/backend/services/shared/types"
)

// Client calls the adapter service.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &Client{baseURL: baseURL, httpClient: httpClient}
}

func (c *Client) GetProviders(ctx context.Context) ([]types.ProviderInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/api/providers", nil)
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

func (c *Client) UploadShard(ctx context.Context, shardID string, provider string, data []byte) (*types.UploadShardResp, error) {
	buf := &bytes.Buffer{}
	writer := multipart.NewWriter(buf)

	if err := writer.WriteField("shard_id", shardID); err != nil {
		return nil, fmt.Errorf("failed to write shard_id: %w", err)
	}
	if err := writer.WriteField("provider", provider); err != nil {
		return nil, fmt.Errorf("failed to write provider: %w", err)
	}
	if err := writer.WriteField("size", fmt.Sprintf("%d", len(data))); err != nil {
		return nil, fmt.Errorf("failed to write size: %w", err)
	}

	part, err := writer.CreateFormFile("file_data", "shard")
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}
	if _, err := part.Write(data); err != nil {
		return nil, fmt.Errorf("failed to write shard data: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to finalize form data: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/shards/upload", buf)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to upload shard: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("adapter returned %d: %s", resp.StatusCode, body)
	}

	var out types.UploadShardResp
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("failed to decode upload response: %w", err)
	}

	return &out, nil
}

func (c *Client) DownloadShard(ctx context.Context, remoteID string, provider string) ([]byte, error) {
	url := fmt.Sprintf("%s/shards/%s?provider=%s", c.baseURL, remoteID, provider)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
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

func (c *Client) DeleteShard(ctx context.Context, remoteID string, provider string) error {
	url := fmt.Sprintf("%s/shards/%s?provider=%s", c.baseURL, remoteID, provider)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
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

// DeleteFile asks the adapter to delete a file (including remote shards when deleteShards is true).
func (c *Client) DeleteFile(ctx context.Context, fileID string, deleteShards bool) error {
	url := fmt.Sprintf("%s/api/v1/files/%s", c.baseURL, fileID)
	if deleteShards {
		url += "?delete_shards=true"
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create delete request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("adapter returned %d: %s", resp.StatusCode, body)
	}
	return nil
}
