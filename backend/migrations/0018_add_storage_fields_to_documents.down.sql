-- SPDX-License-Identifier: AGPL-3.0-or-later

-- Remove storage fields from documents table
DROP INDEX IF EXISTS idx_documents_storage_key;
ALTER TABLE documents DROP CONSTRAINT IF EXISTS documents_storage_provider_check;
ALTER TABLE documents DROP COLUMN IF EXISTS original_filename;
ALTER TABLE documents DROP COLUMN IF EXISTS mime_type;
ALTER TABLE documents DROP COLUMN IF EXISTS file_size;
ALTER TABLE documents DROP COLUMN IF EXISTS storage_provider;
ALTER TABLE documents DROP COLUMN IF EXISTS storage_key;
