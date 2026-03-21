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
	"github.com/vindyang/cs464-project/backend/internal/adapter"
	"github.com/vindyang/cs464-project/backend/internal/adapter/gdrive"
	"github.com/vindyang/cs464-project/backend/internal/adapter/s3"
	"github.com/vindyang/cs464-project/backend/internal/api/handlers"
	"github.com/vindyang/cs464-project/backend/internal/database"
	"github.com/vindyang/cs464-project/backend/internal/repository"
	"github.com/vindyang/cs464-project/backend/internal/service"
)

type App struct {
	Registry *adapter.Registry
}

func main() {
	// Load environment variables from project root (one level up from backend dir)
	if err := godotenv.Load("../.env"); err != nil {
		// Try current directory as fallback
		if err := godotenv.Load(".env"); err != nil {
			log.Printf("Warning: .env file not found in ../.env or .env: %v", err)
		}
	}

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Connect to database
	log.Println("Connecting to database...")
	db, err := database.ConnectFromEnv()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()
	log.Println("Database connected successfully")

	// Initialize repositories
	fileRepo := repository.NewFileRepository(db)
	shardRepo := repository.NewShardRepository(db)

	// Initialize services
	shardingService := service.NewShardingService()
	shardMapService := service.NewShardMapService(fileRepo, shardRepo)

	// Initialize adapter registry
	registry := adapter.NewRegistry()

	// Register adapters (configured via ENV in production)
	registry.Register("awsS3", s3.NewS3Adapter("my-bucket", "us-east-1"))
	registry.Register("googleDrive", gdrive.NewGDriveAdapter("root-folder"))
	log.Println("Registered cloud providers: awsS3, googleDrive")

	// Initialize file operations service
	fileOperationsService := service.NewFileOperationsService(
		shardingService,
		shardMapService,
		registry,
	)

	// Initialize handlers
	shardMapHandler := handlers.NewShardMapHandler(shardMapService)
	fileHandler := handlers.NewFileHandler(fileOperationsService)

	app := &App{Registry: registry}

	// Set up HTTP server with routes
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy"}`))
	})

	// Provider metadata endpoint
	mux.HandleFunc("/api/providers", app.listProviders)

	// Register API routes
	shardMapHandler.RegisterRoutes(mux)
	fileHandler.RegisterRoutes(mux)

	// Wrap with middleware
	handler := corsMiddleware(loggingMiddleware(mux))

	// Create HTTP server
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Server starting on port %s", port)
		log.Printf("Endpoints available:")
		log.Printf("  - GET  /health")
		log.Printf("  - GET  /api/providers")
		log.Printf("  - POST /api/v1/shards/register")
		log.Printf("  - POST /api/v1/shards/record")
		log.Printf("  - GET  /api/v1/shards/file/:fileId")
		log.Printf("  - GET  /api/v1/shards/:shardId")
		log.Printf("  - PUT  /api/v1/shards/:shardId/status")
		log.Printf("  - POST /api/v1/files/upload")
		log.Printf("  - GET  /api/v1/files/:fileId")
		log.Printf("  - GET  /api/v1/files/:fileId/download")
		log.Printf("  - DELETE /api/v1/files/:fileId")

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Server shutting down...")

	// Gracefully shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited gracefully")
}

func (app *App) listProviders(w http.ResponseWriter, r *http.Request) {
	// List all registered providers
	providers := []string{"awsS3", "googleDrive"}
	metadatas := make([]*adapter.ProviderMetadata, 0)

	for _, id := range providers {
		p, _ := app.Registry.Get(id)
		meta, _ := p.GetMetadata(r.Context())
		metadatas = append(metadatas, meta)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metadatas)
}

// loggingMiddleware logs HTTP requests
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap response writer to capture status code
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		duration := time.Since(start)
		log.Printf("%s %s - %d (%v)", r.Method, r.URL.Path, wrapped.statusCode, duration)
	})
}

// corsMiddleware adds CORS headers
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Allow all origins for development (restrict in production)
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Max-Age", "3600")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
