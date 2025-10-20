-- SPDX-License-Identifier: AGPL-3.0-or-later

-- Create checksum_verifications table for tracking document integrity verification attempts
CREATE TABLE checksum_verifications (
    id BIGSERIAL PRIMARY KEY,
    doc_id TEXT NOT NULL,
    verified_by TEXT NOT NULL,
    verified_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    stored_checksum TEXT NOT NULL,
    calculated_checksum TEXT NOT NULL,
    algorithm TEXT NOT NULL CHECK (algorithm IN ('SHA-256', 'SHA-512', 'MD5')),
    is_valid BOOLEAN NOT NULL,
    error_message TEXT,
    CONSTRAINT fk_checksum_verifications_doc_id
        FOREIGN KEY (doc_id)
        REFERENCES documents(doc_id)
        ON DELETE CASCADE
);

COMMENT ON TABLE checksum_verifications IS 'Tracks verification attempts of document checksums for integrity monitoring';
COMMENT ON COLUMN checksum_verifications.id IS 'Unique identifier for the verification record';
COMMENT ON COLUMN checksum_verifications.doc_id IS 'Document identifier (foreign key to documents table)';
COMMENT ON COLUMN checksum_verifications.verified_by IS 'Email of the user who performed the verification';
COMMENT ON COLUMN checksum_verifications.verified_at IS 'Timestamp when verification was performed';
COMMENT ON COLUMN checksum_verifications.stored_checksum IS 'The reference checksum stored in the document metadata at verification time';
COMMENT ON COLUMN checksum_verifications.calculated_checksum IS 'The checksum calculated by the user during verification';
COMMENT ON COLUMN checksum_verifications.algorithm IS 'Algorithm used for checksum calculation (SHA-256, SHA-512, or MD5)';
COMMENT ON COLUMN checksum_verifications.is_valid IS 'True if calculated_checksum matches stored_checksum';
COMMENT ON COLUMN checksum_verifications.error_message IS 'Optional error message if verification failed';

-- Create indexes for efficient querying
CREATE INDEX idx_checksum_verifications_doc_id ON checksum_verifications(doc_id);
CREATE INDEX idx_checksum_verifications_verified_at ON checksum_verifications(verified_at DESC);
CREATE INDEX idx_checksum_verifications_doc_id_verified_at ON checksum_verifications(doc_id, verified_at DESC);

-- Create email_queue table for asynchronous email processing with retry capability
-- This table stores emails to be sent by a background worker with retry logic
CREATE TABLE email_queue (
    id BIGSERIAL PRIMARY KEY,
    -- Email metadata
    to_addresses TEXT[] NOT NULL,
    cc_addresses TEXT[],
    bcc_addresses TEXT[],
    subject TEXT NOT NULL,
    template TEXT NOT NULL,
    locale TEXT NOT NULL DEFAULT 'fr',
    data JSONB NOT NULL DEFAULT '{}',
    headers JSONB,

    -- Queue management
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'processing', 'sent', 'failed', 'cancelled')),
    priority INT NOT NULL DEFAULT 0, -- Higher priority = processed first
    retry_count INT NOT NULL DEFAULT 0,
    max_retries INT NOT NULL DEFAULT 3,

    -- Tracking
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    scheduled_for TIMESTAMPTZ NOT NULL DEFAULT now(), -- When to process (for delayed sends)
    processed_at TIMESTAMPTZ,
    next_retry_at TIMESTAMPTZ,

    -- Error tracking
    last_error TEXT,
    error_details JSONB,

    -- Reference tracking (optional)
    reference_type TEXT, -- e.g., 'reminder', 'notification', etc.
    reference_id TEXT,   -- e.g., doc_id
    created_by TEXT
);

-- Indexes for efficient queue processing
CREATE INDEX idx_email_queue_status_scheduled ON email_queue(status, scheduled_for)
    WHERE status IN ('pending', 'processing');
CREATE INDEX idx_email_queue_priority_scheduled ON email_queue(priority DESC, scheduled_for ASC)
    WHERE status = 'pending';
CREATE INDEX idx_email_queue_retry ON email_queue(next_retry_at)
    WHERE status = 'processing' AND retry_count < max_retries;
CREATE INDEX idx_email_queue_reference ON email_queue(reference_type, reference_id);
CREATE INDEX idx_email_queue_created_at ON email_queue(created_at DESC);

-- Function to calculate next retry time with exponential backoff
CREATE OR REPLACE FUNCTION calculate_next_retry_time(retry_count INT)
RETURNS TIMESTAMPTZ AS $$
BEGIN
    -- Exponential backoff: 1min, 2min, 4min, 8min, 16min, 32min...
    RETURN now() + (interval '1 minute' * power(2, retry_count));
END;
$$ LANGUAGE plpgsql;

-- Trigger to auto-update next_retry_at on status change to processing
CREATE OR REPLACE FUNCTION update_email_queue_retry_time()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.status = 'processing' AND OLD.status != 'processing' THEN
        NEW.next_retry_at = calculate_next_retry_time(NEW.retry_count);
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_email_queue_retry
    BEFORE UPDATE ON email_queue
    FOR EACH ROW
    EXECUTE FUNCTION update_email_queue_retry_time();

-- Add comment explaining the table purpose
COMMENT ON TABLE email_queue IS 'Asynchronous email queue with retry capability for reliable email delivery';
COMMENT ON COLUMN email_queue.priority IS 'Higher values are processed first (0=normal, 10=high, 100=urgent)';
COMMENT ON COLUMN email_queue.scheduled_for IS 'Earliest time to process this email (for delayed sends)';
