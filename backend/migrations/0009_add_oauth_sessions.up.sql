-- SPDX-License-Identifier: AGPL-3.0-or-later

-- Table for storing OAuth refresh tokens securely
CREATE TABLE IF NOT EXISTS oauth_sessions (
    id BIGSERIAL PRIMARY KEY,
    session_id TEXT NOT NULL UNIQUE,           -- Gorilla session ID
    user_sub TEXT NOT NULL,                    -- OAuth user ID (sub claim)
    refresh_token_encrypted BYTEA NOT NULL,    -- AES-256-GCM encrypted refresh token
    access_token_expires_at TIMESTAMPTZ,       -- When the access token expires
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_refreshed_at TIMESTAMPTZ,             -- Last time token was refreshed

    -- Security metadata for session validation
    user_agent TEXT,                           -- User agent for session binding
    ip_address INET                            -- IP address for session binding
);

-- Indexes for fast lookups
CREATE INDEX IF NOT EXISTS idx_oauth_sessions_session_id ON oauth_sessions(session_id);
CREATE INDEX IF NOT EXISTS idx_oauth_sessions_user_sub ON oauth_sessions(user_sub);
CREATE INDEX IF NOT EXISTS idx_oauth_sessions_expires_at ON oauth_sessions(access_token_expires_at);

-- Comment for documentation
COMMENT ON TABLE oauth_sessions IS 'Stores encrypted OAuth refresh tokens for session management';
COMMENT ON COLUMN oauth_sessions.refresh_token_encrypted IS 'Refresh token encrypted with AES-256-GCM';
COMMENT ON COLUMN oauth_sessions.session_id IS 'Links to the gorilla session cookie';
COMMENT ON COLUMN oauth_sessions.user_agent IS 'Used to detect session hijacking';
COMMENT ON COLUMN oauth_sessions.ip_address IS 'Used to detect session hijacking';
