-- SPDX-License-Identifier: AGPL-3.0-or-later
-- Migration: Add 'queued' status to reminder_logs status constraint
-- This allows tracking when emails are queued for async processing

-- Drop the existing constraint
ALTER TABLE reminder_logs DROP CONSTRAINT IF EXISTS reminder_logs_status_check;

-- Add new constraint with 'queued' status
ALTER TABLE reminder_logs ADD CONSTRAINT reminder_logs_status_check
    CHECK (status IN ('sent', 'failed', 'bounced', 'queued'));
