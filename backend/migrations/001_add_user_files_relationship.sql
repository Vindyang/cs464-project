-- Migration: Add user ownership to files
-- Date: 2026-03-21
-- Priority: CRITICAL
-- Description: Establishes relationship between users and their files for access control

BEGIN;

-- Add user_id column to files table
-- Using 'system' as default for any existing rows
ALTER TABLE files
ADD COLUMN user_id TEXT NOT NULL DEFAULT 'system';

-- Add foreign key constraint
ALTER TABLE files
ADD CONSTRAINT files_user_fkey
    FOREIGN KEY (user_id) REFERENCES "user"(id) ON DELETE CASCADE;

-- Create indexes for performance
CREATE INDEX idx_files_user_id ON files(user_id);
CREATE INDEX idx_files_user_status ON files(user_id, status);
CREATE INDEX idx_files_created_at ON files(created_at DESC);

-- Add composite index for common queries (user's recent files)
CREATE INDEX idx_files_user_recent ON files(user_id, created_at DESC)
    INCLUDE (id, original_name, original_size, status);

COMMIT;

-- Verification query (run after migration)
-- SELECT user_id, COUNT(*) as file_count FROM files GROUP BY user_id;
