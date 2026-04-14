package shardmap

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/vindyang/cs464-project/backend/monolith/shared/models"
	"github.com/vindyang/cs464-project/backend/monolith/shared/types"
)

type FileRepository interface {
	Create(file *models.File) error
	GetByID(id uuid.UUID) (*models.File, error)
	GetAll() ([]*models.File, error)
	GetAllWithHealth() ([]*models.FileWithHealth, error)
	UpdateStatus(id uuid.UUID, status models.FileStatus) error
	UpdateLastHealthRefresh(id uuid.UUID, refreshedAt time.Time) error
	Delete(id uuid.UUID) error
}

type ShardRepository interface {
	CreateBatch(shards []*models.Shard) error
	GetByID(id uuid.UUID) (*models.Shard, error)
	GetByFileID(fileID uuid.UUID) ([]*models.Shard, error)
	UpdateStatus(id uuid.UUID, status models.ShardStatus) error
}

type LifecycleRepository interface {
	Insert(event *types.LifecycleEvent) error
	DeleteAll() (int, error)
	GetByFileID(fileID string) ([]types.LifecycleEvent, error)
	GetAll() ([]types.LifecycleEvent, error)
	GetLifecycleSummary(fileID string) (*LifecycleSummary, error)
}

type fileRepository struct {
	db *sqlx.DB
}

type shardRepository struct {
	db *sqlx.DB
}

type lifecycleRepository struct {
	db *sqlx.DB
}

type LifecycleSummary struct {
	FirstCreatedAt   *time.Time
	LastDownloadedAt *time.Time
}

func NewFileRepository(database *sqlx.DB) FileRepository {
	return &fileRepository{db: database}
}

func NewShardRepository(database *sqlx.DB) ShardRepository {
	return &shardRepository{db: database}
}

func NewLifecycleRepository(database *sqlx.DB) LifecycleRepository {
	return &lifecycleRepository{db: database}
}

func (r *fileRepository) Create(file *models.File) error {
	_, err := r.db.Exec(`
		INSERT INTO files (id, original_name, original_size, total_chunks, n, k, shard_size, status, created_at, updated_at, last_health_refresh_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, file.ID.String(), file.OriginalName, file.OriginalSize, file.TotalChunks, file.N, file.K, file.ShardSize, file.Status, file.CreatedAt, file.UpdatedAt, file.LastHealthRefreshAt)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	return nil
}

func (r *fileRepository) GetByID(id uuid.UUID) (*models.File, error) {
	file := &models.File{}
	err := r.db.Get(file, `
		SELECT id, original_name, original_size, total_chunks, n, k, shard_size, status, created_at, updated_at, last_health_refresh_at
		FROM files
		WHERE id = ?
	`, id.String())
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("file not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get file: %w", err)
	}
	return file, nil
}

func (r *fileRepository) GetAll() ([]*models.File, error) {
	files := []*models.File{}
	err := r.db.Select(&files, `
		SELECT id, original_name, original_size, total_chunks, n, k, shard_size, status, created_at, updated_at, last_health_refresh_at
		FROM files
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to get all files: %w", err)
	}
	return files, nil
}

func (r *fileRepository) GetAllWithHealth() ([]*models.FileWithHealth, error) {
	files := []*models.FileWithHealth{}
	err := r.db.Select(&files, `
		SELECT
			f.id, f.original_name, f.original_size, f.total_chunks, f.n, f.k, f.shard_size, f.status, f.created_at, f.updated_at, f.last_health_refresh_at,
			COUNT(CASE WHEN s.status = 'HEALTHY' THEN 1 END) AS healthy_shards,
			COUNT(CASE WHEN s.status = 'CORRUPTED' THEN 1 END) AS corrupted_shards,
			COUNT(CASE WHEN s.status = 'MISSING' THEN 1 END) AS missing_shards,
			COUNT(s.id) AS total_shards
		FROM files f
		LEFT JOIN shards s ON s.file_id = f.id
		GROUP BY f.id, f.original_name, f.original_size, f.total_chunks, f.n, f.k, f.shard_size, f.status, f.created_at, f.updated_at, f.last_health_refresh_at
		ORDER BY f.created_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to get files: %w", err)
	}
	return files, nil
}

func (r *fileRepository) UpdateStatus(id uuid.UUID, status models.FileStatus) error {
	result, err := r.db.Exec(`
		UPDATE files
		SET status = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, status, id.String())
	if err != nil {
		return fmt.Errorf("failed to update file status: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("file not found")
	}
	return nil
}

func (r *fileRepository) UpdateLastHealthRefresh(id uuid.UUID, refreshedAt time.Time) error {
	result, err := r.db.Exec(`
		UPDATE files
		SET last_health_refresh_at = ?
		WHERE id = ?
	`, refreshedAt, id.String())
	if err != nil {
		return fmt.Errorf("failed to update file health refresh time: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("file not found")
	}
	return nil
}

func (r *fileRepository) Delete(id uuid.UUID) error {
	result, err := r.db.Exec(`DELETE FROM files WHERE id = ?`, id.String())
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("file not found")
	}
	return nil
}

func (r *shardRepository) CreateBatch(shards []*models.Shard) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO shards (id, file_id, chunk_index, shard_index, shard_type, remote_id, provider, checksum_sha256, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, shard := range shards {
		if _, err := stmt.Exec(
			shard.ID.String(),
			shard.FileID.String(),
			shard.ChunkIndex,
			shard.ShardIndex,
			shard.ShardType,
			shard.RemoteID,
			shard.Provider,
			shard.ChecksumSHA256,
			shard.Status,
			shard.CreatedAt,
			shard.UpdatedAt,
		); err != nil {
			return fmt.Errorf("failed to create shard in batch: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

func (r *shardRepository) GetByID(id uuid.UUID) (*models.Shard, error) {
	shard := &models.Shard{}
	err := r.db.Get(shard, `
		SELECT id, file_id, chunk_index, shard_index, shard_type, remote_id, provider, checksum_sha256, status, created_at, updated_at
		FROM shards
		WHERE id = ?
	`, id.String())
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("shard not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get shard: %w", err)
	}
	return shard, nil
}

func (r *shardRepository) GetByFileID(fileID uuid.UUID) ([]*models.Shard, error) {
	shards := []*models.Shard{}
	err := r.db.Select(&shards, `
		SELECT id, file_id, chunk_index, shard_index, shard_type, remote_id, provider, checksum_sha256, status, created_at, updated_at
		FROM shards
		WHERE file_id = ?
		ORDER BY chunk_index, shard_index
	`, fileID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get shards by file ID: %w", err)
	}
	return shards, nil
}

func (r *shardRepository) UpdateStatus(id uuid.UUID, status models.ShardStatus) error {
	result, err := r.db.Exec(`
		UPDATE shards
		SET status = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, status, id.String())
	if err != nil {
		return fmt.Errorf("failed to update shard status: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("shard not found")
	}
	return nil
}

var timeFormats = []string{
	time.RFC3339Nano,
	time.RFC3339,
	"2006-01-02T15:04:05Z07:00",
	"2006-01-02 15:04:05.999999999-07:00",
	"2006-01-02 15:04:05.999999999Z07:00",
	"2006-01-02 15:04:05.999999999",
	"2006-01-02 15:04:05",
	"2006-01-02",
}

func parseSQLiteTime(value string) (time.Time, error) {
	for _, format := range timeFormats {
		if parsed, err := time.Parse(format, value); err == nil {
			return parsed, nil
		}
	}
	return time.Time{}, fmt.Errorf("unable to parse time %q", value)
}

func (r *lifecycleRepository) Insert(event *types.LifecycleEvent) error {
	providers := strings.Join(event.Providers, ",")
	_, err := r.db.Exec(`
		INSERT INTO file_lifecycle_log
			(file_id, event_type, file_name, file_size, shard_count, providers, started_at, ended_at, duration_ms, status, error_msg)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, event.FileID, event.EventType, nullableString(event.FileName), nullableInt64(event.FileSize), nullableInt(event.ShardCount), nullableString(providers), event.StartedAt.UTC().Format(time.RFC3339Nano), event.EndedAt.UTC().Format(time.RFC3339Nano), event.DurationMs, event.Status, nullableString(event.ErrorMsg))
	if err != nil {
		return fmt.Errorf("failed to insert lifecycle event: %w", err)
	}
	return nil
}

func (r *lifecycleRepository) DeleteAll() (int, error) {
	result, err := r.db.Exec(`DELETE FROM file_lifecycle_log`)
	if err != nil {
		return 0, fmt.Errorf("failed to delete lifecycle events: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get deleted lifecycle row count: %w", err)
	}
	return int(rows), nil
}

func (r *lifecycleRepository) GetByFileID(fileID string) ([]types.LifecycleEvent, error) {
	rows, err := r.db.Query(`
		SELECT file_id, event_type, file_name, file_size, shard_count, providers, started_at, ended_at, duration_ms, status, error_msg
		FROM file_lifecycle_log
		WHERE file_id = ?
		ORDER BY created_at DESC
	`, fileID)
	if err != nil {
		return nil, fmt.Errorf("failed to query lifecycle events: %w", err)
	}
	defer rows.Close()
	return scanLifecycleRows(rows)
}

func (r *lifecycleRepository) GetAll() ([]types.LifecycleEvent, error) {
	rows, err := r.db.Query(`
		SELECT file_id, event_type, file_name, file_size, shard_count, providers, started_at, ended_at, duration_ms, status, error_msg
		FROM file_lifecycle_log
		ORDER BY created_at DESC
		LIMIT 200
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query lifecycle events: %w", err)
	}
	defer rows.Close()
	return scanLifecycleRows(rows)
}

func (r *lifecycleRepository) GetLifecycleSummary(fileID string) (*LifecycleSummary, error) {
	var firstCreatedRaw sql.NullString
	var lastDownloadedRaw sql.NullString
	err := r.db.QueryRow(`
		SELECT
			MIN(CASE WHEN event_type = 'upload' AND status = 'success' THEN ended_at END) AS first_created_at,
			MAX(CASE WHEN event_type = 'download' AND status = 'success' THEN ended_at END) AS last_downloaded_at
		FROM file_lifecycle_log
		WHERE file_id = ?
	`, fileID).Scan(&firstCreatedRaw, &lastDownloadedRaw)
	if err != nil {
		return nil, fmt.Errorf("failed to query lifecycle summary: %w", err)
	}

	summary := &LifecycleSummary{}
	if firstCreatedRaw.Valid && strings.TrimSpace(firstCreatedRaw.String) != "" {
		parsed, err := parseSQLiteTime(firstCreatedRaw.String)
		if err != nil {
			return nil, fmt.Errorf("failed to parse first_created_at: %w", err)
		}
		utc := parsed.UTC()
		summary.FirstCreatedAt = &utc
	}
	if lastDownloadedRaw.Valid && strings.TrimSpace(lastDownloadedRaw.String) != "" {
		parsed, err := parseSQLiteTime(lastDownloadedRaw.String)
		if err != nil {
			return nil, fmt.Errorf("failed to parse last_downloaded_at: %w", err)
		}
		utc := parsed.UTC()
		summary.LastDownloadedAt = &utc
	}

	return summary, nil
}

func scanLifecycleRows(rows *sql.Rows) ([]types.LifecycleEvent, error) {
	var events []types.LifecycleEvent
	for rows.Next() {
		var event types.LifecycleEvent
		var providers sql.NullString
		var fileName sql.NullString
		var errorMsg sql.NullString
		var fileSize sql.NullInt64
		var shardCount sql.NullInt64
		var startedAtStr string
		var endedAtStr string

		if err := rows.Scan(&event.FileID, &event.EventType, &fileName, &fileSize, &shardCount, &providers, &startedAtStr, &endedAtStr, &event.DurationMs, &event.Status, &errorMsg); err != nil {
			return nil, fmt.Errorf("failed to scan lifecycle row: %w", err)
		}

		startedAt, err := parseSQLiteTime(startedAtStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse started_at: %w", err)
		}
		endedAt, err := parseSQLiteTime(endedAtStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse ended_at: %w", err)
		}
		event.StartedAt = startedAt.UTC()
		event.EndedAt = endedAt.UTC()
		if fileName.Valid {
			event.FileName = fileName.String
		}
		if fileSize.Valid {
			event.FileSize = fileSize.Int64
		}
		if shardCount.Valid {
			event.ShardCount = int(shardCount.Int64)
		}
		if errorMsg.Valid {
			event.ErrorMsg = errorMsg.String
		}
		if providers.Valid && providers.String != "" {
			event.Providers = strings.Split(providers.String, ",")
		}
		events = append(events, event)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("lifecycle row iteration error: %w", err)
	}
	return events, nil
}

func nullableString(value string) interface{} {
	if value == "" {
		return nil
	}
	return value
}

func nullableInt64(value int64) interface{} {
	if value == 0 {
		return nil
	}
	return value
}

func nullableInt(value int) interface{} {
	if value == 0 {
		return nil
	}
	return value
}
