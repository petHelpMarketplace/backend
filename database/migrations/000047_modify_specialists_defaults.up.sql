-- First, update existing null values to false
UPDATE specialists 
SET 
    is_banned = COALESCE(is_banned, false),
    is_deleted = COALESCE(is_deleted, false),
    is_active = COALESCE(is_active, true),
    is_verified = COALESCE(is_verified, false);

-- Then alter the columns to set NOT NULL constraint and default values
ALTER TABLE specialists
    ALTER COLUMN is_banned SET DEFAULT false,
    ALTER COLUMN is_banned SET NOT NULL,
    ALTER COLUMN is_deleted SET DEFAULT false,
    ALTER COLUMN is_deleted SET NOT NULL,
    ALTER COLUMN is_active SET DEFAULT true,
    ALTER COLUMN is_active SET NOT NULL,
    ALTER COLUMN is_verified SET DEFAULT false,
    ALTER COLUMN is_verified SET NOT NULL;