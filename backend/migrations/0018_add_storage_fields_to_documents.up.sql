-- SPDX-License-Identifier: AGPL-3.0-or-later

-- Add storage fields to documents table for file upload support
ALTER TABLE documents ADD COLUMN IF NOT EXISTS storage_key TEXT;
ALTER TABLE documents ADD COLUMN IF NOT EXISTS storage_provider TEXT;
ALTER TABLE documents ADD COLUMN IF NOT EXISTS file_size BIGINT;
ALTER TABLE documents ADD COLUMN IF NOT EXISTS mime_type TEXT;
ALTER TABLE documents ADD COLUMN IF NOT EXISTS original_filename TEXT;

-- Add constraint for storage_provider values
ALTER TABLE documents ADD CONSTRAINT documents_storage_provider_check
    CHECK (storage_provider IS NULL OR storage_provider IN ('local', 's3'));

-- Add index for storage lookups
CREATE INDEX IF NOT EXISTS idx_documents_storage_key ON documents(storage_key) WHERE storage_key IS NOT NULL;

-- Add comments
COMMENT ON COLUMN documents.storage_key IS 'Storage key/path for uploaded files';
COMMENT ON COLUMN documents.storage_provider IS 'Storage provider type: local, s3, or null for URL-based documents';
COMMENT ON COLUMN documents.file_size IS 'File size in bytes for uploaded documents';
COMMENT ON COLUMN documents.mime_type IS 'MIME type of uploaded document';
COMMENT ON COLUMN documents.original_filename IS 'Original filename of uploaded document';
