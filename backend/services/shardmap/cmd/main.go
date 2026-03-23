package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/vindyang/cs464-project/backend/services/shardmap/internal/api/handlers"
	"github.com/vindyang/cs464-project/backend/services/shared/api/middleware"
	"github.com/vindyang/cs464-project/backend/services/shared/database"
	"github.com/vindyang/cs464-project/backend/services/shared/repository"
	"github.com/vindyang/cs464-project/backend/services/shared/service"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	// Connect to database
	db, err := database.ConnectFromEnv()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	log.Println("✅ Database connection established")

	// Initialize repositories
	fileRepo := repository.NewFileRepository(db)
	shardRepo := repository.NewShardRepository(db)

	// Initialize services
	shardMapService := service.NewShardMapService(fileRepo, shardRepo)

	// Initialize handlers
	shardMapHandler := handlers.NewShardMapHandler(shardMapService)

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
