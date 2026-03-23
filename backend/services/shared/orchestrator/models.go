package orchestrator

import "time"

// UploadJob represents a single shard upload task
type UploadJob struct {
	ShardIndex int
	Provider   string
	Data       []byte
	RemoteID   string
}

// UploadResult tracks outcome of each shard upload
type UploadResult struct {
	ShardIndex  int
	Provider    string
	RemoteID    string
	ChecksumSha string
	Error       error
	Success     bool
}

// DownloadJob represents a single shard fetch task
type DownloadJob struct {
	RemoteID string
	Provider string
	Index    int
}

// DownloadResult tracks outcome of each shard download
type DownloadResult struct {
	RemoteID string
	Provider string
	Index    int
	Data     []byte
	Error    error
	Success  bool
	Arrived  time.Time
}
