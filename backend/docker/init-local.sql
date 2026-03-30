CREATE EXTENSION IF NOT EXISTS pgcrypto;

DO $$ BEGIN
    CREATE TYPE file_status AS ENUM ('PENDING', 'UPLOADED', 'DEGRADED', 'DELETED');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
    CREATE TYPE shard_status AS ENUM ('PENDING', 'HEALTHY', 'CORRUPTED', 'MISSING');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
    CREATE TYPE shard_type AS ENUM ('DATA', 'PARITY');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

CREATE TABLE IF NOT EXISTS files (
    id UUID PRIMARY KEY,
    user_id TEXT,
    original_name VARCHAR(255),
    original_size BIGINT NOT NULL,
    total_chunks INT NOT NULL,
    n INT NOT NULL,
    k INT NOT NULL,
    shard_size BIGINT NOT NULL,
    status file_status NOT NULL DEFAULT 'PENDING',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS providers (
    id TEXT PRIMARY KEY,
    display_name TEXT NOT NULL,
    provider_type TEXT NOT NULL DEFAULT 'cloud_storage',
    enabled BOOLEAN NOT NULL DEFAULT true,
    status TEXT NOT NULL DEFAULT 'active',
    quota_used_bytes BIGINT NOT NULL DEFAULT 0,
    quota_total_bytes BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS shards (
    id UUID PRIMARY KEY,
    file_id UUID NOT NULL REFERENCES files(id) ON DELETE CASCADE,
    chunk_index INT NOT NULL,
    shard_index INT NOT NULL,
    shard_type shard_type NOT NULL,
    remote_id VARCHAR(255) NOT NULL,
    provider VARCHAR(255) NOT NULL REFERENCES providers(id),
    checksum_sha256 CHAR(64) NOT NULL,
    status shard_status NOT NULL DEFAULT 'PENDING',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(file_id, chunk_index, shard_index)
);

CREATE TABLE IF NOT EXISTS provider_connections (
    provider_id TEXT PRIMARY KEY REFERENCES providers(id),
    access_token TEXT NOT NULL,
    refresh_token TEXT,
    token_type TEXT,
    expiry TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO providers (id, display_name, provider_type, quota_total_bytes)
VALUES
    ('googleDrive', 'Google Drive', 'oauth', 16106127360),
    ('awsS3', 'AWS S3', 's3', 5368709120),
    ('oneDrive', 'Microsoft OneDrive', 'oauth', 5368709120),
    ('dropbox', 'Dropbox', 'oauth', 2147483648)
ON CONFLICT (id) DO NOTHING;

CREATE INDEX IF NOT EXISTS idx_files_created_at ON files(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_files_user_id ON files(user_id);
CREATE INDEX IF NOT EXISTS idx_shards_file_id ON shards(file_id);
CREATE INDEX IF NOT EXISTS idx_shards_file_status ON shards(file_id, status);
CREATE INDEX IF NOT EXISTS idx_provider_connections_provider_id ON provider_connections(provider_id);
