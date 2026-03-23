// Package db provides a thin wrapper around pgx for Supabase token storage.
package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/oauth2"
)

// DB wraps a pgx connection pool.
type DB struct {
	pool *pgxpool.Pool
}

// New connects to the database using the given PostgreSQL connection URL.
func New(ctx context.Context, connURL string) (*DB, error) {
	pool, err := pgxpool.New(ctx, connURL)
	if err != nil {
		return nil, fmt.Errorf("db: connect: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("db: ping: %w", err)
	}
	return &DB{pool: pool}, nil
}

// Close releases all pool connections.
func (d *DB) Close() {
	d.pool.Close()
}

// UpsertProviderToken inserts or replaces the OAuth2 token for a provider.
func (d *DB) UpsertProviderToken(ctx context.Context, providerID string, tok *oauth2.Token) error {
	_, err := d.pool.Exec(ctx, `
		INSERT INTO provider_connections (provider_id, access_token, refresh_token, token_type, expiry, updated_at)
		VALUES ($1, $2, $3, $4, $5, now())
		ON CONFLICT (provider_id) DO UPDATE
		  SET access_token  = EXCLUDED.access_token,
		      refresh_token = EXCLUDED.refresh_token,
		      token_type    = EXCLUDED.token_type,
		      expiry        = EXCLUDED.expiry,
		      updated_at    = now()
	`, providerID, tok.AccessToken, tok.RefreshToken, tok.TokenType, tok.Expiry.UTC())
	if err != nil {
		return fmt.Errorf("db: upsert token for %q: %w", providerID, err)
	}
	return nil
}

// LoadProviderToken retrieves the stored OAuth2 token for a provider.
// Returns pgx.ErrNoRows if no token exists.
func (d *DB) LoadProviderToken(ctx context.Context, providerID string) (*oauth2.Token, error) {
	var (
		accessToken  string
		refreshToken string
		tokenType    string
		expiry       time.Time
	)
	err := d.pool.QueryRow(ctx, `
		SELECT access_token, refresh_token, token_type, expiry
		FROM provider_connections
		WHERE provider_id = $1
	`, providerID).Scan(&accessToken, &refreshToken, &tokenType, &expiry)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, pgx.ErrNoRows
		}
		return nil, fmt.Errorf("db: load token for %q: %w", providerID, err)
	}
	return &oauth2.Token{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    tokenType,
		Expiry:       expiry,
	}, nil
}

// DeleteProviderToken removes the stored token for a provider.
func (d *DB) DeleteProviderToken(ctx context.Context, providerID string) error {
	_, err := d.pool.Exec(ctx, `
		DELETE FROM provider_connections WHERE provider_id = $1
	`, providerID)
	if err != nil {
		return fmt.Errorf("db: delete token for %q: %w", providerID, err)
	}
	return nil
}
