-- SPDX-License-Identifier: AGPL-3.0-or-later

-- Create documents table for document metadata
CREATE TABLE documents (
    doc_id TEXT PRIMARY KEY,
    title TEXT NOT NULL DEFAULT '',
    url TEXT NOT NULL DEFAULT '',
    checksum TEXT NOT NULL DEFAULT '',
    checksum_algorithm TEXT NOT NULL DEFAULT 'SHA-256' CHECK (checksum_algorithm IN ('SHA-256', 'SHA-512', 'MD5')),
    description TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by TEXT NOT NULL DEFAULT ''
);

COMMENT ON TABLE documents IS 'Stores document metadata including URL, checksum, and description';
COMMENT ON COLUMN documents.doc_id IS 'Document identifier (references signatures.doc_id)';
COMMENT ON COLUMN documents.title IS 'Optional document title';
COMMENT ON COLUMN documents.url IS 'URL or path to the document';
COMMENT ON COLUMN documents.checksum IS 'Checksum/hash of the document for integrity verification';
COMMENT ON COLUMN documents.checksum_algorithm IS 'Algorithm used for checksum (SHA-256, SHA-512, or MD5)';
COMMENT ON COLUMN documents.description IS 'Optional document description';
COMMENT ON COLUMN documents.created_at IS 'Timestamp when document metadata was created';
COMMENT ON COLUMN documents.updated_at IS 'Timestamp when document metadata was last updated';
COMMENT ON COLUMN documents.created_by IS 'Email of user who created the document metadata';

-- Create index on created_at for sorting
CREATE INDEX idx_documents_created_at ON documents(created_at DESC);

-- Create trigger to automatically update updated_at timestamp
CREATE OR REPLACE FUNCTION update_documents_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_documents_updated_at
    BEFORE UPDATE ON documents
    FOR EACH ROW
    EXECUTE FUNCTION update_documents_updated_at();
