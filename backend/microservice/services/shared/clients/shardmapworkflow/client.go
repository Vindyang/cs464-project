package shardmapworkflow

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/vindyang/cs464-project/backend/services/shared/types"
)

// Client calls the shardmap service using workflow-level DTOs.
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

func (c *Client) RegisterFile(ctx context.Context, req *types.RegisterFileReq) (*types.RegisterFileResp, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v1/shards/register", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to register file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("shard map returned %d: %s", resp.StatusCode, respBody)
	}

	var out types.RegisterFileResp
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &out, nil
}

func (c *Client) RecordShards(ctx context.Context, req *types.RecordShardReq) (*types.RecordShardResp, error) {
	if len(req.Shards) == 0 {
		return nil, fmt.Errorf("must provide at least 1 shard")
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v1/shards/record", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to record shards: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("shard map returned %d: %s", resp.StatusCode, respBody)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Some shard-map implementations return 201 with an empty body for record calls.
	// Treat that as success for compatibility instead of failing with EOF.
	if len(strings.TrimSpace(string(respBody))) == 0 {
		return &types.RecordShardResp{FileID: req.FileID}, nil
	}

	var out types.RecordShardResp
	if err := json.Unmarshal(respBody, &out); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &out, nil
}

func (c *Client) GetShardMap(ctx context.Context, fileID string) (*types.GetShardMapResp, error) {
	url := fmt.Sprintf("%s/api/v1/shards/file/%s", c.baseURL, fileID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get shard map: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("shard map returned %d: %s", resp.StatusCode, respBody)
	}

	var out types.GetShardMapResp
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &out, nil
}

func (c *Client) ListFiles(ctx context.Context) ([]types.FileMetadata, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/api/v1/files", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("shard map returned %d: %s", resp.StatusCode, respBody)
	}

	var out []types.FileMetadata
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return out, nil
}

func (c *Client) MarkShardStatus(ctx context.Context, shardID string, status string) error {
	payload := map[string]string{"status": status}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/shards/%s/status", c.baseURL, shardID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to mark shard status: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("shard map returned %d: %s", resp.StatusCode, respBody)
	}
	return nil
}

func (c *Client) UpdateFileHealthRefresh(ctx context.Context, fileID string, refreshedAt time.Time) error {
	payload := map[string]string{"refreshed_at": refreshedAt.UTC().Format(time.RFC3339)}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal refresh timestamp: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/files/%s/health-refresh", c.baseURL, fileID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create health refresh request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to update file health refresh time: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("shard map returned %d: %s", resp.StatusCode, respBody)
	}
	return nil
}

// LogLifecycleEvent sends a lifecycle event (upload / download) to the shardmap
// service for persistence. The caller should treat failures as non-fatal.
func (c *Client) LogLifecycleEvent(ctx context.Context, event *types.LifecycleEvent) error {
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal lifecycle event: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v1/lifecycle", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create lifecycle request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to log lifecycle event: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusNoContent {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("shardmap lifecycle returned %d: %s", resp.StatusCode, respBody)
	}
	return nil
}

// GetFileHistory retrieves all lifecycle events for a given file from the shardmap service.
func (c *Client) GetFileHistory(ctx context.Context, fileID string) (*types.FileHistoryResp, error) {
	url := fmt.Sprintf("%s/api/v1/lifecycle/%s", c.baseURL, fileID)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create history request: %w", err)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to get file history: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("shardmap history returned %d: %s", resp.StatusCode, respBody)
	}

	var out types.FileHistoryResp
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("failed to decode history response: %w", err)
	}
	return &out, nil
}

// GetAllHistory retrieves lifecycle events across all files.
func (c *Client) GetAllHistory(ctx context.Context) (*types.GlobalHistoryResp, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/api/v1/lifecycle", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create global history request: %w", err)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to get global lifecycle history: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("shardmap global history returned %d: %s", resp.StatusCode, respBody)
	}

	var out types.GlobalHistoryResp
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("failed to decode global history response: %w", err)
	}
	return &out, nil
}

// DeleteAllHistory removes all lifecycle events from the shardmap service.
func (c *Client) DeleteAllHistory(ctx context.Context) (int, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.baseURL+"/api/v1/lifecycle", nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create delete history request: %w", err)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return 0, fmt.Errorf("failed to delete lifecycle history: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		respBody, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("shardmap delete lifecycle returned %d: %s", resp.StatusCode, respBody)
	}

	var out struct {
		DeletedEvents int `json:"deleted_events"`
	}
	if resp.StatusCode == http.StatusNoContent {
		return 0, nil
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return 0, fmt.Errorf("failed to decode delete lifecycle response: %w", err)
	}
	return out.DeletedEvents, nil
}
