-- Shard Map Service Database Schema
-- PostgreSQL 14+

-- File status enum
CREATE TYPE file_status AS ENUM ('PENDING', 'UPLOADED', 'DEGRADED', 'DELETED');

-- Shard status enum
CREATE TYPE shard_status AS ENUM ('PENDING', 'HEALTHY', 'CORRUPTED', 'MISSING');

-- Shard type enum
CREATE TYPE shard_type AS ENUM ('DATA', 'PARITY');

-- Files table: Stores metadata for each uploaded file
CREATE TABLE files (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    original_name VARCHAR(255),
    original_size BIGINT NOT NULL,
    total_chunks INT NOT NULL,
    n INT NOT NULL,  -- Total shards per chunk
    k INT NOT NULL,  -- Data shards per chunk (minimum required for reconstruction)
    shard_size BIGINT NOT NULL,
    status file_status NOT NULL DEFAULT 'PENDING',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Shards table: Stores location and health of each shard
CREATE TABLE shards (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    file_id UUID NOT NULL REFERENCES files(id) ON DELETE CASCADE,
    chunk_index INT NOT NULL,  -- Which chunk this shard belongs to
    shard_index INT NOT NULL,  -- Position within the chunk (0 to N-1)
    shard_type shard_type NOT NULL,
    remote_id VARCHAR(512) NOT NULL,  -- Cloud provider's ID for this shard
    provider VARCHAR(50) NOT NULL,  -- e.g., 'aws_s3', 'google_drive', 'onedrive'
    checksum_sha256 CHAR(64) NOT NULL,  -- SHA-256 hash for corruption detection
    status shard_status NOT NULL DEFAULT 'PENDING',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    -- Ensure each shard position is unique per file and chunk
    UNIQUE(file_id, chunk_index, shard_index)
);

-- Indexes for performance optimization
CREATE INDEX idx_shards_file_id ON shards(file_id);
CREATE INDEX idx_shards_file_status ON shards(file_id, status);
CREATE INDEX idx_files_status ON files(status);
CREATE INDEX idx_shards_provider ON shards(provider);
CREATE INDEX idx_shards_remote_id ON shards(remote_id);

-- Trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_files_updated_at BEFORE UPDATE ON files
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_shards_updated_at BEFORE UPDATE ON shards
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- View for easy querying of file health
CREATE VIEW file_health_summary AS
SELECT
    f.id,
    f.original_name,
    f.status as file_status,
    COUNT(DISTINCT s.chunk_index) as total_chunks_stored,
    COUNT(s.id) as total_shards,
    COUNT(CASE WHEN s.status = 'HEALTHY' THEN 1 END) as healthy_shards,
    COUNT(CASE WHEN s.status = 'CORRUPTED' THEN 1 END) as corrupted_shards,
    COUNT(CASE WHEN s.status = 'MISSING' THEN 1 END) as missing_shards,
    f.k as min_shards_required_per_chunk,
    f.created_at,
    f.updated_at
FROM files f
LEFT JOIN shards s ON f.id = s.file_id
GROUP BY f.id, f.original_name, f.status, f.k, f.created_at, f.updated_at;
