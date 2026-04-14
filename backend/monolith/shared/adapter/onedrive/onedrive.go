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

	"github.com/vindyang/cs464-project/backend/monolith/shared/adapter"
	"github.com/vindyang/cs464-project/backend/monolith/shared/db"
)

const (
	graphBase                = "https://graph.microsoft.com/v1.0"
	OmnishardFolderConfigKey = "onedrive_omnishard_folder_id"
)

type OneDriveAdapter struct {
	client       *http.Client
	store        *db.Store
	mu           sync.Mutex
	folderID     string
	fileFolderMu sync.Mutex
	fileFolders  map[string]string
}

func NewOneDriveAdapter(config *oauth2.Config, token *oauth2.Token, store *db.Store) (*OneDriveAdapter, error) {
	client := config.Client(context.Background(), token)
	return &OneDriveAdapter{client: client, store: store, fileFolders: make(map[string]string)}, nil
}

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

type driveItem struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type driveItemList struct {
	Value []driveItem `json:"value"`
}

func (o *OneDriveAdapter) GetMetadata(ctx context.Context) (*adapter.ProviderMetadata, error) {
	start := time.Now()

	var dr driveResponse
	if err := o.graphGet(ctx, "/me/drive", &dr); err != nil {
		log.Printf("onedrive metadata probe failed: err=%v", err)
		return nil, fmt.Errorf("onedrive: get drive: %w", err)
	}

	latency := time.Since(start).Milliseconds()
	log.Printf("onedrive metadata probe ok: drive_id=%q drive_type=%q owner=%q owner_id=%q quota_used=%d quota_total=%d latency_ms=%d", dr.ID, dr.DriveType, dr.Owner.User.DisplayName, dr.Owner.User.ID, dr.Quota.Used, dr.Quota.Total, latency)

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

	var item driveItem
	err := o.graphGet(ctx, "/me/drive/root:/Omnishard:", &item)
	if err == nil {
		o.folderID = item.ID
		_ = o.store.SetConfig(OmnishardFolderConfigKey, item.ID)
		return item.ID, nil
	}

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

func (o *OneDriveAdapter) UploadShard(ctx context.Context, fileID string, index int, data io.Reader) (string, error) {
	folderID, err := o.ensureFileFolder(ctx, fileID)
	if err != nil {
		return "", err
	}

	name := fmt.Sprintf("shard_%03d", index)
	requestURL := fmt.Sprintf("%s/me/drive/items/%s:/%s:/content", graphBase, folderID, name)

	body, err := io.ReadAll(data)
	if err != nil {
		return "", fmt.Errorf("onedrive: read shard data: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, requestURL, bytes.NewReader(body))
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

func (o *OneDriveAdapter) DownloadShard(ctx context.Context, remoteID string) (io.ReadCloser, error) {
	start := time.Now()
	remoteDriveID := driveIDFromRemoteID(remoteID)
	escapedRemoteID := url.PathEscape(remoteID)
	contentURL := fmt.Sprintf("%s/me/drive/items/%s/content", graphBase, escapedRemoteID)
	log.Printf("onedrive download start: remote_id=%q remote_drive_prefix=%q content_url=%q", remoteID, remoteDriveID, contentURL)

	redirectClient := &http.Client{Transport: o.client.Transport, CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}}

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
		log.Printf("onedrive download failed: remote_id=%q remote_drive_prefix=%q status=%d duration_ms=%d body=%s", remoteID, remoteDriveID, resp.StatusCode, time.Since(start).Milliseconds(), strings.TrimSpace(string(b)))
		return nil, fmt.Errorf("onedrive: download shard %q: status %d: %s", remoteID, resp.StatusCode, b)
	}

	log.Printf("onedrive download ok: remote_id=%q remote_drive_prefix=%q duration_ms=%d", remoteID, remoteDriveID, time.Since(start).Milliseconds())

	return resp.Body, nil
}

func (o *OneDriveAdapter) DeleteShard(ctx context.Context, remoteID string) error {
	var item struct {
		ParentReference struct {
			ID string `json:"id"`
		} `json:"parentReference"`
	}
	parentID := ""
	if err := o.graphGet(ctx, fmt.Sprintf("/me/drive/items/%s", remoteID), &item); err == nil {
		parentID = item.ParentReference.ID
	}

	requestURL := fmt.Sprintf("%s/me/drive/items/%s", graphBase, remoteID)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, requestURL, nil)
	if err != nil {
		return fmt.Errorf("onedrive: build delete request: %w", err)
	}

	resp, err := o.client.Do(req)
	if err != nil {
		return fmt.Errorf("onedrive: delete shard %q: %w", remoteID, err)
	}
	resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil
	}
	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("onedrive: delete shard %q: status %d", remoteID, resp.StatusCode)
	}

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

func (o *OneDriveAdapter) HealthCheck(ctx context.Context) error {
	var dr driveResponse
	if err := o.graphGet(ctx, "/me/drive", &dr); err != nil {
		log.Printf("onedrive health check failed: err=%v", err)
		return fmt.Errorf("onedrive: health check: %w", err)
	}
	log.Printf("onedrive health check ok: drive_id=%q drive_type=%q owner=%q owner_id=%q quota_used=%d quota_total=%d", dr.ID, dr.DriveType, dr.Owner.User.DisplayName, dr.Owner.User.ID, dr.Quota.Used, dr.Quota.Total)
	return nil
}

func (o *OneDriveAdapter) graphGet(ctx context.Context, path string, dest any) error {
	requestURL := graphBase + path
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
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

func (o *OneDriveAdapter) createFolder(ctx context.Context, parentID, name string) (string, error) {
	var requestURL string
	if parentID == "root" {
		requestURL = fmt.Sprintf("%s/me/drive/root/children", graphBase)
	} else {
		requestURL = fmt.Sprintf("%s/me/drive/items/%s/children", graphBase, parentID)
	}

	body, _ := json.Marshal(map[string]any{"name": name, "folder": map[string]any{}, "@microsoft.graph.conflictBehavior": "rename"})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, requestURL, bytes.NewReader(body))
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
