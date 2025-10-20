-- Create expected_signers table for tracking who should sign a document
-- This table allows administrators to manage expected signers and track completion rates
CREATE TABLE expected_signers (
    id BIGSERIAL PRIMARY KEY,
    doc_id TEXT NOT NULL,
    email TEXT NOT NULL,
    added_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    added_by TEXT NOT NULL,
    notes TEXT,
    UNIQUE (doc_id, email)
);

-- Create indexes for efficient queries
CREATE INDEX idx_expected_signers_doc_id ON expected_signers(doc_id);
CREATE INDEX idx_expected_signers_email ON expected_signers(email);

-- Add comment explaining the table purpose
COMMENT ON TABLE expected_signers IS 'Tracks expected signers for documents to monitor completion rates';
