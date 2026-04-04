//go:build e2e

package e2e_test

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/vindyang/cs464-project/backend/services/shared/types"
)

const (
	orchestratorTestPort = "19082"
	e2eComposePath       = "../docker-compose.test.yml"
	e2eProfile           = "e2e"
	adapterMockPort      = "19080"
)

// testStack holds the in-process adapter mock and the Docker stack teardown.
type testStack struct {
	orchestratorURL string
	shutdown        func()
}

// startTestStack binds an in-process adapter mock and brings up the Docker
// Compose e2e profile (shardmap, sharding, orchestrator).
func startTestStack(t *testing.T) *testStack {
	t.Helper()

	// Shared shard storage for the mock adapter.
	var mu sync.Mutex
	shards := map[string][]byte{}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/providers", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]types.ProviderInfo{
			{ProviderID: "mock-a", DisplayName: "Mock A", Status: "connected", LatencyMs: 5},
			{ProviderID: "mock-b", DisplayName: "Mock B", Status: "connected", LatencyMs: 8},
		})
	})
	mux.HandleFunc("/shards/upload", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseMultipartForm(32 << 20)
		shardID := r.FormValue("shard_id")
		remoteID := "remote-" + shardID
		if file, _, err := r.FormFile("file_data"); err == nil {
			data, _ := readAllBytes(file)
			mu.Lock()
			shards[remoteID] = data
			mu.Unlock()
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(types.UploadShardResp{
			RemoteID:    remoteID,
			ChecksumSha: "mock-" + shardID,
		})
	})
	mux.HandleFunc("/shards/", func(w http.ResponseWriter, r *http.Request) {
		remoteID := strings.TrimPrefix(r.URL.Path, "/shards/")
		switch r.Method {
		case http.MethodGet:
			mu.Lock()
			data, ok := shards[remoteID]
			mu.Unlock()
			if !ok {
				http.NotFound(w, r)
				return
			}
			_, _ = w.Write(data)
		case http.MethodDelete:
			mu.Lock()
			delete(shards, remoteID)
			mu.Unlock()
			w.WriteHeader(http.StatusNoContent)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	ln, err := net.Listen("tcp", ":"+adapterMockPort)
	if err != nil {
		t.Fatalf("failed to bind adapter mock on :%s — %v", adapterMockPort, err)
	}
	srv := &httptest.Server{Listener: ln, Config: &http.Server{Handler: mux}}
	srv.Start()
	t.Logf("Adapter mock: http://localhost:%s", adapterMockPort)

	// Compose up.
	up := exec.Command("docker", "compose",
		"-f", e2eComposePath,
		"--profile", e2eProfile,
		"up", "-d", "--build", "--wait",
	)
	up.Stdout = os.Stdout
	up.Stderr = os.Stderr
	if err := up.Run(); err != nil {
		srv.Close()
		t.Fatalf("docker compose up failed: %v", err)
	}

	orchURL := "http://localhost:" + orchestratorTestPort
	if err := waitForHTTP(orchURL+"/health", 90*time.Second); err != nil {
		shutdownCompose(t)
		srv.Close()
		t.Fatalf("orchestrator did not become healthy: %v", err)
	}
	t.Logf("Orchestrator: %s", orchURL)

	return &testStack{
		orchestratorURL: orchURL,
		shutdown: func() {
			srv.Close()
			shutdownCompose(t)
		},
	}
}

func shutdownCompose(t *testing.T) {
	t.Helper()
	cmd := exec.Command("docker", "compose",
		"-f", e2eComposePath,
		"--profile", e2eProfile,
		"down", "-v",
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	_ = cmd.Run()
}

func waitForHTTP(url string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		resp, err := http.Get(url)
		if err == nil && resp.StatusCode < 500 {
			_ = resp.Body.Close()
			return nil
		}
		time.Sleep(500 * time.Millisecond)
	}
	return fmt.Errorf("timed out waiting for %s after %s", url, timeout)
}

func readAllBytes(r interface{ Read([]byte) (int, error) }) ([]byte, error) {
	var buf []byte
	tmp := make([]byte, 4096)
	for {
		n, readErr := r.Read(tmp)
		if n > 0 {
			buf = append(buf, tmp[:n]...)
		}
		if readErr != nil {
			break
		}
	}
	return buf, nil
}
