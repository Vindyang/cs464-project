package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/vindyang/cs464-project/backend/services/shared/models"
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

type fileRepository struct {
	db *sqlx.DB
}

func NewFileRepository(db *sqlx.DB) FileRepository {
	return &fileRepository{db: db}
}

func (r *fileRepository) Create(file *models.File) error {
	query := `
		INSERT INTO files (id, original_name, original_size, total_chunks, n, k, shard_size, status, created_at, updated_at, last_health_refresh_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.Exec(
		query,
		file.ID.String(),
		file.OriginalName,
		file.OriginalSize,
		file.TotalChunks,
		file.N,
		file.K,
		file.ShardSize,
		file.Status,
		file.CreatedAt,
		file.UpdatedAt,
		file.LastHealthRefreshAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}

	return nil
}

func (r *fileRepository) GetByID(id uuid.UUID) (*models.File, error) {
	file := &models.File{}
	query := `
		SELECT id, original_name, original_size, total_chunks, n, k, shard_size, status, created_at, updated_at, last_health_refresh_at
		FROM files
		WHERE id = ?
	`

	err := r.db.Get(file, query, id.String())
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
	query := `
		SELECT id, original_name, original_size, total_chunks, n, k, shard_size, status, created_at, updated_at, last_health_refresh_at
		FROM files
		ORDER BY created_at DESC
	`

	err := r.db.Select(&files, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all files: %w", err)
	}

	return files, nil
}

func (r *fileRepository) UpdateStatus(id uuid.UUID, status models.FileStatus) error {
	query := `
		UPDATE files
		SET status = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`

	result, err := r.db.Exec(query, status, id.String())
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
	query := `
		UPDATE files
		SET last_health_refresh_at = ?
		WHERE id = ?
	`

	result, err := r.db.Exec(query, refreshedAt, id.String())
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
	query := `DELETE FROM files WHERE id = ?`

	result, err := r.db.Exec(query, id.String())
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

func (r *fileRepository) GetAllWithHealth() ([]*models.FileWithHealth, error) {
	query := `
		SELECT
			f.id, f.original_name, f.original_size, f.total_chunks, f.n, f.k, f.shard_size, f.status, f.created_at, f.updated_at, f.last_health_refresh_at,
			COUNT(CASE WHEN s.status = 'HEALTHY'   THEN 1 END) AS healthy_shards,
			COUNT(CASE WHEN s.status = 'CORRUPTED' THEN 1 END) AS corrupted_shards,
			COUNT(CASE WHEN s.status = 'MISSING'   THEN 1 END) AS missing_shards,
			COUNT(s.id)                                         AS total_shards
		FROM files f
		LEFT JOIN shards s ON s.file_id = f.id
		GROUP BY f.id, f.original_name, f.original_size, f.total_chunks, f.n, f.k, f.shard_size, f.status, f.created_at, f.updated_at, f.last_health_refresh_at
		ORDER BY f.created_at DESC
	`

	var files []*models.FileWithHealth
	err := r.db.Select(&files, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get files: %w", err)
	}

	return files, nil
}
