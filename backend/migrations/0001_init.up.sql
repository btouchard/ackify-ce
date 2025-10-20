-- Create signatures table for Community Edition
CREATE TABLE signatures (
    id BIGSERIAL PRIMARY KEY,
    doc_id TEXT NOT NULL,
    user_sub TEXT NOT NULL,
    user_email TEXT NOT NULL,
    user_name TEXT,
    signed_at TIMESTAMPTZ NOT NULL,
    payload_hash TEXT NOT NULL,
    signature TEXT NOT NULL,
    nonce TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    referer TEXT,
    prev_hash TEXT,
    UNIQUE (doc_id, user_sub)
);

-- Create indexes for efficient queries
CREATE INDEX idx_signatures_user ON signatures(user_sub);
CREATE INDEX idx_signatures_doc_id ON signatures(doc_id);

-- Create trigger to prevent modification of created_at
CREATE OR REPLACE FUNCTION prevent_created_at_update()
RETURNS TRIGGER AS $$
BEGIN
    IF OLD.created_at IS DISTINCT FROM NEW.created_at THEN
        RAISE EXCEPTION 'Cannot modify created_at timestamp';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_prevent_created_at_update
    BEFORE UPDATE ON signatures
    FOR EACH ROW
    EXECUTE FUNCTION prevent_created_at_update();