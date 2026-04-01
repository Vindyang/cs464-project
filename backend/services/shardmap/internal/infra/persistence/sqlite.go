package persistence

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

// ConnectSQLite opens a local SQLite DB for shardmap metadata and ensures schema exists.
func ConnectSQLite(path string) (*sqlx.DB, error) {
	db, err := sqlx.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite db: %w", err)
	}

	if _, err := db.Exec("PRAGMA journal_mode=WAL;"); err != nil {
		return nil, fmt.Errorf("set WAL mode: %w", err)
	}
	if _, err := db.Exec("PRAGMA foreign_keys=ON;"); err != nil {
		return nil, fmt.Errorf("enable foreign keys: %w", err)
	}
	if _, err := db.Exec("PRAGMA busy_timeout=5000;"); err != nil {
		return nil, fmt.Errorf("set busy timeout: %w", err)
	}

	if err := migrate(db); err != nil {
		return nil, err
	}

	return db, nil
}

func migrate(db *sqlx.DB) error {
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
	updated_at DATETIME NOT NULL
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

CREATE INDEX IF NOT EXISTS idx_shards_file_id ON shards(file_id);
CREATE INDEX IF NOT EXISTS idx_shards_file_chunk ON shards(file_id, chunk_index);
`

	if _, err := db.Exec(schema); err != nil {
		return fmt.Errorf("migrate sqlite schema: %w", err)
	}
	return nil
}
