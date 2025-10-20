-- SPDX-License-Identifier: AGPL-3.0-or-later
-- Rollback: Remove 'queued' status from reminder_logs status constraint

-- Drop the constraint with 'queued'
ALTER TABLE reminder_logs DROP CONSTRAINT IF EXISTS reminder_logs_status_check;

-- Restore original constraint without 'queued'
ALTER TABLE reminder_logs ADD CONSTRAINT reminder_logs_status_check
    CHECK (status IN ('sent', 'failed', 'bounced'));
