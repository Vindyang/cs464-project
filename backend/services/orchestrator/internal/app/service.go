package app

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/vindyang/cs464-project/backend/services/shared/types"
	"golang.org/x/sync/errgroup"
)

const (
	TotalShards    = 6
	RequiredShards = 4
)

type Service struct {
	adapter  AdapterClient
	shardMap ShardMapClient
	sharding ShardingClient
}

type AdapterClient interface {
	GetProviders(ctx context.Context) ([]types.ProviderInfo, error)
	UploadShard(ctx context.Context, shardID string, provider string, data []byte) (*types.UploadShardResp, error)
	DownloadShard(ctx context.Context, remoteID string, provider string) ([]byte, error)
	DeleteShard(ctx context.Context, remoteID string, provider string) error
	DeleteFile(ctx context.Context, fileID string, deleteShards bool) error
}

type ShardMapClient interface {
	RegisterFile(ctx context.Context, req *types.RegisterFileReq) (*types.RegisterFileResp, error)
	RecordShards(ctx context.Context, req *types.RecordShardReq) (*types.RecordShardResp, error)
	GetShardMap(ctx context.Context, fileID string) (*types.GetShardMapResp, error)
	MarkShardStatus(ctx context.Context, shardID string, status string) error
	// LogLifecycleEvent records a file operation lifecycle event in the shardmap service.
	// Callers treat failures as non-fatal (fire-and-forget).
	LogLifecycleEvent(ctx context.Context, event *types.LifecycleEvent) error
	// GetFileHistory retrieves all lifecycle events for a file from the shardmap service.
	GetFileHistory(ctx context.Context, fileID string) (*types.FileHistoryResp, error)
}

type ShardingClient interface {
	EncodeChunk(chunkData []byte, k, n int) ([][]byte, error)
	DecodeChunk(shards [][]byte, k, n int) ([]byte, error)
}

func NewService(adapter AdapterClient, shardMap ShardMapClient) *Service {
	return &Service{
		adapter:  adapter,
		shardMap: shardMap,
	}
}

func NewServiceWithSharding(adapter AdapterClient, shardMap ShardMapClient, sharding ShardingClient) *Service {
	return &Service{
		adapter:  adapter,
		shardMap: shardMap,
		sharding: sharding,
	}
}

// logEvent sends a lifecycle event to the shardmap service.
// Failures are logged but never propagate to the caller.
func (s *Service) logEvent(event *types.LifecycleEvent) {
	// Use a background context so a cancelled request context doesn't prevent logging.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.shardMap.LogLifecycleEvent(ctx, event); err != nil {
		log.Printf("[lifecycle] failed to log %s event for file %s: %v", event.EventType, event.FileID, err)
	}
}

// GetFileHistory returns the lifecycle event history for a file by proxying
// the request through to the shardmap service.
func (s *Service) GetFileHistory(ctx context.Context, fileID string) (*types.FileHistoryResp, error) {
	return s.shardMap.GetFileHistory(ctx, fileID)
}

// DeleteFile deletes a file via the adapter service and logs a lifecycle event.
func (s *Service) DeleteFile(ctx context.Context, fileID string, deleteShards bool) error {
	startedAt := time.Now()

	// Fetch metadata before deletion so we can record it in the lifecycle event.
	var fileName string
	var fileSize int64
	var shardCount int
	var providers []string
	if shardMap, err := s.shardMap.GetShardMap(ctx, fileID); err == nil {
		fileName = shardMap.OriginalName
		shardCount = len(shardMap.Shards)
		// Collect unique providers.
		seen := map[string]struct{}{}
		for _, shard := range shardMap.Shards {
			p := strings.TrimSpace(shard.Provider)
			if p != "" {
				if _, ok := seen[p]; !ok {
					seen[p] = struct{}{}
					providers = append(providers, p)
				}
			}
		}
	}

	// Perform the actual delete.
	err := s.adapter.DeleteFile(ctx, fileID, deleteShards)

	endedAt := time.Now()
	durationMs := endedAt.Sub(startedAt).Milliseconds()

	status := "success"
	errMsg := ""
	if err != nil {
		status = "failed"
		errMsg = err.Error()
	}

	event := &types.LifecycleEvent{
		FileID:     fileID,
		EventType:  "delete",
		FileName:   fileName,
		FileSize:   fileSize,
		ShardCount: shardCount,
		Providers:  providers,
		StartedAt:  startedAt,
		EndedAt:    endedAt,
		DurationMs: durationMs,
		Status:     status,
		ErrorMsg:   errMsg,
	}
	s.logEvent(event)

	return err
}

func (s *Service) UploadRawFile(ctx context.Context, fileName string, fileData []byte, k int, n int) (*types.UploadResp, error) {
	if s.sharding == nil {
		return nil, fmt.Errorf("sharding client is not configured")
	}

	if k <= 0 || n <= 0 || k > n {
		return nil, fmt.Errorf("invalid erasure coding parameters: k=%d, n=%d", k, n)
	}

	if len(fileData) == 0 {
		return nil, fmt.Errorf("file data cannot be empty")
	}

	shards, err := s.sharding.EncodeChunk(fileData, k, n)
	if err != nil {
		return nil, fmt.Errorf("failed to shard file data: %w", err)
	}

	isDataShard := make([]bool, len(shards))
	for i := range isDataShard {
		isDataShard[i] = i < k
	}

	return s.UploadFile(ctx, fileName, shards, isDataShard)
}

func (s *Service) UploadFile(ctx context.Context, fileName string, shards [][]byte, isDataShard []bool) (*types.UploadResp, error) {
	startedAt := time.Now()

	resp, err := s.uploadFileInternal(ctx, fileName, shards, isDataShard)

	endedAt := time.Now()
	durationMs := endedAt.Sub(startedAt).Milliseconds()

	// Determine the file ID for logging (may be empty on early failure).
	fileID := ""
	if resp != nil {
		fileID = resp.FileID
	}

	// Build and fire the lifecycle event (fire-and-forget).
	if fileID != "" {
		status := "success"
		errMsg := ""
		if err != nil {
			status = "failed"
			errMsg = err.Error()
		} else if resp != nil && resp.Status == "failed" {
			status = "failed"
			errMsg = resp.Error
		}

		providers := uniqueProviders(shards, resp)
		event := &types.LifecycleEvent{
			FileID:     fileID,
			EventType:  "upload",
			FileName:   fileName,
			FileSize:   totalShardSize(shards),
			ShardCount: len(shards),
			Providers:  providers,
			StartedAt:  startedAt,
			EndedAt:    endedAt,
			DurationMs: durationMs,
			Status:     status,
			ErrorMsg:   errMsg,
		}
		s.logEvent(event)
	}

	return resp, err
}

// uploadFileInternal contains the original upload logic, separated so timing wraps it cleanly.
func (s *Service) uploadFileInternal(ctx context.Context, fileName string, shards [][]byte, isDataShard []bool) (*types.UploadResp, error) {
	if len(shards) != TotalShards {
		return nil, fmt.Errorf("expected %d shards, got %d", TotalShards, len(shards))
	}
	if len(isDataShard) != len(shards) {
		return nil, fmt.Errorf("isDataShard length (%d) must match shards length (%d)", len(isDataShard), len(shards))
	}

	originalSize := int64(0)
	for _, shard := range shards {
		originalSize += int64(len(shard))
	}

	registerReq := &types.RegisterFileReq{
		OriginalName: fileName,
		OriginalSize: originalSize,
		TotalChunks:  TotalShards,
		N:            TotalShards,
		K:            RequiredShards,
		ShardSize:    int64(len(shards[0])),
	}

	fileResp, err := s.shardMap.RegisterFile(ctx, registerReq)
	if err != nil {
		return nil, fmt.Errorf("failed to register file: %w", err)
	}
	fileID := fileResp.FileID

	providers, err := s.adapter.GetProviders(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get providers: %w", err)
	}

	healthyProviders := []string{}
	for _, p := range providers {
		if p.Status == "connected" {
			healthyProviders = append(healthyProviders, p.ProviderID)
		}
	}

	if len(healthyProviders) < 1 {
		return nil, fmt.Errorf("insufficient healthy providers: need 1, have %d", len(healthyProviders))
	}

	allocation := make(map[int]string)
	for i := 0; i < TotalShards; i++ {
		allocation[i] = healthyProviders[i%len(healthyProviders)]
	}

	uploadResults := s.uploadShardsParallel(ctx, fileID, shards, allocation)

	successCount := 0
	for _, res := range uploadResults {
		if res.Success {
			successCount++
		}
	}

	if successCount < TotalShards {
		s.rollbackUpload(ctx, uploadResults)
		return &types.UploadResp{
			FileID: fileID,
			Status: "failed",
			Error:  fmt.Sprintf("only %d/%d shards succeeded, need all %d", successCount, TotalShards, TotalShards),
		}, nil
	}

	recordReq := &types.RecordShardReq{
		FileID: fileID,
		Shards: []types.ShardInfo{},
	}

	for i, res := range uploadResults {
		recordReq.Shards = append(recordReq.Shards, types.ShardInfo{
			ChunkIndex: i,
			ShardIndex: i,
			Type: func() string {
				if isDataShard[i] {
					return "DATA"
				}
				return "PARITY"
			}(),
			RemoteID:    res.RemoteID,
			Provider:    res.Provider,
			ChecksumSha: res.ChecksumSha,
		})
	}

	recordResp, err := s.shardMap.RecordShards(ctx, recordReq)
	if err != nil {
		s.rollbackUpload(ctx, uploadResults)
		return &types.UploadResp{
			FileID: fileID,
			Status: "failed",
			Error:  fmt.Sprintf("failed to record shards: %v", err),
		}, nil
	}

	for _, shard := range recordResp.Shards {
		if shard.ShardID == "" {
			return &types.UploadResp{
				FileID: fileID,
				Status: "failed",
				Error:  "failed to mark shard health: missing shard_id in record response",
			}, nil
		}
		if err := s.shardMap.MarkShardStatus(ctx, shard.ShardID, "HEALTHY"); err != nil {
			return &types.UploadResp{
				FileID: fileID,
				Status: "failed",
				Error:  fmt.Sprintf("failed to mark shard %s healthy: %v", shard.ShardID, err),
			}, nil
		}
	}

	shardInfos := []types.ShardInfo{}
	for i, res := range uploadResults {
		shardInfos = append(shardInfos, types.ShardInfo{
			ChunkIndex: i,
			ShardIndex: i,
			Type: func() string {
				if isDataShard[i] {
					return "data"
				}
				return "parity"
			}(),
			RemoteID:    res.RemoteID,
			Provider:    res.Provider,
			ChecksumSha: res.ChecksumSha,
		})
	}

	return &types.UploadResp{
		FileID: fileID,
		Status: "committed",
		Shards: shardInfos,
	}, nil
}

func (s *Service) uploadShardsParallel(ctx context.Context, fileID string, shards [][]byte, allocation map[int]string) []UploadResult {
	results := make([]UploadResult, len(shards))
	var wg sync.WaitGroup
	var mu sync.Mutex

	for i, shardData := range shards {
		wg.Add(1)
		go func(idx int, data []byte, provider string) {
			defer wg.Done()

			shardID := fmt.Sprintf("%s-shard-%d", fileID, idx)
			uploadResp, err := s.adapter.UploadShard(ctx, shardID, provider, data)

			mu.Lock()
			results[idx] = UploadResult{
				ShardIndex: idx,
				Provider:   provider,
				Success:    err == nil,
				Error:      err,
			}
			if err == nil {
				results[idx].RemoteID = uploadResp.RemoteID
				results[idx].ChecksumSha = uploadResp.ChecksumSha
			}
			mu.Unlock()
		}(i, shardData, allocation[i])
	}

	wg.Wait()
	return results
}

func (s *Service) rollbackUpload(ctx context.Context, results []UploadResult) {
	for _, res := range results {
		if res.Success {
			if err := s.adapter.DeleteShard(ctx, res.RemoteID, res.Provider); err != nil {
				fmt.Printf("rollback: failed to delete shard %s: %v\n", res.RemoteID, err)
			}
		}
	}
}

func (s *Service) DownloadFile(ctx context.Context, fileID string) (*types.DownloadResp, error) {
	startedAt := time.Now()

	resp, err := s.downloadFileInternal(ctx, fileID)

	endedAt := time.Now()
	durationMs := endedAt.Sub(startedAt).Milliseconds()

	// Fire-and-forget lifecycle log.
	status := "success"
	errMsg := ""
	if err != nil {
		status = "failed"
		errMsg = err.Error()
	}
	fileName := ""
	if resp != nil {
		fileName = resp.FileName
	}
	event := &types.LifecycleEvent{
		FileID:     fileID,
		EventType:  "download",
		FileName:   fileName,
		StartedAt:  startedAt,
		EndedAt:    endedAt,
		DurationMs: durationMs,
		Status:     status,
		ErrorMsg:   errMsg,
	}
	s.logEvent(event)

	return resp, err
}

// downloadFileInternal contains the original download logic.
func (s *Service) downloadFileInternal(ctx context.Context, fileID string) (*types.DownloadResp, error) {
	shardMap, err := s.shardMap.GetShardMap(ctx, fileID)
	if err != nil {
		return nil, fmt.Errorf("file not found: %w", err)
	}

	if len(shardMap.Shards) < shardMap.K {
		return nil, fmt.Errorf("insufficient shards available: have %d, need %d", len(shardMap.Shards), shardMap.K)
	}

	shards := s.downloadShardsParallelEarlyExit(ctx, shardMap.Shards, shardMap.K)

	if len(shards) < shardMap.K {
		return nil, fmt.Errorf("failed to download sufficient shards: got %d, need %d", len(shards), shardMap.K)
	}

	reconstructedData := []byte{}
	if s.sharding != nil {
		indexedShards := make([][]byte, shardMap.N)
		for _, shard := range shards {
			if shard.Index >= 0 && shard.Index < shardMap.N {
				indexedShards[shard.Index] = shard.Data
			}
		}

		reconstructedData, err = s.sharding.DecodeChunk(indexedShards, shardMap.K, shardMap.N)
		if err != nil {
			return nil, fmt.Errorf("failed to reconstruct file: %w", err)
		}
	} else {
		for i := 0; i < shardMap.K && i < len(shards); i++ {
			reconstructedData = append(reconstructedData, shards[i].Data...)
		}
	}

	return &types.DownloadResp{
		FileID:   fileID,
		FileName: shardMap.OriginalName,
		Shards:   [][]byte{reconstructedData},
	}, nil
}

func (s *Service) downloadShardsParallelEarlyExit(ctx context.Context, shardEntries []types.ShardMapEntry, k int) []DownloadResult {
	results := make([]DownloadResult, 0, len(shardEntries))
	resultsChan := make(chan DownloadResult, len(shardEntries))
	received := 0
	mu := sync.Mutex{}

	eg, egCtx := errgroup.WithContext(ctx)

	for _, entry := range shardEntries {
		entry := entry
		eg.Go(func() error {
			data, err := s.adapter.DownloadShard(egCtx, entry.RemoteID, entry.Provider)
			result := DownloadResult{
				RemoteID: entry.RemoteID,
				Provider: entry.Provider,
				Index:    entry.ShardIndex,
				Success:  err == nil,
				Error:    err,
				Arrived:  time.Now(),
			}
			if err == nil {
				result.Data = data
			}
			resultsChan <- result
			return nil
		})
	}

	go func() {
		eg.Wait()
		close(resultsChan)
	}()

	for result := range resultsChan {
		if result.Success {
			mu.Lock()
			received++
			results = append(results, result)
			mu.Unlock()

			if received >= k {
				break
			}
		}
	}

	for result := range resultsChan {
		if result.Success {
			results = append(results, result)
		}
	}

	return results
}

// totalShardSize returns the sum of all shard lengths (used as an approximation of file size).
func totalShardSize(shards [][]byte) int64 {
	var total int64
	for _, s := range shards {
		total += int64(len(s))
	}
	return total
}

// uniqueProviders extracts the distinct provider IDs used in a committed upload response.
func uniqueProviders(_ [][]byte, resp *types.UploadResp) []string {
	if resp == nil {
		return nil
	}
	seen := map[string]struct{}{}
	var out []string
	for _, si := range resp.Shards {
		p := si.Provider
		if p == "" {
			continue
		}
		// Trim any stray whitespace
		p = strings.TrimSpace(p)
		if _, ok := seen[p]; !ok {
			seen[p] = struct{}{}
			out = append(out, p)
		}
	}
	return out
}
