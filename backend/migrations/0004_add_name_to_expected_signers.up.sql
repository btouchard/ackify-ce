-- Add name field to expected_signers table
-- This allows storing optional display names for expected readers
-- Supports formats like "Benjamin Touchard <benjamin@kolapsis.com>"
ALTER TABLE expected_signers ADD COLUMN name TEXT NOT NULL DEFAULT '';

COMMENT ON COLUMN expected_signers.name IS 'Optional display name for personalized reminder emails';
