package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/vindyang/cs464-project/backend/services/sharding/internal/api/handlers"
	"github.com/vindyang/cs464-project/backend/services/sharding/internal/app"
)
// CD testing again
func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8083"
	}

	shardingService := app.NewShardingService()
	shardingHandler := handlers.NewShardingHandler(shardingService)

	mux := http.NewServeMux()
	shardingHandler.RegisterRoutes(mux)

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("Sharding service starting on :%s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Sharding service failed to start: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Sharding service shutting down...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Sharding service forced to shutdown: %v", err)
	}

	log.Println("Sharding service exited gracefully")
}
