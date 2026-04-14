package mock

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	sharedadapter "github.com/vindyang/cs464-project/backend/monolith/shared/adapter"
)

// Provider stores shards in memory so the monolith can run end-to-end without cloud credentials.
type Provider struct {
	mu     sync.RWMutex
	shards map[string][]byte
}

func NewProvider() *Provider {
	return &Provider{shards: map[string][]byte{}}
}

func (p *Provider) GetMetadata(context.Context) (*sharedadapter.ProviderMetadata, error) {
	p.mu.RLock()
	count := len(p.shards)
	p.mu.RUnlock()

	return &sharedadapter.ProviderMetadata{
		ProviderID:  "mockLocal",
		DisplayName: "Mock Local Provider",
		Status:      "connected",
		LatencyMs:   1,
		Region:      "local",
		Capabilities: map[string]interface{}{
			"mock_mode": true,
			"storage":   "in_memory",
		},
		QuotaTotal: int64(1024 * 1024 * 1024),
		QuotaUsed:  int64(count),
		LastCheck:  time.Now().UTC().Format(time.RFC3339),
	}, nil
}

func (p *Provider) UploadShard(_ context.Context, fileID string, index int, data io.Reader) (string, error) {
	payload, err := io.ReadAll(data)
	if err != nil {
		return "", fmt.Errorf("failed to read shard payload: %w", err)
	}

	remoteID := fmt.Sprintf("mock/%s/%d/%d", fileID, index, time.Now().UnixNano())

	p.mu.Lock()
	p.shards[remoteID] = payload
	p.mu.Unlock()

	return remoteID, nil
}

func (p *Provider) DownloadShard(_ context.Context, remoteID string) (io.ReadCloser, error) {
	p.mu.RLock()
	payload, ok := p.shards[remoteID]
	p.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("%w: %s", sharedadapter.ErrShardNotFound, remoteID)
	}
	return io.NopCloser(bytes.NewReader(payload)), nil
}

func (p *Provider) DeleteShard(_ context.Context, remoteID string) error {
	p.mu.Lock()
	delete(p.shards, remoteID)
	p.mu.Unlock()
	return nil
}

func (p *Provider) HealthCheck(context.Context) error {
	return nil
}
