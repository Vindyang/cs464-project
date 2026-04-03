package repository

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/vindyang/cs464-project/backend/services/shared/types"
)

// LifecycleRepository persists and queries file lifecycle events.
type LifecycleRepository interface {
	// EnsureSchema creates the lifecycle table and indexes if they don't already exist.
	EnsureSchema() error
	// Insert persists a new lifecycle event row.
	Insert(event *types.LifecycleEvent) error
	// GetByFileID returns all lifecycle events for a file, newest first.
	GetByFileID(fileID string) ([]types.LifecycleEvent, error)
}

type lifecycleRepository struct {
	db *sqlx.DB
}

// NewLifecycleRepository creates a new LifecycleRepository backed by the given DB.
func NewLifecycleRepository(db *sqlx.DB) LifecycleRepository {
	return &lifecycleRepository{db: db}
}

// EnsureSchema creates the lifecycle table and indexes idempotently.
// Called once at service startup. Safe to call even if the table already exists.
// NOTE: When shardmap migrates to SQLite, replace BIGINT→INTEGER, TIMESTAMP→DATETIME,
// and switch SQL placeholders from $N to ?.
func (r *lifecycleRepository) EnsureSchema() error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS file_lifecycle_log (
			id          BIGSERIAL PRIMARY KEY,
			file_id     TEXT NOT NULL,
			event_type  TEXT NOT NULL,
			file_name   TEXT,
			file_size   BIGINT,
			shard_count INTEGER,
			providers   TEXT,
			started_at  TIMESTAMP WITH TIME ZONE NOT NULL,
			ended_at    TIMESTAMP WITH TIME ZONE NOT NULL,
			duration_ms BIGINT NOT NULL,
			status      TEXT NOT NULL,
			error_msg   TEXT,
			created_at  TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_lifecycle_file_id
			ON file_lifecycle_log(file_id)`,
		`CREATE INDEX IF NOT EXISTS idx_lifecycle_event_created
			ON file_lifecycle_log(event_type, created_at DESC)`,
	}
	for _, stmt := range statements {
		if _, err := r.db.Exec(stmt); err != nil {
			return fmt.Errorf("lifecycle schema migration failed: %w", err)
		}
	}
	return nil
}

// Insert saves a lifecycle event to the database.
func (r *lifecycleRepository) Insert(event *types.LifecycleEvent) error {
	providers := strings.Join(event.Providers, ",")
	_, err := r.db.Exec(`
		INSERT INTO file_lifecycle_log
			(file_id, event_type, file_name, file_size, shard_count, providers,
			 started_at, ended_at, duration_ms, status, error_msg)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		event.FileID,
		event.EventType,
		nullableString(event.FileName),
		nullableInt64(event.FileSize),
		nullableInt(event.ShardCount),
		nullableString(providers),
		event.StartedAt.UTC(),
		event.EndedAt.UTC(),
		event.DurationMs,
		event.Status,
		nullableString(event.ErrorMsg),
	)
	if err != nil {
		return fmt.Errorf("failed to insert lifecycle event: %w", err)
	}
	return nil
}

// GetByFileID returns all lifecycle events for a file ordered newest first.
func (r *lifecycleRepository) GetByFileID(fileID string) ([]types.LifecycleEvent, error) {
	rows, err := r.db.Query(`
		SELECT file_id, event_type, file_name, file_size, shard_count, providers,
		       started_at, ended_at, duration_ms, status, error_msg
		FROM file_lifecycle_log
		WHERE file_id = $1
		ORDER BY created_at DESC`,
		fileID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query lifecycle events: %w", err)
	}
	defer rows.Close()

	var events []types.LifecycleEvent
	for rows.Next() {
		var e types.LifecycleEvent
		var providers sql.NullString
		var fileName, errorMsg sql.NullString
		var fileSize sql.NullInt64
		var shardCount sql.NullInt64

		if err := rows.Scan(
			&e.FileID,
			&e.EventType,
			&fileName,
			&fileSize,
			&shardCount,
			&providers,
			&e.StartedAt,
			&e.EndedAt,
			&e.DurationMs,
			&e.Status,
			&errorMsg,
		); err != nil {
			return nil, fmt.Errorf("failed to scan lifecycle row: %w", err)
		}
		if fileName.Valid {
			e.FileName = fileName.String
		}
		if fileSize.Valid {
			e.FileSize = fileSize.Int64
		}
		if shardCount.Valid {
			e.ShardCount = int(shardCount.Int64)
		}
		if errorMsg.Valid {
			e.ErrorMsg = errorMsg.String
		}
		if providers.Valid && providers.String != "" {
			e.Providers = strings.Split(providers.String, ",")
		}
		// Normalise timestamps to UTC.
		e.StartedAt = e.StartedAt.In(time.UTC)
		e.EndedAt = e.EndedAt.In(time.UTC)
		events = append(events, e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("lifecycle row iteration error: %w", err)
	}
	return events, nil
}

// nullableString returns nil for empty strings so they are stored as NULL.
func nullableString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

// nullableInt64 returns nil for zero values.
func nullableInt64(v int64) interface{} {
	if v == 0 {
		return nil
	}
	return v
}

// nullableInt returns nil for zero values.
func nullableInt(v int) interface{} {
	if v == 0 {
		return nil
	}
	return v
}
