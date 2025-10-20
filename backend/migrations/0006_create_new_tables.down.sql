-- SPDX-License-Identifier: AGPL-3.0-or-later

-- Drop email queue table and related functions
DROP TRIGGER IF EXISTS trigger_update_email_queue_retry ON email_queue;
DROP FUNCTION IF EXISTS update_email_queue_retry_time();
DROP FUNCTION IF EXISTS calculate_next_retry_time(INT);
DROP TABLE IF EXISTS email_queue;

-- Drop checksum_verifications indexes
DROP INDEX IF EXISTS idx_checksum_verifications_doc_id_verified_at;
DROP INDEX IF EXISTS idx_checksum_verifications_verified_at;
DROP INDEX IF EXISTS idx_checksum_verifications_doc_id;

-- Drop the checksum_verifications table
DROP TABLE IF EXISTS checksum_verifications;
