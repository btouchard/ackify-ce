-- Create reminder_logs table for tracking email reminders sent to expected signers
-- This table logs all reminder attempts, both successful and failed
CREATE TABLE reminder_logs (
    id BIGSERIAL PRIMARY KEY,
    doc_id TEXT NOT NULL,
    recipient_email TEXT NOT NULL,
    sent_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    sent_by TEXT NOT NULL,
    template_used TEXT NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('sent', 'failed', 'bounced')),
    error_message TEXT,
    FOREIGN KEY (doc_id, recipient_email) REFERENCES expected_signers(doc_id, email) ON DELETE CASCADE
);

-- Create indexes for efficient queries
CREATE INDEX idx_reminder_logs_doc_id ON reminder_logs(doc_id);
CREATE INDEX idx_reminder_logs_recipient ON reminder_logs(recipient_email);
CREATE INDEX idx_reminder_logs_sent_at ON reminder_logs(sent_at);
CREATE INDEX idx_reminder_logs_doc_email_sent ON reminder_logs(doc_id, recipient_email, sent_at DESC);

-- Add comment explaining the table purpose
COMMENT ON TABLE reminder_logs IS 'Logs all email reminder attempts sent to expected signers, including failures';
