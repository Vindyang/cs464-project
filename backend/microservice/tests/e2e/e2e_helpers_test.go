//go:build e2e

package e2e_test

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

const (
	orchestratorTestPort = "19082"
	gatewayTestPort      = "19084"
	e2eProfile           = "e2e"
)

type testStack struct {
	orchestratorURL string
	gatewayURL      string
	shutdown        func()
}

// startTestStack binds an in-process adapter mock and brings up the Docker
// Compose e2e profile (shardmap, sharding, orchestrator).
func startTestStack(t *testing.T) *testStack {
	t.Helper()

	up := exec.Command("docker", "compose",
		"-f", e2eComposePath(t),
		"--profile", e2eProfile,
		"up", "-d", "--build", "--wait",
	)
	up.Stdout = os.Stdout
	up.Stderr = os.Stderr
	if err := up.Run(); err != nil {
		t.Fatalf("docker compose up failed: %v", err)
	}

	orchURL := "http://localhost:" + orchestratorTestPort
	if err := waitForHTTP(orchURL+"/health", 90*time.Second); err != nil {
		shutdownCompose(t)
		t.Fatalf("orchestrator did not become healthy: %v", err)
	}
	gatewayURL := "http://localhost:" + gatewayTestPort
	if err := waitForHTTP(gatewayURL+"/", 90*time.Second); err != nil {
		shutdownCompose(t)
		t.Fatalf("gateway did not become healthy: %v", err)
	}
	t.Logf("Orchestrator: %s", orchURL)
	t.Logf("Gateway: %s", gatewayURL)

	return &testStack{
		orchestratorURL: orchURL,
		gatewayURL:      gatewayURL,
		shutdown: func() {
			shutdownCompose(t)
		},
	}
}

func restartService(t *testing.T, service string) {
	t.Helper()
	cmd := exec.Command("docker", "compose", "-f", e2eComposePath(t), "--profile", e2eProfile, "restart", service)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("docker compose restart %s failed: %v", service, err)
	}
}

func waitForServiceHealthy(t *testing.T, url string, timeout time.Duration) {
	t.Helper()
	if err := waitForHTTP(url, timeout); err != nil {
		t.Fatalf("service did not become healthy: %v", err)
	}
}

func shutdownCompose(t *testing.T) {
	t.Helper()
	cmd := exec.Command("docker", "compose",
		"-f", e2eComposePath(t),
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

func e2eComposePath(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to resolve e2e helper path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", "docker-compose.test.yml"))
}
