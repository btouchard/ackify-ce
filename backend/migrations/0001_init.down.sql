-- Drop trigger and function
DROP TRIGGER IF EXISTS trigger_prevent_created_at_update ON signatures;
DROP FUNCTION IF EXISTS prevent_created_at_update();

-- Drop indexes
DROP INDEX IF EXISTS idx_signatures_user;

-- Drop signatures table
DROP TABLE IF EXISTS signatures;