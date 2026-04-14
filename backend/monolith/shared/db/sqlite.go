package db

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"golang.org/x/oauth2"
	_ "modernc.org/sqlite"
)

var ErrNotFound = errors.New("db: record not found")

type Store struct {
	db *sql.DB
}

type CredentialRecord struct {
	ProviderID  string    `json:"provider_id"`
	ClientID    string    `json:"client_id"`
	RedirectURI string    `json:"redirect_uri"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func NewStore(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("db: open %q: %w", path, err)
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	if _, err := db.Exec("PRAGMA journal_mode=WAL;"); err != nil {
		return nil, fmt.Errorf("db: set WAL mode: %w", err)
	}
	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		return nil, fmt.Errorf("db: migrate: %w", err)
	}
	return s, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

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

func (s *Store) LoadToken(providerID string) (*oauth2.Token, error) {
	var accessToken string
	var refreshToken string
	var tokenType string
	var expiry time.Time
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
	return &oauth2.Token{AccessToken: accessToken, RefreshToken: refreshToken, TokenType: tokenType, Expiry: expiry}, nil
}

func (s *Store) DeleteToken(providerID string) error {
	_, err := s.db.Exec(`DELETE FROM provider_tokens WHERE provider_id = ?`, providerID)
	if err != nil {
		return fmt.Errorf("db: delete token for %q: %w", providerID, err)
	}
	return nil
}

func (s *Store) DeleteAllTokens() (int, error) {
	result, err := s.db.Exec(`DELETE FROM provider_tokens`)
	if err != nil {
		return 0, fmt.Errorf("db: delete all tokens: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("db: rows affected deleting tokens: %w", err)
	}
	return int(rows), nil
}

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

func (s *Store) DeleteCredentials(providerID string) error {
	_, err := s.db.Exec(`DELETE FROM credentials WHERE provider_id = ?`, providerID)
	if err != nil {
		return fmt.Errorf("db: delete credentials for %q: %w", providerID, err)
	}
	return nil
}

func (s *Store) DeleteAllCredentials() (int, error) {
	result, err := s.db.Exec(`DELETE FROM credentials`)
	if err != nil {
		return 0, fmt.Errorf("db: delete all credentials: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("db: rows affected deleting credentials: %w", err)
	}
	return int(rows), nil
}

func (s *Store) ListCredentials() ([]CredentialRecord, error) {
	rows, err := s.db.Query(`
		SELECT provider_id, client_id, redirect_uri, updated_at
		FROM credentials
		ORDER BY provider_id ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("db: list credentials: %w", err)
	}
	defer rows.Close()

	out := make([]CredentialRecord, 0)
	for rows.Next() {
		var rec CredentialRecord
		if err := rows.Scan(&rec.ProviderID, &rec.ClientID, &rec.RedirectURI, &rec.UpdatedAt); err != nil {
			return nil, fmt.Errorf("db: scan credential row: %w", err)
		}
		out = append(out, rec)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("db: iterate credentials rows: %w", err)
	}
	return out, nil
}

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
