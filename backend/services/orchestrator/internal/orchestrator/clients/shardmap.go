package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/vindyang/cs464-project/backend/services/orchestrator/internal/types"
)

// ShardMapClient calls Shard Map Service at SHARDMAP_URL (:8081)
type ShardMapClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewShardMapClient(baseURL string, httpClient *http.Client) *ShardMapClient {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &ShardMapClient{
		baseURL:    baseURL,
		httpClient: httpClient,
	}
}

// RegisterFile calls POST /api/v1/shards/register
func (c *ShardMapClient) RegisterFile(ctx context.Context, req *types.RegisterFileReq) (*types.RegisterFileResp, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/v1/shards/register", bytes.NewReader(body))
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

	var regResp types.RegisterFileResp
	if err := json.NewDecoder(resp.Body).Decode(&regResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &regResp, nil
}

// RecordShards calls POST /api/v1/shards/record (with ALL N shards)
func (c *ShardMapClient) RecordShards(ctx context.Context, req *types.RecordShardReq) error {
	if len(req.Shards) == 0 {
		return fmt.Errorf("must provide at least 1 shard")
	}

	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/v1/shards/record", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to record shards: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("shard map returned %d: %s", resp.StatusCode, respBody)
	}
	return nil
}

// GetShardMap calls GET /api/v1/shards/file/{fileId}
func (c *ShardMapClient) GetShardMap(ctx context.Context, fileID string) (*types.GetShardMapResp, error) {
	url := fmt.Sprintf("%s/api/v1/shards/file/%s", c.baseURL, fileID)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
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

	var shardMap types.GetShardMapResp
	if err := json.NewDecoder(resp.Body).Decode(&shardMap); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &shardMap, nil
}

// MarkShardStatus calls PUT /api/v1/shards/{shardId}/status
func (c *ShardMapClient) MarkShardStatus(ctx context.Context, shardID string, status string) error {
	payload := map[string]string{"status": status}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/shards/%s/status", c.baseURL, shardID)
	req, err := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewReader(body))
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
