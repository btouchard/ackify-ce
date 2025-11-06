-- SPDX-License-Identifier: AGPL-3.0-or-later

-- Rollback: Remove purpose and doc_id columns from magic_link_tokens table

DROP INDEX IF EXISTS idx_magic_link_tokens_doc_id;
DROP INDEX IF EXISTS idx_magic_link_tokens_purpose;

ALTER TABLE magic_link_tokens
  DROP COLUMN IF EXISTS doc_id,
  DROP COLUMN IF EXISTS purpose;
