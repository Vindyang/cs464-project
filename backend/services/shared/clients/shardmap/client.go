package shardmap

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/google/uuid"
	"github.com/vindyang/cs464-project/backend/services/shared/api/dto"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	return &Client{
		baseURL:    baseURL,
		httpClient: httpClient,
	}
}

func (c *Client) RegisterFile(req *dto.RegisterFileRequest) (*dto.RegisterFileResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal register request: %w", err)
	}

	httpReq, err := http.NewRequest(http.MethodPost, c.baseURL+"/api/v1/shards/register", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create register request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("call shardmap service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("shardmap service returned %d: %s", resp.StatusCode, string(respBody))
	}

	var out dto.RegisterFileResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("decode register response: %w", err)
	}

	return &out, nil
}

func (c *Client) RecordShards(req *dto.RecordShardsRequest) (*dto.RecordShardsResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal record request: %w", err)
	}

	httpReq, err := http.NewRequest(http.MethodPost, c.baseURL+"/api/v1/shards/record", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create record request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("call shardmap service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("shardmap service returned %d: %s", resp.StatusCode, string(respBody))
	}

	var out dto.RecordShardsResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("decode record response: %w", err)
	}

	return &out, nil
}

func (c *Client) GetShardMap(fileID uuid.UUID) (*dto.GetShardMapResponse, error) {
	httpReq, err := http.NewRequest(http.MethodGet, c.baseURL+"/api/v1/shards/file/"+fileID.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("create get shard map request: %w", err)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("call shardmap service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("shardmap service returned %d: %s", resp.StatusCode, string(respBody))
	}

	var out dto.GetShardMapResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("decode get shard map response: %w", err)
	}

	return &out, nil
}

func (c *Client) GetShardByID(shardID uuid.UUID) (*dto.ShardInfo, error) {
	httpReq, err := http.NewRequest(http.MethodGet, c.baseURL+"/api/v1/shards/"+shardID.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("create get shard request: %w", err)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("call shardmap service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("shardmap service returned %d: %s", resp.StatusCode, string(respBody))
	}

	var out dto.ShardInfo
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("decode get shard response: %w", err)
	}

	return &out, nil
}

func (c *Client) MarkShardStatus(shardID uuid.UUID, req *dto.MarkShardStatusRequest) error {
	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshal status request: %w", err)
	}

	httpReq, err := http.NewRequest(http.MethodPut, c.baseURL+"/api/v1/shards/"+shardID.String()+"/status", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create status request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("call shardmap service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("shardmap service returned %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}
