// cmd/gdrive-auth is a one-time CLI tool that runs the OAuth2 authorization
// code flow for Google Drive and saves the resulting token to a local file.
//
// Run once per machine before starting the server:
//
//	go run ./cmd/gdrive-auth/main.go
//
// Prerequisites (set in .env or environment):
//
//	GDRIVE_OAUTH_CREDENTIALS_FILE  path to OAuth2 client credentials JSON
//	                               (GCP Console → Credentials → OAuth 2.0 Client ID, type: Desktop app)
//	GDRIVE_TOKEN_FILE              path where the token will be saved
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const driveScope = "https://www.googleapis.com/auth/drive.file"

func main() {
	// Load .env if present so credentials paths don't need to be set manually.
	_ = godotenv.Load()

	oauthCredentialsFile := os.Getenv("GDRIVE_OAUTH_CREDENTIALS_FILE")
	tokenFile := os.Getenv("GDRIVE_TOKEN_FILE")

	if oauthCredentialsFile == "" || tokenFile == "" {
		log.Fatal("GDRIVE_OAUTH_CREDENTIALS_FILE and GDRIVE_TOKEN_FILE must be set in .env or environment")
	}

	raw, err := os.ReadFile(oauthCredentialsFile)
	if err != nil {
		log.Fatalf("read OAuth credentials file: %v", err)
	}

	config, err := google.ConfigFromJSON(raw, driveScope)
	if err != nil {
		log.Fatalf("parse OAuth credentials: %v", err)
	}

	// Generate the URL the user must open to grant access.
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Println("Open this URL in your browser and grant access:")
	fmt.Println()
	fmt.Println(authURL)
	fmt.Println()
	fmt.Print("Paste the authorization code here: ")

	var code string
	if _, err := fmt.Scan(&code); err != nil {
		log.Fatalf("read authorization code: %v", err)
	}

	token, err := config.Exchange(context.Background(), code)
	if err != nil {
		log.Fatalf("exchange authorization code: %v", err)
	}

	if err := saveTokenFile(tokenFile, token); err != nil {
		log.Fatalf("save token: %v", err)
	}

	fmt.Printf("\nToken saved to %s\n", tokenFile)
	fmt.Println("You can now run the server with: go run ./cmd/server/main.go")
}

func saveTokenFile(path string, token *oauth2.Token) error {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("create token file %q: %w", path, err)
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(token)
}
