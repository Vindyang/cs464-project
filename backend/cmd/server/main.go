package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/vindyang/cs464-project/backend/internal/adapter"
	"github.com/vindyang/cs464-project/backend/internal/adapter/gdrive"
	"github.com/vindyang/cs464-project/backend/internal/adapter/s3"
)

type App struct {
	Registry *adapter.Registry
}

func main() {
	// Load .env if present; silently ignored if absent (real env vars take precedence).
	_ = godotenv.Load()

	registry := adapter.NewRegistry()

	// AWS S3 adapter (stubbed; no credentials required yet)
	registry.Register("awsS3", s3.NewS3Adapter("my-bucket", "us-east-1"))

	// Google Drive adapter — enabled only when OAuth2 credentials are provided.
	// Set GDRIVE_OAUTH_CREDENTIALS_FILE (path to OAuth2 Desktop app credentials JSON),
	// GDRIVE_TOKEN_FILE (path to stored token — run cmd/gdrive-auth to generate it),
	// and GDRIVE_FOLDER_ID (Drive folder ID).
	oauthCredsFile := os.Getenv("GDRIVE_OAUTH_CREDENTIALS_FILE")
	tokenFile := os.Getenv("GDRIVE_TOKEN_FILE")
	folderID := os.Getenv("GDRIVE_FOLDER_ID")
	if oauthCredsFile == "" || tokenFile == "" || folderID == "" {
		log.Println("Warning: GDRIVE_OAUTH_CREDENTIALS_FILE, GDRIVE_TOKEN_FILE, or GDRIVE_FOLDER_ID not set; Google Drive adapter disabled")
	} else {
		gda, err := gdrive.NewGDriveAdapter(oauthCredsFile, tokenFile, folderID)
		if err != nil {
			log.Fatalf("Failed to initialize Google Drive adapter: %v", err)
		}
		registry.Register("googleDrive", gda)
	}

	app := &App{Registry: registry}

	http.HandleFunc("/api/providers", app.listProviders)

	log.Println("Adapter service starting on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func (app *App) listProviders(w http.ResponseWriter, r *http.Request) {
	metadatas := make([]*adapter.ProviderMetadata, 0)

	for _, id := range app.Registry.IDs() {
		p, err := app.Registry.Get(id)
		if err != nil {
			continue
		}
		meta, err := p.GetMetadata(r.Context())
		if err != nil {
			log.Printf("GetMetadata(%s): %v", id, err)
			continue
		}
		metadatas = append(metadatas, meta)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metadatas)
}
