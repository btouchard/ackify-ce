-- SPDX-License-Identifier: AGPL-3.0-or-later

-- Rollback: Remove Tenant Support
-- This migration reverts the tenant support changes

-- Step 1: Drop triggers for tenant_id immutability
DROP TRIGGER IF EXISTS tr_documents_tenant_id_immutable ON documents;
DROP TRIGGER IF EXISTS tr_signatures_tenant_id_immutable ON signatures;
DROP TRIGGER IF EXISTS tr_expected_signers_tenant_id_immutable ON expected_signers;
DROP TRIGGER IF EXISTS tr_webhooks_tenant_id_immutable ON webhooks;
DROP TRIGGER IF EXISTS tr_reminder_logs_tenant_id_immutable ON reminder_logs;
DROP TRIGGER IF EXISTS tr_email_queue_tenant_id_immutable ON email_queue;
DROP TRIGGER IF EXISTS tr_checksum_verifications_tenant_id_immutable ON checksum_verifications;
DROP TRIGGER IF EXISTS tr_webhook_deliveries_tenant_id_immutable ON webhook_deliveries;
DROP TRIGGER IF EXISTS tr_instance_metadata_id_immutable ON instance_metadata;

-- Step 2: Drop the trigger functions
DROP FUNCTION IF EXISTS prevent_tenant_id_modification();
DROP FUNCTION IF EXISTS prevent_instance_metadata_modification();

-- Step 3: Drop indexes
DROP INDEX IF EXISTS idx_documents_tenant_id;
DROP INDEX IF EXISTS idx_signatures_tenant_id;
DROP INDEX IF EXISTS idx_expected_signers_tenant_id;
DROP INDEX IF EXISTS idx_webhooks_tenant_id;
DROP INDEX IF EXISTS idx_reminder_logs_tenant_id;
DROP INDEX IF EXISTS idx_email_queue_tenant_id;
DROP INDEX IF EXISTS idx_checksum_verifications_tenant_id;
DROP INDEX IF EXISTS idx_webhook_deliveries_tenant_id;

-- Step 4: Drop tenant_id columns from all tables
ALTER TABLE documents DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE signatures DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE expected_signers DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE webhooks DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE reminder_logs DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE email_queue DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE checksum_verifications DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE webhook_deliveries DROP COLUMN IF EXISTS tenant_id;

-- Step 5: Drop instance_metadata table
DROP TABLE IF EXISTS instance_metadata;
