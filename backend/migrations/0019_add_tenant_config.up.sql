-- SPDX-License-Identifier: AGPL-3.0-or-later

-- ============================================================================
-- Migration: Add Tenant Configuration Storage
-- ============================================================================
-- This migration creates a table for storing tenant-specific configuration
-- with support for:
--   - Category-based JSONB storage for flexibility
--   - Encrypted secrets storage (separate from main config)
--   - Optimistic locking via version field
--   - Audit trail (updated_by, updated_at)
--   - Tenant isolation via RLS
-- ============================================================================

-- Step 1: Create tenant_config table
CREATE TABLE tenant_config (
    id BIGSERIAL PRIMARY KEY,
    tenant_id UUID NOT NULL,
    category TEXT NOT NULL CHECK (category IN ('general', 'oidc', 'magiclink', 'smtp', 'storage')),
    config JSONB NOT NULL DEFAULT '{}',
    secrets_encrypted BYTEA,
    version INT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_by TEXT,
    UNIQUE(tenant_id, category)
);

COMMENT ON TABLE tenant_config IS 'Stores tenant-specific configuration with JSONB per category';
COMMENT ON COLUMN tenant_config.tenant_id IS 'Tenant identifier (references instance_metadata.id)';
COMMENT ON COLUMN tenant_config.category IS 'Configuration category: general, oidc, magiclink, smtp, storage';
COMMENT ON COLUMN tenant_config.config IS 'JSONB configuration data (secrets excluded)';
COMMENT ON COLUMN tenant_config.secrets_encrypted IS 'AES-256-GCM encrypted secrets blob';
COMMENT ON COLUMN tenant_config.version IS 'Optimistic locking version (incremented on each update)';
COMMENT ON COLUMN tenant_config.updated_by IS 'Email of the user who last updated this config';

-- Step 2: Add indexes
CREATE INDEX idx_tenant_config_tenant_category ON tenant_config(tenant_id, category);

-- Step 3: Add config_seeded_at column to instance_metadata
ALTER TABLE instance_metadata ADD COLUMN IF NOT EXISTS config_seeded_at TIMESTAMPTZ;

COMMENT ON COLUMN instance_metadata.config_seeded_at IS 'Timestamp when configuration was first seeded from environment variables';

-- Step 4: Create trigger for automatic updated_at and version increment
CREATE OR REPLACE FUNCTION update_tenant_config_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    NEW.version = OLD.version + 1;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION update_tenant_config_timestamp() IS 'Automatically updates updated_at and increments version on tenant_config updates';

CREATE TRIGGER tr_tenant_config_update_timestamp
    BEFORE UPDATE ON tenant_config
    FOR EACH ROW EXECUTE FUNCTION update_tenant_config_timestamp();

-- Step 5: Add tenant_id immutability trigger
CREATE TRIGGER tr_tenant_config_tenant_id_immutable
    BEFORE UPDATE ON tenant_config
    FOR EACH ROW EXECUTE FUNCTION prevent_tenant_id_modification();

-- Step 6: Enable Row Level Security
ALTER TABLE tenant_config ENABLE ROW LEVEL SECURITY;
ALTER TABLE tenant_config FORCE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS tenant_isolation_tenant_config ON tenant_config;
CREATE POLICY tenant_isolation_tenant_config ON tenant_config
    USING (tenant_id = current_tenant_id())
    WITH CHECK (tenant_id = current_tenant_id());

-- Step 7: Grant permissions to ackify_app role
GRANT SELECT, INSERT, UPDATE, DELETE ON tenant_config TO ackify_app;
GRANT USAGE, SELECT ON SEQUENCE tenant_config_id_seq TO ackify_app;

-- Step 8: Grant UPDATE on instance_metadata for config_seeded_at column
GRANT UPDATE (config_seeded_at) ON instance_metadata TO ackify_app;
