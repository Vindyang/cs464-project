package adapter

import (
	"context"
	"errors"
	"fmt"
	"io"
)

var ErrShardNotFound = errors.New("shard not found")

type ProviderMetadata struct {
	ProviderID   string                 `json:"providerId"`
	DisplayName  string                 `json:"displayName"`
	Status       string                 `json:"status"`
	LatencyMs    int64                  `json:"latencyMs"`
	Region       string                 `json:"region"`
	Capabilities map[string]interface{} `json:"capabilities"`
	QuotaTotal   int64                  `json:"quotaTotalBytes"`
	QuotaUsed    int64                  `json:"quotaUsedBytes"`
	LastCheck    string                 `json:"lastHealthCheckAt"`
}

type StorageProvider interface {
	GetMetadata(ctx context.Context) (*ProviderMetadata, error)
	UploadShard(ctx context.Context, fileID string, index int, data io.Reader) (string, error)
	DownloadShard(ctx context.Context, remoteID string) (io.ReadCloser, error)
	DeleteShard(ctx context.Context, remoteID string) error
	HealthCheck(ctx context.Context) error
}

type Registry struct {
	providers map[string]StorageProvider
}

func NewRegistry() *Registry {
	return &Registry{providers: make(map[string]StorageProvider)}
}

func (r *Registry) Register(id string, provider StorageProvider) {
	r.providers[id] = provider
}

func (r *Registry) Unregister(id string) {
	delete(r.providers, id)
}

func (r *Registry) Get(id string) (StorageProvider, error) {
	p, ok := r.providers[id]
	if !ok {
		return nil, fmt.Errorf("provider %s not found", id)
	}
	return p, nil
}

func (r *Registry) IDs() []string {
	ids := make([]string, 0, len(r.providers))
	for id := range r.providers {
		ids = append(ids, id)
	}
	return ids
}
