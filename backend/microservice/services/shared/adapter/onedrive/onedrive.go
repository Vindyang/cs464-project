package onedrive

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"golang.org/x/oauth2"

	"github.com/vindyang/cs464-project/backend/services/shared/adapter"
	"github.com/vindyang/cs464-project/backend/services/shared/db"
)

const (
	graphBase                = "https://graph.microsoft.com/v1.0"
	OmnishardFolderConfigKey = "onedrive_omnishard_folder_id"
)

// OneDriveAdapter implements StorageProvider using the Microsoft Graph API.
// Authenticates via OAuth2 user credentials (Azure AD flow).
type OneDriveAdapter struct {
	client       *http.Client
	store        *db.Store
	mu           sync.Mutex
	folderID     string // in-memory cache of the Omnishard root item ID
	fileFolderMu sync.Mutex
	fileFolders  map[string]string // fileID → OneDrive item ID
}

// NewOneDriveAdapter constructs an OneDriveAdapter from an OAuth2 config and token.
func NewOneDriveAdapter(config *oauth2.Config, token *oauth2.Token, store *db.Store) (*OneDriveAdapter, error) {
	client := config.Client(context.Background(), token)
	return &OneDriveAdapter{
		client:      client,
		store:       store,
		fileFolders: make(map[string]string),
	}, nil
}

// driveResponse is the top-level response from GET /me/drive.
type driveResponse struct {
	ID        string `json:"id"`
	DriveType string `json:"driveType"`
	Quota     struct {
		Total int64 `json:"total"`
		Used  int64 `json:"used"`
	} `json:"quota"`
	Owner struct {
		User struct {
			DisplayName string `json:"displayName"`
			ID          string `json:"id"`
		} `json:"user"`
	} `json:"owner"`
}

// driveItem is a minimal Graph driveItem (file or folder).
type driveItem struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type driveItemDownloadInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DownloadURL string `json:"@microsoft.graph.downloadUrl"`
}

// driveItemList is the response from a children list or search.
type driveItemList struct {
	Value []driveItem `json:"value"`
}

// GetMetadata fetches real quota and latency from the Graph API.
func (o *OneDriveAdapter) GetMetadata(ctx context.Context) (*adapter.ProviderMetadata, error) {
	start := time.Now()

	var dr driveResponse
	if err := o.graphGet(ctx, "/me/drive", &dr); err != nil {
		log.Printf("onedrive metadata probe failed: err=%v", err)
		return nil, fmt.Errorf("onedrive: get drive: %w", err)
	}

	latency := time.Since(start).Milliseconds()
	log.Printf(
		"onedrive metadata probe ok: drive_id=%q drive_type=%q owner=%q owner_id=%q quota_used=%d quota_total=%d latency_ms=%d",
		dr.ID,
		dr.DriveType,
		dr.Owner.User.DisplayName,
		dr.Owner.User.ID,
		dr.Quota.Used,
		dr.Quota.Total,
		latency,
	)

	return &adapter.ProviderMetadata{
		ProviderID:   "oneDrive",
		DisplayName:  "OneDrive",
		Status:       "connected",
		LatencyMs:    latency,
		Region:       "global",
		Capabilities: map[string]any{"supportsVersioning": false},
		QuotaTotal:   dr.Quota.Total,
		QuotaUsed:    dr.Quota.Used,
		LastCheck:    time.Now().UTC().Format(time.RFC3339),
	}, nil
}

// ensureOmnishardFolder returns the item ID of the "Omnishard" root folder,
// creating it if it doesn't exist. Result is cached in memory and SQLite.
func (o *OneDriveAdapter) ensureOmnishardFolder(ctx context.Context) (string, error) {
	o.mu.Lock()
	defer o.mu.Unlock()

	if o.folderID != "" {
		return o.folderID, nil
	}

	if id, err := o.store.GetConfig(OmnishardFolderConfigKey); err == nil {
		o.folderID = id
		return id, nil
	}

	// Try to get the existing folder via path shorthand.
	var item driveItem
	err := o.graphGet(ctx, "/me/drive/root:/Omnishard:", &item)
	if err == nil {
		o.folderID = item.ID
		_ = o.store.SetConfig(OmnishardFolderConfigKey, item.ID)
		return item.ID, nil
	}

	// Create the folder under root.
	id, err := o.createFolder(ctx, "root", "Omnishard")
	if err != nil {
		return "", fmt.Errorf("onedrive: create Omnishard folder: %w", err)
	}

	if err := o.store.SetConfig(OmnishardFolderConfigKey, id); err != nil {
		return "", fmt.Errorf("onedrive: persist Omnishard folder id: %w", err)
	}
	o.folderID = id
	return id, nil
}

// ensureFileFolder returns the item ID for a per-file subfolder inside Omnishard.
// Creates it if it doesn't exist; serialized to prevent duplicate creation.
func (o *OneDriveAdapter) ensureFileFolder(ctx context.Context, fileID string) (string, error) {
	o.fileFolderMu.Lock()
	defer o.fileFolderMu.Unlock()

	if id, ok := o.fileFolders[fileID]; ok {
		return id, nil
	}

	rootID, err := o.ensureOmnishardFolder(ctx)
	if err != nil {
		return "", err
	}

	// Try to get existing subfolder.
	var item driveItem
	err = o.graphGet(ctx, fmt.Sprintf("/me/drive/items/%s:/%s:", rootID, fileID), &item)
	if err == nil {
		o.fileFolders[fileID] = item.ID
		return item.ID, nil
	}

	id, err := o.createFolder(ctx, rootID, fileID)
	if err != nil {
		return "", fmt.Errorf("onedrive: create file folder %s: %w", fileID, err)
	}
	o.fileFolders[fileID] = id
	return id, nil
}

// UploadShard uploads shard data into the per-file subfolder inside Omnishard.
// Returns the Graph item ID as the remoteID.
func (o *OneDriveAdapter) UploadShard(ctx context.Context, fileID string, index int, data io.Reader) (string, error) {
	folderID, err := o.ensureFileFolder(ctx, fileID)
	if err != nil {
		return "", err
	}

	name := fmt.Sprintf("shard_%03d", index)
	url := fmt.Sprintf("%s/me/drive/items/%s:/%s:/content", graphBase, folderID, name)

	body, err := io.ReadAll(data)
	if err != nil {
		return "", fmt.Errorf("onedrive: read shard data: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("onedrive: build upload request: %w", err)
	}
	req.Header.Set("Content-Type", "application/octet-stream")

	resp, err := o.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("onedrive: upload shard %s[%d]: %w", fileID, index, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("onedrive: upload shard %s[%d]: status %d: %s", fileID, index, resp.StatusCode, b)
	}

	var item driveItem
	if err := json.NewDecoder(resp.Body).Decode(&item); err != nil {
		return "", fmt.Errorf("onedrive: decode upload response: %w", err)
	}
	return item.ID, nil
}

// DownloadShard downloads a shard by its Graph item ID.
// The returned ReadCloser must be closed by the caller.
func (o *OneDriveAdapter) DownloadShard(ctx context.Context, remoteID string) (io.ReadCloser, error) {
	start := time.Now()
	remoteDriveID := driveIDFromRemoteID(remoteID)
	escapedRemoteID := url.PathEscape(remoteID)
	contentURL := fmt.Sprintf("%s/me/drive/items/%s/content", graphBase, escapedRemoteID)
	log.Printf("onedrive download start: remote_id=%q remote_drive_prefix=%q content_url=%q", remoteID, remoteDriveID, contentURL)

	redirectClient := &http.Client{
		Transport: o.client.Transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, contentURL, nil)
	if err != nil {
		log.Printf("onedrive download request build failed: remote_id=%q err=%v", remoteID, err)
		return nil, fmt.Errorf("onedrive: build download request: %w", err)
	}

	resp, err := redirectClient.Do(req)
	if err != nil {
		log.Printf("onedrive download transport failed: remote_id=%q remote_drive_prefix=%q err=%v", remoteID, remoteDriveID, err)
		return nil, fmt.Errorf("onedrive: download shard %q: %w", remoteID, err)
	}

	if resp.StatusCode == http.StatusNotFound {
		log.Printf("onedrive download not found: remote_id=%q remote_drive_prefix=%q duration_ms=%d", remoteID, remoteDriveID, time.Since(start).Milliseconds())
		resp.Body.Close()
		return nil, fmt.Errorf("%w: onedrive shard %q", adapter.ErrShardNotFound, remoteID)
	}
	if resp.StatusCode >= 300 && resp.StatusCode < 400 {
		redirectURL := resp.Header.Get("Location")
		resp.Body.Close()
		if redirectURL == "" {
			log.Printf("onedrive download redirect missing location: remote_id=%q remote_drive_prefix=%q status=%d duration_ms=%d", remoteID, remoteDriveID, resp.StatusCode, time.Since(start).Milliseconds())
			return nil, fmt.Errorf("onedrive: download shard %q: redirect missing location", remoteID)
		}

		log.Printf("onedrive download redirect resolved: remote_id=%q remote_drive_prefix=%q status=%d redirect_target=%q", remoteID, remoteDriveID, resp.StatusCode, sanitizeURLForLog(redirectURL))

		downloadReq, err := http.NewRequestWithContext(ctx, http.MethodGet, redirectURL, nil)
		if err != nil {
			log.Printf("onedrive redirect download request build failed: remote_id=%q err=%v", remoteID, err)
			return nil, fmt.Errorf("onedrive: build redirect download request: %w", err)
		}

		resp, err = http.DefaultClient.Do(downloadReq)
		if err != nil {
			log.Printf("onedrive redirect download transport failed: remote_id=%q remote_drive_prefix=%q err=%v", remoteID, remoteDriveID, err)
			return nil, fmt.Errorf("onedrive: download shard %q via redirect: %w", remoteID, err)
		}
	}
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		log.Printf(
			"onedrive download failed: remote_id=%q remote_drive_prefix=%q status=%d duration_ms=%d body=%s",
			remoteID,
			remoteDriveID,
			resp.StatusCode,
			time.Since(start).Milliseconds(),
			strings.TrimSpace(string(b)),
		)
		return nil, fmt.Errorf("onedrive: download shard %q: status %d: %s", remoteID, resp.StatusCode, b)
	}

	log.Printf("onedrive download ok: remote_id=%q remote_drive_prefix=%q duration_ms=%d", remoteID, remoteDriveID, time.Since(start).Milliseconds())

	return resp.Body, nil
}

// DeleteShard permanently deletes a shard by its Graph item ID.
// Treats 404 as success for idempotent rollback behavior.
// After deletion, removes the parent folder if it's empty.
func (o *OneDriveAdapter) DeleteShard(ctx context.Context, remoteID string) error {
	// Fetch parent ID before deleting.
	var item struct {
		ParentReference struct {
			ID string `json:"id"`
		} `json:"parentReference"`
	}
	parentID := ""
	if err := o.graphGet(ctx, fmt.Sprintf("/me/drive/items/%s", remoteID), &item); err == nil {
		parentID = item.ParentReference.ID
	}

	url := fmt.Sprintf("%s/me/drive/items/%s", graphBase, remoteID)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("onedrive: build delete request: %w", err)
	}

	resp, err := o.client.Do(req)
	if err != nil {
		return fmt.Errorf("onedrive: delete shard %q: %w", remoteID, err)
	}
	resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil // already deleted; treat as success
	}
	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("onedrive: delete shard %q: status %d", remoteID, resp.StatusCode)
	}

	// Clean up empty parent folder.
	if parentID != "" {
		o.fileFolderMu.Lock()
		defer o.fileFolderMu.Unlock()

		var children driveItemList
		if err := o.graphGet(ctx, fmt.Sprintf("/me/drive/items/%s/children", parentID), &children); err == nil && len(children.Value) == 0 {
			delURL := fmt.Sprintf("%s/me/drive/items/%s", graphBase, parentID)
			delReq, _ := http.NewRequestWithContext(ctx, http.MethodDelete, delURL, nil)
			delResp, _ := o.client.Do(delReq)
			if delResp != nil {
				delResp.Body.Close()
			}
			// Evict from cache.
			for k, v := range o.fileFolders {
				if v == parentID {
					delete(o.fileFolders, k)
					break
				}
			}
		}
	}

	return nil
}

// HealthCheck performs a lightweight liveness ping against the Graph API.
func (o *OneDriveAdapter) HealthCheck(ctx context.Context) error {
	var dr driveResponse
	if err := o.graphGet(ctx, "/me/drive", &dr); err != nil {
		log.Printf("onedrive health check failed: err=%v", err)
		return fmt.Errorf("onedrive: health check: %w", err)
	}
	log.Printf(
		"onedrive health check ok: drive_id=%q drive_type=%q owner=%q owner_id=%q quota_used=%d quota_total=%d",
		dr.ID,
		dr.DriveType,
		dr.Owner.User.DisplayName,
		dr.Owner.User.ID,
		dr.Quota.Used,
		dr.Quota.Total,
	)
	return nil
}

// graphGet performs a GET request to the Graph API and decodes the JSON response into dest.
func (o *OneDriveAdapter) graphGet(ctx context.Context, path string, dest any) error {
	url := graphBase + path
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	resp, err := o.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("not found")
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("status %d: %s", resp.StatusCode, b)
	}

	return json.NewDecoder(resp.Body).Decode(dest)
}

// createFolder creates a new folder as a child of parentID (or "root") and returns its item ID.
func (o *OneDriveAdapter) createFolder(ctx context.Context, parentID, name string) (string, error) {
	var url string
	if parentID == "root" {
		url = fmt.Sprintf("%s/me/drive/root/children", graphBase)
	} else {
		url = fmt.Sprintf("%s/me/drive/items/%s/children", graphBase, parentID)
	}

	body, _ := json.Marshal(map[string]any{
		"name":                              name,
		"folder":                            map[string]any{},
		"@microsoft.graph.conflictBehavior": "rename",
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := o.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("status %d: %s", resp.StatusCode, b)
	}

	var item driveItem
	if err := json.NewDecoder(resp.Body).Decode(&item); err != nil {
		return "", err
	}
	return item.ID, nil
}

func driveIDFromRemoteID(remoteID string) string {
	if remoteID == "" {
		return ""
	}
	if idx := strings.Index(remoteID, "!"); idx > 0 {
		return remoteID[:idx]
	}
	return ""
}

func sanitizeURLForLog(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "invalid-url"
	}
	parsed.RawQuery = ""
	parsed.Fragment = ""
	return parsed.String()
}
