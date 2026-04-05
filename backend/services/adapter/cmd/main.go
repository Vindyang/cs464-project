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

	"github.com/joho/godotenv"
	"github.com/vindyang/cs464-project/backend/services/adapter/internal/api/handlers"
	mockprovider "github.com/vindyang/cs464-project/backend/services/adapter/internal/mock"
	"github.com/vindyang/cs464-project/backend/services/shared/adapter"
	"github.com/vindyang/cs464-project/backend/services/shared/db"
	"github.com/vindyang/cs464-project/backend/services/shared/oauthhandler"
	"github.com/vindyang/cs464-project/backend/services/shared/s3handler"
)

//test CD
type App struct {
	Registry *adapter.Registry
}

func main() {
	_ = godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Open local SQLite store (token + credential + config persistence).
	dbPath := os.Getenv("Omnishard_DB_PATH")
	if dbPath == "" {
		dbPath = "Omnishard.db"
	}
	store, err := db.NewStore(dbPath)
	if err != nil {
		log.Fatalf("Failed to open local store: %v", err)
	}
	defer store.Close()

	// Initialize adapter registry
	registry := adapter.NewRegistry()

	mockModeEnabled := os.Getenv("ADAPTER_MOCK_MODE") == "true"
	if mockModeEnabled {
		registry.Register("mockLocal", mockprovider.NewProvider())
		log.Println("Adapter mock mode enabled: registered mockLocal provider")
	} else {
		// Restore Google Drive adapter from stored token if available
		if err := tryRestoreGDriveAdapter(store, registry); err != nil {
			log.Printf("Google Drive adapter not restored: %v", err)
		}

		// Restore S3 adapter from stored credentials if available
		s3Handler := s3handler.New(store, registry)
		if err := s3Handler.RestoreAdapter(); err != nil {
			log.Printf("S3 adapter not restored: %v", err)
		}
	}

	// OAuth handler for Google Drive
	oauthHandler := oauthhandler.New(store, registry)
	s3Handler := s3handler.New(store, registry)

	credentialsHandler := handlers.NewCredentialsHandler(store)
	settingsHandler := handlers.NewSettingsHandler(store)
	shardIOHandler := handlers.NewShardIOHandler(registry)

	shardmapURL := os.Getenv("SHARDMAP_URL")
	if shardmapURL == "" {
		shardmapURL = "http://localhost:8081"
	}
	fileHandler := handlers.NewFileHandler(shardmapURL, registry)

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
	mux.HandleFunc("/api/providers/awsS3/connect", s3Handler.Connect)
	mux.HandleFunc("/api/providers/awsS3/disconnect", s3Handler.Disconnect)
	credentialsHandler.RegisterRoutes(mux)
	settingsHandler.RegisterRoutes(mux)
	shardIOHandler.RegisterRoutes(mux)
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

// tryRestoreGDriveAdapter loads a stored token from the local store and re-registers the adapter.
// Non-fatal: logs and returns nil if no token is found.
func tryRestoreGDriveAdapter(store *db.Store, registry *adapter.Registry) error {
	tok, err := store.LoadToken("googleDrive")
	if err != nil {
		// No token stored yet — user hasn't connected. Not an error.
		log.Println("No stored Google Drive token — connect via UI")
		return nil
	}

	h := oauthhandler.New(store, registry)
	if err := h.RestoreAdapter(registry, tok); err != nil {
		return err
	}
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
