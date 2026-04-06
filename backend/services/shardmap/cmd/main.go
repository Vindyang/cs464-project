package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/vindyang/cs464-project/backend/services/shardmap/internal/api/handlers"
	"github.com/vindyang/cs464-project/backend/services/shardmap/internal/app"
	"github.com/vindyang/cs464-project/backend/services/shardmap/internal/infra/persistence"
	"github.com/vindyang/cs464-project/backend/services/shardmap/internal/infra/repository"
	"github.com/vindyang/cs464-project/backend/services/shared/api/middleware"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	// Connect to local SQLite database for shard map/file metadata.
	dbPath := os.Getenv("Omnishard_SHARDMAP_DB_PATH")
	if dbPath == "" {
		dbPath = "Omnishard-shardmap.db"
	}
	db, err := persistence.ConnectSQLite(dbPath)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	log.Println("✅ Database connection established")

	// Initialize repositories
	fileRepo := repository.NewFileRepository(db)
	shardRepo := repository.NewShardRepository(db)
	lifecycleRepo := repository.NewLifecycleRepository(db)

	// Ensure lifecycle schema exists (idempotent, safe on every startup)
	if err := lifecycleRepo.EnsureSchema(); err != nil {
		log.Fatalf("Failed to initialize lifecycle schema: %v", err)
	}
	log.Println("✅ Lifecycle schema ready")

	// Initialize services
	shardMapService := app.NewShardMapService(fileRepo, shardRepo, lifecycleRepo)
	lifecycleService := app.NewLifecycleService(lifecycleRepo)

	// Initialize handlers
	shardMapHandler := handlers.NewShardMapHandler(shardMapService)
	lifecycleHandler := handlers.NewLifecycleHandler(lifecycleService)

	// Set up HTTP mux
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "healthy",
			"service": "shard-map-service",
		})
	})

	// Register API routes
	shardMapHandler.RegisterRoutes(mux)
	lifecycleHandler.RegisterRoutes(mux)

	// Apply middleware
	handler := middleware.Logger(middleware.CORS(middleware.Recovery(mux)))

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	// Start server
	log.Printf("🚀 Shard Map Service starting on port %s", port)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
