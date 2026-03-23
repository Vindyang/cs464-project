package repository

import (
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/vindyang/cs464-project/backend/services/shared/models"
)

type ShardRepository interface {
	Create(shard *models.Shard) error
	CreateBatch(shards []*models.Shard) error
	GetByID(id uuid.UUID) (*models.Shard, error)
	GetByFileID(fileID uuid.UUID) ([]*models.Shard, error)
	GetByFileAndChunk(fileID uuid.UUID, chunkIndex int) ([]*models.Shard, error)
	UpdateStatus(id uuid.UUID, status models.ShardStatus) error
	Delete(id uuid.UUID) error
}

type shardRepository struct {
	db *sqlx.DB
}

func NewShardRepository(db *sqlx.DB) ShardRepository {
	return &shardRepository{db: db}
}

func (r *shardRepository) Create(shard *models.Shard) error {
	query := `
		INSERT INTO shards (id, file_id, chunk_index, shard_index, shard_type, remote_id, provider, checksum_sha256, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRow(
		query,
		shard.ID,
		shard.FileID,
		shard.ChunkIndex,
		shard.ShardIndex,
		shard.ShardType,
		shard.RemoteID,
		shard.Provider,
		shard.ChecksumSHA256,
		shard.Status,
		shard.CreatedAt,
		shard.UpdatedAt,
	).Scan(&shard.ID, &shard.CreatedAt, &shard.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create shard: %w", err)
	}

	return nil
}

func (r *shardRepository) CreateBatch(shards []*models.Shard) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	query := `
		INSERT INTO shards (id, file_id, chunk_index, shard_index, shard_type, remote_id, provider, checksum_sha256, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, created_at, updated_at
	`

	stmt, err := tx.Prepare(query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, shard := range shards {
		err := stmt.QueryRow(
			shard.ID,
			shard.FileID,
			shard.ChunkIndex,
			shard.ShardIndex,
			shard.ShardType,
			shard.RemoteID,
			shard.Provider,
			shard.ChecksumSHA256,
			shard.Status,
			shard.CreatedAt,
			shard.UpdatedAt,
		).Scan(&shard.ID, &shard.CreatedAt, &shard.UpdatedAt)

		if err != nil {
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
	query := `
		SELECT id, file_id, chunk_index, shard_index, shard_type, remote_id, provider, checksum_sha256, status, created_at, updated_at
		FROM shards
		WHERE id = $1
	`

	err := r.db.Get(shard, query, id)
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
	query := `
		SELECT id, file_id, chunk_index, shard_index, shard_type, remote_id, provider, checksum_sha256, status, created_at, updated_at
		FROM shards
		WHERE file_id = $1
		ORDER BY chunk_index, shard_index
	`

	err := r.db.Select(&shards, query, fileID)
	if err != nil {
		return nil, fmt.Errorf("failed to get shards by file ID: %w", err)
	}

	return shards, nil
}

func (r *shardRepository) GetByFileAndChunk(fileID uuid.UUID, chunkIndex int) ([]*models.Shard, error) {
	shards := []*models.Shard{}
	query := `
		SELECT id, file_id, chunk_index, shard_index, shard_type, remote_id, provider, checksum_sha256, status, created_at, updated_at
		FROM shards
		WHERE file_id = $1 AND chunk_index = $2
		ORDER BY shard_index
	`

	err := r.db.Select(&shards, query, fileID, chunkIndex)
	if err != nil {
		return nil, fmt.Errorf("failed to get shards by file and chunk: %w", err)
	}

	return shards, nil
}

func (r *shardRepository) UpdateStatus(id uuid.UUID, status models.ShardStatus) error {
	query := `
		UPDATE shards
		SET status = $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
	`

	result, err := r.db.Exec(query, status, id)
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

func (r *shardRepository) Delete(id uuid.UUID) error {
	query := `DELETE FROM shards WHERE id = $1`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete shard: %w", err)
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
