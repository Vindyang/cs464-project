package sharding

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
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

func (c *Client) ChunkFile(data []byte, chunkSize int64) ([][]byte, error) {
	if chunkSize <= 0 {
		return nil, fmt.Errorf("chunk size must be positive")
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("data cannot be empty")
	}

	var chunks [][]byte
	for offset := int64(0); offset < int64(len(data)); offset += chunkSize {
		end := offset + chunkSize
		if end > int64(len(data)) {
			end = int64(len(data))
		}

		chunk := make([]byte, end-offset)
		copy(chunk, data[offset:end])
		chunks = append(chunks, chunk)
	}

	return chunks, nil
}

func (c *Client) EncodeChunk(chunkData []byte, k, n int) ([][]byte, error) {
	reqBody := map[string]interface{}{
		"fileId":   uuid.NewString(),
		"fileData": chunkData,
		"n":        n,
		"k":        k,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal shard request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/api/sharding/shard", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create shard request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call sharding service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errPayload map[string]interface{}
		_ = json.NewDecoder(resp.Body).Decode(&errPayload)
		return nil, fmt.Errorf("sharding service returned %d: %v", resp.StatusCode, errPayload)
	}

	var out struct {
		Shards []struct {
			ShardData       []byte `json:"shardData"`
			LegacyShardData []byte `json:"shard_data"`
		} `json:"shards"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("decode shard response: %w", err)
	}

	shards := make([][]byte, 0, len(out.Shards))
	for _, shard := range out.Shards {
		if len(shard.ShardData) > 0 {
			shards = append(shards, shard.ShardData)
			continue
		}
		shards = append(shards, shard.LegacyShardData)
	}

	return shards, nil
}

func (c *Client) DecodeChunk(shards [][]byte, k, n int) ([]byte, error) {
	available := make([]map[string]interface{}, 0, len(shards))
	for idx, shard := range shards {
		if shard == nil {
			continue
		}

		shardType := "data"
		if idx >= k {
			shardType = "parity"
		}

		available = append(available, map[string]interface{}{
			"shard_index": idx,
			"shard_type":  shardType,
			"shard_data":  shard,
		})
	}

	reqBody := map[string]interface{}{
		"file_id":          uuid.NewString(),
		"N":                n,
		"K":                k,
		"available_shards": available,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal reconstruct request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/api/sharding/reconstruct", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create reconstruct request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call sharding service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errPayload map[string]interface{}
		_ = json.NewDecoder(resp.Body).Decode(&errPayload)
		return nil, fmt.Errorf("sharding service returned %d: %v", resp.StatusCode, errPayload)
	}

	var out struct {
		ReconstructedFile []byte `json:"reconstructed_file"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("decode reconstruct response: %w", err)
	}

	return out.ReconstructedFile, nil
}

func (c *Client) CalculateChecksum(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}
