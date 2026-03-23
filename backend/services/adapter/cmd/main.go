package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
	"github.com/vindyang/cs464-project/backend/services/adapter/internal/adapter"
	"github.com/vindyang/cs464-project/backend/services/adapter/internal/adapter/gdrive"
	"github.com/vindyang/cs464-project/backend/services/adapter/internal/api/handlers"
	"github.com/vindyang/cs464-project/backend/services/adapter/internal/database"
	"github.com/vindyang/cs464-project/backend/services/adapter/internal/db"
	"github.com/vindyang/cs464-project/backend/services/adapter/internal/oauthhandler"
	"github.com/vindyang/cs464-project/backend/services/adapter/internal/repository"
	"github.com/vindyang/cs464-project/backend/services/adapter/internal/service"
	"golang.org/x/oauth2/google"
)

const driveScope = "https://www.googleapis.com/auth/drive.file"

type App struct {
	Registry *adapter.Registry
}

func main() {
	_ = godotenv.Load()

	ctx := context.Background()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Connect to Supabase via pgx (OAuth token storage)
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL must be set")
	}
	tokenDB, err := db.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to token database: %v", err)
	}
	defer tokenDB.Close()

	// Connect to database via sqlx (file/shard metadata)
	log.Println("Connecting to database...")
	sqlDB, err := database.ConnectFromEnv()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer sqlDB.Close()
	log.Println("Database connected successfully")

	// Initialize repositories and services
	fileRepo := repository.NewFileRepository(sqlDB)
	shardRepo := repository.NewShardRepository(sqlDB)
	shardingService := service.NewShardingService()
	shardMapService := service.NewShardMapService(fileRepo, shardRepo)

	// Initialize adapter registry
	registry := adapter.NewRegistry()

	// Restore Google Drive adapter from stored token if available
	oauthCredsFile := os.Getenv("GDRIVE_OAUTH_CREDENTIALS_FILE")
	redirectURI := os.Getenv("GDRIVE_OAUTH_REDIRECT_URI")
	folderID := os.Getenv("GDRIVE_FOLDER_ID")
	if oauthCredsFile != "" && redirectURI != "" && folderID != "" {
		if err := tryRestoreGDriveAdapter(ctx, tokenDB, registry, oauthCredsFile, redirectURI, folderID); err != nil {
			log.Printf("Google Drive adapter not restored: %v", err)
		}
	}

	// OAuth handler for Google Drive
	oauthHandler, err := oauthhandler.New(tokenDB, registry)
	if err != nil {
		log.Fatalf("Failed to initialize OAuth handler: %v", err)
	}

	// File operations service and handlers
	fileOperationsService := service.NewFileOperationsService(shardingService, shardMapService, registry)
	shardMapHandler := handlers.NewShardMapHandler(shardMapService)
	fileHandler := handlers.NewFileHandler(fileOperationsService)

	app := &App{Registry: registry}

	// Set up routes
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy"}`))
	})
	mux.HandleFunc("/api/providers", app.listProviders)
	mux.HandleFunc("/api/oauth/gdrive/authorize", oauthHandler.Authorize)
	mux.HandleFunc("/api/oauth/gdrive/callback", oauthHandler.Callback)
	mux.HandleFunc("/api/oauth/gdrive/disconnect", oauthHandler.Disconnect)
	shardMapHandler.RegisterRoutes(mux)
	fileHandler.RegisterRoutes(mux)

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      corsMiddleware(loggingMiddleware(mux)),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("Adapter service starting on :%s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Server shutting down...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
	log.Println("Server exited gracefully")
}

// tryRestoreGDriveAdapter loads a stored token from DB and registers the adapter.
func tryRestoreGDriveAdapter(ctx context.Context, tokenDB *db.DB, registry *adapter.Registry, credsFile, redirectURI, folderID string) error {
	tok, err := tokenDB.LoadProviderToken(ctx, "googleDrive")
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

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
