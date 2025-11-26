# Expected Signers

Tracking expected signers with email reminders.

## Overview

The "Expected Signers" feature allows you to:
- Define who should sign a document
- Track completion rate
- Send automatic email reminders
- Detect unexpected signatures

## Adding Signers

### Via Admin Dashboard

1. Go to `/admin`
2. Select a document
3. Click "Expected Signers"
4. Paste email list:

```
Alice Smith <alice@company.com>
bob@company.com
charlie@company.com
```

**Supported formats**:
- One email per line
- Comma-separated emails
- Semicolon-separated emails
- Format with name: `Alice Smith <alice@company.com>`

### Via API

```http
POST /api/v1/admin/documents/policy_2025/signers
Content-Type: application/json
X-CSRF-Token: abc123

{
  "email": "alice@company.com",
  "name": "Alice Smith",
  "notes": "Engineering team lead"
}
```

### CSV Import (Recommended)

The most efficient method to add many signers is the native CSV import.

**Via Admin Dashboard:**
1. Go to `/admin` and select a document
2. In the "Expected Readers" section, click **Import CSV**
3. Select a CSV file
4. Preview entries (valid, existing, invalid)
5. Confirm the import

**Supported CSV format:**
```csv
email,name
alice@company.com,Alice Smith
bob@company.com,Bob Jones
charlie@company.com,Charlie Brown
```

**Auto-detected features:**
- **Separator**: comma (`,`) or semicolon (`;`)
- **Header**: automatic detection of `email` and `name` columns
- **Column order**: flexible (email/name or name/email)
- **Name column**: optional

**Examples of valid formats:**
```csv
# With header, comma separator
email,name
alice@company.com,Alice Smith

# Without header (email only)
bob@company.com
charlie@company.com

# French headers, semicolon separator
courriel;nom
alice@company.com;Alice Smith

# Reversed columns
name,email
Bob Jones,bob@company.com
```

**Configurable limit:**
```bash
# Default: 500 signers max per import
ACKIFY_IMPORT_MAX_SIGNERS=1000
```

**Via API:**

1. **Preview** (CSV analysis):
```http
POST /api/v1/admin/documents/{docId}/signers/preview-csv
Content-Type: multipart/form-data
X-CSRF-Token: abc123

file: [CSV file]
```

Response:
```json
{
  "signers": [
    {"lineNumber": 2, "email": "alice@company.com", "name": "Alice Smith"},
    {"lineNumber": 3, "email": "bob@company.com", "name": "Bob Jones"}
  ],
  "errors": [
    {"lineNumber": 5, "content": "invalid-email", "error": "invalid_email_format"}
  ],
  "totalLines": 4,
  "validCount": 2,
  "invalidCount": 1,
  "hasHeader": true,
  "existingEmails": ["charlie@company.com"],
  "maxSigners": 500
}
```

2. **Import** (after validation):
```http
POST /api/v1/admin/documents/{docId}/signers/import
Content-Type: application/json
X-CSRF-Token: abc123

{
  "signers": [
    {"email": "alice@company.com", "name": "Alice Smith"},
    {"email": "bob@company.com", "name": "Bob Jones"}
  ]
}
```

Response:
```json
{
  "message": "Import completed",
  "imported": 2,
  "skipped": 0,
  "total": 2
}
```

## Completion Tracking

### Admin Dashboard

Displays:
- **Progress bar** - Visual with percentage
- **Signer list**:
  - ✓ Email (signed on MM/DD/YYYY HH:MM)
  - ⏳ Email (pending)
- **Statistics**:
  - Expected: 50
  - Signed: 42
  - Pending: 8
  - Completion: 84%

### Via API

```http
GET /api/v1/documents/policy_2025/expected-signers
```

**Response**:
```json
{
  "docId": "policy_2025",
  "expectedSigners": [
    {
      "email": "alice@company.com",
      "name": "Alice Smith",
      "addedAt": "2025-01-15T10:00:00Z",
      "hasSigned": true,
      "signedAt": "2025-01-15T14:30:00Z"
    },
    {
      "email": "bob@company.com",
      "name": "Bob Jones",
      "addedAt": "2025-01-15T10:00:00Z",
      "hasSigned": false
    }
  ],
  "completionStats": {
    "expected": 50,
    "signed": 42,
    "pending": 8,
    "completionPercentage": 84.0
  }
}
```

## Email Reminders

### Sending Reminders

**Via Dashboard**:
1. Select recipients (or "Select all pending")
2. Choose language (fr, en, es, de, it)
3. Click "Send Reminders"

**Via API**:
```http
POST /api/v1/admin/documents/policy_2025/reminders
Content-Type: application/json
X-CSRF-Token: abc123

{
  "emails": ["bob@company.com", "charlie@company.com"],
  "locale": "fr"
}
```

**Response**:
```json
{
  "sent": 2,
  "failed": 0,
  "errors": []
}
```

### Email Content

Templates are in `/backend/templates/emails/{locale}/reminder.html`:

```html
Hello {{.RecipientName}},

You are expected to sign the document "{{.DocumentTitle}}".

[Button: Sign now] → {{.SignURL}}

Document available here: {{.DocumentURL}}

Best regards,
{{.OrganisationName}}
```

**Available variables**:
- `RecipientName` - Recipient name
- `DocumentTitle` - Document title
- `DocumentURL` - Document URL (metadata)
- `SignURL` - Direct link to signature page
- `OrganisationName` - Your organization name

### Reminder History

```http
GET /api/v1/admin/documents/policy_2025/reminders
```

**Response**:
```json
{
  "reminders": [
    {
      "recipientEmail": "bob@company.com",
      "sentAt": "2025-01-15T15:00:00Z",
      "sentBy": "admin@company.com",
      "status": "sent",
      "templateUsed": "reminder"
    },
    {
      "recipientEmail": "charlie@company.com",
      "sentAt": "2025-01-15T15:00:05Z",
      "sentBy": "admin@company.com",
      "status": "failed",
      "errorMessage": "SMTP timeout"
    }
  ]
}
```

**Statuses**:
- `sent` - Successfully sent
- `failed` - Send failure
- `bounced` - Invalid email (bounce)

## Unexpected Signatures

Automatically detects users who signed **without being expected**.

### Via Dashboard

"Unexpected Signatures" section displays:
```
⚠️ 3 unexpected signatures detected
- stranger@external.com (signed on 01/15/2025)
- unknown@gmail.com (signed on 01/16/2025)
```

### Via API

SQL query to detect:
```sql
SELECT s.user_email, s.signed_at
FROM signatures s
LEFT JOIN expected_signers e ON s.user_email = e.email AND s.doc_id = e.doc_id
WHERE s.doc_id = 'policy_2025' AND e.id IS NULL;
```

## Use Cases

### Mandatory Training

```
Document: "GDPR Training 2025"
Expected: All employees (CSV import)
```

**Workflow**:
1. Import CSV with employee emails
2. Send signature link to everyone
3. Automatic reminder on D+7 to non-signers
4. Final export for HR

### Security Policy

```
Document: "Security Policy v3"
Expected: Engineers + DevOps (50 people)
```

**Features used**:
- Real-time tracking (dashboard)
- Selective reminders (only some)
- Document metadata (URL + checksum)

### Contractual

```
Document: "NDA 2025"
Expected: External contractors (manual list)
```

**Particularity**:
- Restricted OAuth domain disabled
- Allows external emails to sign
- Unexpected signature detection crucial

## Removing a Signer

```http
DELETE /api/v1/admin/documents/policy_2025/signers/alice@company.com
X-CSRF-Token: abc123
```

**Behavior**:
- Removes from expected_signers list
- Signature (if exists) remains in database
- Completion rate is recalculated

## Email Configuration

For reminders to work, configure SMTP:

```bash
ACKIFY_MAIL_HOST=smtp.gmail.com
ACKIFY_MAIL_PORT=587
ACKIFY_MAIL_USERNAME=noreply@company.com
ACKIFY_MAIL_PASSWORD=app_password
ACKIFY_MAIL_FROM=noreply@company.com
```

See [Email Setup](../configuration/email-setup.md) for more details.

## Best Practices

### CSV Import

Use the native CSV import (see "CSV Import" section above) for bulk imports. Benefits:
- Preview before import with error detection
- Automatic duplicate detection
- Configurable limit (`ACKIFY_IMPORT_MAX_SIGNERS`)
- Multi-format support (comma/semicolon, with/without header)

### Customization

For more personalized reminders:
1. Modify templates in `/backend/templates/emails/`
2. Add custom variables in email service
3. Rebuild Docker image

### Monitoring

Monitor `reminder_logs` to detect:
- High bounce rate (invalid emails)
- Repeated SMTP failures
- Reminder effectiveness (conversion rate)

## Limitations

- Maximum **1000 expected signers** per document (soft limit)
- Reminders sent **synchronously** (no queue)
- No automatic scheduled reminders (manual only)

## API Reference

See [API Documentation](../api.md#expected-signers-admin) for all endpoints.
