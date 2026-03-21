-- Migration: Add data integrity constraints
-- Date: 2026-03-21
-- Priority: MODERATE
-- Description: Adds validation constraints to prevent invalid data

BEGIN;

-- Files table constraints
ALTER TABLE files
ADD CONSTRAINT check_k_less_than_n CHECK (k <= n),
ADD CONSTRAINT check_k_positive CHECK (k > 0),
ADD CONSTRAINT check_n_positive CHECK (n > 0),
ADD CONSTRAINT check_positive_size CHECK (original_size > 0),
ADD CONSTRAINT check_positive_shard_size CHECK (shard_size > 0),
ADD CONSTRAINT check_positive_chunks CHECK (total_chunks > 0);

-- Shards table constraints
ALTER TABLE shards
ADD CONSTRAINT check_valid_chunk CHECK (chunk_index >= 0),
ADD CONSTRAINT check_valid_shard CHECK (shard_index >= 0),
ADD CONSTRAINT check_non_empty_remote_id CHECK (length(remote_id) > 0),
ADD CONSTRAINT check_non_empty_provider CHECK (length(provider) > 0);

-- Provider connections constraints
ALTER TABLE provider_connections
ADD CONSTRAINT check_valid_expiry
    CHECK (expiry IS NULL OR expiry > created_at);

-- Providers table constraints
ALTER TABLE providers
ADD CONSTRAINT check_positive_quota CHECK (quota_total_bytes >= 0),
ADD CONSTRAINT check_used_within_total CHECK (quota_used_bytes <= quota_total_bytes),
ADD CONSTRAINT check_positive_latency CHECK (latency_ms IS NULL OR latency_ms >= 0),
ADD CONSTRAINT check_non_negative_errors CHECK (error_count >= 0);

COMMIT;

-- Verification: Try to insert invalid data (should fail)
-- INSERT INTO files (original_size, total_chunks, n, k, shard_size, user_id)
-- VALUES (-100, 1, 3, 5, 1024, 'test'); -- Should fail: k > n
