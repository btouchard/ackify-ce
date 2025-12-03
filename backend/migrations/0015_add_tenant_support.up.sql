-- SPDX-License-Identifier: AGPL-3.0-or-later

-- ============================================================================
-- Migration: Add Tenant Support (tenant-ready)
-- ============================================================================
-- This migration prepares Ackify CE for multi-tenancy by:
-- 1. Creating an instance_metadata table with a unique tenant UUID
-- 2. Adding tenant_id columns to all business and auth tables
-- 3. Backfilling existing data with the instance tenant UUID
-- 4. Adding immutability triggers to prevent tenant_id modification
-- ============================================================================

-- Step 1: Create instance_metadata table to store the unique instance tenant UUID
CREATE TABLE IF NOT EXISTS instance_metadata (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE instance_metadata IS 'Stores the unique tenant UUID for this Ackify instance (one row per instance)';
COMMENT ON COLUMN instance_metadata.id IS 'The unique tenant identifier for this instance';

-- Ensure exactly one row exists (the instance tenant)
INSERT INTO instance_metadata DEFAULT VALUES
ON CONFLICT DO NOTHING;

-- Step 2: Add nullable tenant_id columns to all tables
-- NOTE: We use UUID type to match instance_metadata.id

-- Business tables
ALTER TABLE documents ADD COLUMN IF NOT EXISTS tenant_id UUID;
ALTER TABLE signatures ADD COLUMN IF NOT EXISTS tenant_id UUID;
ALTER TABLE expected_signers ADD COLUMN IF NOT EXISTS tenant_id UUID;
ALTER TABLE webhooks ADD COLUMN IF NOT EXISTS tenant_id UUID;
ALTER TABLE reminder_logs ADD COLUMN IF NOT EXISTS tenant_id UUID;
ALTER TABLE email_queue ADD COLUMN IF NOT EXISTS tenant_id UUID;
ALTER TABLE checksum_verifications ADD COLUMN IF NOT EXISTS tenant_id UUID;
ALTER TABLE webhook_deliveries ADD COLUMN IF NOT EXISTS tenant_id UUID;

-- Authentication tables (tenant_id may be NULL for pre-auth operations)
ALTER TABLE oauth_sessions ADD COLUMN IF NOT EXISTS tenant_id UUID;
ALTER TABLE magic_link_tokens ADD COLUMN IF NOT EXISTS tenant_id UUID;
ALTER TABLE magic_link_auth_attempts ADD COLUMN IF NOT EXISTS tenant_id UUID;

-- Step 3: Backfill all existing rows with the instance tenant UUID
-- This ensures all existing data belongs to the single instance tenant

UPDATE documents SET tenant_id = (SELECT id FROM instance_metadata LIMIT 1) WHERE tenant_id IS NULL;
UPDATE signatures SET tenant_id = (SELECT id FROM instance_metadata LIMIT 1) WHERE tenant_id IS NULL;
UPDATE expected_signers SET tenant_id = (SELECT id FROM instance_metadata LIMIT 1) WHERE tenant_id IS NULL;
UPDATE webhooks SET tenant_id = (SELECT id FROM instance_metadata LIMIT 1) WHERE tenant_id IS NULL;
UPDATE reminder_logs SET tenant_id = (SELECT id FROM instance_metadata LIMIT 1) WHERE tenant_id IS NULL;
UPDATE email_queue SET tenant_id = (SELECT id FROM instance_metadata LIMIT 1) WHERE tenant_id IS NULL;
UPDATE checksum_verifications SET tenant_id = (SELECT id FROM instance_metadata LIMIT 1) WHERE tenant_id IS NULL;
UPDATE webhook_deliveries SET tenant_id = (SELECT id FROM instance_metadata LIMIT 1) WHERE tenant_id IS NULL;
UPDATE oauth_sessions SET tenant_id = (SELECT id FROM instance_metadata LIMIT 1) WHERE tenant_id IS NULL;
UPDATE magic_link_tokens SET tenant_id = (SELECT id FROM instance_metadata LIMIT 1) WHERE tenant_id IS NULL;
UPDATE magic_link_auth_attempts SET tenant_id = (SELECT id FROM instance_metadata LIMIT 1) WHERE tenant_id IS NULL;

-- Step 4: Add indexes for tenant_id columns
CREATE INDEX IF NOT EXISTS idx_documents_tenant_id ON documents(tenant_id);
CREATE INDEX IF NOT EXISTS idx_signatures_tenant_id ON signatures(tenant_id);
CREATE INDEX IF NOT EXISTS idx_expected_signers_tenant_id ON expected_signers(tenant_id);
CREATE INDEX IF NOT EXISTS idx_webhooks_tenant_id ON webhooks(tenant_id);
CREATE INDEX IF NOT EXISTS idx_reminder_logs_tenant_id ON reminder_logs(tenant_id);
CREATE INDEX IF NOT EXISTS idx_email_queue_tenant_id ON email_queue(tenant_id);
CREATE INDEX IF NOT EXISTS idx_checksum_verifications_tenant_id ON checksum_verifications(tenant_id);
CREATE INDEX IF NOT EXISTS idx_webhook_deliveries_tenant_id ON webhook_deliveries(tenant_id);
CREATE INDEX IF NOT EXISTS idx_oauth_sessions_tenant_id ON oauth_sessions(tenant_id);
CREATE INDEX IF NOT EXISTS idx_magic_link_tokens_tenant_id ON magic_link_tokens(tenant_id);
CREATE INDEX IF NOT EXISTS idx_magic_link_auth_attempts_tenant_id ON magic_link_auth_attempts(tenant_id);

-- Step 5: Add comments explaining the tenant_id columns
COMMENT ON COLUMN documents.tenant_id IS 'Tenant identifier (references instance_metadata.id in CE mode)';
COMMENT ON COLUMN signatures.tenant_id IS 'Tenant identifier (references instance_metadata.id in CE mode)';
COMMENT ON COLUMN expected_signers.tenant_id IS 'Tenant identifier (references instance_metadata.id in CE mode)';
COMMENT ON COLUMN webhooks.tenant_id IS 'Tenant identifier (references instance_metadata.id in CE mode)';
COMMENT ON COLUMN reminder_logs.tenant_id IS 'Tenant identifier (references instance_metadata.id in CE mode)';
COMMENT ON COLUMN email_queue.tenant_id IS 'Tenant identifier (references instance_metadata.id in CE mode)';
COMMENT ON COLUMN checksum_verifications.tenant_id IS 'Tenant identifier (references instance_metadata.id in CE mode)';
COMMENT ON COLUMN webhook_deliveries.tenant_id IS 'Tenant identifier (references instance_metadata.id in CE mode)';
COMMENT ON COLUMN oauth_sessions.tenant_id IS 'Tenant identifier (NOT NULL after auth)';
COMMENT ON COLUMN magic_link_tokens.tenant_id IS 'Tenant identifier (NULL for login requests, set for admin reminders)';
COMMENT ON COLUMN magic_link_auth_attempts.tenant_id IS 'Tenant identifier (may be NULL before authentication)';

-- Step 6: Create trigger to prevent tenant_id modification after creation (immutability)
-- This ensures data cannot be moved between tenants once created
CREATE OR REPLACE FUNCTION prevent_tenant_id_modification()
RETURNS TRIGGER AS $$
BEGIN
    IF OLD.tenant_id IS NOT NULL AND NEW.tenant_id IS DISTINCT FROM OLD.tenant_id THEN
        RAISE EXCEPTION 'tenant_id cannot be modified after creation';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION prevent_tenant_id_modification() IS 'Prevents modification of tenant_id column after initial assignment';

-- Apply trigger to all tables with tenant_id
CREATE TRIGGER tr_documents_tenant_id_immutable
    BEFORE UPDATE ON documents FOR EACH ROW EXECUTE FUNCTION prevent_tenant_id_modification();

CREATE TRIGGER tr_signatures_tenant_id_immutable
    BEFORE UPDATE ON signatures FOR EACH ROW EXECUTE FUNCTION prevent_tenant_id_modification();

CREATE TRIGGER tr_expected_signers_tenant_id_immutable
    BEFORE UPDATE ON expected_signers FOR EACH ROW EXECUTE FUNCTION prevent_tenant_id_modification();

CREATE TRIGGER tr_webhooks_tenant_id_immutable
    BEFORE UPDATE ON webhooks FOR EACH ROW EXECUTE FUNCTION prevent_tenant_id_modification();

CREATE TRIGGER tr_reminder_logs_tenant_id_immutable
    BEFORE UPDATE ON reminder_logs FOR EACH ROW EXECUTE FUNCTION prevent_tenant_id_modification();

CREATE TRIGGER tr_email_queue_tenant_id_immutable
    BEFORE UPDATE ON email_queue FOR EACH ROW EXECUTE FUNCTION prevent_tenant_id_modification();

CREATE TRIGGER tr_checksum_verifications_tenant_id_immutable
    BEFORE UPDATE ON checksum_verifications FOR EACH ROW EXECUTE FUNCTION prevent_tenant_id_modification();

CREATE TRIGGER tr_webhook_deliveries_tenant_id_immutable
    BEFORE UPDATE ON webhook_deliveries FOR EACH ROW EXECUTE FUNCTION prevent_tenant_id_modification();

CREATE TRIGGER tr_oauth_sessions_tenant_id_immutable
    BEFORE UPDATE ON oauth_sessions FOR EACH ROW EXECUTE FUNCTION prevent_tenant_id_modification();

CREATE TRIGGER tr_magic_link_tokens_tenant_id_immutable
    BEFORE UPDATE ON magic_link_tokens FOR EACH ROW EXECUTE FUNCTION prevent_tenant_id_modification();

CREATE TRIGGER tr_magic_link_auth_attempts_tenant_id_immutable
    BEFORE UPDATE ON magic_link_auth_attempts FOR EACH ROW EXECUTE FUNCTION prevent_tenant_id_modification();

-- Step 7: Protect instance_metadata.id from modifications
CREATE OR REPLACE FUNCTION prevent_instance_metadata_modification()
RETURNS TRIGGER AS $$
BEGIN
    IF OLD.id IS NOT NULL AND NEW.id IS DISTINCT FROM OLD.id THEN
        RAISE EXCEPTION 'instance_metadata.id cannot be modified after creation';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER tr_instance_metadata_id_immutable
    BEFORE UPDATE ON instance_metadata FOR EACH ROW EXECUTE FUNCTION prevent_instance_metadata_modification();
