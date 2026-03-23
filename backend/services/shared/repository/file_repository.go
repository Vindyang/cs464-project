package repository

import (
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/vindyang/cs464-project/backend/services/shared/models"
)

// FileRepository defines the interface for file database operations
type FileRepository interface {
	Create(file *models.File) error
	GetByID(id uuid.UUID) (*models.File, error)
	GetAll() ([]*models.File, error)
	UpdateStatus(id uuid.UUID, status models.FileStatus) error
	Delete(id uuid.UUID) error
}

// fileRepository implements FileRepository
type fileRepository struct {
	db *sqlx.DB
}

// NewFileRepository creates a new FileRepository instance
func NewFileRepository(db *sqlx.DB) FileRepository {
	return &fileRepository{db: db}
}

// Create inserts a new file record
func (r *fileRepository) Create(file *models.File) error {
	query := `
		INSERT INTO files (id, original_name, original_size, total_chunks, n, k, shard_size, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRow(
		query,
		file.ID,
		file.OriginalName,
		file.OriginalSize,
		file.TotalChunks,
		file.N,
		file.K,
		file.ShardSize,
		file.Status,
		file.CreatedAt,
		file.UpdatedAt,
	).Scan(&file.ID, &file.CreatedAt, &file.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}

	return nil
}

// GetByID retrieves a file by its ID
func (r *fileRepository) GetByID(id uuid.UUID) (*models.File, error) {
	file := &models.File{}
	query := `
		SELECT id, original_name, original_size, total_chunks, n, k, shard_size, status, created_at, updated_at
		FROM files
		WHERE id = $1
	`

	err := r.db.Get(file, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("file not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get file: %w", err)
	}

	return file, nil
}

// GetAll retrieves all files
func (r *fileRepository) GetAll() ([]*models.File, error) {
	files := []*models.File{}
	query := `
		SELECT id, original_name, original_size, total_chunks, n, k, shard_size, status, created_at, updated_at
		FROM files
		ORDER BY created_at DESC
	`

	err := r.db.Select(&files, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all files: %w", err)
	}

	return files, nil
}

// UpdateStatus updates a file's status
func (r *fileRepository) UpdateStatus(id uuid.UUID, status models.FileStatus) error {
	query := `
		UPDATE files
		SET status = $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
	`

	result, err := r.db.Exec(query, status, id)
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

// Delete removes a file record
func (r *fileRepository) Delete(id uuid.UUID) error {
	query := `DELETE FROM files WHERE id = $1`

	result, err := r.db.Exec(query, id)
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
