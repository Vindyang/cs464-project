package orchestrator

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/vindyang/cs464-project/backend/monolith/shared/types"
	"golang.org/x/sync/errgroup"
)

const (
	TotalShards    = 6
	RequiredShards = 4
)

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
	ListFiles(ctx context.Context) ([]types.FileMetadata, error)
	GetShardMap(ctx context.Context, fileID string) (*types.GetShardMapResp, error)
	MarkShardStatus(ctx context.Context, shardID string, status string) error
	UpdateFileHealthRefresh(ctx context.Context, fileID string, refreshedAt time.Time) error
	LogLifecycleEvent(ctx context.Context, event *types.LifecycleEvent) error
	GetFileHistory(ctx context.Context, fileID string) (*types.FileHistoryResp, error)
	GetAllHistory(ctx context.Context) (*types.GlobalHistoryResp, error)
}

type ShardingClient interface {
	EncodeChunk(chunkData []byte, k, n int) ([][]byte, error)
	DecodeChunk(shards [][]byte, k, n int) ([]byte, error)
}

type Service struct {
	adapter  AdapterClient
	shardMap ShardMapClient
	sharding ShardingClient
}

type RecoverabilityError struct {
	FileID    string
	Available int
	Required  int
	Cause     string
}

func (e *RecoverabilityError) Error() string {
	if e.Cause == "" {
		return fmt.Sprintf("file %s cannot be reconstructed: only %d shards available, need %d", e.FileID, e.Available, e.Required)
	}
	return fmt.Sprintf("file %s cannot be reconstructed: %s (available %d, required %d)", e.FileID, e.Cause, e.Available, e.Required)
}

func NewServiceWithSharding(adapter AdapterClient, shardMap ShardMapClient, sharding ShardingClient) *Service {
	return &Service{adapter: adapter, shardMap: shardMap, sharding: sharding}
}

func (s *Service) logEvent(event *types.LifecycleEvent) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.shardMap.LogLifecycleEvent(ctx, event); err != nil {
		log.Printf("[lifecycle] failed to log %s event for file %s: %v", event.EventType, event.FileID, err)
	}
}

func (s *Service) GetFileHistory(ctx context.Context, fileID string) (*types.FileHistoryResp, error) {
	return s.shardMap.GetFileHistory(ctx, fileID)
}

func (s *Service) GetAllHistory(ctx context.Context) (*types.GlobalHistoryResp, error) {
	return s.shardMap.GetAllHistory(ctx)
}

func (s *Service) RefreshAllFileHealth(ctx context.Context) (*types.HealthRefreshSummary, error) {
	files, err := s.shardMap.ListFiles(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	summary := &types.HealthRefreshSummary{
		FilesScanned:  len(files),
		ErrorMessages: []string{},
		RefreshedAt:   time.Now().UTC().Format(time.RFC3339),
	}

	for _, file := range files {
		fileSummary, err := s.RefreshFileHealth(ctx, file.FileID)
		if err != nil {
			summary.FailedFiles++
			summary.ErrorMessages = append(summary.ErrorMessages, fmt.Sprintf("file %s: %v", file.FileID, err))
			continue
		}
		summary.ShardsChecked += fileSummary.ShardsChecked
		summary.MarkedHealthy += fileSummary.MarkedHealthy
		summary.MarkedMissing += fileSummary.MarkedMissing
		summary.SkippedErrors += fileSummary.SkippedErrors
		summary.ErrorMessages = append(summary.ErrorMessages, fileSummary.ErrorMessages...)
	}

	if len(summary.ErrorMessages) == 0 {
		summary.ErrorMessages = nil
	}
	return summary, nil
}

func (s *Service) RefreshFileHealth(ctx context.Context, fileID string) (*types.HealthRefreshSummary, error) {
	shardMap, err := s.shardMap.GetShardMap(ctx, fileID)
	if err != nil {
		return nil, fmt.Errorf("failed to load shard map: %w", err)
	}

	summary := &types.HealthRefreshSummary{
		FilesScanned:  1,
		ErrorMessages: []string{},
		RefreshedAt:   time.Now().UTC().Format(time.RFC3339),
	}

	for _, shard := range shardMap.Shards {
		summary.ShardsChecked++
		_, err := s.adapter.DownloadShard(ctx, shard.RemoteID, shard.Provider)
		if err == nil {
			if markErr := s.shardMap.MarkShardStatus(ctx, shard.ShardID, "HEALTHY"); markErr != nil {
				summary.SkippedErrors++
				summary.ErrorMessages = append(summary.ErrorMessages, fmt.Sprintf("shard %s: failed marking healthy: %v", shard.ShardID, markErr))
				continue
			}
			summary.MarkedHealthy++
			continue
		}

		if isAdapterNotFoundError(err) {
			if markErr := s.shardMap.MarkShardStatus(ctx, shard.ShardID, "MISSING"); markErr != nil {
				summary.SkippedErrors++
				summary.ErrorMessages = append(summary.ErrorMessages, fmt.Sprintf("shard %s: failed marking missing: %v", shard.ShardID, markErr))
				continue
			}
			summary.MarkedMissing++
			continue
		}

		summary.SkippedErrors++
		summary.ErrorMessages = append(summary.ErrorMessages, fmt.Sprintf("shard %s: probe error skipped: %v", shard.ShardID, err))
	}

	if len(summary.ErrorMessages) == 0 {
		summary.ErrorMessages = nil
	}
	if refreshedAt, err := time.Parse(time.RFC3339, summary.RefreshedAt); err == nil {
		if touchErr := s.shardMap.UpdateFileHealthRefresh(ctx, fileID, refreshedAt); touchErr != nil {
			summary.SkippedErrors++
			summary.ErrorMessages = append(summary.ErrorMessages, fmt.Sprintf("file %s: failed persisting health refresh time: %v", fileID, touchErr))
		}
	}
	return summary, nil
}

func (s *Service) DeleteFile(ctx context.Context, fileID string, deleteShards bool) error {
	startedAt := time.Now()

	var fileName string
	var fileSize int64
	var shardCount int
	var providers []string
	if shardMap, err := s.shardMap.GetShardMap(ctx, fileID); err == nil {
		fileName = shardMap.OriginalName
		shardCount = len(shardMap.Shards)
		seen := map[string]struct{}{}
		for _, shard := range shardMap.Shards {
			provider := strings.TrimSpace(shard.Provider)
			if provider == "" {
				continue
			}
			if _, ok := seen[provider]; ok {
				continue
			}
			seen[provider] = struct{}{}
			providers = append(providers, provider)
		}
	}

	err := s.adapter.DeleteFile(ctx, fileID, deleteShards)

	endedAt := time.Now()
	durationMs := endedAt.Sub(startedAt).Milliseconds()
	status := "success"
	errorMsg := ""
	if err != nil {
		status = "failed"
		errorMsg = err.Error()
	}

	s.logEvent(&types.LifecycleEvent{
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
		ErrorMsg:   errorMsg,
	})

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
	for index := range isDataShard {
		isDataShard[index] = index < k
	}

	return s.UploadFile(ctx, fileName, shards, isDataShard)
}

func (s *Service) UploadFile(ctx context.Context, fileName string, shards [][]byte, isDataShard []bool) (*types.UploadResp, error) {
	startedAt := time.Now()

	resp, err := s.uploadFileInternal(ctx, fileName, shards, isDataShard)

	endedAt := time.Now()
	durationMs := endedAt.Sub(startedAt).Milliseconds()
	fileID := ""
	if resp != nil {
		fileID = resp.FileID
	}

	if fileID != "" {
		status := "success"
		errorMsg := ""
		if err != nil {
			status = "failed"
			errorMsg = err.Error()
		} else if resp != nil && resp.Status == "failed" {
			status = "failed"
			errorMsg = resp.Error
		}

		s.logEvent(&types.LifecycleEvent{
			FileID:     fileID,
			EventType:  "upload",
			FileName:   fileName,
			FileSize:   totalShardSize(shards),
			ShardCount: len(shards),
			Providers:  uniqueProviders(resp),
			StartedAt:  startedAt,
			EndedAt:    endedAt,
			DurationMs: durationMs,
			Status:     status,
			ErrorMsg:   errorMsg,
		})
	}

	return resp, err
}

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
	for _, provider := range providers {
		if provider.Status == "connected" {
			healthyProviders = append(healthyProviders, provider.ProviderID)
		}
	}
	if len(healthyProviders) < 1 {
		return nil, fmt.Errorf("insufficient healthy providers: need 1, have %d", len(healthyProviders))
	}

	allocation := make(map[int]string)
	for index := 0; index < TotalShards; index++ {
		allocation[index] = healthyProviders[index%len(healthyProviders)]
	}

	uploadResults := s.uploadShardsParallel(ctx, fileID, shards, allocation)
	successCount := 0
	for _, result := range uploadResults {
		if result.Success {
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

	recordReq := &types.RecordShardReq{FileID: fileID, Shards: []types.ShardInfo{}}
	for index, result := range uploadResults {
		shardType := "PARITY"
		if isDataShard[index] {
			shardType = "DATA"
		}
		recordReq.Shards = append(recordReq.Shards, types.ShardInfo{
			ChunkIndex:  index,
			ShardIndex:  index,
			Type:        shardType,
			RemoteID:    result.RemoteID,
			Provider:    result.Provider,
			ChecksumSha: result.ChecksumSha,
		})
	}

	recordResp, err := s.shardMap.RecordShards(ctx, recordReq)
	if err != nil {
		s.rollbackUpload(ctx, uploadResults)
		return &types.UploadResp{FileID: fileID, Status: "failed", Error: fmt.Sprintf("failed to record shards: %v", err)}, nil
	}

	for _, shard := range recordResp.Shards {
		if shard.ShardID == "" {
			return &types.UploadResp{FileID: fileID, Status: "failed", Error: "failed to mark shard health: missing shard_id in record response"}, nil
		}
		if err := s.shardMap.MarkShardStatus(ctx, shard.ShardID, "HEALTHY"); err != nil {
			return &types.UploadResp{FileID: fileID, Status: "failed", Error: fmt.Sprintf("failed to mark shard %s healthy: %v", shard.ShardID, err)}, nil
		}
	}

	shardInfos := []types.ShardInfo{}
	for index, result := range uploadResults {
		shardType := "parity"
		if isDataShard[index] {
			shardType = "data"
		}
		shardInfos = append(shardInfos, types.ShardInfo{
			ChunkIndex:  index,
			ShardIndex:  index,
			Type:        shardType,
			RemoteID:    result.RemoteID,
			Provider:    result.Provider,
			ChecksumSha: result.ChecksumSha,
		})
	}

	return &types.UploadResp{FileID: fileID, Status: "committed", Shards: shardInfos}, nil
}

func (s *Service) uploadShardsParallel(ctx context.Context, fileID string, shards [][]byte, allocation map[int]string) []UploadResult {
	results := make([]UploadResult, len(shards))
	var wg sync.WaitGroup
	var mu sync.Mutex

	for index, shardData := range shards {
		wg.Add(1)
		go func(idx int, data []byte, provider string) {
			defer wg.Done()
			shardID := fmt.Sprintf("%s-shard-%d", fileID, idx)
			uploadResp, err := s.adapter.UploadShard(ctx, shardID, provider, data)

			mu.Lock()
			defer mu.Unlock()
			results[idx] = UploadResult{ShardIndex: idx, Provider: provider, Success: err == nil, Error: err}
			if err == nil {
				results[idx].RemoteID = uploadResp.RemoteID
				results[idx].ChecksumSha = uploadResp.ChecksumSha
			}
		}(index, shardData, allocation[index])
	}

	wg.Wait()
	return results
}

func (s *Service) rollbackUpload(ctx context.Context, results []UploadResult) {
	for _, result := range results {
		if result.Success {
			if err := s.adapter.DeleteShard(ctx, result.RemoteID, result.Provider); err != nil {
				log.Printf("rollback: failed to delete shard %s: %v", result.RemoteID, err)
			}
		}
	}
}

func (s *Service) DownloadFile(ctx context.Context, fileID string) (*types.DownloadResp, error) {
	startedAt := time.Now()
	resp, err := s.downloadFileInternal(ctx, fileID)
	endedAt := time.Now()
	durationMs := endedAt.Sub(startedAt).Milliseconds()

	status := "success"
	errorMsg := ""
	fileName := ""
	if err != nil {
		status = "failed"
		errorMsg = err.Error()
	}
	if resp != nil {
		fileName = resp.FileName
	}

	s.logEvent(&types.LifecycleEvent{
		FileID:     fileID,
		EventType:  "download",
		FileName:   fileName,
		StartedAt:  startedAt,
		EndedAt:    endedAt,
		DurationMs: durationMs,
		Status:     status,
		ErrorMsg:   errorMsg,
	})

	return resp, err
}

func (s *Service) downloadFileInternal(ctx context.Context, fileID string) (*types.DownloadResp, error) {
	shardMap, err := s.shardMap.GetShardMap(ctx, fileID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve shard map: %w", err)
	}

	healthyShards := countHealthyShards(shardMap.Shards)
	if healthyShards < shardMap.K {
		return nil, &RecoverabilityError{FileID: fileID, Available: healthyShards, Required: shardMap.K, Cause: "local shard health state shows too few healthy shards"}
	}
	if len(shardMap.Shards) < shardMap.K {
		return nil, &RecoverabilityError{FileID: fileID, Available: len(shardMap.Shards), Required: shardMap.K, Cause: "shard map contains fewer shard records than required"}
	}

	shards := s.downloadShardsParallelEarlyExit(ctx, shardMap.Shards, shardMap.K)
	if len(shards) < shardMap.K {
		return nil, &RecoverabilityError{FileID: fileID, Available: len(shards), Required: shardMap.K, Cause: "not enough shards could be downloaded from currently reachable providers"}
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
		for index := 0; index < shardMap.K && index < len(shards); index++ {
			reconstructedData = append(reconstructedData, shards[index].Data...)
		}
	}

	return &types.DownloadResp{FileID: fileID, FileName: shardMap.OriginalName, Shards: [][]byte{reconstructedData}}, nil
}

func (s *Service) downloadShardsParallelEarlyExit(ctx context.Context, shardEntries []types.ShardMapEntry, k int) []DownloadResult {
	results := make([]DownloadResult, 0, len(shardEntries))
	resultsChan := make(chan DownloadResult, len(shardEntries))
	received := 0
	var mu sync.Mutex

	eg, egCtx := errgroup.WithContext(ctx)
	for _, entry := range shardEntries {
		entry := entry
		eg.Go(func() error {
			data, err := s.adapter.DownloadShard(egCtx, entry.RemoteID, entry.Provider)
			result := DownloadResult{RemoteID: entry.RemoteID, Provider: entry.Provider, Index: entry.ShardIndex, Success: err == nil, Error: err, Arrived: time.Now()}
			if err == nil {
				result.Data = data
			}
			resultsChan <- result
			return nil
		})
	}

	go func() {
		_ = eg.Wait()
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

func totalShardSize(shards [][]byte) int64 {
	var total int64
	for _, shard := range shards {
		total += int64(len(shard))
	}
	return total
}

func uniqueProviders(resp *types.UploadResp) []string {
	if resp == nil {
		return nil
	}
	seen := map[string]struct{}{}
	var out []string
	for _, shard := range resp.Shards {
		provider := strings.TrimSpace(shard.Provider)
		if provider == "" {
			continue
		}
		if _, ok := seen[provider]; ok {
			continue
		}
		seen[provider] = struct{}{}
		out = append(out, provider)
	}
	return out
}

func isAdapterNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "adapter returned 404") || strings.Contains(msg, "not found") || strings.Contains(msg, "no such key") || strings.Contains(msg, "nosuchkey")
}

func countHealthyShards(entries []types.ShardMapEntry) int {
	healthy := 0
	for _, entry := range entries {
		if strings.EqualFold(entry.Status, "HEALTHY") {
			healthy++
		}
	}
	return healthy
}
