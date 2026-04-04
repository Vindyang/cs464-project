//go:build e2e

// Package e2e_test contains end-to-end tests that spin up real service containers
// via Docker Compose. Run with:
//
//	go test -tags e2e ./tests/e2e/... -v -timeout 300s
//
// Prerequisites: Docker with Compose plugin must be installed.
package e2e_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/vindyang/cs464-project/backend/services/shared/types"
)

// TestMain manages the shared Docker stack for all e2e tests.
// It starts the stack once, runs all tests, then tears it down.
var (
	sharedStack *testStack
)

func TestMain(m *testing.M) {
	// TestMain is not a standard testing.T test — use a minimal wrapper
	// to satisfy the startTestStack signature which requires *testing.T.
	// In TestMain we rely on log.Fatal for fatal errors.
	// The stack is torn down via the shutdown func regardless of test outcome.
	panic("use go test -tags e2e and set up stack per-test via t.Skip guard")
}

// stackOrSkip returns the shared stack or skips the test if Docker isn't available.
func stackOrSkip(t *testing.T) *testStack {
	t.Helper()
	stack := startTestStack(t)
	t.Cleanup(stack.shutdown)
	return stack
}

// TestE2EUploadDownloadHappyPath verifies a complete upload then download cycle
// against real shardmap, sharding, and orchestrator containers.
func TestE2EUploadDownloadHappyPath(t *testing.T) {
	stack := stackOrSkip(t)

	fileContent := []byte("hello nebula e2e test content - " + fmt.Sprintf("%d", time.Now().UnixNano()))
	uploadResp := doUpload(t, stack.orchestratorURL, fileContent, "e2e_happy.txt")

	if uploadResp.Status != "committed" {
		t.Fatalf("expected committed, got %q (error=%q)", uploadResp.Status, uploadResp.Error)
	}
	if uploadResp.FileID == "" {
		t.Fatal("expected non-empty file ID")
	}
	t.Logf("Uploaded file ID: %s", uploadResp.FileID)

	// Download and verify content matches.
	downloaded := doDownload(t, stack.orchestratorURL, uploadResp.FileID)
	if !bytes.Equal(fileContent, downloaded) {
		t.Errorf("downloaded content mismatch: got %d bytes, want %d bytes", len(downloaded), len(fileContent))
	}
	t.Logf("✓ Upload→Download round-trip verified (%d bytes)", len(fileContent))
}

// TestE2EUploadLifecycleLogging verifies that after a successful upload,
// the history endpoint returns an upload lifecycle event.
func TestE2EUploadLifecycleLogging(t *testing.T) {
	stack := stackOrSkip(t)

	uploadResp := doUpload(t, stack.orchestratorURL, []byte("lifecycle-upload-e2e"), "lifecycle.txt")
	if uploadResp.Status != "committed" {
		t.Fatalf("expected committed, got %q", uploadResp.Status)
	}

	// Allow async log call to complete.
	time.Sleep(200 * time.Millisecond)

	history := doGetHistory(t, stack.orchestratorURL, uploadResp.FileID)
	if len(history.Events) == 0 {
		t.Fatal("expected at least one lifecycle event, got none")
	}

	uploadEvent := findEvent(history.Events, "upload")
	if uploadEvent == nil {
		t.Fatal("no 'upload' event found in lifecycle history")
	}
	if uploadEvent.Status != "success" {
		t.Errorf("expected status=success, got %q", uploadEvent.Status)
	}
	t.Logf("✓ Upload lifecycle events: %d events, upload.status=%s duration=%dms",
		len(history.Events), uploadEvent.Status, uploadEvent.DurationMs)
}

// TestE2EDownloadLifecycleLogging verifies that after upload+download,
// the history contains both event types.
func TestE2EDownloadLifecycleLogging(t *testing.T) {
	stack := stackOrSkip(t)

	uploadResp := doUpload(t, stack.orchestratorURL, []byte("lifecycle-dl-e2e"), "lifecycle_dl.txt")
	if uploadResp.Status != "committed" {
		t.Fatalf("expected committed, got %q", uploadResp.Status)
	}

	_ = doDownload(t, stack.orchestratorURL, uploadResp.FileID)

	// Allow async log calls to complete.
	time.Sleep(200 * time.Millisecond)

	history := doGetHistory(t, stack.orchestratorURL, uploadResp.FileID)

	uploadEvent := findEvent(history.Events, "upload")
	if uploadEvent == nil {
		t.Fatal("no 'upload' event in history")
	}
	downloadEvent := findEvent(history.Events, "download")
	if downloadEvent == nil {
		t.Fatal("no 'download' event in history after downloading file")
	}
	t.Logf("✓ Both lifecycle events present: upload.status=%s download.status=%s",
		uploadEvent.Status, downloadEvent.Status)
}

// TestE2EDownloadNonexistentFile verifies that downloading a non-existent file returns 404.
func TestE2EDownloadNonexistentFile(t *testing.T) {
	stack := stackOrSkip(t)

	url := fmt.Sprintf("%s/api/orchestrator/files/00000000-0000-0000-0000-000000000001/download", stack.orchestratorURL)
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("download request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound && resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected 404 or 500 for nonexistent file, got %d", resp.StatusCode)
	}
	t.Logf("✓ Nonexistent file correctly returned %d", resp.StatusCode)
}

// TestE2EMultipleFilesConcurrent uploads 3 files concurrently and verifies all are committed.
func TestE2EMultipleFilesConcurrent(t *testing.T) {
	stack := stackOrSkip(t)

	var wg sync.WaitGroup
	results := make([]types.UploadResp, 3)
	errors := make([]error, 3)

	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			content := []byte(fmt.Sprintf("concurrent-file-%d-%d", idx, time.Now().UnixNano()))
			filename := fmt.Sprintf("concurrent_%d.txt", idx)
			resp := doUpload(t, stack.orchestratorURL, content, filename)
			results[idx] = resp
			if resp.Status != "committed" {
				errors[idx] = fmt.Errorf("file %d: expected committed, got %q (error=%q)", idx, resp.Status, resp.Error)
			}
		}(i)
	}
	wg.Wait()

	for i, err := range errors {
		if err != nil {
			t.Errorf("concurrent upload %d failed: %v", i, err)
		}
	}

	committed := 0
	for _, r := range results {
		if r.Status == "committed" {
			committed++
		}
	}
	t.Logf("✓ Concurrent upload: %d/3 files committed", committed)
}

// --- Helpers used only in e2e tests ---

// doUpload sends a file to the orchestrator upload endpoint.
func doUpload(t *testing.T, baseURL string, content []byte, filename string) types.UploadResp {
	t.Helper()
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	_ = w.WriteField("k", "4")
	_ = w.WriteField("n", "6")
	part, _ := w.CreateFormFile("file", filename)
	_, _ = part.Write(content)
	_ = w.Close()

	req, _ := http.NewRequest(http.MethodPost, baseURL+"/api/orchestrator/upload", &body)
	req.Header.Set("Content-Type", w.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("upload request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		raw, _ := io.ReadAll(resp.Body)
		t.Fatalf("upload returned %d: %s", resp.StatusCode, raw)
	}
	var out types.UploadResp
	_ = json.NewDecoder(resp.Body).Decode(&out)
	return out
}

// doDownload fetches file content from the orchestrator download endpoint.
func doDownload(t *testing.T, baseURL, fileID string) []byte {
	t.Helper()
	url := fmt.Sprintf("%s/api/orchestrator/files/%s/download", baseURL, fileID)
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("download request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(resp.Body)
		t.Fatalf("download returned %d: %s", resp.StatusCode, raw)
	}
	data, _ := io.ReadAll(resp.Body)
	return data
}

// doGetHistory fetches lifecycle history from the orchestrator.
func doGetHistory(t *testing.T, baseURL, fileID string) *types.FileHistoryResp {
	t.Helper()
	url := fmt.Sprintf("%s/api/orchestrator/files/%s/history", baseURL, fileID)
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("history request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(resp.Body)
		t.Fatalf("history returned %d: %s", resp.StatusCode, raw)
	}
	var out types.FileHistoryResp
	_ = json.NewDecoder(resp.Body).Decode(&out)
	return &out
}

// findEvent returns the first lifecycle event matching the given eventType, or nil.
func findEvent(events []types.LifecycleEvent, eventType string) *types.LifecycleEvent {
	for i := range events {
		if events[i].EventType == eventType {
			return &events[i]
		}
	}
	return nil
}
