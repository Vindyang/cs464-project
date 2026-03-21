-- Migration: Create audit logs table
-- Date: 2026-03-21
-- Priority: MODERATE
-- Description: Enables comprehensive audit trail for security and compliance

BEGIN;

-- Create audit action enum
CREATE TYPE audit_action AS ENUM (
    'file_upload',
    'file_download',
    'file_delete',
    'file_share',
    'file_unshare',
    'shard_upload',
    'shard_corruption',
    'shard_recovery',
    'provider_connect',
    'provider_disconnect',
    'provider_failure',
    'user_login',
    'user_logout',
    'permission_grant',
    'permission_revoke'
);

-- Create audit logs table (simple, non-partitioned)
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id TEXT REFERENCES "user"(id) ON DELETE SET NULL, -- NULL for system events
    action audit_action NOT NULL,
    resource_type TEXT NOT NULL, -- 'file', 'shard', 'provider', 'user'
    resource_id TEXT, -- UUID or ID of the resource
    metadata JSONB DEFAULT '{}', -- Flexible field for action-specific data
    ip_address INET,
    user_agent TEXT,
    success BOOLEAN NOT NULL DEFAULT true,
    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for common queries
CREATE INDEX idx_audit_logs_user ON audit_logs(user_id, created_at DESC);
CREATE INDEX idx_audit_logs_resource ON audit_logs(resource_type, resource_id, created_at DESC);
CREATE INDEX idx_audit_logs_action ON audit_logs(action, created_at DESC);
CREATE INDEX idx_audit_logs_created ON audit_logs(created_at DESC);
CREATE INDEX idx_audit_logs_failures ON audit_logs(success, created_at DESC) WHERE success = false;

-- Create a function to log file operations automatically
CREATE OR REPLACE FUNCTION log_file_operation()
RETURNS TRIGGER AS $$
BEGIN
    IF (TG_OP = 'DELETE') THEN
        INSERT INTO audit_logs (user_id, action, resource_type, resource_id, metadata)
        VALUES (OLD.user_id, 'file_delete', 'file', OLD.id::text,
                jsonb_build_object('filename', OLD.original_name, 'size', OLD.original_size));
        RETURN OLD;
    ELSIF (TG_OP = 'INSERT') THEN
        INSERT INTO audit_logs (user_id, action, resource_type, resource_id, metadata)
        VALUES (NEW.user_id, 'file_upload', 'file', NEW.id::text,
                jsonb_build_object('filename', NEW.original_name, 'size', NEW.original_size));
        RETURN NEW;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- Create triggers for automatic logging
CREATE TRIGGER log_file_insert
    AFTER INSERT ON files
    FOR EACH ROW
    EXECUTE FUNCTION log_file_operation();

CREATE TRIGGER log_file_delete
    AFTER DELETE ON files
    FOR EACH ROW
    EXECUTE FUNCTION log_file_operation();

COMMIT;

-- Verification query (run after migration)
-- SELECT action, COUNT(*) FROM audit_logs GROUP BY action ORDER BY COUNT(*) DESC;
