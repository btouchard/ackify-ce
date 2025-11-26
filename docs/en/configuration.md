# Configuration

Complete configuration guide for Ackify via environment variables.

## Required Variables

These variables are **required** to start Ackify:

```bash
# Public URL of your instance (used for OAuth callbacks)
APP_DNS=sign.your-domain.com
ACKIFY_BASE_URL=https://sign.your-domain.com

# Your organization name (displayed in the interface)
ACKIFY_ORGANISATION="Your Organization Name"

# PostgreSQL configuration
POSTGRES_USER=ackifyr
POSTGRES_PASSWORD=your_secure_password
POSTGRES_DB=ackify

# OAuth2 Provider
ACKIFY_OAUTH_PROVIDER=google  # or github, gitlab, or empty for custom
ACKIFY_OAUTH_CLIENT_ID=your_oauth_client_id
ACKIFY_OAUTH_CLIENT_SECRET=your_oauth_client_secret

# Secret to encrypt session cookies (generate with: openssl rand -base64 32)
ACKIFY_OAUTH_COOKIE_SECRET=your_base64_encoded_secret_key
```

## Optional Variables

### Server

```bash
# HTTP listening address (default: :8080)
ACKIFY_LISTEN_ADDR=:8080

# Log level: debug, info, warn, error (default: info)
ACKIFY_LOG_LEVEL=info
```

### Security & OAuth2

```bash
# Restrict access to a specific email domain
ACKIFY_OAUTH_ALLOWED_DOMAIN=@company.com

# Enable silent auto-login (default: false)
ACKIFY_OAUTH_AUTO_LOGIN=false

# Custom logout URL (optional)
ACKIFY_OAUTH_LOGOUT_URL=https://your-provider.com/logout

# Custom OAuth2 scopes (default: openid,email,profile)
ACKIFY_OAUTH_SCOPES=openid,email,profile
```

### Authentication Methods

**Important**: At least ONE authentication method must be enabled (OAuth or MagicLink).

```bash
# Force enable/disable OAuth (default: auto-detected from credentials)
ACKIFY_AUTH_OAUTH_ENABLED=true

# Enable MagicLink passwordless authentication (default: false)
# Requires ACKIFY_MAIL_HOST to be configured
ACKIFY_AUTH_MAGICLINK_ENABLED=true
```

**Auto-detection**:
- **OAuth** is automatically enabled if `ACKIFY_OAUTH_CLIENT_ID` and `ACKIFY_OAUTH_CLIENT_SECRET` are set
- **MagicLink** requires explicit activation with `ACKIFY_AUTH_MAGICLINK_ENABLED=true` + SMTP configuration
- **SMTP/Email service** is automatically enabled when `ACKIFY_MAIL_HOST` is configured (independent of MagicLink)

**Note**: SMTP and MagicLink are two distinct features:
- **SMTP** = Email reminder service for expected signers (auto-detected)
- **MagicLink** = Passwordless email authentication (requires explicit activation + SMTP)

### Administration

```bash
# Admin email list (comma-separated)
ACKIFY_ADMIN_EMAILS=admin@company.com,admin2@company.com

# Restrict document creation to admins only (default: false)
ACKIFY_ONLY_ADMIN_CAN_CREATE=false
```

Admins have access to:
- Admin dashboard (`/admin`)
- Document metadata management
- Expected signers tracking
- Email reminders sending
- Document deletion

When `ACKIFY_ONLY_ADMIN_CAN_CREATE` is enabled:
- ✅ Only admin users can create new documents
- ✅ Non-admin users will see an error message when attempting to create documents
- ✅ Both API endpoints (`POST /documents` and `GET /documents/find-or-create`) are protected

### Rate Limiting

Configure API rate limits to prevent abuse and control request rates:

```bash
# Magic Link authentication rate limits (per time window)
ACKIFY_AUTH_MAGICLINK_RATE_LIMIT_EMAIL=3   # Max requests per email (default: 3)
ACKIFY_AUTH_MAGICLINK_RATE_LIMIT_IP=10     # Max requests per IP (default: 10)

# General API rate limits (requests per minute)
ACKIFY_AUTH_RATE_LIMIT=5          # Authentication endpoints (default: 5/min)
ACKIFY_DOCUMENT_RATE_LIMIT=10     # Document creation (default: 10/min)
ACKIFY_GENERAL_RATE_LIMIT=100     # General API requests (default: 100/min)

# CSV Import
ACKIFY_IMPORT_MAX_SIGNERS=500     # Max signers per CSV import (default: 500)
```

**When to adjust**:
- **Testing/CI**: Increase limits (e.g., `1000`) to prevent 429 errors during automated tests
- **High traffic**: Increase `GENERAL_RATE_LIMIT` for production workloads
- **Security**: Lower `AUTH_RATE_LIMIT` to prevent brute-force attacks

### Logging

```bash
# Log level: debug, info, warn, error (default: info)
ACKIFY_LOG_LEVEL=info

# Log format: classic or json (default: classic)
ACKIFY_LOG_FORMAT=classic
```

**Log formats**:
- `classic`: Human-readable format for development and simple deployments
- `json`: Structured JSON for log aggregators (Datadog, ELK, Splunk)

**Example JSON output**:
```json
{"time":"2025-11-24T10:00:00Z","level":"INFO","msg":"Server started","port":8080}
```

### Document Checksum (Optional)

Configuration for automatic checksum computation when creating documents from URLs:

```bash
# Maximum file size to download for checksum calculation (default: 10485760 = 10MB)
ACKIFY_CHECKSUM_MAX_BYTES=10485760

# Timeout for checksum download in milliseconds (default: 5000ms = 5s)
ACKIFY_CHECKSUM_TIMEOUT_MS=5000

# Maximum number of HTTP redirects to follow (default: 3)
ACKIFY_CHECKSUM_MAX_REDIRECTS=3

# Comma-separated list of allowed MIME types (default includes PDF, images, Office docs, ODF)
ACKIFY_CHECKSUM_ALLOWED_TYPES=application/pdf,image/*,application/msword,application/vnd.openxmlformats-officedocument.wordprocessingml.document,application/vnd.ms-excel,application/vnd.openxmlformats-officedocument.spreadsheetml.sheet,application/vnd.oasis.opendocument.*
```

**Note**: These settings only apply when admins create documents via the admin dashboard with a remote URL. The system will attempt to download and calculate the SHA-256 checksum automatically.

**Testing variables** (⚠️ **NEVER use in production**):
```bash
# Disable SSRF protection (testing only)
ACKIFY_CHECKSUM_SKIP_SSRF_CHECK=false

# Skip TLS certificate verification (testing only)
ACKIFY_CHECKSUM_INSECURE_SKIP_VERIFY=false
```

These variables disable critical security protections and should **only** be used in isolated test environments.

## Advanced Configuration

### OAuth2 Providers

See [OAuth Providers](configuration/oauth-providers.md) for detailed configuration of:
- Google OAuth2
- GitHub OAuth2
- GitLab OAuth2 (public + self-hosted)
- Custom OAuth2 provider

### Email (SMTP)

See [Email Setup](configuration/email-setup.md) to configure email reminders sending.

## Complete Example

Example `.env` for a production installation:

```bash
# Application
APP_DNS=sign.company.com
ACKIFY_BASE_URL=https://sign.company.com
ACKIFY_ORGANISATION="ACME Corporation"
ACKIFY_LOG_LEVEL=info
ACKIFY_LISTEN_ADDR=:8080

# Database
POSTGRES_USER=ackifyr
POSTGRES_PASSWORD=super_secure_password_123
POSTGRES_DB=ackify

# OAuth2 (Google)
ACKIFY_OAUTH_PROVIDER=google
ACKIFY_OAUTH_CLIENT_ID=123456789-abc.apps.googleusercontent.com
ACKIFY_OAUTH_CLIENT_SECRET=GOCSPX-xyz123
ACKIFY_OAUTH_ALLOWED_DOMAIN=@company.com

# Security
ACKIFY_OAUTH_COOKIE_SECRET=ZXhhbXBsZV9iYXNlNjRfc2VjcmV0X2tleQ==

# Administration
ACKIFY_ADMIN_EMAILS=admin@company.com,cto@company.com

# Email (optional - omit MAIL_HOST to disable)
ACKIFY_MAIL_HOST=smtp.gmail.com
ACKIFY_MAIL_PORT=587
ACKIFY_MAIL_USERNAME=noreply@company.com
ACKIFY_MAIL_PASSWORD=app_specific_password
ACKIFY_MAIL_FROM=noreply@company.com
ACKIFY_MAIL_FROM_NAME="Ackify - ACME"
ACKIFY_MAIL_TEMPLATE_DIR=templates/emails
ACKIFY_MAIL_DEFAULT_LOCALE=en

# Document Checksum (optional - for auto-checksum from URLs)
ACKIFY_CHECKSUM_MAX_BYTES=10485760
ACKIFY_CHECKSUM_TIMEOUT_MS=5000
ACKIFY_CHECKSUM_MAX_REDIRECTS=3
```

## Configuration Validation

After modifying `.env`, restart:

```bash
docker compose restart ackify-ce
```

Check logs:

```bash
docker compose logs -f ackify-ce
```

Test the health check:

```bash
curl http://localhost:8080/api/v1/health
```

## Production Variables

**Production security checklist**:

- ✅ Use HTTPS (`ACKIFY_BASE_URL=https://...`)
- ✅ Generate strong secrets (64+ characters)
- ✅ Restrict OAuth domain (`ACKIFY_OAUTH_ALLOWED_DOMAIN`)
- ✅ Configure admin emails (`ACKIFY_ADMIN_EMAILS`)
- ✅ Use PostgreSQL with SSL in production
- ✅ Log in `info` mode (not `debug`)
- ✅ Regularly backup the database

See [Deployment](deployment.md) for more details on production deployment.
