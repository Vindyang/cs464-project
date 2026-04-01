// Package db provides a SQLite-backed local store for OAuth tokens,
// provider credentials, and miscellaneous key-value config.
package db

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"golang.org/x/oauth2"
	_ "modernc.org/sqlite"
)

// ErrNotFound is returned when a requested record does not exist.
var ErrNotFound = errors.New("db: record not found")

// Store is a SQLite-backed local store.
type Store struct {
	db *sql.DB
}

// NewStore opens (or creates) a SQLite database at the given file path
// and runs the schema migrations.
func NewStore(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("db: open %q: %w", path, err)
	}
	// SQLite only supports one writer at a time; WAL mode improves concurrency.
	if _, err := db.Exec("PRAGMA journal_mode=WAL;"); err != nil {
		return nil, fmt.Errorf("db: set WAL mode: %w", err)
	}
	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		return nil, fmt.Errorf("db: migrate: %w", err)
	}
	return s, nil
}

// Close releases the database connection.
func (s *Store) Close() error {
	return s.db.Close()
}

// migrate creates tables if they don't exist.
func (s *Store) migrate() error {
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS provider_tokens (
			provider_id   TEXT PRIMARY KEY,
			access_token  TEXT NOT NULL,
			refresh_token TEXT NOT NULL DEFAULT '',
			token_type    TEXT NOT NULL DEFAULT 'Bearer',
			expiry        DATETIME,
			updated_at    DATETIME DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS credentials (
			provider_id   TEXT PRIMARY KEY,
			client_id     TEXT NOT NULL,
			client_secret TEXT NOT NULL,
			redirect_uri  TEXT NOT NULL,
			updated_at    DATETIME DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS provider_config (
			key        TEXT PRIMARY KEY,
			value      TEXT NOT NULL,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
	`)
	return err
}

// ── Tokens ────────────────────────────────────────────────────────────────────

// UpsertToken saves (or replaces) the OAuth2 token for a provider.
func (s *Store) UpsertToken(providerID string, tok *oauth2.Token) error {
	_, err := s.db.Exec(`
		INSERT INTO provider_tokens (provider_id, access_token, refresh_token, token_type, expiry, updated_at)
		VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(provider_id) DO UPDATE SET
			access_token  = excluded.access_token,
			refresh_token = excluded.refresh_token,
			token_type    = excluded.token_type,
			expiry        = excluded.expiry,
			updated_at    = CURRENT_TIMESTAMP
	`, providerID, tok.AccessToken, tok.RefreshToken, tok.TokenType, tok.Expiry.UTC())
	if err != nil {
		return fmt.Errorf("db: upsert token for %q: %w", providerID, err)
	}
	return nil
}

// LoadToken retrieves the stored OAuth2 token for a provider.
// Returns ErrNotFound if no token exists.
func (s *Store) LoadToken(providerID string) (*oauth2.Token, error) {
	var (
		accessToken  string
		refreshToken string
		tokenType    string
		expiry       time.Time
	)
	err := s.db.QueryRow(`
		SELECT access_token, refresh_token, token_type, expiry
		FROM provider_tokens
		WHERE provider_id = ?
	`, providerID).Scan(&accessToken, &refreshToken, &tokenType, &expiry)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("db: load token for %q: %w", providerID, err)
	}
	return &oauth2.Token{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    tokenType,
		Expiry:       expiry,
	}, nil
}

// DeleteToken removes the stored token for a provider.
func (s *Store) DeleteToken(providerID string) error {
	_, err := s.db.Exec(`DELETE FROM provider_tokens WHERE provider_id = ?`, providerID)
	if err != nil {
		return fmt.Errorf("db: delete token for %q: %w", providerID, err)
	}
	return nil
}

// ── Credentials ───────────────────────────────────────────────────────────────

// UpsertCredentials saves (or replaces) OAuth client credentials for a provider.
func (s *Store) UpsertCredentials(providerID, clientID, clientSecret, redirectURI string) error {
	_, err := s.db.Exec(`
		INSERT INTO credentials (provider_id, client_id, client_secret, redirect_uri, updated_at)
		VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(provider_id) DO UPDATE SET
			client_id     = excluded.client_id,
			client_secret = excluded.client_secret,
			redirect_uri  = excluded.redirect_uri,
			updated_at    = CURRENT_TIMESTAMP
	`, providerID, clientID, clientSecret, redirectURI)
	if err != nil {
		return fmt.Errorf("db: upsert credentials for %q: %w", providerID, err)
	}
	return nil
}

// LoadCredentials retrieves the OAuth client credentials for a provider.
// Returns ErrNotFound if no credentials exist.
func (s *Store) LoadCredentials(providerID string) (clientID, clientSecret, redirectURI string, err error) {
	err = s.db.QueryRow(`
		SELECT client_id, client_secret, redirect_uri
		FROM credentials
		WHERE provider_id = ?
	`, providerID).Scan(&clientID, &clientSecret, &redirectURI)
	if errors.Is(err, sql.ErrNoRows) {
		return "", "", "", ErrNotFound
	}
	if err != nil {
		return "", "", "", fmt.Errorf("db: load credentials for %q: %w", providerID, err)
	}
	return clientID, clientSecret, redirectURI, nil
}

// DeleteCredentials removes the stored credentials for a provider.
func (s *Store) DeleteCredentials(providerID string) error {
	_, err := s.db.Exec(`DELETE FROM credentials WHERE provider_id = ?`, providerID)
	if err != nil {
		return fmt.Errorf("db: delete credentials for %q: %w", providerID, err)
	}
	return nil
}

// ── Key-Value Config ──────────────────────────────────────────────────────────

// SetConfig stores an arbitrary key-value pair (e.g. "gdrive_nebula_folder_id").
func (s *Store) SetConfig(key, value string) error {
	_, err := s.db.Exec(`
		INSERT INTO provider_config (key, value, updated_at)
		VALUES (?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(key) DO UPDATE SET
			value      = excluded.value,
			updated_at = CURRENT_TIMESTAMP
	`, key, value)
	if err != nil {
		return fmt.Errorf("db: set config %q: %w", key, err)
	}
	return nil
}

// GetConfig retrieves a config value by key.
// Returns ErrNotFound if the key does not exist.
func (s *Store) GetConfig(key string) (string, error) {
	var value string
	err := s.db.QueryRow(`SELECT value FROM provider_config WHERE key = ?`, key).Scan(&value)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrNotFound
	}
	if err != nil {
		return "", fmt.Errorf("db: get config %q: %w", key, err)
	}
	return value, nil
}
