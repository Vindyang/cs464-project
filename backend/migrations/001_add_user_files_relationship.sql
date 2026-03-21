-- Migration: Add user ownership to files
-- Date: 2026-03-21
-- Priority: CRITICAL
-- Description: Establishes relationship between users and their files for access control

BEGIN;

-- Step 1: Add user_id column as nullable first
ALTER TABLE files
ADD COLUMN user_id TEXT;

-- Step 2: Assign existing files to the first user in the database
-- (In production, you'd want a more sophisticated approach)
UPDATE files
SET user_id = (SELECT id FROM "user" LIMIT 1)
WHERE user_id IS NULL;

-- Step 3: Make user_id NOT NULL now that all rows have a value
ALTER TABLE files
ALTER COLUMN user_id SET NOT NULL;

-- Step 4: Add foreign key constraint
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
