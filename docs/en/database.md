# Database

PostgreSQL schema, migrations, and integrity guarantees.

## Overview

Ackify uses **PostgreSQL 16+** with:
- Versioned SQL migrations
- Strict integrity constraints
- Triggers for immutability
- Indexes for performance

## Main Schema

### Table `signatures`

Stores Ed25519 cryptographic signatures.

```sql
CREATE TABLE signatures (
    id BIGSERIAL PRIMARY KEY,
    doc_id TEXT NOT NULL,
    user_sub TEXT NOT NULL,                 -- OAuth user ID (sub claim)
    user_email TEXT NOT NULL,
    user_name TEXT,                         -- User name (optional)
    signed_at TIMESTAMPTZ NOT NULL,
    payload_hash TEXT NOT NULL,             -- SHA-256 of payload
    signature TEXT NOT NULL,                -- Ed25519 signature (base64)
    nonce TEXT NOT NULL,                    -- Anti-replay attack
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    referer TEXT,                           -- Source (optional)
    prev_hash TEXT,                         -- Hash of previous signature (chaining)
    UNIQUE (doc_id, user_sub)              -- ONE signature per user/document
);

CREATE INDEX idx_signatures_doc_id ON signatures(doc_id);
CREATE INDEX idx_signatures_user_sub ON signatures(user_sub);
```

**Guarantees**:
- ✅ One signature per user/document (UNIQUE constraint)
- ✅ Immutable timestamp via PostgreSQL trigger
- ✅ Hash chaining (blockchain-like) via `prev_hash`
- ✅ Cryptographic non-repudiation (Ed25519)

### Table `documents`

Document metadata.

```sql
CREATE TABLE documents (
    doc_id TEXT PRIMARY KEY,
    title TEXT NOT NULL DEFAULT '',
    url TEXT NOT NULL DEFAULT '',           -- Source document URL
    checksum TEXT NOT NULL DEFAULT '',      -- SHA-256, SHA-512, or MD5
    checksum_algorithm TEXT NOT NULL DEFAULT 'SHA-256',
    description TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by TEXT NOT NULL DEFAULT ''     -- Creator admin's user_sub
);
```

**Usage**:
- Title, description displayed in interface
- URL included in reminder emails
- Checksum for integrity verification (optional)

### Table `expected_signers`

Expected signers for tracking.

```sql
CREATE TABLE expected_signers (
    id BIGSERIAL PRIMARY KEY,
    doc_id TEXT NOT NULL,
    email TEXT NOT NULL,
    name TEXT NOT NULL DEFAULT '',          -- Name for personalization
    added_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    added_by TEXT NOT NULL,                 -- Admin who added
    notes TEXT,
    UNIQUE (doc_id, email)
);

CREATE INDEX idx_expected_signers_doc_id ON expected_signers(doc_id);
```

**Features**:
- Completion tracking (% signed)
- Email reminder sending
- Unexpected signature detection

### Table `reminder_logs`

Email reminder history.

```sql
CREATE TABLE reminder_logs (
    id BIGSERIAL PRIMARY KEY,
    doc_id TEXT NOT NULL,
    recipient_email TEXT NOT NULL,
    sent_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    sent_by TEXT NOT NULL,                  -- Admin who sent
    template_used TEXT NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('sent', 'failed', 'bounced')),
    error_message TEXT,
    FOREIGN KEY (doc_id, recipient_email)
        REFERENCES expected_signers(doc_id, email)
);

CREATE INDEX idx_reminder_logs_doc_id ON reminder_logs(doc_id);
```

### Table `checksum_verifications`

Integrity verification history.

```sql
CREATE TABLE checksum_verifications (
    id BIGSERIAL PRIMARY KEY,
    doc_id TEXT NOT NULL,
    verified_by TEXT NOT NULL,
    verified_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    stored_checksum TEXT NOT NULL,
    calculated_checksum TEXT NOT NULL,
    algorithm TEXT NOT NULL,
    is_valid BOOLEAN NOT NULL,
    error_message TEXT,
    FOREIGN KEY (doc_id) REFERENCES documents(doc_id)
);

CREATE INDEX idx_checksum_verifications_doc_id ON checksum_verifications(doc_id);
```

### Table `oauth_sessions`

OAuth2 sessions with encrypted refresh tokens.

```sql
CREATE TABLE oauth_sessions (
    id BIGSERIAL PRIMARY KEY,
    session_id TEXT NOT NULL UNIQUE,           -- Gorilla session ID
    user_sub TEXT NOT NULL,                    -- OAuth user ID
    refresh_token_encrypted BYTEA NOT NULL,    -- Encrypted AES-256-GCM
    access_token_expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_refreshed_at TIMESTAMPTZ,
    user_agent TEXT,
    ip_address INET
);

CREATE INDEX idx_oauth_sessions_session_id ON oauth_sessions(session_id);
CREATE INDEX idx_oauth_sessions_user_sub ON oauth_sessions(user_sub);
CREATE INDEX idx_oauth_sessions_updated_at ON oauth_sessions(updated_at);
```

**Security**:
- Encrypted refresh tokens (AES-256-GCM)
- Automatic cleanup after 37 days
- IP + User-Agent tracking to detect theft

### Table `email_queue`

Asynchronous email queue with retry mechanism.

```sql
CREATE TABLE email_queue (
    id BIGSERIAL PRIMARY KEY,

    -- Email metadata
    to_addresses TEXT[] NOT NULL,              -- Recipient email addresses
    cc_addresses TEXT[],                       -- CC addresses (optional)
    bcc_addresses TEXT[],                      -- BCC addresses (optional)
    subject TEXT NOT NULL,                     -- Email subject
    template TEXT NOT NULL,                    -- Template name (e.g., 'reminder')
    locale TEXT NOT NULL DEFAULT 'fr',         -- Email language (en, fr, es, de, it)
    data JSONB NOT NULL DEFAULT '{}',          -- Template variables
    headers JSONB,                             -- Custom email headers (optional)

    -- Queue management
    status TEXT NOT NULL DEFAULT 'pending'     -- pending, processing, sent, failed, cancelled
        CHECK (status IN ('pending', 'processing', 'sent', 'failed', 'cancelled')),
    priority INT NOT NULL DEFAULT 0,           -- Higher = processed first (0=normal, 10=high, 100=urgent)
    retry_count INT NOT NULL DEFAULT 0,        -- Number of retry attempts
    max_retries INT NOT NULL DEFAULT 3,        -- Maximum retry limit

    -- Tracking
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    scheduled_for TIMESTAMPTZ NOT NULL DEFAULT now(),  -- Earliest processing time
    processed_at TIMESTAMPTZ,                  -- When email was sent
    next_retry_at TIMESTAMPTZ,                 -- Calculated retry time (exponential backoff)

    -- Error tracking
    last_error TEXT,                           -- Last error message
    error_details JSONB,                       -- Detailed error information

    -- Reference tracking (optional)
    reference_type TEXT,                       -- e.g., 'reminder', 'notification'
    reference_id TEXT,                         -- e.g., doc_id
    created_by TEXT                            -- User who queued the email
);

-- Indexes for efficient queue processing
CREATE INDEX idx_email_queue_status_scheduled
    ON email_queue(status, scheduled_for)
    WHERE status IN ('pending', 'processing');

CREATE INDEX idx_email_queue_priority_scheduled
    ON email_queue(priority DESC, scheduled_for ASC)
    WHERE status = 'pending';

CREATE INDEX idx_email_queue_retry
    ON email_queue(next_retry_at)
    WHERE status = 'processing' AND retry_count < max_retries;

CREATE INDEX idx_email_queue_reference
    ON email_queue(reference_type, reference_id);

CREATE INDEX idx_email_queue_created_at
    ON email_queue(created_at DESC);
```

**Features**:
- **Asynchronous processing**: Emails processed by background worker
- **Retry mechanism**: Exponential backoff (1min, 2min, 4min, 8min, 16min, 32min...)
- **Priority support**: High-priority emails processed first
- **Scheduled sending**: Delay email delivery with `scheduled_for`
- **Error tracking**: Detailed error logging and retry history
- **Reference tracking**: Link emails to documents or other entities

**Automatic retry calculation**:
```sql
-- Function to calculate next retry time with exponential backoff
CREATE OR REPLACE FUNCTION calculate_next_retry_time(retry_count INT)
RETURNS TIMESTAMPTZ AS $$
BEGIN
    -- Exponential backoff: 1min, 2min, 4min, 8min, 16min, 32min...
    RETURN now() + (interval '1 minute' * power(2, retry_count));
END;
$$ LANGUAGE plpgsql;
```

**Worker configuration**:
- Batch size: 10 emails per batch
- Poll interval: 5 seconds
- Concurrent sends: 5 simultaneous emails
- Old email cleanup: 7 days retention for sent/failed emails

## Migrations

### Migration Management

Migrations are in `/backend/migrations/` with format:

```
XXXX_description.up.sql     # "up" migration
XXXX_description.down.sql   # "down" rollback
```

**Current files**:
- `0001_init.up.sql` - Signatures table
- `0002_expected_signers.up.sql` - Expected signers
- `0003_reminder_logs.up.sql` - Reminder logs
- `0004_add_name_to_expected_signers.up.sql` - Signer names
- `0005_create_documents_table.up.sql` - Documents metadata
- `0006_create_new_tables.up.sql` - Checksum verifications and email queue
- `0007_oauth_sessions.up.sql` - OAuth sessions with refresh tokens

### Applying Migrations

**Via Docker Compose** (automatic):

```bash
docker compose up -d
# The ackify-migrate service applies migrations on startup
```

**Manually**:

```bash
cd backend
go run ./cmd/migrate up
```

**Rollback last migration**:

```bash
go run ./cmd/migrate down
```

### Custom Migrations

To create a new migration:

1. Create `XXXX_my_feature.up.sql`:
```sql
-- Migration up
ALTER TABLE signatures ADD COLUMN new_field TEXT;
```

2. Create `XXXX_my_feature.down.sql`:
```sql
-- Rollback
ALTER TABLE signatures DROP COLUMN new_field;
```

3. Apply:
```bash
go run ./cmd/migrate up
```

## PostgreSQL Triggers

### Immutability of `created_at`

Trigger preventing `created_at` modification:

```sql
CREATE OR REPLACE FUNCTION prevent_created_at_update()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.created_at <> OLD.created_at THEN
        RAISE EXCEPTION 'created_at cannot be modified';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER prevent_signatures_created_at_update
    BEFORE UPDATE ON signatures
    FOR EACH ROW
    EXECUTE FUNCTION prevent_created_at_update();
```

**Guarantee**: No signature can be backdated.

### Auto-update of `updated_at`

For tables with `updated_at`:

```sql
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_documents_updated_at
    BEFORE UPDATE ON documents
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
```

## Useful Queries

### View document signatures

```sql
SELECT
    user_email,
    user_name,
    signed_at,
    payload_hash,
    signature
FROM signatures
WHERE doc_id = 'my_document'
ORDER BY signed_at DESC;
```

### Completion status

```sql
WITH expected AS (
    SELECT COUNT(*) as total
    FROM expected_signers
    WHERE doc_id = 'my_document'
),
signed AS (
    SELECT COUNT(*) as count
    FROM signatures s
    INNER JOIN expected_signers e ON s.user_email = e.email AND s.doc_id = e.doc_id
    WHERE s.doc_id = 'my_document'
)
SELECT
    e.total as expected,
    s.count as signed,
    ROUND(100.0 * s.count / NULLIF(e.total, 0), 2) as completion_pct
FROM expected e, signed s;
```

### Missing signers

```sql
SELECT
    e.email,
    e.name,
    e.added_at
FROM expected_signers e
LEFT JOIN signatures s ON e.email = s.user_email AND e.doc_id = s.doc_id
WHERE e.doc_id = 'my_document' AND s.id IS NULL
ORDER BY e.added_at;
```

### Unexpected signatures

```sql
SELECT
    s.user_email,
    s.signed_at
FROM signatures s
LEFT JOIN expected_signers e ON s.user_email = e.email AND s.doc_id = e.doc_id
WHERE s.doc_id = 'my_document' AND e.id IS NULL
ORDER BY s.signed_at DESC;
```

### Email queue status

```sql
-- View pending emails
SELECT
    id,
    to_addresses,
    subject,
    status,
    priority,
    retry_count,
    scheduled_for,
    created_at
FROM email_queue
WHERE status IN ('pending', 'processing')
ORDER BY priority DESC, scheduled_for ASC
LIMIT 20;

-- Failed emails needing attention
SELECT
    id,
    to_addresses,
    subject,
    retry_count,
    max_retries,
    last_error,
    next_retry_at
FROM email_queue
WHERE status = 'failed'
ORDER BY created_at DESC;

-- Email statistics by status
SELECT
    status,
    COUNT(*) as count,
    MIN(created_at) as oldest,
    MAX(created_at) as newest
FROM email_queue
GROUP BY status
ORDER BY status;
```

## Backup & Restore

### PostgreSQL Backup

```bash
# Full backup
docker compose exec ackify-db pg_dump -U ackifyr ackify > backup.sql

# Compressed backup
docker compose exec ackify-db pg_dump -U ackifyr ackify | gzip > backup.sql.gz
```

### Restore

```bash
# Restore from backup
cat backup.sql | docker compose exec -T ackify-db psql -U ackifyr ackify

# Restore from compressed backup
gunzip -c backup.sql.gz | docker compose exec -T ackify-db psql -U ackifyr ackify
```

### Automated Backup

Example cron for daily backup:

```bash
0 2 * * * docker compose -f /path/to/compose.yml exec -T ackify-db pg_dump -U ackifyr ackify | gzip > /backups/ackify-$(date +\%Y\%m\%d).sql.gz
```

## Performance

### Indexes

Indexes are automatically created for:
- `signatures(doc_id)` - Document queries
- `signatures(user_sub)` - User queries
- `expected_signers(doc_id)` - Completion tracking
- `oauth_sessions(session_id)` - Session lookups

### Connection Pooling

The Go backend automatically handles connection pooling:
- Max open connections: 25
- Max idle connections: 5
- Connection max lifetime: 5 minutes

### Vacuum & Analyze

PostgreSQL handles automatically via `autovacuum`. To force:

```sql
VACUUM ANALYZE signatures;
VACUUM ANALYZE documents;
```

## Monitoring

### Table sizes

```sql
SELECT
    schemaname,
    tablename,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS size
FROM pg_tables
WHERE schemaname = 'public'
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;
```

### Statistics

```sql
SELECT * FROM pg_stat_user_tables WHERE schemaname = 'public';
```

### Active connections

```sql
SELECT
    datname,
    usename,
    application_name,
    client_addr,
    state,
    query
FROM pg_stat_activity
WHERE datname = 'ackify';
```

## Security

### In Production

- ✅ Use SSL: `?sslmode=require` in DSN
- ✅ Strong password for PostgreSQL
- ✅ Restrict network connections
- ✅ Encrypted backups
- ✅ Regular secret rotation

### SSL Configuration

```bash
# In .env
ACKIFY_DB_DSN=postgres://user:pass@host:5432/ackify?sslmode=require
```

### Audit Trail

All important operations are tracked:
- `signatures.created_at` - Signature timestamp
- `expected_signers.added_by` - Who added
- `reminder_logs.sent_by` - Who sent reminder
- `checksum_verifications.verified_by` - Who verified

## Troubleshooting

### Blocked migrations

```bash
# Check status
docker compose logs ackify-migrate

# Force rollback
docker compose exec ackify-ce /app/migrate down
docker compose exec ackify-ce /app/migrate up
```

### UNIQUE constraint violated

Error: `duplicate key value violates unique constraint`

**Cause**: User already signed this document.

**Solution**: This is normal behavior (one signature per user/doc).

### Connection refused

Verify PostgreSQL is started:

```bash
docker compose ps ackify-db
docker compose logs ackify-db
```
