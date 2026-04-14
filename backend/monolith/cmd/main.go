package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	monolithapp "github.com/vindyang/cs464-project/backend/monolith/internal/app"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	application, err := monolithapp.New(monolithapp.ConfigFromEnv())
	if err != nil {
		log.Fatalf("failed to initialize monolith app: %v", err)
	}
	defer func() {
		if err := application.Close(); err != nil {
			log.Printf("error while closing app resources: %v", err)
		}
	}()

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      application.Handler(),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("Monolith service starting on :%s", port)
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
