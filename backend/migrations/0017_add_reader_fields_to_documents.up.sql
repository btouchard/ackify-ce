-- SPDX-License-Identifier: AGPL-3.0-or-later

-- Add reader configuration fields to documents table
-- These fields control the integrated document reader behavior

-- read_mode: 'external' (link to external URL) or 'integrated' (embedded reader)
ALTER TABLE documents ADD COLUMN read_mode TEXT NOT NULL DEFAULT 'integrated' CHECK (read_mode IN ('external', 'integrated'));

-- allow_download: Whether users can download the document
ALTER TABLE documents ADD COLUMN allow_download BOOLEAN NOT NULL DEFAULT true;

-- require_full_read: Whether user must scroll through entire document before signing
ALTER TABLE documents ADD COLUMN require_full_read BOOLEAN NOT NULL DEFAULT false;

-- verify_checksum: Whether to verify document checksum on each signature
ALTER TABLE documents ADD COLUMN verify_checksum BOOLEAN NOT NULL DEFAULT true;

COMMENT ON COLUMN documents.read_mode IS 'Reading mode: external (link) or integrated (embedded reader)';
COMMENT ON COLUMN documents.allow_download IS 'Whether document download is allowed';
COMMENT ON COLUMN documents.require_full_read IS 'Whether full document read is required before signing';
COMMENT ON COLUMN documents.verify_checksum IS 'Whether to verify document checksum on signature';