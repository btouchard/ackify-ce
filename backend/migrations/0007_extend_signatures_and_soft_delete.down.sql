-- SPDX-License-Identifier: AGPL-3.0-or-later

-- Drop trigger and function for soft delete
DROP TRIGGER IF EXISTS trigger_mark_signatures_on_document_soft_delete ON documents;
DROP FUNCTION IF EXISTS mark_signatures_on_document_soft_delete();

-- Drop index and remove deleted_at column from documents table
DROP INDEX IF EXISTS idx_documents_deleted_at;
ALTER TABLE documents DROP COLUMN IF EXISTS deleted_at;

-- Drop index and remove doc_deleted_at column from signatures table
DROP INDEX IF EXISTS idx_signatures_doc_deleted_at;
ALTER TABLE signatures DROP COLUMN IF EXISTS doc_deleted_at;

-- Drop index and remove hash_version column from signatures table
DROP INDEX IF EXISTS idx_signatures_hash_version;
ALTER TABLE signatures DROP COLUMN IF EXISTS hash_version;

-- Drop index and remove doc_checksum column from signatures table
DROP INDEX IF EXISTS idx_signatures_doc_checksum;
ALTER TABLE signatures DROP COLUMN IF EXISTS doc_checksum;
