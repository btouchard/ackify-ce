-- SPDX-License-Identifier: AGPL-3.0-or-later

-- Add doc_checksum column to signatures table
-- This ensures signatures are tied to a specific version of the document
-- The checksum is included in the cryptographic signature payload for integrity verification
ALTER TABLE signatures ADD COLUMN doc_checksum TEXT;

-- Add index for efficient checksum-based queries
CREATE INDEX idx_signatures_doc_checksum ON signatures(doc_checksum) WHERE doc_checksum IS NOT NULL;

-- Add comment explaining the column
COMMENT ON COLUMN signatures.doc_checksum IS 'SHA-256 checksum of the document at time of signature. Included in Ed25519 signature payload to prove signature applies to specific document version.';

-- Add hash_version column to support hash algorithm evolution
-- Version 1: pipe-separated format (legacy)
-- Version 2: JSON canonical format (recommended)
ALTER TABLE signatures ADD COLUMN hash_version INT NOT NULL DEFAULT 1;

-- Add index for queries by hash version
CREATE INDEX idx_signatures_hash_version ON signatures(hash_version);

-- Add comment explaining the versioning
COMMENT ON COLUMN signatures.hash_version IS 'Hash algorithm version used for ComputeRecordHash. 1=pipe-separated (legacy), 2=JSON canonical format. Allows backward compatibility while supporting improved hash formats.';

-- Add doc_deleted_at column to track when the referenced document was deleted
-- This allows keeping signature history even after document deletion
ALTER TABLE signatures ADD COLUMN doc_deleted_at TIMESTAMPTZ;

-- Add index for efficient queries filtering deleted/non-deleted docs
CREATE INDEX idx_signatures_doc_deleted_at ON signatures(doc_deleted_at) WHERE doc_deleted_at IS NOT NULL;

-- Add comment explaining the column
COMMENT ON COLUMN signatures.doc_deleted_at IS 'Timestamp when the referenced document was deleted. NULL means document still exists. Allows preserving signature history.';

-- Add deleted_at column for soft delete on documents
ALTER TABLE documents ADD COLUMN deleted_at TIMESTAMPTZ;

-- Add index for efficient queries filtering deleted/non-deleted documents
CREATE INDEX idx_documents_deleted_at ON documents(deleted_at) WHERE deleted_at IS NOT NULL;

-- Add comment explaining the column
COMMENT ON COLUMN documents.deleted_at IS 'Timestamp when the document was soft-deleted. NULL means document is active. Allows preserving document metadata and signature history.';

-- Create trigger function for soft delete
CREATE OR REPLACE FUNCTION mark_signatures_on_document_soft_delete()
RETURNS TRIGGER AS $$
BEGIN
    -- When a document is soft deleted (deleted_at is set), mark all signatures
    IF NEW.deleted_at IS NOT NULL AND OLD.deleted_at IS NULL THEN
        UPDATE signatures
        SET doc_deleted_at = NEW.deleted_at
        WHERE doc_id = NEW.doc_id
          AND doc_deleted_at IS NULL;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger that fires on UPDATE
CREATE TRIGGER trigger_mark_signatures_on_document_soft_delete
    AFTER UPDATE ON documents
    FOR EACH ROW
    EXECUTE FUNCTION mark_signatures_on_document_soft_delete();

COMMENT ON FUNCTION mark_signatures_on_document_soft_delete() IS 'Marks all signatures of a document as deleted when the document is soft-deleted';
