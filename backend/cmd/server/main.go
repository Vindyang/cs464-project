package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/vindyang/cs464-project/backend/internal/adapter"
	"github.com/vindyang/cs464-project/backend/internal/adapter/gdrive"
	"github.com/vindyang/cs464-project/backend/internal/adapter/s3"
)

type App struct {
	Registry *adapter.Registry
}

func main() {
	registry := adapter.NewRegistry()

	// Initialize adapters (configured via ENV in production)
	registry.Register("awsS3", s3.NewS3Adapter("my-bucket", "us-east-1"))
	registry.Register("googleDrive", gdrive.NewGDriveAdapter("root-folder"))

	app := &App{Registry: registry}

	http.HandleFunc("/api/providers", app.listProviders)
	
	log.Println("Adapter service starting on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func (app *App) listProviders(w http.ResponseWriter, r *http.Request) {
	// Simple mock implementation for now
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
