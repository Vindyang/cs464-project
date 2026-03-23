package app

import "time"

// UploadResult tracks outcome of each shard upload.
type UploadResult struct {
	ShardIndex  int
	Provider    string
	RemoteID    string
	ChecksumSha string
	Error       error
	Success     bool
}

// DownloadResult tracks outcome of each shard download.
type DownloadResult struct {
	RemoteID string
	Provider string
	Index    int
	Data     []byte
	Error    error
	Success  bool
	Arrived  time.Time
}
