-- SPDX-License-Identifier: AGPL-3.0-or-later

-- Rollback: Remove Tenant Configuration Storage

-- Revoke permissions
REVOKE SELECT, INSERT, UPDATE, DELETE ON tenant_config FROM ackify_app;
REVOKE USAGE, SELECT ON SEQUENCE tenant_config_id_seq FROM ackify_app;
REVOKE UPDATE (config_seeded_at) ON instance_metadata FROM ackify_app;

-- Drop RLS policy
DROP POLICY IF EXISTS tenant_isolation_tenant_config ON tenant_config;

-- Drop triggers
DROP TRIGGER IF EXISTS tr_tenant_config_tenant_id_immutable ON tenant_config;
DROP TRIGGER IF EXISTS tr_tenant_config_update_timestamp ON tenant_config;

-- Drop function
DROP FUNCTION IF EXISTS update_tenant_config_timestamp();

-- Remove column from instance_metadata
ALTER TABLE instance_metadata DROP COLUMN IF EXISTS config_seeded_at;

-- Drop table
DROP TABLE IF EXISTS tenant_config;
