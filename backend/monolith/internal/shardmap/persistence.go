package shardmap

import (
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

func ConnectSQLite(path string) (*sqlx.DB, error) {
	database, err := sqlx.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite db: %w", err)
	}
	database.SetMaxOpenConns(1)
	database.SetMaxIdleConns(1)

	if _, err := database.Exec("PRAGMA journal_mode=WAL;"); err != nil {
		return nil, fmt.Errorf("set WAL mode: %w", err)
	}
	if _, err := database.Exec("PRAGMA foreign_keys=ON;"); err != nil {
		return nil, fmt.Errorf("enable foreign keys: %w", err)
	}
	if _, err := database.Exec("PRAGMA busy_timeout=5000;"); err != nil {
		return nil, fmt.Errorf("set busy timeout: %w", err)
	}

	if err := migrate(database); err != nil {
		return nil, err
	}

	return database, nil
}

func migrate(database *sqlx.DB) error {
	const schema = `
CREATE TABLE IF NOT EXISTS files (
	id TEXT PRIMARY KEY,
	original_name TEXT,
	original_size INTEGER NOT NULL,
	total_chunks INTEGER NOT NULL,
	n INTEGER NOT NULL,
	k INTEGER NOT NULL,
	shard_size INTEGER NOT NULL,
	status TEXT NOT NULL,
	created_at DATETIME NOT NULL,
	updated_at DATETIME NOT NULL,
	last_health_refresh_at DATETIME
);

CREATE TABLE IF NOT EXISTS shards (
	id TEXT PRIMARY KEY,
	file_id TEXT NOT NULL,
	chunk_index INTEGER NOT NULL,
	shard_index INTEGER NOT NULL,
	shard_type TEXT NOT NULL,
	remote_id TEXT NOT NULL,
	provider TEXT NOT NULL,
	checksum_sha256 TEXT NOT NULL,
	status TEXT NOT NULL,
	created_at DATETIME NOT NULL,
	updated_at DATETIME NOT NULL,
	FOREIGN KEY(file_id) REFERENCES files(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS file_lifecycle_log (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	file_id TEXT NOT NULL,
	event_type TEXT NOT NULL,
	file_name TEXT,
	file_size INTEGER,
	shard_count INTEGER,
	providers TEXT,
	started_at DATETIME NOT NULL,
	ended_at DATETIME NOT NULL,
	duration_ms INTEGER NOT NULL,
	status TEXT NOT NULL,
	error_msg TEXT,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_shards_file_id ON shards(file_id);
CREATE INDEX IF NOT EXISTS idx_shards_file_chunk ON shards(file_id, chunk_index);
CREATE INDEX IF NOT EXISTS idx_lifecycle_file_id ON file_lifecycle_log(file_id);
CREATE INDEX IF NOT EXISTS idx_lifecycle_event_created ON file_lifecycle_log(event_type, created_at DESC);
`

	if _, err := database.Exec(schema); err != nil {
		return fmt.Errorf("migrate sqlite schema: %w", err)
	}
	if _, err := database.Exec(`ALTER TABLE files ADD COLUMN last_health_refresh_at DATETIME`); err != nil {
		if !strings.Contains(strings.ToLower(err.Error()), "duplicate column name") {
			return fmt.Errorf("migrate sqlite schema add last_health_refresh_at: %w", err)
		}
	}
	return nil
}
