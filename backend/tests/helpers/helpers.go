package helpers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/vindyang/cs464-project/backend/services/shared/types"
)

func backendRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatalf("failed to determine caller file path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "../.."))
}

// StartOrchestrator launches the real orchestrator process wired to mock service URLs.
// It returns the base URL and a shutdown function that must be deferred by callers.
func StartOrchestrator(t *testing.T, adapterURL, shardMapURL, shardingURL string) (string, func()) {
	t.Helper()

	port := freePort(t)
	orchestratorURL := "http://127.0.0.1:" + strconv.Itoa(port)

	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, "go", "run", "./services/orchestrator/cmd/main.go")
	cmd.Dir = backendRoot(t)
	cmd.Env = append(
		os.Environ(),
		"ADAPTER_URL="+adapterURL,
		"SHARDMAP_URL="+shardMapURL,
		"SHARDING_URL="+shardingURL,
		"PORT="+strconv.Itoa(port),
	)

	if err := cmd.Start(); err != nil {
		cancel()
		t.Fatalf("failed to start orchestrator: %v", err)
	}

	if err := waitForHTTP(orchestratorURL+"/api/orchestrator/upload", 15*time.Second); err != nil {
		cancel()
		_ = cmd.Wait()
		t.Fatalf("orchestrator did not start: %v", err)
	}

	shutdown := func() {
		cancel()
		_ = cmd.Wait()
	}
	return orchestratorURL, shutdown
}

// UploadFile performs a successful multipart upload request and decodes the JSON response.
func UploadFile(t *testing.T, baseURL string, payload []byte) types.UploadResp {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	_ = writer.WriteField("k", "4")
	_ = writer.WriteField("n", "6")
	part, _ := writer.CreateFormFile("file", "contract.txt")
	_, _ = part.Write(payload)
	_ = writer.Close()

	req, _ := http.NewRequest(http.MethodPost, baseURL+"/api/orchestrator/upload", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	httpResp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("upload request failed: %v", err)
	}
	defer httpResp.Body.Close()
	if httpResp.StatusCode != http.StatusCreated {
		b, _ := io.ReadAll(httpResp.Body)
		t.Fatalf("upload failed: status=%d body=%s", httpResp.StatusCode, string(b))
	}
	var out types.UploadResp
	_ = json.NewDecoder(httpResp.Body).Decode(&out)
	return out
}

// UploadFileRaw sends the same multipart payload as UploadFile but intentionally
// does not assert success status so failure-path tests can validate error responses.
func UploadFileRaw(t *testing.T, baseURL string, payload []byte) (*http.Response, []byte) {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	_ = writer.WriteField("k", "4")
	_ = writer.WriteField("n", "6")
	part, _ := writer.CreateFormFile("file", "contract.txt")
	_, _ = part.Write(payload)
	_ = writer.Close()

	req, _ := http.NewRequest(http.MethodPost, baseURL+"/api/orchestrator/upload", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	httpResp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("upload request failed: %v", err)
	}
	respBody, _ := io.ReadAll(httpResp.Body)
	return httpResp, respBody
}

// DownloadFile performs a download request for the given fileID and returns the raw bytes.
func DownloadFile(t *testing.T, baseURL string, fileID string) ([]byte, int) {
	t.Helper()
	url := fmt.Sprintf("%s/api/orchestrator/files/%s/download", baseURL, fileID)
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("download request failed: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return body, resp.StatusCode
}

// GetFileHistory queries the orchestrator's lifecycle history endpoint for a file.
func GetFileHistory(t *testing.T, baseURL string, fileID string) *types.FileHistoryResp {
	t.Helper()
	url := fmt.Sprintf("%s/api/orchestrator/files/%s/history", baseURL, fileID)
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("history request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("history endpoint returned %d: %s", resp.StatusCode, body)
	}
	var out types.FileHistoryResp
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatalf("failed to decode history response: %v", err)
	}
	return &out
}

// waitForHTTP polls an endpoint until it responds or the timeout expires.
func waitForHTTP(url string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		resp, err := http.Get(url)
		if err == nil {
			_ = resp.Body.Close()
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("timeout waiting for %s", url)
}

// freePort allocates an ephemeral localhost TCP port for test process startup.
func freePort(t *testing.T) int {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("freePort: %v", err)
	}
	defer ln.Close()
	return ln.Addr().(*net.TCPAddr).Port
}

type AdapterMockConfig struct {
	OnGetProviders  func(http.ResponseWriter, *http.Request)
	OnUploadShard   func(http.ResponseWriter, *http.Request)
	OnDownloadShard func(http.ResponseWriter, *http.Request)
	OnDeleteShard   func(http.ResponseWriter, *http.Request)
}

func NewAdapterMock(t *testing.T, cfg AdapterMockConfig) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/providers":
			if cfg.OnGetProviders != nil {
				cfg.OnGetProviders(w, r)
				return
			}
		case r.Method == http.MethodPost && r.URL.Path == "/shards/upload":
			if cfg.OnUploadShard != nil {
				cfg.OnUploadShard(w, r)
				return
			}
		case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/shards/"):
			if cfg.OnDownloadShard != nil {
				cfg.OnDownloadShard(w, r)
				return
			}
		case r.Method == http.MethodDelete && strings.HasPrefix(r.URL.Path, "/shards/"):
			if cfg.OnDeleteShard != nil {
				cfg.OnDeleteShard(w, r)
				return
			}
		}
		http.NotFound(w, r)
	}))
}

type ShardMapMockConfig struct {
	OnRegisterFile  func(http.ResponseWriter, *http.Request)
	OnRecordShards  func(http.ResponseWriter, *http.Request)
	OnGetShardMap   func(http.ResponseWriter, *http.Request)
	OnLogLifecycle  func(http.ResponseWriter, *http.Request) // POST /api/v1/lifecycle
	OnGetLifecycle  func(http.ResponseWriter, *http.Request) // GET  /api/v1/lifecycle/{fileId}
}

func NewShardMapMock(t *testing.T, cfg ShardMapMockConfig) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/shards/register":
			if cfg.OnRegisterFile != nil {
				cfg.OnRegisterFile(w, r)
				return
			}
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/shards/record":
			if cfg.OnRecordShards != nil {
				cfg.OnRecordShards(w, r)
				return
			}
		case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/api/v1/shards/file/"):
			if cfg.OnGetShardMap != nil {
				cfg.OnGetShardMap(w, r)
				return
			}
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/lifecycle":
			if cfg.OnLogLifecycle != nil {
				cfg.OnLogLifecycle(w, r)
				return
			}
			// Default: accept silently so tests that don't care still pass.
			w.WriteHeader(http.StatusCreated)
			return
		case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/api/v1/lifecycle/"):
			if cfg.OnGetLifecycle != nil {
				cfg.OnGetLifecycle(w, r)
				return
			}
		}
		http.NotFound(w, r)
	}))
}

type ShardingMockConfig struct {
	OnShard       func(http.ResponseWriter, *http.Request)
	OnReconstruct func(http.ResponseWriter, *http.Request)
}

func NewShardingMock(t *testing.T, cfg ShardingMockConfig) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/sharding/shard":
			if cfg.OnShard != nil {
				cfg.OnShard(w, r)
				return
			}
		case r.Method == http.MethodPost && r.URL.Path == "/api/sharding/reconstruct":
			if cfg.OnReconstruct != nil {
				cfg.OnReconstruct(w, r)
				return
			}
		}
		http.NotFound(w, r)
	}))
}
