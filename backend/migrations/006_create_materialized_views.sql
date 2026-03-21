-- Migration: Create materialized views for analytics
-- Date: 2026-03-21
-- Priority: LOW
-- Description: Optimizes dashboard queries with pre-computed statistics

BEGIN;

-- User storage statistics
CREATE MATERIALIZED VIEW user_storage_stats AS
SELECT
    u.id as user_id,
    u.email,
    u.name,
    COUNT(f.id) as total_files,
    COALESCE(SUM(f.original_size), 0) as total_bytes_stored,
    COUNT(CASE WHEN f.status = 'UPLOADED' THEN 1 END) as active_files,
    COUNT(CASE WHEN f.status = 'DEGRADED' THEN 1 END) as degraded_files,
    COUNT(CASE WHEN f.status = 'DELETED' THEN 1 END) as deleted_files,
    MAX(f.created_at) as last_upload_at,
    u."createdAt" as user_created_at
FROM "user" u
LEFT JOIN files f ON u.id = f.user_id
GROUP BY u.id, u.email, u.name, u."createdAt";

-- Create unique index for concurrent refresh
CREATE UNIQUE INDEX idx_user_storage_stats_user ON user_storage_stats(user_id);

-- Provider health statistics
CREATE MATERIALIZED VIEW provider_health_stats AS
SELECT
    p.id as provider_id,
    p.display_name,
    p.status,
    p.quota_used_bytes,
    p.quota_total_bytes,
    p.latency_ms,
    COUNT(s.id) as total_shards,
    COUNT(CASE WHEN s.status = 'HEALTHY' THEN 1 END) as healthy_shards,
    COUNT(CASE WHEN s.status = 'CORRUPTED' THEN 1 END) as corrupted_shards,
    COUNT(CASE WHEN s.status = 'MISSING' THEN 1 END) as missing_shards,
    COUNT(DISTINCT s.file_id) as files_using_provider,
    p.last_health_check_at,
    p.updated_at
FROM providers p
LEFT JOIN shards s ON p.id = s.provider
GROUP BY p.id, p.display_name, p.status, p.quota_used_bytes, p.quota_total_bytes,
         p.latency_ms, p.last_health_check_at, p.updated_at;

-- Create unique index
CREATE UNIQUE INDEX idx_provider_health_stats_provider ON provider_health_stats(provider_id);

-- System-wide statistics
CREATE MATERIALIZED VIEW system_stats AS
SELECT
    (SELECT COUNT(*) FROM "user") as total_users,
    (SELECT COUNT(*) FROM files WHERE status = 'UPLOADED') as active_files,
    (SELECT COALESCE(SUM(original_size), 0) FROM files WHERE status = 'UPLOADED') as total_bytes_stored,
    (SELECT COUNT(*) FROM shards WHERE status = 'HEALTHY') as healthy_shards,
    (SELECT COUNT(*) FROM shards WHERE status = 'CORRUPTED') as corrupted_shards,
    (SELECT COUNT(*) FROM providers WHERE enabled = true) as active_providers,
    CURRENT_TIMESTAMP as last_updated;

-- Create a function to refresh all materialized views
CREATE OR REPLACE FUNCTION refresh_all_stats()
RETURNS void AS $$
BEGIN
    REFRESH MATERIALIZED VIEW CONCURRENTLY user_storage_stats;
    REFRESH MATERIALIZED VIEW CONCURRENTLY provider_health_stats;
    REFRESH MATERIALIZED VIEW system_stats;
END;
$$ LANGUAGE plpgsql;

-- Create a scheduled job to refresh stats every 5 minutes
-- Note: This requires pg_cron extension (available on Supabase)
-- SELECT cron.schedule('refresh-stats', '*/5 * * * *', 'SELECT refresh_all_stats()');

COMMIT;

-- Manual refresh command (run after migration and periodically)
-- SELECT refresh_all_stats();

-- Verification queries
-- SELECT * FROM user_storage_stats ORDER BY total_bytes_stored DESC LIMIT 10;
-- SELECT * FROM provider_health_stats;
-- SELECT * FROM system_stats;
