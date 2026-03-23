package orchestrator

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/vindyang/cs464-project/backend/services/adapter/internal/orchestrator/clients"
	"github.com/vindyang/cs464-project/backend/services/adapter/internal/types"
	"golang.org/x/sync/errgroup"
)

const (
	// Configuration (from spec: N=6, K=4)
	TotalShards    = 6
	RequiredShards = 4
)

// Service coordinates upload/download operations
type Service struct {
	adapter  *clients.AdapterClient
	shardMap *clients.ShardMapClient
}

func NewService(adapter *clients.AdapterClient, shardMap *clients.ShardMapClient) *Service {
	return &Service{
		adapter:  adapter,
		shardMap: shardMap,
	}
}

// UploadFile orchestrates a full upload: register → upload shards in parallel → record locations
func (s *Service) UploadFile(ctx context.Context, fileName string, shards [][]byte, isDataShard []bool) (*types.UploadResp, error) {
	if len(shards) != TotalShards {
		return nil, fmt.Errorf("expected %d shards, got %d", TotalShards, len(shards))
	}

	// Step 1: Register file with Shard Map Service
	registerReq := &types.RegisterFileReq{
		OriginalName: fileName,
		OriginalSize: 0, // TODO: calculate from shards
		TotalChunks:  TotalShards,
		N:            TotalShards,
		K:            RequiredShards,
		ShardSize:    int64(len(shards[0])), // assume uniform
	}

	fileResp, err := s.shardMap.RegisterFile(ctx, registerReq)
	if err != nil {
		return nil, fmt.Errorf("failed to register file: %w", err)
	}
	fileID := fileResp.FileID

	// Step 2: Get provider health/quota to decide routing
	providers, err := s.adapter.GetProviders(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get providers: %w", err)
	}

	// Filter healthy providers (status == "connected")
	healthyProviders := []string{}
	for _, p := range providers {
		if p.Status == "connected" {
			healthyProviders = append(healthyProviders, p.ProviderID)
		}
	}

	if len(healthyProviders) < 2 {
		return nil, fmt.Errorf("insufficient healthy providers: need 2, have %d", len(healthyProviders))
	}

	// Step 3: Round-robin distribute shards across healthy providers
	allocation := make(map[int]string)
	for i := 0; i < TotalShards; i++ {
		allocation[i] = healthyProviders[i%len(healthyProviders)]
	}

	// Step 4: Upload all shards in parallel
	uploadResults := s.uploadShardsParallel(ctx, fileID, shards, allocation)

	// Step 5: Check results — need ALL N shards to succeed
	successCount := 0
	for _, res := range uploadResults {
		if res.Success {
			successCount++
		}
	}

	if successCount < TotalShards {
		// Rollback: delete succeeded shards
		s.rollbackUpload(ctx, uploadResults)
		return &types.UploadResp{
			FileID: fileID,
			Status: "failed",
			Error:  fmt.Sprintf("only %d/%d shards succeeded, need all %d", successCount, TotalShards, TotalShards),
		}, nil
	}

	// Step 6: Record all shard locations in Shard Map Service (with ALL N shards, failed ones have empty RemoteID)
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
					return "data"
				} else {
					return "parity"
				}
			}(),
			RemoteID:    res.RemoteID,
			Provider:    res.Provider,
			ChecksumSha: res.ChecksumSha,
		})
	}

	if err := s.shardMap.RecordShards(ctx, recordReq); err != nil {
		// Record failed; try to clean up
		s.rollbackUpload(ctx, uploadResults)
		return &types.UploadResp{
			FileID: fileID,
			Status: "failed",
			Error:  fmt.Sprintf("failed to record shards: %v", err),
		}, nil
	}

	// Step 7: Return success with all shards
	shardInfos := []types.ShardInfo{}
	for i, res := range uploadResults {
		shardInfos = append(shardInfos, types.ShardInfo{
			ChunkIndex: i,
			ShardIndex: i,
			Type: func() string {
				if isDataShard[i] {
					return "data"
				} else {
					return "parity"
				}
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

// uploadShardsParallel uploads all shards concurrently
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

// rollbackUpload deletes all successfully uploaded shards (Saga pattern)
func (s *Service) rollbackUpload(ctx context.Context, results []UploadResult) {
	for _, res := range results {
		if res.Success {
			// Attempt delete, but log errors — don't block rollback
			if err := s.adapter.DeleteShard(ctx, res.RemoteID, res.Provider); err != nil {
				// TODO: log this error for observability
				fmt.Printf("rollback: failed to delete shard %s: %v\n", res.RemoteID, err)
			}
		}
	}
}

// DownloadFile orchestrates a full download: query shard map → fetch shards → early exit once K arrive
func (s *Service) DownloadFile(ctx context.Context, fileID string) (*types.DownloadResp, error) {
	// Step 1: Query Shard Map for shard locations
	shardMap, err := s.shardMap.GetShardMap(ctx, fileID)
	if err != nil {
		return nil, fmt.Errorf("file not found: %w", err)
	}

	if len(shardMap.Shards) < shardMap.K {
		return nil, fmt.Errorf("insufficient shards available: have %d, need %d", len(shardMap.Shards), shardMap.K)
	}

	// Step 2: Download shards in parallel with early exit once K arrive
	shards := s.downloadShardsParallelEarlyExit(ctx, shardMap.Shards, shardMap.K)

	if len(shards) < shardMap.K {
		return nil, fmt.Errorf("failed to download sufficient shards: got %d, need %d", len(shards), shardMap.K)
	}

	// Step 3: Return first K shards (early exit means we don't have all, but enough for reconstruction)
	// NOTE: Actual Reed-Solomon reconstruction happens in frontend; here we just return the shards
	reconstructedData := []byte{}
	for i := 0; i < shardMap.K && i < len(shards); i++ {
		reconstructedData = append(reconstructedData, shards[i].Data...)
	}

	return &types.DownloadResp{
		FileID:   fileID,
		FileName: shardMap.OriginalName,
		Shards:   [][]byte{reconstructedData},
	}, nil
}

// downloadShardsParallelEarlyExit fetches shards concurrently and exits once K shards arrive
func (s *Service) downloadShardsParallelEarlyExit(ctx context.Context, shardEntries []types.ShardMapEntry, k int) []DownloadResult {
	results := make([]DownloadResult, 0, len(shardEntries))
	resultsChan := make(chan DownloadResult, len(shardEntries))
	received := 0
	mu := sync.Mutex{}

	eg, egCtx := errgroup.WithContext(ctx)

	// Launch goroutine for each shard
	for _, entry := range shardEntries {
		entry := entry // capture loop variable
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
			return nil // don't let errors block others
		})
	}

	// Collect results from channel until we have K or all goroutines done
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
				// Early exit: we have enough shards, cancel remaining goroutines
				// (but let them finish naturally — we just stop waiting)
				break
			}
		}
	}

	// Collect any remaining results that arrive (to let goroutines finish cleanly)
	for result := range resultsChan {
		if result.Success {
			results = append(results, result)
		}
	}

	return results
}
