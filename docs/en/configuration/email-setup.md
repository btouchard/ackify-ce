# Email Setup

SMTP configuration for sending email reminders to expected signers.

## Overview

Ackify's email service allows sending automatic reminders to users who have not yet signed a document.

**Features**:
- Multilingual reminder sending (fr, en, es, de, it)
- HTML and plain text templates
- Send history in PostgreSQL
- TLS/STARTTLS support
- Configurable timeout

**Note**: Email service is **optional**. If `ACKIFY_MAIL_HOST` is not defined, emails are disabled.

## Basic Configuration

### Required Variables

```bash
# SMTP server
ACKIFY_MAIL_HOST=smtp.gmail.com
ACKIFY_MAIL_PORT=587
ACKIFY_MAIL_USERNAME=your-email@gmail.com
ACKIFY_MAIL_PASSWORD=your-app-password

# Sender address
ACKIFY_MAIL_FROM=noreply@company.com
```

### Optional Variables

```bash
# Displayed sender name (default: ACKIFY_ORGANISATION)
ACKIFY_MAIL_FROM_NAME="Ackify - ACME Corporation"

# Email subject prefix (optional)
ACKIFY_MAIL_SUBJECT_PREFIX="[Ackify]"

# Enable TLS (default: true)
ACKIFY_MAIL_TLS=true

# Enable STARTTLS (default: true)
ACKIFY_MAIL_STARTTLS=true

# Disable TLS certificate verification (default: false)
# Useful for self-signed certificates in development/testing
# /!\ DO NOT USE IN PRODUCTION
ACKIFY_MAIL_INSECURE_SKIP_VERIFY=false

# Connection timeout (default: 10s)
ACKIFY_MAIL_TIMEOUT=10s

# Email template directory (default: templates/emails)
ACKIFY_MAIL_TEMPLATE_DIR=templates/emails

# Default email language/locale (default: en)
# Supported: en, fr, es, de, it
ACKIFY_MAIL_DEFAULT_LOCALE=en
```

## Popular SMTP Providers

### Gmail

**Configuration**:
```bash
ACKIFY_MAIL_HOST=smtp.gmail.com
ACKIFY_MAIL_PORT=587
ACKIFY_MAIL_USERNAME=your-email@gmail.com
ACKIFY_MAIL_PASSWORD=your-app-password
ACKIFY_MAIL_TLS=true
ACKIFY_MAIL_STARTTLS=true
```

**Prerequisites**:
1. Enable 2-step verification on your Google account
2. Generate an "App Password": https://myaccount.google.com/apppasswords
3. Use this password in `ACKIFY_MAIL_PASSWORD`

### SendGrid

```bash
ACKIFY_MAIL_HOST=smtp.sendgrid.net
ACKIFY_MAIL_PORT=587
ACKIFY_MAIL_USERNAME=apikey
ACKIFY_MAIL_PASSWORD=your-sendgrid-api-key
ACKIFY_MAIL_FROM=noreply@your-domain.com
ACKIFY_MAIL_TLS=true
```

### Amazon SES

```bash
ACKIFY_MAIL_HOST=email-smtp.us-east-1.amazonaws.com
ACKIFY_MAIL_PORT=587
ACKIFY_MAIL_USERNAME=your-smtp-username
ACKIFY_MAIL_PASSWORD=your-smtp-password
ACKIFY_MAIL_FROM=noreply@verified-domain.com
ACKIFY_MAIL_TLS=true
```

**Important**: Verify your domain in AWS SES before sending.

### Mailgun

```bash
ACKIFY_MAIL_HOST=smtp.mailgun.org
ACKIFY_MAIL_PORT=587
ACKIFY_MAIL_USERNAME=postmaster@your-domain.mailgun.org
ACKIFY_MAIL_PASSWORD=your-mailgun-smtp-password
ACKIFY_MAIL_FROM=noreply@your-domain.com
ACKIFY_MAIL_TLS=true
```

### Custom SMTP (Self-hosted)

```bash
ACKIFY_MAIL_HOST=mail.company.com
ACKIFY_MAIL_PORT=587
ACKIFY_MAIL_USERNAME=ackify@company.com
ACKIFY_MAIL_PASSWORD=secure_password
ACKIFY_MAIL_FROM=ackify@company.com
ACKIFY_MAIL_TLS=true
ACKIFY_MAIL_STARTTLS=true
# For self-signed certificates only (/!\ not in production)
# ACKIFY_MAIL_INSECURE_SKIP_VERIFY=true
```

## Email Templates

Templates are in `/backend/templates/emails/` with multilingual support.

### Structure

```
templates/emails/
├── fr/
│   ├── reminder.html      # French HTML template
│   └── reminder.txt       # French plain text template
├── en/
│   ├── reminder.html      # English HTML template
│   └── reminder.txt       # English plain text template
└── ...
```

### Available Variables

In templates, you can use:

```go
{{.RecipientName}}     // Recipient name
{{.DocumentID}}        // Document ID
{{.DocumentTitle}}     // Document title
{{.DocumentURL}}       // Document URL (if defined in metadata)
{{.SignURL}}           // URL to sign
{{.OrganisationName}}  // Organization name
{{.SenderName}}        // Sender name (admin)
```

### HTML Template Example

```html
<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8">
  <title>Signature reminder</title>
</head>
<body>
  <h1>Hello {{.RecipientName}},</h1>
  <p>
    You are expected to sign the document
    <strong>{{.DocumentTitle}}</strong>.
  </p>
  <p>
    <a href="{{.SignURL}}" style="background: #0066cc; color: white; padding: 10px 20px; text-decoration: none; border-radius: 5px;">
      Sign now
    </a>
  </p>
  <p>
    Best regards,<br>
    {{.OrganisationName}}
  </p>
</body>
</html>
```

### Customize Templates

To use custom templates:

```bash
ACKIFY_MAIL_TEMPLATE_DIR=/custom/path/to/email/templates
```

Make sure to maintain the same directory structure (locale/reminder.html).

## Sending Reminders

### Via Admin Dashboard

1. Go to `/admin`
2. Select a document
3. Click "Expected Signers"
4. Select recipients
5. Click "Send Reminders"

### Via API

```bash
curl -X POST http://localhost:8080/api/v1/admin/documents/doc_id/reminders \
  -b cookies.txt \
  -H "Content-Type: application/json" \
  -H "X-CSRF-Token: YOUR_TOKEN" \
  -d '{
    "emails": ["user1@company.com", "user2@company.com"],
    "locale": "fr"
  }'
```

**Response**:
```json
{
  "sent": 2,
  "failed": 0,
  "errors": []
}
```

## Reminder History

Sends are tracked in the `reminder_logs` table:

```sql
SELECT
  recipient_email,
  sent_at,
  status,
  error_message
FROM reminder_logs
WHERE doc_id = 'my_document'
ORDER BY sent_at DESC;
```

**Possible statuses**:
- `sent` - Successfully sent
- `failed` - Send failure
- `bounced` - Bounced (invalid email)

## Testing the Configuration

### Manual Test via API

```bash
# 1. Login as admin
# 2. Add an expected signer with your email
curl -X POST http://localhost:8080/api/v1/admin/documents/test_doc/signers \
  -b cookies.txt \
  -H "Content-Type: application/json" \
  -H "X-CSRF-Token: YOUR_TOKEN" \
  -d '{
    "email": "your-email@company.com",
    "name": "Test User"
  }'

# 3. Send a test reminder
curl -X POST http://localhost:8080/api/v1/admin/documents/test_doc/reminders \
  -b cookies.txt \
  -H "Content-Type: application/json" \
  -H "X-CSRF-Token: YOUR_TOKEN" \
  -d '{
    "emails": ["your-email@company.com"],
    "locale": "en"
  }'
```

### Check Logs

```bash
docker compose logs -f ackify-ce | grep -i mail
```

You should see:
```
INFO  Email sent successfully to: your-email@company.com
```

## Troubleshooting

### Error "SMTP connection failed"

Verify:
- `ACKIFY_MAIL_HOST` and `ACKIFY_MAIL_PORT` are correct
- Your server allows outgoing connections on the SMTP port
- `ACKIFY_MAIL_TLS=true` if the server requires TLS

### Error "tls: failed to verify certificate: x509: certificate signed by unknown authority"

This error occurs with self-signed certificates. **For development/testing environments only**:

```bash
ACKIFY_MAIL_INSECURE_SKIP_VERIFY=true
```

/!\ **Warning**: This option disables TLS certificate verification. NEVER use in production!

### Error "Authentication failed"

Verify:
- `ACKIFY_MAIL_USERNAME` and `ACKIFY_MAIL_PASSWORD` are correct
- For Gmail: use an "App Password", not your main password
- For SendGrid: username must be `apikey`

### Email not received but status "sent"

Verify:
- Spam/junk folder
- SPF/DKIM/DMARC of your domain (to avoid spam filters)
- The `ACKIFY_MAIL_FROM` address is verified with your provider

### Template not found

Verify:
- `ACKIFY_MAIL_TEMPLATE_DIR` points to the correct directory
- The structure `{locale}/reminder.html` exists
- Files have the correct permissions (readable)

### Timeout during send

Increase timeout:
```bash
ACKIFY_MAIL_TIMEOUT=30s
```

## Best Practices

### Production

- ✅ Use a dedicated SMTP service (SendGrid, Mailgun, SES)
- ✅ Verify your domain (SPF, DKIM, DMARC)
- ✅ Use a `noreply@` address for `ACKIFY_MAIL_FROM`
- ✅ Monitor `reminder_logs` to detect failures
- ✅ Regularly test email sending

### Security

- ✅ Never commit `ACKIFY_MAIL_PASSWORD` to git
- ✅ Use Docker secrets or environment variables
- ✅ Restrict SMTP account permissions
- ✅ Enable TLS/STARTTLS in production

### Performance

- Emails are sent **synchronously** during API call
- For large volumes, consider an asynchronous queue
- Limit number of recipients per batch (recommended: < 100)

## Disabling Emails

To completely disable the email service:

```bash
# Remove or comment out ACKIFY_MAIL_HOST
# ACKIFY_MAIL_HOST=
```

The admin dashboard will no longer display reminder sending options.
