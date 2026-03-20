package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
	"github.com/vindyang/cs464-project/backend/internal/adapter"
	"github.com/vindyang/cs464-project/backend/internal/adapter/gdrive"
	"github.com/vindyang/cs464-project/backend/internal/db"
	"github.com/vindyang/cs464-project/backend/internal/oauthhandler"
	"golang.org/x/oauth2/google"
)

const driveScope = "https://www.googleapis.com/auth/drive.file"

type App struct {
	Registry *adapter.Registry
}

func main() {
	_ = godotenv.Load()

	ctx := context.Background()
	registry := adapter.NewRegistry()


	// Connect to Supabase
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL must be set")
	}
	database, err := db.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	// Restore Google Drive adapter from stored token if available
	oauthCredsFile := os.Getenv("GDRIVE_OAUTH_CREDENTIALS_FILE")
	redirectURI := os.Getenv("GDRIVE_OAUTH_REDIRECT_URI")
	folderID := os.Getenv("GDRIVE_FOLDER_ID")
	if oauthCredsFile != "" && redirectURI != "" && folderID != "" {
		if err := tryRestoreGDriveAdapter(ctx, database, registry, oauthCredsFile, redirectURI, folderID); err != nil {
			log.Printf("Google Drive adapter not restored: %v", err)
		}
	}

	// OAuth handler for Google Drive
	oauthHandler, err := oauthhandler.New(database, registry)
	if err != nil {
		log.Fatalf("Failed to initialize OAuth handler: %v", err)
	}

	app := &App{Registry: registry}

	http.HandleFunc("/api/providers", corsMiddleware(app.listProviders))
	http.HandleFunc("/api/oauth/gdrive/authorize", corsMiddleware(oauthHandler.Authorize))
	http.HandleFunc("/api/oauth/gdrive/callback", oauthHandler.Callback)
	http.HandleFunc("/api/oauth/gdrive/disconnect", corsMiddleware(oauthHandler.Disconnect))

	log.Println("Adapter service starting on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

// tryRestoreGDriveAdapter loads a stored token from DB and registers the adapter.
func tryRestoreGDriveAdapter(ctx context.Context, database *db.DB, registry *adapter.Registry, credsFile, redirectURI, folderID string) error {
	tok, err := database.LoadProviderToken(ctx, "googleDrive")
	if err != nil {
		if err == pgx.ErrNoRows {
			log.Println("No stored Google Drive token — connect via UI")
			return nil
		}
		return err
	}

	raw, err := os.ReadFile(credsFile)
	if err != nil {
		return err
	}

	config, err := google.ConfigFromJSON(raw, driveScope)
	if err != nil {
		return err
	}
	config.RedirectURL = redirectURI

	gda, err := gdrive.NewGDriveAdapter(config, tok, folderID)
	if err != nil {
		return err
	}
	registry.Register("googleDrive", gda)
	log.Println("Google Drive adapter restored from stored token")
	return nil
}

func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next(w, r)
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
