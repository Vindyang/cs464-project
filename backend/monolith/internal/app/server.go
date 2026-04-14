package app

import (
	"bytes"
	"context"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/vindyang/cs464-project/backend/monolith/internal/orchestrator"
	"github.com/vindyang/cs464-project/backend/monolith/internal/sharding"
	"github.com/vindyang/cs464-project/backend/monolith/internal/shardmap"
	sharedadapter "github.com/vindyang/cs464-project/backend/monolith/shared/adapter"
	"github.com/vindyang/cs464-project/backend/monolith/shared/api/dto"
	"github.com/vindyang/cs464-project/backend/monolith/shared/db"
	"github.com/vindyang/cs464-project/backend/monolith/shared/oauthhandler"
	"github.com/vindyang/cs464-project/backend/monolith/shared/onedrivehandler"
	"github.com/vindyang/cs464-project/backend/monolith/shared/s3handler"
	"github.com/vindyang/cs464-project/backend/monolith/shared/transport/httpx"
	"github.com/vindyang/cs464-project/backend/monolith/shared/types"
)

//go:embed docs/openapi.yml
var openAPISpec []byte

const (
	defaultStorePath    = "Omnishard.db"
	defaultShardMapPath = "Omnishard-shardmap.db"
	defaultMaxUploadMB  = 30
	uploadBodyOverhead  = 2 << 20
)

type Config struct {
	StorePath    string
	ShardMapPath string
}

type App struct {
	store               *db.Store
	shardMapDB          *sqlx.DB
	registry            *sharedadapter.Registry
	shardMapService     shardmap.ShardMapService
	lifecycleService    shardmap.LifecycleService
	shardingService     sharding.Service
	orchestratorService *orchestrator.Service
	handler             http.Handler
}

type FileDeleteSummary struct {
	DeletedFiles       int  `json:"deleted_files"`
	DeletedShards      int  `json:"deleted_shards"`
	FailedShardDeletes int  `json:"failed_shard_deletes"`
	DeleteShards       bool `json:"delete_shards"`
}

type CredentialResetSummary struct {
	DeletedCredentials    int `json:"deleted_credentials"`
	DeletedTokens         int `json:"deleted_tokens"`
	DisconnectedProviders int `json:"disconnected_providers"`
}

func ConfigFromEnv() Config {
	storePath := strings.TrimSpace(os.Getenv("Omnishard_DB_PATH"))
	if storePath == "" {
		storePath = defaultStorePath
	}

	shardMapPath := strings.TrimSpace(os.Getenv("Omnishard_SHARDMAP_DB_PATH"))
	if shardMapPath == "" {
		shardMapPath = defaultShardMapPath
	}

	return Config{
		StorePath:    storePath,
		ShardMapPath: shardMapPath,
	}
}

func New(config Config) (*App, error) {
	if strings.TrimSpace(config.StorePath) == "" {
		config.StorePath = defaultStorePath
	}
	if strings.TrimSpace(config.ShardMapPath) == "" {
		config.ShardMapPath = defaultShardMapPath
	}

	store, err := db.NewStore(config.StorePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open local store: %w", err)
	}

	shardMapDB, err := shardmap.ConnectSQLite(config.ShardMapPath)
	if err != nil {
		_ = store.Close()
		return nil, fmt.Errorf("failed to connect shardmap db: %w", err)
	}

	registry := sharedadapter.NewRegistry()
	if err := tryRestoreGDriveAdapter(store, registry); err != nil {
		log.Printf("Google Drive adapter not restored: %v", err)
	}
	s3Handler := s3handler.New(store, registry)
	if err := s3Handler.RestoreAdapter(); err != nil {
		log.Printf("S3 adapter not restored: %v", err)
	}
	oneDriveHandler := onedrivehandler.New(store, registry)
	if err := tryRestoreOneDriveAdapter(store, registry, oneDriveHandler); err != nil {
		log.Printf("OneDrive adapter not restored: %v", err)
	}

	fileRepo := shardmap.NewFileRepository(shardMapDB)
	shardRepo := shardmap.NewShardRepository(shardMapDB)
	lifecycleRepo := shardmap.NewLifecycleRepository(shardMapDB)
	shardMapService := shardmap.NewShardMapService(fileRepo, shardRepo, lifecycleRepo)
	lifecycleService := shardmap.NewLifecycleService(lifecycleRepo)
	shardingService := sharding.NewService()

	app := &App{
		store:            store,
		shardMapDB:       shardMapDB,
		registry:         registry,
		shardMapService:  shardMapService,
		lifecycleService: lifecycleService,
		shardingService:  shardingService,
	}

	adapterClient := &localAdapterClient{app: app}
	shardMapClient := &localShardMapClient{app: app}
	app.orchestratorService = orchestrator.NewServiceWithSharding(adapterClient, shardMapClient, shardingService)
	app.handler = corsMiddleware(loggingMiddleware(app.routes()))

	return app, nil
}

func (a *App) Close() error {
	var errs []error
	if a.store != nil {
		if err := a.store.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if a.shardMapDB != nil {
		if err := a.shardMapDB.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func (a *App) Handler() http.Handler {
	return a.handler
}

func (a *App) routes() http.Handler {
	mux := http.NewServeMux()
	oauth := oauthhandler.New(a.store, a.registry)
	oneDrive := onedrivehandler.New(a.store, a.registry)
	s3 := s3handler.New(a.store, a.registry)

	mux.HandleFunc("GET /health", a.health)
	mux.HandleFunc("GET /api/v1/health", a.health)
	mux.HandleFunc("GET /api/v1/docs", a.docs)
	mux.HandleFunc("GET /api/v1/docs/openapi.yml", a.docs)

	mux.HandleFunc("GET /api/providers", a.listProviders)
	mux.HandleFunc("GET /api/v1/providers", a.listProviders)
	mux.HandleFunc("GET /api/oauth/gdrive/authorize", oauth.Authorize)
	mux.HandleFunc("GET /api/oauth/gdrive/callback", oauth.Callback)
	mux.HandleFunc("POST /api/oauth/gdrive/disconnect", oauth.Disconnect)
	mux.HandleFunc("GET /api/oauth/onedrive/authorize", oneDrive.Authorize)
	mux.HandleFunc("GET /api/oauth/onedrive/callback", oneDrive.Callback)
	mux.HandleFunc("POST /api/oauth/onedrive/disconnect", oneDrive.Disconnect)
	mux.HandleFunc("POST /api/providers/awsS3/connect", s3.Connect)
	mux.HandleFunc("POST /api/providers/awsS3/disconnect", s3.Disconnect)

	mux.HandleFunc("GET /api/credentials", a.listCredentials)
	mux.HandleFunc("GET /api/credentials/status", a.credentialsStatus)
	mux.HandleFunc("GET /api/credentials/{provider}", a.getCredential)
	mux.HandleFunc("PUT /api/credentials/{provider}", a.upsertCredential)
	mux.HandleFunc("DELETE /api/credentials/{provider}", a.deleteCredential)
	mux.HandleFunc("GET /api/credentials/{provider}/secret", a.revealCredential)

	mux.HandleFunc("GET /api/settings", a.getSettings)
	mux.HandleFunc("PUT /api/settings", a.putSettings)
	mux.HandleFunc("POST /api/settings/reset", a.resetSettings)

	mux.HandleFunc("GET /api/v1/files", a.listFiles)
	mux.HandleFunc("GET /api/v1/files/{fileId}", a.getFileMetadata)
	mux.HandleFunc("DELETE /api/v1/files/{fileId}", a.deleteFile)
	mux.HandleFunc("GET /api/v1/shards/file/{fileId}", a.getShardMap)

	mux.HandleFunc("POST /api/v1/upload", a.uploadWorkflow)
	mux.HandleFunc("POST /api/orchestrator/upload", a.uploadWorkflow)
	mux.HandleFunc("GET /api/v1/download/{fileId}", a.downloadWorkflow)
	mux.HandleFunc("GET /api/orchestrator/files/{fileId}/download", a.downloadWorkflow)
	mux.HandleFunc("GET /api/v1/history", a.globalHistory)
	mux.HandleFunc("GET /api/orchestrator/history", a.globalHistory)
	mux.HandleFunc("GET /api/v1/history/{fileId}", a.fileHistory)
	mux.HandleFunc("GET /api/orchestrator/files/{fileId}/history", a.fileHistory)
	mux.HandleFunc("POST /api/v1/files/health/refresh", a.refreshAllFileHealth)
	mux.HandleFunc("POST /api/orchestrator/files/health/refresh", a.refreshAllFileHealth)
	mux.HandleFunc("POST /api/v1/files/{fileId}/health/refresh", a.refreshOneFileHealth)
	mux.HandleFunc("POST /api/orchestrator/files/{fileId}/health/refresh", a.refreshOneFileHealth)
	mux.HandleFunc("DELETE /api/orchestrator/files/{fileId}", a.deleteFile)

	return mux
}

func (a *App) health(w http.ResponseWriter, r *http.Request) {
	providers, err := a.listProviderMetadata(r.Context())
	if err != nil {
		httpx.WriteErrorWithCode(w, http.StatusInternalServerError, "failed to enumerate providers", "UNKNOWN_ERROR", err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"status":    "healthy",
		"service":   "monolith",
		"providers": providers,
	})
}

func (a *App) docs(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/yaml")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(openAPISpec)
}

func (a *App) listProviders(w http.ResponseWriter, r *http.Request) {
	providers, err := a.listProviderMetadata(r.Context())
	if err != nil {
		httpx.WriteErrorWithCode(w, http.StatusInternalServerError, "failed to enumerate providers", "UNKNOWN_ERROR", err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, providers)
}

func (a *App) listProviderMetadata(ctx context.Context) ([]*sharedadapter.ProviderMetadata, error) {
	metadata := make([]*sharedadapter.ProviderMetadata, 0)
	for _, id := range a.registry.IDs() {
		provider, err := a.registry.Get(id)
		if err != nil {
			continue
		}
		meta, err := provider.GetMetadata(ctx)
		if err != nil {
			log.Printf("GetMetadata(%s): %v", id, err)
			continue
		}
		metadata = append(metadata, meta)
	}
	return metadata, nil
}

func (a *App) listCredentials(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}
	records, err := a.store.ListCredentials()
	if err != nil {
		httpx.WriteErrorWithCode(w, http.StatusInternalServerError, "failed to list credentials", "UNKNOWN_ERROR", err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, records)
}

func (a *App) credentialsStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}
	records, err := a.store.ListCredentials()
	if err != nil {
		httpx.WriteErrorWithCode(w, http.StatusInternalServerError, "failed to list credentials status", "UNKNOWN_ERROR", err)
		return
	}
	providers := make([]string, 0, len(records))
	for _, record := range records {
		providers = append(providers, record.ProviderID)
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"configured": len(records) > 0,
		"count":      len(records),
		"providers":  providers,
	})
}

func (a *App) getCredential(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("provider")
	if id == "" {
		httpx.WriteError(w, http.StatusBadRequest, "provider is required", nil)
		return
	}
	clientID, _, redirectURI, err := a.store.LoadCredentials(id)
	if errors.Is(err, db.ErrNotFound) {
		httpx.WriteErrorWithCode(w, http.StatusNotFound, "no credentials configured for provider", "PROVIDER_NOT_FOUND", nil)
		return
	}
	if err != nil {
		httpx.WriteErrorWithCode(w, http.StatusInternalServerError, "failed to load credentials", "UNKNOWN_ERROR", err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]string{
		"provider_id":  id,
		"client_id":    clientID,
		"redirect_uri": redirectURI,
	})
}

func (a *App) revealCredential(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("provider")
	if id == "" {
		httpx.WriteError(w, http.StatusBadRequest, "provider is required", nil)
		return
	}
	clientID, clientSecret, redirectURI, err := a.store.LoadCredentials(id)
	if errors.Is(err, db.ErrNotFound) {
		httpx.WriteErrorWithCode(w, http.StatusNotFound, "no credentials configured for provider", "PROVIDER_NOT_FOUND", nil)
		return
	}
	if err != nil {
		httpx.WriteErrorWithCode(w, http.StatusInternalServerError, "failed to load credentials", "UNKNOWN_ERROR", err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]string{
		"provider_id":   id,
		"client_id":     clientID,
		"client_secret": clientSecret,
		"redirect_uri":  redirectURI,
	})
}

func (a *App) upsertCredential(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("provider")
	if id == "" {
		httpx.WriteError(w, http.StatusBadRequest, "provider is required", nil)
		return
	}
	var body struct {
		ClientID     string `json:"client_id"`
		ClientSecret string `json:"client_secret"`
		RedirectURI  string `json:"redirect_uri"`
	}
	if err := httpx.DecodeJSON(r, &body, 1<<20); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", err)
		return
	}
	body.ClientID = strings.TrimSpace(body.ClientID)
	body.ClientSecret = strings.TrimSpace(body.ClientSecret)
	body.RedirectURI = strings.TrimSpace(body.RedirectURI)
	if body.ClientID == "" || body.ClientSecret == "" || body.RedirectURI == "" {
		httpx.WriteError(w, http.StatusBadRequest, "client_id, client_secret, and redirect_uri are required", nil)
		return
	}
	if id == "awsS3" {
		if err := s3handler.ValidateRegion(body.RedirectURI); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, "invalid AWS region", err)
			return
		}
	}
	if err := a.store.UpsertCredentials(id, body.ClientID, body.ClientSecret, body.RedirectURI); err != nil {
		httpx.WriteErrorWithCode(w, http.StatusInternalServerError, "failed to save credentials", "UNKNOWN_ERROR", err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"provider_id": id, "status": "credentials saved"})
}

func (a *App) deleteCredential(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("provider")
	if id == "" {
		httpx.WriteError(w, http.StatusBadRequest, "provider is required", nil)
		return
	}
	if err := a.deleteProviderData(id); err != nil {
		httpx.WriteErrorWithCode(w, http.StatusInternalServerError, "failed to delete credentials", "UNKNOWN_ERROR", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *App) DeleteAllCredentials() (CredentialResetSummary, error) {
	records, err := a.store.ListCredentials()
	if err != nil {
		return CredentialResetSummary{}, err
	}
	deletedCredentials, err := a.store.DeleteAllCredentials()
	if err != nil {
		return CredentialResetSummary{}, err
	}
	deletedTokens, err := a.store.DeleteAllTokens()
	if err != nil {
		return CredentialResetSummary{}, err
	}
	disconnected := 0
	for _, record := range records {
		a.registry.Unregister(record.ProviderID)
		disconnected++
	}
	return CredentialResetSummary{DeletedCredentials: deletedCredentials, DeletedTokens: deletedTokens, DisconnectedProviders: disconnected}, nil
}

func (a *App) deleteProviderData(id string) error {
	if err := a.store.DeleteCredentials(id); err != nil {
		return err
	}
	if err := a.store.DeleteToken(id); err != nil {
		return err
	}
	a.registry.Unregister(id)
	return nil
}

func (a *App) getSettings(w http.ResponseWriter, _ *http.Request) {
	redundancy, err := a.store.GetConfig("settings_redundancy")
	if errors.Is(err, db.ErrNotFound) {
		redundancy = "(6,4)"
	} else if err != nil {
		httpx.WriteErrorWithCode(w, http.StatusInternalServerError, "failed to load settings", "UNKNOWN_ERROR", err)
		return
	}
	encryptDefault, err := a.getBoolConfig("settings_encrypt_default", true)
	if err != nil {
		httpx.WriteErrorWithCode(w, http.StatusInternalServerError, "failed to load settings", "UNKNOWN_ERROR", err)
		return
	}
	autoDelete, err := a.getBoolConfig("settings_auto_delete", false)
	if err != nil {
		httpx.WriteErrorWithCode(w, http.StatusInternalServerError, "failed to load settings", "UNKNOWN_ERROR", err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"redundancy":      redundancy,
		"encrypt_default": encryptDefault,
		"auto_delete":     autoDelete,
	})
}

func (a *App) putSettings(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Redundancy     string `json:"redundancy"`
		EncryptDefault *bool  `json:"encrypt_default"`
		AutoDelete     *bool  `json:"auto_delete"`
	}
	if err := httpx.DecodeJSON(r, &body, 1<<20); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", err)
		return
	}
	if body.Redundancy != "(6,4)" && body.Redundancy != "(8,4)" && body.Redundancy != "(10,8)" {
		httpx.WriteError(w, http.StatusBadRequest, "invalid redundancy value", nil)
		return
	}
	if body.EncryptDefault == nil || body.AutoDelete == nil {
		httpx.WriteError(w, http.StatusBadRequest, "encrypt_default and auto_delete are required", nil)
		return
	}
	if err := a.store.SetConfig("settings_redundancy", body.Redundancy); err != nil {
		httpx.WriteErrorWithCode(w, http.StatusInternalServerError, "failed to save settings", "UNKNOWN_ERROR", err)
		return
	}
	if err := a.store.SetConfig("settings_encrypt_default", strconv.FormatBool(*body.EncryptDefault)); err != nil {
		httpx.WriteErrorWithCode(w, http.StatusInternalServerError, "failed to save settings", "UNKNOWN_ERROR", err)
		return
	}
	if err := a.store.SetConfig("settings_auto_delete", strconv.FormatBool(*body.AutoDelete)); err != nil {
		httpx.WriteErrorWithCode(w, http.StatusInternalServerError, "failed to save settings", "UNKNOWN_ERROR", err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "settings saved"})
}

func (a *App) resetSettings(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Scope        string `json:"scope"`
		DeleteShards *bool  `json:"delete_shards,omitempty"`
	}
	if err := httpx.DecodeJSON(r, &body, 1<<20); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body", err)
		return
	}
	deleteShards := true
	if body.DeleteShards != nil {
		deleteShards = *body.DeleteShards
	}
	response := map[string]any{"scope": body.Scope, "delete_shards": deleteShards}

	switch body.Scope {
	case "files":
		fileSummary, err := a.DeleteAllFiles(r.Context(), deleteShards)
		if err != nil {
			httpx.WriteErrorWithCode(w, http.StatusInternalServerError, "failed to delete file data", "UNKNOWN_ERROR", err)
			return
		}
		response["file_summary"] = fileSummary
	case "credentials":
		credentialSummary, err := a.DeleteAllCredentials()
		if err != nil {
			httpx.WriteErrorWithCode(w, http.StatusInternalServerError, "failed to delete credentials", "UNKNOWN_ERROR", err)
			return
		}
		response["credential_summary"] = credentialSummary
	case "all_data":
		fileSummary, err := a.DeleteAllFiles(r.Context(), deleteShards)
		if err != nil {
			httpx.WriteErrorWithCode(w, http.StatusInternalServerError, "failed to delete file data", "UNKNOWN_ERROR", err)
			return
		}
		credentialSummary, err := a.DeleteAllCredentials()
		if err != nil {
			httpx.WriteErrorWithCode(w, http.StatusInternalServerError, "failed to delete credentials", "UNKNOWN_ERROR", err)
			return
		}
		deletedEvents, err := a.lifecycleService.DeleteAllHistory()
		if err != nil {
			httpx.WriteErrorWithCode(w, http.StatusInternalServerError, "failed to delete lifecycle history", "UNKNOWN_ERROR", err)
			return
		}
		response["file_summary"] = fileSummary
		response["credential_summary"] = credentialSummary
		response["lifecycle_summary"] = map[string]int{"deleted_events": deletedEvents}
	default:
		httpx.WriteError(w, http.StatusBadRequest, "invalid reset scope", nil)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, response)
}

func (a *App) getBoolConfig(key string, fallback bool) (bool, error) {
	raw, err := a.store.GetConfig(key)
	if errors.Is(err, db.ErrNotFound) {
		return fallback, nil
	}
	if err != nil {
		return false, err
	}
	parsed, err := strconv.ParseBool(raw)
	if err != nil {
		return fallback, nil
	}
	return parsed, nil
}

func (a *App) listFiles(w http.ResponseWriter, _ *http.Request) {
	files, err := a.shardMapService.ListFiles()
	if err != nil {
		httpx.WriteErrorWithCode(w, http.StatusInternalServerError, "Failed to list files", "UNKNOWN_ERROR", err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, files)
}

func (a *App) getFileMetadata(w http.ResponseWriter, r *http.Request) {
	fileID, err := parseUUIDPathValue(r, "fileId")
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "Invalid file ID format", err)
		return
	}
	file, err := a.shardMapService.GetFileMetadata(fileID)
	if err != nil {
		status := http.StatusInternalServerError
		code := "UNKNOWN_ERROR"
		message := "Failed to get file metadata"
		if isNotFoundError(err) {
			status = http.StatusNotFound
			code = "FILE_NOT_FOUND"
			message = "File not found"
		}
		httpx.WriteErrorWithCode(w, status, message, code, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, file)
}

func (a *App) getShardMap(w http.ResponseWriter, r *http.Request) {
	fileID, err := parseUUIDPathValue(r, "fileId")
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "Invalid file ID format", err)
		return
	}
	shardMap, err := a.shardMapService.GetShardMap(fileID)
	if err != nil {
		status := http.StatusInternalServerError
		code := "UNKNOWN_ERROR"
		message := "Failed to get shard map"
		if isNotFoundError(err) {
			status = http.StatusNotFound
			code = "FILE_NOT_FOUND"
			message = "File not found"
		}
		httpx.WriteErrorWithCode(w, status, message, code, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, shardMap)
}

func (a *App) deleteFile(w http.ResponseWriter, r *http.Request) {
	fileID := r.PathValue("fileId")
	summary, err := a.deleteFileData(r.Context(), fileID, r.URL.Query().Get("delete_shards") == "true")
	if err != nil {
		httpx.WriteErrorWithCode(w, http.StatusBadGateway, "Failed to delete file", "UNKNOWN_ERROR", err)
		return
	}
	if strings.HasPrefix(r.URL.Path, "/api/orchestrator/") {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"success":              true,
		"file_id":              fileID,
		"message":              "File deleted successfully",
		"shards_deleted":       summary.DeletedShards,
		"failed_shard_deletes": summary.FailedShardDeletes,
	})
}

func (a *App) DeleteAllFiles(ctx context.Context, deleteShards bool) (FileDeleteSummary, error) {
	files, err := a.shardMapService.ListFiles()
	if err != nil {
		return FileDeleteSummary{}, err
	}
	summary := FileDeleteSummary{DeleteShards: deleteShards}
	for _, file := range files {
		result, err := a.deleteFileData(ctx, file.FileID, deleteShards)
		if err != nil {
			return summary, err
		}
		summary.DeletedFiles += result.DeletedFiles
		summary.DeletedShards += result.DeletedShards
		summary.FailedShardDeletes += result.FailedShardDeletes
	}
	return summary, nil
}

func (a *App) deleteFileData(ctx context.Context, fileID string, deleteShards bool) (FileDeleteSummary, error) {
	summary := FileDeleteSummary{DeleteShards: deleteShards}
	parsedFileID, err := uuid.Parse(fileID)
	if err != nil {
		return summary, err
	}

	if deleteShards {
		shardMap, err := a.shardMapService.GetShardMap(parsedFileID)
		if err != nil {
			return summary, err
		}
		for _, shard := range shardMap.Shards {
			provider, err := a.registry.Get(shard.Provider)
			if err != nil {
				summary.FailedShardDeletes++
				continue
			}
			if err := provider.DeleteShard(ctx, shard.RemoteID); err != nil {
				if !errors.Is(err, sharedadapter.ErrShardNotFound) {
					log.Printf("DeleteShard(%s, %s): %v", shard.Provider, shard.RemoteID, err)
					summary.FailedShardDeletes++
					continue
				}
			}
			summary.DeletedShards++
		}
	}

	if err := a.shardMapService.DeleteFile(parsedFileID); err != nil {
		return summary, err
	}
	summary.DeletedFiles = 1
	return summary, nil
}

func (a *App) uploadWorkflow(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, resolveMaxUploadBytes()+uploadBodyOverhead)
	fileName, fileData, k, n, err := parseUploadMultipart(r)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "request body too large") {
			httpx.WriteJSON(w, http.StatusRequestEntityTooLarge, map[string]any{
				"error":         "Upload limit exceeded",
				"details":       fmt.Sprintf("Maximum upload size is %d MB", resolveMaxUploadBytes()/(1<<20)),
				"max_upload_mb": resolveMaxUploadBytes() / (1 << 20),
			})
			return
		}
		httpx.WriteError(w, http.StatusBadRequest, "Invalid upload payload", err)
		return
	}
	if int64(len(fileData)) > resolveMaxUploadBytes() {
		httpx.WriteJSON(w, http.StatusRequestEntityTooLarge, map[string]any{
			"error":         "Upload limit exceeded",
			"details":       fmt.Sprintf("Maximum upload size is %d MB", resolveMaxUploadBytes()/(1<<20)),
			"max_upload_mb": resolveMaxUploadBytes() / (1 << 20),
		})
		return
	}

	resp, err := a.orchestratorService.UploadRawFile(r.Context(), fileName, fileData, k, n)
	if err != nil {
		httpx.WriteErrorWithCode(w, http.StatusInternalServerError, "Failed to upload file", httpx.ClassifyUploadError(err.Error()), err)
		return
	}
	if resp.Status == "failed" {
		httpx.WriteErrorWithCode(w, http.StatusInternalServerError, resp.Error, httpx.ClassifyUploadError(resp.Error), nil)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, resp)
}

func (a *App) downloadWorkflow(w http.ResponseWriter, r *http.Request) {
	fileID := r.PathValue("fileId")
	resp, err := a.orchestratorService.DownloadFile(r.Context(), fileID)
	if err != nil {
		var recoverabilityErr *orchestrator.RecoverabilityError
		if errors.As(err, &recoverabilityErr) {
			httpx.WriteErrorWithCode(w, http.StatusConflict, "File cannot be reconstructed", "SHARD_NOT_RECOVERABLE", err)
			return
		}
		if strings.Contains(err.Error(), "404") || strings.Contains(strings.ToLower(err.Error()), "not found") {
			httpx.WriteErrorWithCode(w, http.StatusNotFound, "File not found", "FILE_NOT_FOUND", err)
			return
		}
		httpx.WriteErrorWithCode(w, http.StatusInternalServerError, "Failed to download file", "UNKNOWN_ERROR", err)
		return
	}
	if len(resp.Shards) == 0 {
		httpx.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "No file data reconstructed"})
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+resp.FileName+"\"")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(resp.Shards[0])
}

func (a *App) globalHistory(w http.ResponseWriter, r *http.Request) {
	history, err := a.orchestratorService.GetAllHistory(r.Context())
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "Failed to get lifecycle history", err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, history)
}

func (a *App) fileHistory(w http.ResponseWriter, r *http.Request) {
	fileID := r.PathValue("fileId")
	history, err := a.orchestratorService.GetFileHistory(r.Context(), fileID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "Failed to get file history", err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, history)
}

func (a *App) refreshAllFileHealth(w http.ResponseWriter, r *http.Request) {
	summary, err := a.orchestratorService.RefreshAllFileHealth(r.Context())
	if err != nil {
		httpx.WriteErrorWithCode(w, http.StatusInternalServerError, "Failed to refresh file health", "HEALTH_REFRESH_FAILED", err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, summary)
}

func (a *App) refreshOneFileHealth(w http.ResponseWriter, r *http.Request) {
	fileID := r.PathValue("fileId")
	summary, err := a.orchestratorService.RefreshFileHealth(r.Context(), fileID)
	if err != nil {
		httpx.WriteErrorWithCode(w, http.StatusInternalServerError, "Failed to refresh file health", "HEALTH_REFRESH_FAILED", err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, summary)
}

func tryRestoreGDriveAdapter(store *db.Store, registry *sharedadapter.Registry) error {
	tok, err := store.LoadToken("googleDrive")
	if err != nil {
		log.Println("No stored Google Drive token — connect via UI")
		return nil
	}
	handler := oauthhandler.New(store, registry)
	if err := handler.RestoreAdapter(registry, tok); err != nil {
		return err
	}
	log.Println("Google Drive adapter restored from stored token")
	return nil
}

func tryRestoreOneDriveAdapter(store *db.Store, registry *sharedadapter.Registry, handler *onedrivehandler.OneDriveHandler) error {
	tok, err := store.LoadToken("oneDrive")
	if err != nil {
		log.Println("No stored OneDrive token — connect via UI")
		return nil
	}
	if err := handler.RestoreAdapter(registry, tok); err != nil {
		return err
	}
	log.Println("OneDrive adapter restored from stored token")
	return nil
}

func parseUploadMultipart(r *http.Request) (string, []byte, int, int, error) {
	reader, err := r.MultipartReader()
	if err != nil {
		return "", nil, 0, 0, fmt.Errorf("failed to open multipart reader: %w", err)
	}

	var fileName string
	var fileData []byte
	var kRaw string
	var nRaw string

	for {
		part, partErr := reader.NextPart()
		if errors.Is(partErr, io.EOF) {
			break
		}
		if partErr != nil {
			return "", nil, 0, 0, fmt.Errorf("failed to read multipart part: %w", partErr)
		}

		switch part.FormName() {
		case "file":
			fileName = part.FileName()
			fileData, err = io.ReadAll(part)
			if err != nil {
				return "", nil, 0, 0, fmt.Errorf("failed to read uploaded file: %w", err)
			}
		case "k":
			value, readErr := io.ReadAll(part)
			if readErr != nil {
				return "", nil, 0, 0, fmt.Errorf("failed to read k value: %w", readErr)
			}
			kRaw = strings.TrimSpace(string(value))
		case "n":
			value, readErr := io.ReadAll(part)
			if readErr != nil {
				return "", nil, 0, 0, fmt.Errorf("failed to read n value: %w", readErr)
			}
			nRaw = strings.TrimSpace(string(value))
		default:
			_, _ = io.Copy(io.Discard, part)
		}
		_ = part.Close()
	}

	if fileName == "" || len(fileData) == 0 {
		return "", nil, 0, 0, fmt.Errorf("missing file field")
	}
	parsedK, err := strconv.Atoi(kRaw)
	if err != nil {
		return "", nil, 0, 0, fmt.Errorf("invalid k value: %w", err)
	}
	parsedN, err := strconv.Atoi(nRaw)
	if err != nil {
		return "", nil, 0, 0, fmt.Errorf("invalid n value: %w", err)
	}
	return fileName, fileData, parsedK, parsedN, nil
}

func resolveMaxUploadBytes() int64 {
	raw := strings.TrimSpace(os.Getenv("ORCHESTRATOR_MAX_UPLOAD_MB"))
	if raw == "" {
		return defaultMaxUploadMB << 20
	}
	mb, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || mb <= 0 {
		log.Printf("invalid ORCHESTRATOR_MAX_UPLOAD_MB=%q, using default %dMB", raw, defaultMaxUploadMB)
		return defaultMaxUploadMB << 20
	}
	return mb << 20
}

func parseUUIDPathValue(r *http.Request, key string) (uuid.UUID, error) {
	return uuid.Parse(r.PathValue(key))
}

func isNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	lower := strings.ToLower(err.Error())
	return strings.Contains(lower, "not found") || strings.Contains(lower, "no rows")
}

func parseFileID(shardID string) string {
	if idx := strings.LastIndex(shardID, "-shard-"); idx != -1 {
		return shardID[:idx]
	}
	return shardID
}

func parseShardIndex(shardID string) int {
	parts := strings.Split(shardID, "-")
	if len(parts) == 0 {
		return -1
	}
	index, err := strconv.Atoi(parts[len(parts)-1])
	if err != nil {
		return -1
	}
	return index
}

type localAdapterClient struct {
	app *App
}

func (c *localAdapterClient) GetProviders(ctx context.Context) ([]types.ProviderInfo, error) {
	metadata, err := c.app.listProviderMetadata(ctx)
	if err != nil {
		return nil, err
	}
	providers := make([]types.ProviderInfo, 0, len(metadata))
	for _, meta := range metadata {
		providers = append(providers, types.ProviderInfo{
			ProviderID:      meta.ProviderID,
			DisplayName:     meta.DisplayName,
			Status:          meta.Status,
			LatencyMs:       meta.LatencyMs,
			QuotaTotalBytes: meta.QuotaTotal,
			QuotaUsedBytes:  meta.QuotaUsed,
		})
	}
	return providers, nil
}

func (c *localAdapterClient) UploadShard(ctx context.Context, shardID string, providerID string, data []byte) (*types.UploadShardResp, error) {
	provider, err := c.app.registry.Get(providerID)
	if err != nil {
		return nil, err
	}
	remoteID, err := provider.UploadShard(ctx, parseFileID(shardID), parseShardIndex(shardID), bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	hash := sha256.Sum256(data)
	return &types.UploadShardResp{RemoteID: remoteID, ChecksumSha: hex.EncodeToString(hash[:])}, nil
}

func (c *localAdapterClient) DownloadShard(ctx context.Context, remoteID string, providerID string) ([]byte, error) {
	provider, err := c.app.registry.Get(providerID)
	if err != nil {
		return nil, err
	}
	reader, err := provider.DownloadShard(ctx, remoteID)
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	return io.ReadAll(reader)
}

func (c *localAdapterClient) DeleteShard(ctx context.Context, remoteID string, providerID string) error {
	provider, err := c.app.registry.Get(providerID)
	if err != nil {
		return err
	}
	return provider.DeleteShard(ctx, remoteID)
}

func (c *localAdapterClient) DeleteFile(ctx context.Context, fileID string, deleteShards bool) error {
	_, err := c.app.deleteFileData(ctx, fileID, deleteShards)
	return err
}

type localShardMapClient struct {
	app *App
}

func (c *localShardMapClient) RegisterFile(_ context.Context, req *types.RegisterFileReq) (*types.RegisterFileResp, error) {
	resp, err := c.app.shardMapService.RegisterFile(&dto.RegisterFileRequest{
		OriginalName: req.OriginalName,
		OriginalSize: req.OriginalSize,
		TotalChunks:  req.TotalChunks,
		N:            req.N,
		K:            req.K,
		ShardSize:    req.ShardSize,
	})
	if err != nil {
		return nil, err
	}
	return &types.RegisterFileResp{FileID: resp.FileID, Status: resp.Status}, nil
}

func (c *localShardMapClient) RecordShards(_ context.Context, req *types.RecordShardReq) (*types.RecordShardResp, error) {
	shards := make([]dto.ShardInfo, 0, len(req.Shards))
	for _, shard := range req.Shards {
		shards = append(shards, dto.ShardInfo{
			ShardID:        shard.ShardID,
			ChunkIndex:     shard.ChunkIndex,
			ShardIndex:     shard.ShardIndex,
			Type:           shard.Type,
			RemoteID:       shard.RemoteID,
			Provider:       shard.Provider,
			ChecksumSHA256: shard.ChecksumSha,
			Status:         shard.Type,
		})
	}
	resp, err := c.app.shardMapService.RecordShards(&dto.RecordShardsRequest{FileID: req.FileID, Shards: shards})
	if err != nil {
		return nil, err
	}
	out := &types.RecordShardResp{FileID: resp.FileID, Shards: make([]types.ShardInfo, 0, len(resp.Shards))}
	for _, shard := range resp.Shards {
		out.Shards = append(out.Shards, types.ShardInfo{
			ShardID:     shard.ShardID,
			ChunkIndex:  shard.ChunkIndex,
			ShardIndex:  shard.ShardIndex,
			Type:        shard.Type,
			RemoteID:    shard.RemoteID,
			Provider:    shard.Provider,
			ChecksumSha: shard.ChecksumSHA256,
		})
	}
	return out, nil
}

func (c *localShardMapClient) ListFiles(_ context.Context) ([]types.FileMetadata, error) {
	files, err := c.app.shardMapService.ListFiles()
	if err != nil {
		return nil, err
	}
	out := make([]types.FileMetadata, 0, len(files))
	for _, file := range files {
		out = append(out, types.FileMetadata{
			FileID:              file.FileID,
			OriginalName:        file.OriginalName,
			OriginalSize:        file.OriginalSize,
			TotalChunks:         file.TotalChunks,
			TotalShards:         file.TotalShards,
			N:                   file.N,
			K:                   file.K,
			ChunkSize:           file.ChunkSize,
			ShardSize:           file.ShardSize,
			Status:              file.Status,
			CreatedAt:           file.CreatedAt,
			UpdatedAt:           file.UpdatedAt,
			LastHealthRefreshAt: file.LastHealthRefreshAt,
			FirstCreatedAt:      file.FirstCreatedAt,
			LastDownloadedAt:    file.LastDownloadedAt,
		})
	}
	return out, nil
}

func (c *localShardMapClient) GetShardMap(_ context.Context, fileID string) (*types.GetShardMapResp, error) {
	parsedID, err := uuid.Parse(fileID)
	if err != nil {
		return nil, err
	}
	resp, err := c.app.shardMapService.GetShardMap(parsedID)
	if err != nil {
		return nil, err
	}
	out := &types.GetShardMapResp{
		FileID:           resp.FileID,
		OriginalName:     resp.OriginalName,
		N:                resp.N,
		K:                resp.K,
		Status:           resp.Status,
		FirstCreatedAt:   resp.FirstCreatedAt,
		LastDownloadedAt: resp.LastDownloadedAt,
		Shards:           make([]types.ShardMapEntry, 0, len(resp.Shards)),
	}
	for _, shard := range resp.Shards {
		out.Shards = append(out.Shards, types.ShardMapEntry{
			ShardID:    shard.ShardID,
			ShardIndex: shard.ShardIndex,
			RemoteID:   shard.RemoteID,
			Provider:   shard.Provider,
			Status:     shard.Status,
			Checksum:   shard.ChecksumSHA256,
		})
	}
	return out, nil
}

func (c *localShardMapClient) MarkShardStatus(_ context.Context, shardID string, status string) error {
	parsedID, err := uuid.Parse(shardID)
	if err != nil {
		return err
	}
	return c.app.shardMapService.MarkShardStatus(parsedID, &dto.MarkShardStatusRequest{Status: status})
}

func (c *localShardMapClient) UpdateFileHealthRefresh(_ context.Context, fileID string, refreshedAt time.Time) error {
	parsedID, err := uuid.Parse(fileID)
	if err != nil {
		return err
	}
	return c.app.shardMapService.UpdateFileHealthRefresh(parsedID, refreshedAt)
}

func (c *localShardMapClient) LogLifecycleEvent(_ context.Context, event *types.LifecycleEvent) error {
	return c.app.lifecycleService.RecordEvent(event)
}

func (c *localShardMapClient) GetFileHistory(_ context.Context, fileID string) (*types.FileHistoryResp, error) {
	return c.app.lifecycleService.GetFileHistory(fileID)
}

func (c *localShardMapClient) GetAllHistory(_ context.Context) (*types.GlobalHistoryResp, error) {
	return c.app.lifecycleService.GetAllHistory()
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(wrapped, r)
		log.Printf("%s %s - %d (%v)", r.Method, r.URL.Path, wrapped.statusCode, time.Since(start))
	})
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
