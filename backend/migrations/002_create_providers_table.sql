-- Migration: Create providers table
-- Date: 2026-03-21
-- Priority: CRITICAL
-- Description: Centralizes provider configuration and metadata

BEGIN;

-- Create provider status enum
CREATE TYPE provider_status AS ENUM ('active', 'degraded', 'disabled', 'error');

-- Create providers table
CREATE TABLE providers (
    id TEXT PRIMARY KEY, -- e.g., 'google_drive', 'aws_s3'
    display_name TEXT NOT NULL,
    provider_type TEXT NOT NULL DEFAULT 'cloud_storage', -- 'oauth', 's3', 'webdav', etc.
    enabled BOOLEAN NOT NULL DEFAULT true,
    status provider_status NOT NULL DEFAULT 'active',

    -- Quota tracking
    quota_used_bytes BIGINT NOT NULL DEFAULT 0,
    quota_total_bytes BIGINT NOT NULL DEFAULT 0,

    -- Health monitoring
    latency_ms INT,
    last_health_check_at TIMESTAMP WITH TIME ZONE,
    error_count INT NOT NULL DEFAULT 0,

    -- Metadata
    region TEXT,
    capabilities JSONB DEFAULT '{}', -- {versioning: true, encryption: false, etc.}

    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Insert common providers (using camelCase to match existing shards data)
INSERT INTO providers (id, display_name, provider_type, quota_total_bytes, region) VALUES
    ('googleDrive', 'Google Drive', 'oauth', 16106127360, 'global'),
    ('awsS3', 'AWS S3', 's3', 5368709120, 'us-east-1'),
    ('oneDrive', 'Microsoft OneDrive', 'oauth', 5368709120, 'global'),
    ('dropbox', 'Dropbox', 'oauth', 2147483648, 'global');

-- Create indexes
CREATE INDEX idx_providers_status ON providers(status);
CREATE INDEX idx_providers_enabled ON providers(enabled);
CREATE INDEX idx_providers_health ON providers(last_health_check_at DESC);

-- Add trigger to update updated_at
CREATE TRIGGER update_providers_updated_at
    BEFORE UPDATE ON providers
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Add foreign key to shards table
ALTER TABLE shards
ADD CONSTRAINT shards_provider_fkey
    FOREIGN KEY (provider) REFERENCES providers(id);

-- Add user_id to provider_connections
ALTER TABLE provider_connections
ADD COLUMN user_id TEXT REFERENCES "user"(id) ON DELETE CASCADE,
ADD COLUMN created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP;

-- Add foreign key from provider_connections to providers
ALTER TABLE provider_connections
ADD CONSTRAINT provider_connections_provider_fkey
    FOREIGN KEY (provider_id) REFERENCES providers(id);

-- Create index
CREATE INDEX idx_provider_connections_user_id ON provider_connections(user_id);

COMMIT;

-- Verification queries (run after migration)
-- SELECT * FROM providers;
-- SELECT COUNT(*) FROM shards WHERE provider NOT IN (SELECT id FROM providers);
