-- Migration: Create file permissions table
-- Date: 2026-03-21
-- Priority: MODERATE
-- Description: Enables file sharing and granular access control

BEGIN;

-- Create permission level enum
CREATE TYPE permission_level AS ENUM ('viewer', 'editor', 'owner');

-- Create file permissions table
CREATE TABLE file_permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    file_id UUID NOT NULL REFERENCES files(id) ON DELETE CASCADE,
    user_id TEXT NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    permission permission_level NOT NULL DEFAULT 'viewer',
    granted_by TEXT REFERENCES "user"(id) ON DELETE SET NULL,
    granted_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP WITH TIME ZONE,

    -- Ensure each user has at most one permission per file
    UNIQUE(file_id, user_id)
);

-- Create indexes
CREATE INDEX idx_file_permissions_user ON file_permissions(user_id);
CREATE INDEX idx_file_permissions_file ON file_permissions(file_id);
CREATE INDEX idx_file_permissions_expires ON file_permissions(expires_at)
    WHERE expires_at IS NOT NULL;

-- Add check constraint
ALTER TABLE file_permissions
ADD CONSTRAINT check_valid_expiry
    CHECK (expires_at IS NULL OR expires_at > granted_at);

-- Create a function to automatically grant owner permission when file is created
CREATE OR REPLACE FUNCTION grant_owner_permission()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO file_permissions (file_id, user_id, permission, granted_by)
    VALUES (NEW.id, NEW.user_id, 'owner', NEW.user_id);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger to auto-grant owner permission
CREATE TRIGGER auto_grant_owner_permission
    AFTER INSERT ON files
    FOR EACH ROW
    EXECUTE FUNCTION grant_owner_permission();

COMMIT;

-- Verification query (run after migration)
-- SELECT f.original_name, u.email, fp.permission
-- FROM file_permissions fp
-- JOIN files f ON fp.file_id = f.id
-- JOIN "user" u ON fp.user_id = u.id;
