package orchestrator

import "time"

type UploadResult struct {
	ShardIndex  int
	Provider    string
	RemoteID    string
	ChecksumSha string
	Error       error
	Success     bool
}

type DownloadResult struct {
	RemoteID string
	Provider string
	Index    int
	Data     []byte
	Error    error
	Success  bool
	Arrived  time.Time
}
