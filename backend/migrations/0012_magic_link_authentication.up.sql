-- Migration 0012: Magic Link Authentication
-- Adds tables for passwordless email authentication

-- Table pour stocker les tokens Magic Link
CREATE TABLE magic_link_tokens (
    id BIGSERIAL PRIMARY KEY,
    token TEXT NOT NULL UNIQUE,
    email TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at TIMESTAMPTZ NOT NULL,
    used_at TIMESTAMPTZ,
    used_by_ip INET,
    used_by_user_agent TEXT,
    redirect_to TEXT NOT NULL DEFAULT '/',
    created_by_ip INET NOT NULL,
    created_by_user_agent TEXT
);

-- Index pour requêtes fréquentes
CREATE INDEX idx_magic_link_tokens_token ON magic_link_tokens(token) WHERE used_at IS NULL;
CREATE INDEX idx_magic_link_tokens_email ON magic_link_tokens(email);
CREATE INDEX idx_magic_link_tokens_expires ON magic_link_tokens(expires_at) WHERE used_at IS NULL;

-- Index pour cleanup des tokens expirés
CREATE INDEX idx_magic_link_tokens_cleanup ON magic_link_tokens(created_at) WHERE used_at IS NULL;

COMMENT ON TABLE magic_link_tokens IS 'Tokens de connexion par Magic Link (usage unique, expiration 15min)';
COMMENT ON COLUMN magic_link_tokens.token IS 'Token cryptographiquement sécurisé (base64url, 32 bytes)';
COMMENT ON COLUMN magic_link_tokens.used_at IS 'Timestamp d''utilisation (NULL = non utilisé)';
COMMENT ON COLUMN magic_link_tokens.redirect_to IS 'URL de destination après authentification (ex: /?doc=xxx)';

-- Table pour logs des tentatives d'authentification Magic Link
CREATE TABLE magic_link_auth_attempts (
    id BIGSERIAL PRIMARY KEY,
    email TEXT NOT NULL,
    success BOOLEAN NOT NULL,
    failure_reason TEXT,
    ip_address INET NOT NULL,
    user_agent TEXT,
    attempted_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Index pour rate limiting
CREATE INDEX idx_magic_link_attempts_ip_time ON magic_link_auth_attempts(ip_address, attempted_at);
CREATE INDEX idx_magic_link_attempts_email_time ON magic_link_auth_attempts(email, attempted_at);

COMMENT ON TABLE magic_link_auth_attempts IS 'Logs des tentatives d''authentification Magic Link (rate limiting + audit)';
