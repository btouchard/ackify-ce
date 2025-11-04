-- Rollback migration 0012: Magic Link Authentication

DROP INDEX IF EXISTS idx_magic_link_attempts_email_time;
DROP INDEX IF EXISTS idx_magic_link_attempts_ip_time;
DROP TABLE IF EXISTS magic_link_auth_attempts;

DROP INDEX IF EXISTS idx_magic_link_tokens_cleanup;
DROP INDEX IF EXISTS idx_magic_link_tokens_expires;
DROP INDEX IF EXISTS idx_magic_link_tokens_email;
DROP INDEX IF EXISTS idx_magic_link_tokens_token;
DROP TABLE IF EXISTS magic_link_tokens;
