-- SPDX-License-Identifier: AGPL-3.0-or-later

-- Rollback reader configuration fields from documents table
ALTER TABLE documents DROP COLUMN IF EXISTS verify_checksum;
ALTER TABLE documents DROP COLUMN IF EXISTS require_full_read;
ALTER TABLE documents DROP COLUMN IF EXISTS allow_download;
ALTER TABLE documents DROP COLUMN IF EXISTS read_mode;