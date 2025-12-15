-- SPDX-License-Identifier: AGPL-3.0-or-later

-- ============================================================================
-- Migration: Add Row Level Security (RLS) Policies
-- ============================================================================
-- This migration enables PostgreSQL Row Level Security for tenant isolation.
-- It ensures that all queries are automatically filtered by tenant_id,
-- eliminating the risk of data leakage if application code forgets the filter.
--
-- Prerequisites:
--   - Migration 0015 must have run (tenant_id columns exist)
--   - A non-superuser role 'ackify_app' should be used for runtime queries
--
-- How it works:
--   1. current_tenant_id() reads 'app.tenant_id' from session config
--   2. RLS policies filter rows where tenant_id = current_tenant_id()
--   3. FORCE ROW LEVEL SECURITY ensures policies apply even to table owners
--   4. Application sets app.tenant_id via: SELECT set_config('app.tenant_id', $1, true)
-- ============================================================================

-- Create helper function to get current tenant from session
-- The function returns NULL if app.tenant_id is not set, which means
-- RLS policies will filter out ALL rows (secure by default).
CREATE OR REPLACE FUNCTION current_tenant_id() RETURNS UUID AS $$
DECLARE
    tenant_id_str TEXT;
BEGIN
    tenant_id_str := current_setting('app.tenant_id', true);
    IF tenant_id_str IS NULL OR tenant_id_str = '' THEN
        RETURN NULL;
    END IF;
    RETURN tenant_id_str::UUID;
EXCEPTION WHEN OTHERS THEN
    -- Invalid UUID format - return NULL for safety
    RAISE WARNING 'current_tenant_id(): Invalid UUID format: %', tenant_id_str;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql STABLE;

COMMENT ON FUNCTION current_tenant_id() IS 'Returns the current tenant UUID from session config (app.tenant_id). Returns NULL if not set.';

-- IMPORTANT: The ackify_app role is created by the migrate tool before running migrations.
-- Set ACKIFY_APP_PASSWORD environment variable to enable RLS support.
-- The migrate tool will create the role with the specified password.

-- ============================================================================
-- Enable RLS and create policies for each tenant-aware table
-- ============================================================================

-- ----- DOCUMENTS -----
ALTER TABLE documents ENABLE ROW LEVEL SECURITY;
ALTER TABLE documents FORCE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS tenant_isolation_documents ON documents;
CREATE POLICY tenant_isolation_documents ON documents
    USING (tenant_id = current_tenant_id())
    WITH CHECK (tenant_id = current_tenant_id());

GRANT SELECT, INSERT, UPDATE, DELETE ON documents TO ackify_app;

-- ----- SIGNATURES -----
ALTER TABLE signatures ENABLE ROW LEVEL SECURITY;
ALTER TABLE signatures FORCE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS tenant_isolation_signatures ON signatures;
CREATE POLICY tenant_isolation_signatures ON signatures
    USING (tenant_id = current_tenant_id())
    WITH CHECK (tenant_id = current_tenant_id());

GRANT SELECT, INSERT, UPDATE, DELETE ON signatures TO ackify_app;
GRANT USAGE, SELECT ON SEQUENCE signatures_id_seq TO ackify_app;

-- ----- EXPECTED_SIGNERS -----
ALTER TABLE expected_signers ENABLE ROW LEVEL SECURITY;
ALTER TABLE expected_signers FORCE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS tenant_isolation_expected_signers ON expected_signers;
CREATE POLICY tenant_isolation_expected_signers ON expected_signers
    USING (tenant_id = current_tenant_id())
    WITH CHECK (tenant_id = current_tenant_id());

GRANT SELECT, INSERT, UPDATE, DELETE ON expected_signers TO ackify_app;
GRANT USAGE, SELECT ON SEQUENCE expected_signers_id_seq TO ackify_app;

-- ----- WEBHOOKS -----
ALTER TABLE webhooks ENABLE ROW LEVEL SECURITY;
ALTER TABLE webhooks FORCE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS tenant_isolation_webhooks ON webhooks;
CREATE POLICY tenant_isolation_webhooks ON webhooks
    USING (tenant_id = current_tenant_id())
    WITH CHECK (tenant_id = current_tenant_id());

GRANT SELECT, INSERT, UPDATE, DELETE ON webhooks TO ackify_app;
GRANT USAGE, SELECT ON SEQUENCE webhooks_id_seq TO ackify_app;

-- ----- REMINDER_LOGS -----
ALTER TABLE reminder_logs ENABLE ROW LEVEL SECURITY;
ALTER TABLE reminder_logs FORCE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS tenant_isolation_reminder_logs ON reminder_logs;
CREATE POLICY tenant_isolation_reminder_logs ON reminder_logs
    USING (tenant_id = current_tenant_id())
    WITH CHECK (tenant_id = current_tenant_id());

GRANT SELECT, INSERT, UPDATE, DELETE ON reminder_logs TO ackify_app;
GRANT USAGE, SELECT ON SEQUENCE reminder_logs_id_seq TO ackify_app;

-- ----- EMAIL_QUEUE -----
ALTER TABLE email_queue ENABLE ROW LEVEL SECURITY;
ALTER TABLE email_queue FORCE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS tenant_isolation_email_queue ON email_queue;
CREATE POLICY tenant_isolation_email_queue ON email_queue
    USING (tenant_id = current_tenant_id())
    WITH CHECK (tenant_id = current_tenant_id());

GRANT SELECT, INSERT, UPDATE, DELETE ON email_queue TO ackify_app;
GRANT USAGE, SELECT ON SEQUENCE email_queue_id_seq TO ackify_app;

-- ----- CHECKSUM_VERIFICATIONS -----
ALTER TABLE checksum_verifications ENABLE ROW LEVEL SECURITY;
ALTER TABLE checksum_verifications FORCE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS tenant_isolation_checksum_verifications ON checksum_verifications;
CREATE POLICY tenant_isolation_checksum_verifications ON checksum_verifications
    USING (tenant_id = current_tenant_id())
    WITH CHECK (tenant_id = current_tenant_id());

GRANT SELECT, INSERT, UPDATE, DELETE ON checksum_verifications TO ackify_app;
GRANT USAGE, SELECT ON SEQUENCE checksum_verifications_id_seq TO ackify_app;

-- ----- WEBHOOK_DELIVERIES -----
ALTER TABLE webhook_deliveries ENABLE ROW LEVEL SECURITY;
ALTER TABLE webhook_deliveries FORCE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS tenant_isolation_webhook_deliveries ON webhook_deliveries;
CREATE POLICY tenant_isolation_webhook_deliveries ON webhook_deliveries
    USING (tenant_id = current_tenant_id())
    WITH CHECK (tenant_id = current_tenant_id());

GRANT SELECT, INSERT, UPDATE, DELETE ON webhook_deliveries TO ackify_app;
GRANT USAGE, SELECT ON SEQUENCE webhook_deliveries_id_seq TO ackify_app;

-- ----- OAUTH_SESSIONS -----
ALTER TABLE oauth_sessions ENABLE ROW LEVEL SECURITY;
ALTER TABLE oauth_sessions FORCE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS tenant_isolation_oauth_sessions ON oauth_sessions;
CREATE POLICY tenant_isolation_oauth_sessions ON oauth_sessions
    USING (tenant_id = current_tenant_id())
    WITH CHECK (tenant_id = current_tenant_id());

GRANT SELECT, INSERT, UPDATE, DELETE ON oauth_sessions TO ackify_app;
GRANT USAGE, SELECT ON SEQUENCE oauth_sessions_id_seq TO ackify_app;

-- ----- MAGIC_LINK_TOKENS -----
-- Note: Magic link tokens may have NULL tenant_id for login requests
-- Policy allows NULL tenant_id OR matching tenant_id
ALTER TABLE magic_link_tokens ENABLE ROW LEVEL SECURITY;
ALTER TABLE magic_link_tokens FORCE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS tenant_isolation_magic_link_tokens ON magic_link_tokens;
CREATE POLICY tenant_isolation_magic_link_tokens ON magic_link_tokens
    USING (tenant_id IS NULL OR tenant_id = current_tenant_id())
    WITH CHECK (tenant_id IS NULL OR tenant_id = current_tenant_id());

GRANT SELECT, INSERT, UPDATE, DELETE ON magic_link_tokens TO ackify_app;
GRANT USAGE, SELECT ON SEQUENCE magic_link_tokens_id_seq TO ackify_app;

-- ----- MAGIC_LINK_AUTH_ATTEMPTS -----
-- Note: Auth attempts may have NULL tenant_id before authentication
ALTER TABLE magic_link_auth_attempts ENABLE ROW LEVEL SECURITY;
ALTER TABLE magic_link_auth_attempts FORCE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS tenant_isolation_magic_link_auth_attempts ON magic_link_auth_attempts;
CREATE POLICY tenant_isolation_magic_link_auth_attempts ON magic_link_auth_attempts
    USING (tenant_id IS NULL OR tenant_id = current_tenant_id())
    WITH CHECK (tenant_id IS NULL OR tenant_id = current_tenant_id());

GRANT SELECT, INSERT, UPDATE, DELETE ON magic_link_auth_attempts TO ackify_app;
GRANT USAGE, SELECT ON SEQUENCE magic_link_auth_attempts_id_seq TO ackify_app;

-- ----- INSTANCE_METADATA -----
-- This table is read-only for the app (tenant ID source)
-- No RLS needed as it contains only one row per instance
GRANT SELECT ON instance_metadata TO ackify_app;

-- ============================================================================
-- Set default privileges for future tables
-- ============================================================================
ALTER DEFAULT PRIVILEGES IN SCHEMA public
    GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO ackify_app;

ALTER DEFAULT PRIVILEGES IN SCHEMA public
    GRANT USAGE, SELECT ON SEQUENCES TO ackify_app;
