-- SPDX-License-Identifier: AGPL-3.0-or-later

-- Add purpose and doc_id columns to magic_link_tokens table
-- This allows magic links to be used for both login and reminder authentication

ALTER TABLE magic_link_tokens
  ADD COLUMN purpose VARCHAR(50) NOT NULL DEFAULT 'login',
  ADD COLUMN doc_id TEXT;

-- Add index on purpose for faster queries
CREATE INDEX idx_magic_link_tokens_purpose ON magic_link_tokens(purpose);

-- Add index on doc_id for reminder lookups
CREATE INDEX idx_magic_link_tokens_doc_id ON magic_link_tokens(doc_id) WHERE doc_id IS NOT NULL;

-- Add comment explaining the purpose column
COMMENT ON COLUMN magic_link_tokens.purpose IS 'Type of magic link: login (15 min validity) or reminder_auth (24h validity)';
COMMENT ON COLUMN magic_link_tokens.doc_id IS 'Document ID for reminder_auth links (NULL for login links)';
