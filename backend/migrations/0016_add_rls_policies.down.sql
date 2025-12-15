-- SPDX-License-Identifier: AGPL-3.0-or-later

-- ============================================================================
-- Migration Rollback: Remove Row Level Security (RLS) Policies
-- ============================================================================
-- This rollback disables RLS and removes all tenant isolation policies.
-- WARNING: After this rollback, tenant isolation relies solely on application
-- code (WHERE tenant_id = ...). Use with caution in production.
-- ============================================================================

-- Step 1: Drop policies and disable RLS on all tables

-- ----- DOCUMENTS -----
DROP POLICY IF EXISTS tenant_isolation_documents ON documents;
ALTER TABLE documents DISABLE ROW LEVEL SECURITY;

-- ----- SIGNATURES -----
DROP POLICY IF EXISTS tenant_isolation_signatures ON signatures;
ALTER TABLE signatures DISABLE ROW LEVEL SECURITY;

-- ----- EXPECTED_SIGNERS -----
DROP POLICY IF EXISTS tenant_isolation_expected_signers ON expected_signers;
ALTER TABLE expected_signers DISABLE ROW LEVEL SECURITY;

-- ----- WEBHOOKS -----
DROP POLICY IF EXISTS tenant_isolation_webhooks ON webhooks;
ALTER TABLE webhooks DISABLE ROW LEVEL SECURITY;

-- ----- REMINDER_LOGS -----
DROP POLICY IF EXISTS tenant_isolation_reminder_logs ON reminder_logs;
ALTER TABLE reminder_logs DISABLE ROW LEVEL SECURITY;

-- ----- EMAIL_QUEUE -----
DROP POLICY IF EXISTS tenant_isolation_email_queue ON email_queue;
ALTER TABLE email_queue DISABLE ROW LEVEL SECURITY;

-- ----- CHECKSUM_VERIFICATIONS -----
DROP POLICY IF EXISTS tenant_isolation_checksum_verifications ON checksum_verifications;
ALTER TABLE checksum_verifications DISABLE ROW LEVEL SECURITY;

-- ----- WEBHOOK_DELIVERIES -----
DROP POLICY IF EXISTS tenant_isolation_webhook_deliveries ON webhook_deliveries;
ALTER TABLE webhook_deliveries DISABLE ROW LEVEL SECURITY;

-- ----- OAUTH_SESSIONS -----
DROP POLICY IF EXISTS tenant_isolation_oauth_sessions ON oauth_sessions;
ALTER TABLE oauth_sessions DISABLE ROW LEVEL SECURITY;

-- ----- MAGIC_LINK_TOKENS -----
DROP POLICY IF EXISTS tenant_isolation_magic_link_tokens ON magic_link_tokens;
ALTER TABLE magic_link_tokens DISABLE ROW LEVEL SECURITY;

-- ----- MAGIC_LINK_AUTH_ATTEMPTS -----
DROP POLICY IF EXISTS tenant_isolation_magic_link_auth_attempts ON magic_link_auth_attempts;
ALTER TABLE magic_link_auth_attempts DISABLE ROW LEVEL SECURITY;

-- Step 2: Revoke privileges from ackify_app role
-- Note: We don't DROP the role as it might be in use by other connections
REVOKE SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public FROM ackify_app;
REVOKE USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public FROM ackify_app;
REVOKE USAGE ON SCHEMA public FROM ackify_app;
-- REVOKE CONNECT is not done to avoid breaking active connections

-- Step 3: Remove default privileges
ALTER DEFAULT PRIVILEGES IN SCHEMA public
    REVOKE SELECT, INSERT, UPDATE, DELETE ON TABLES FROM ackify_app;

ALTER DEFAULT PRIVILEGES IN SCHEMA public
    REVOKE USAGE, SELECT ON SEQUENCES FROM ackify_app;

-- Step 4: Drop the helper function
DROP FUNCTION IF EXISTS current_tenant_id();

-- Note: The ackify_app role is NOT dropped to avoid breaking existing connections.
-- To fully remove it, run: DROP ROLE IF EXISTS ackify_app;
-- after ensuring no active connections use this role.
