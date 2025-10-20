# 🔐 Ackify

> **Proof of Read. Compliance made simple.**

Secure document reading validation service with cryptographic traceability and irrefutable proof.

[![Build](https://github.com/btouchard/ackify-ce/actions/workflows/ci.yml/badge.svg)](https://github.com/btouchard/ackify-ce/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/btouchard/ackify-ce/branch/main/graph/badge.svg)](https://codecov.io/gh/btouchard/ackify-ce)
[![Security](https://img.shields.io/badge/crypto-Ed25519-blue.svg)](https://en.wikipedia.org/wiki/EdDSA)
[![Go](https://img.shields.io/badge/go-1.24.5-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/License-AGPL_v3-blue.svg)](LICENSE)

> 🇫🇷 [Version française disponible ici](README_FR.md)

### Visit our website here : https://www.ackify.eu

## 🎯 Why Ackify?

**Problem**: How to prove that a collaborator has actually read and understood an important document?

**Solution**: Ed25519 cryptographic signatures with immutable timestamps and complete traceability.

### Real-world use cases
- ✅ Security policy validation
- ✅ Mandatory training attestations
- ✅ GDPR acknowledgment
- ✅ Contractual acknowledgments
- ✅ Quality and compliance procedures

### Key Features

**Core Functionality**:
- Ed25519 cryptographic signatures with hash chain validation
- One signature per user per document (enforced by database constraints)
- OAuth2 authentication (Google, GitHub, GitLab, or custom provider)
- Public embeddable widgets for Notion, Outline, Google Docs, etc.

**Document Management**:
- Document metadata with title, URL, and description
- Checksum verification (SHA-256, SHA-512, MD5) for integrity tracking
- Verification history with timestamped audit trail
- Client-side checksum calculation with Web Crypto API

**Tracking & Reminders**:
- Expected signers list with completion tracking
- Email reminders in user's preferred language (fr, en, es, de, it)
- Visual progress bars and completion percentages
- Automatic detection of unexpected signatures

**Admin Dashboard**:
- Modern Vue.js 3 interface with dark mode
- Document management with bulk operations
- Signature tracking and analytics
- Expected signers management
- Email reminder system with history

**Integration & Embedding**:
- oEmbed support for automatic unfurling (Slack, Teams, etc.)
- Dynamic Open Graph and Twitter Card meta tags
- Public embed pages with signature buttons
- RESTful API v1 with OpenAPI specification
- PNG badges for README files and documentation

**Security & Compliance**:
- Immutable audit trail with PostgreSQL triggers
- CSRF protection and rate limiting (5 auth attempts/min, 10 document creates/min, 100 general requests/min)
- Encrypted sessions with secure cookies
- Content Security Policy (CSP) headers
- HTTPS enforcement in production

---

## 📸 Vidéos


Click to GIFs for open videos WebM in your browser.

<table>
<tr>
  <td align="center">
    <strong>1) Create sign</strong><br>
    <a href="screenshots/videos/1-initialize-sign.webm" target="_blank">
      <img src="screenshots/videos/1-initialize-sign.gif" width="380" alt="Initialisation d’une signature">
    </a>
  </td>
  <td align="center">
    <strong>2) User sign flow</strong><br>
    <a href="screenshots/videos/2-user-sign-flow.webm" target="_blank">
      <img src="screenshots/videos/2-user-sign-flow.gif" width="380" alt="Parcours de signature utilisateur">
    </a>
  </td>
  
</tr>
</table>

## 📸 Screenshots

<table>
<tr>
<td align="center">
<strong>Home page</strong><br>
<a href="screenshots/1-home.png"><img src="screenshots/1-home.png" width="200" alt="Home page"></a>
</td>
<td align="center">
<strong>Signing request</strong><br>
<a href="screenshots/2-signing-request.png"><img src="screenshots/2-signing-request.png" width="200" alt="Signing request"></a>
</td>
<td align="center">
<strong>Signature confirmed</strong><br>
<a href="screenshots/3-signing-ok.png"><img src="screenshots/3-signing-ok.png" width="200" alt="Signature confirmed"></a>
</td>
</tr>
<tr>
<td align="center">
<strong>Signatures list</strong><br>
<a href="screenshots/4-sign-list.png"><img src="screenshots/4-sign-list.png" width="200" alt="Signatures list"></a>
</td>
<td align="center">
<strong>Outline integration</strong><br>
<a href="screenshots/5-integrated-to-outline.png"><img src="screenshots/5-integrated-to-outline.png" width="200" alt="Outline integration"></a>
</td>
<td align="center">
<strong>Google Docs integration</strong><br>
<a href="screenshots/6-integrated-to-google-doc.png"><img src="screenshots/6-integrated-to-google-doc.png" width="200" alt="Google Docs integration"></a>
</td>
</tr>
</table>

---

## ⚡ Quick Start

### With Docker (recommended)
```bash
# Clone repository
git clone https://github.com/btouchard/ackify-ce.git
cd ackify-ce

# Configure environment
cp .env.example .env
# Edit .env with your OAuth2 settings (see configuration section below)

# Start services (PostgreSQL + Ackify)
docker compose up -d

# View logs
docker compose logs -f ackify-ce

# Verify deployment
curl http://localhost:8080/api/v1/health
# Expected: {"status": "healthy", "database": "connected"}

# Access web interface
open http://localhost:8080
# Modern Vue.js 3 SPA with dark mode support
```

**What's included**:
- PostgreSQL 16 with automatic migrations
- Ackify backend (Go) with embedded frontend
- Health monitoring endpoint
- Admin dashboard at `/admin`
- API documentation at `/api/openapi.yaml`

### Required Environment Variables

```bash
# Application Base URL (required - used for OAuth callbacks and embed URLs)
ACKIFY_BASE_URL="https://your-domain.com"

# Organization Name (required - used in email templates and display)
ACKIFY_ORGANISATION="Your Organization Name"

# OAuth2 Configuration (required)
ACKIFY_OAUTH_CLIENT_ID="your-oauth-client-id"
ACKIFY_OAUTH_CLIENT_SECRET="your-oauth-client-secret"

# Database Connection (required)
ACKIFY_DB_DSN="postgres://user:password@localhost/ackify?sslmode=disable"

# Session Security (required - generate with: openssl rand -base64 32)
ACKIFY_OAUTH_COOKIE_SECRET="$(openssl rand -base64 32)"
```

### Optional Environment Variables

**Email Notifications (SMTP)**:
```bash
ACKIFY_MAIL_HOST="smtp.gmail.com"              # SMTP server (if empty, email is disabled)
ACKIFY_MAIL_PORT="587"                         # SMTP port (default: 587)
ACKIFY_MAIL_USERNAME="your-email@gmail.com"    # SMTP authentication username
ACKIFY_MAIL_PASSWORD="your-app-password"       # SMTP authentication password
ACKIFY_MAIL_TLS="true"                         # Enable TLS (default: true)
ACKIFY_MAIL_STARTTLS="true"                    # Enable STARTTLS (default: true)
ACKIFY_MAIL_TIMEOUT="10s"                      # Connection timeout (default: 10s)
ACKIFY_MAIL_FROM="noreply@company.com"         # Email sender address
ACKIFY_MAIL_FROM_NAME="Ackify"                 # Email sender display name
ACKIFY_MAIL_SUBJECT_PREFIX=""                  # Optional prefix for email subjects
ACKIFY_MAIL_TEMPLATE_DIR="templates/emails"    # Email template directory (default: templates/emails)
ACKIFY_MAIL_DEFAULT_LOCALE="en"                # Default email locale (default: en)
```

**Server Configuration**:
```bash
ACKIFY_LISTEN_ADDR=":8080"                     # HTTP listen address (default: :8080)
ACKIFY_LOG_LEVEL="info"                        # Log level: debug, info, warn, error (default: info)
```

**Admin Access**:
```bash
ACKIFY_ADMIN_EMAILS="alice@company.com,bob@company.com"  # Comma-separated admin emails
```

**Cryptographic Keys**:
```bash
ACKIFY_ED25519_PRIVATE_KEY="$(openssl rand -base64 64)"  # Ed25519 signing key (optional, auto-generated if empty)
```

**OAuth2 Advanced**:
```bash
ACKIFY_OAUTH_AUTO_LOGIN="true"                 # Enable silent authentication (default: false)
ACKIFY_OAUTH_ALLOWED_DOMAIN="@company.com"     # Restrict to specific email domain
ACKIFY_OAUTH_LOGOUT_URL=""                     # Custom OAuth provider logout URL (optional)
```

**Templates & Locales**:
```bash
ACKIFY_TEMPLATES_DIR="/custom/path/to/templates"  # Custom template directory (optional)
ACKIFY_LOCALES_DIR="/custom/path/to/locales"      # Custom locales directory (optional)
```

---

## 🚀 Simple Usage

### 1. Request a signature
```
https://your-domain.com/?doc=security_procedure_2025
```
→ User authenticates via OAuth2 and validates their reading

### 2. Integrate into your pages

**Embeddable widget** (with signature button):
```html
<!-- The SPA handles the display -->
<iframe src="https://your-domain.com/?doc=security_procedure_2025"
        width="600" height="200"
        frameborder="0"
        style="border: 1px solid #ddd; border-radius: 6px;"></iframe>
```

**oEmbed support** (automatic unfurling in Notion, Outline, Confluence, etc.):
```html
<!-- Just paste the URL - platforms with oEmbed support will auto-discover and embed -->
https://your-domain.com/?doc=security_procedure_2025
```

The oEmbed endpoint (`/oembed`) is automatically discovered via the `<link rel="alternate" type="application/json+oembed">` meta tag.

**Manual oEmbed**:
```javascript
fetch('/oembed?url=https://your-domain.com/?doc=security_procedure_2025')
  .then(r => r.json())
  .then(data => {
    console.log(data.html);  // <iframe src="..." width="100%" height="200"></iframe>
    console.log(data.title); // Document title with signature count
  });
```

### 3. Dynamic Metadata for Unfurling

Ackify automatically generates **dynamic Open Graph, Twitter Card, and oEmbed discovery meta tags**:

```html
<!-- Auto-generated meta tags for /?doc=doc_id -->
<meta property="og:title" content="Document: security_procedure_2025 - 3 confirmations" />
<meta property="og:description" content="3 personnes ont confirmé avoir lu le document" />
<meta property="og:url" content="https://your-domain.com/?doc=doc_id" />
<meta property="og:type" content="website" />
<meta name="twitter:card" content="summary" />
<link rel="alternate" type="application/json+oembed"
      href="https://your-domain.com/oembed?url=https://your-domain.com/?doc=doc_id"
      title="Document: security_procedure_2025 - 3 confirmations" />
```

**Result**: When you paste a document URL in Slack, Teams, Discord, Notion, Outline, or social media:
- **Open Graph/Twitter**: Rich preview with title, description, signature count
- **oEmbed** (Notion, Outline, Confluence): Full interactive widget embedded in the page
- **No authentication required** on the public page, making it perfect for sharing progress publicly

---

## 🔧 OAuth2 Configuration

### Supported providers

| Provider | Configuration |
|----------|---------------|
| **Google** | `ACKIFY_OAUTH_PROVIDER=google` |
| **GitHub** | `ACKIFY_OAUTH_PROVIDER=github` |
| **GitLab** | `ACKIFY_OAUTH_PROVIDER=gitlab` + `ACKIFY_OAUTH_GITLAB_URL` |
| **Custom** | Custom endpoints |

### Custom provider
```bash
# Leave ACKIFY_OAUTH_PROVIDER empty
ACKIFY_OAUTH_AUTH_URL="https://auth.company.com/oauth/authorize"
ACKIFY_OAUTH_TOKEN_URL="https://auth.company.com/oauth/token"
ACKIFY_OAUTH_USERINFO_URL="https://auth.company.com/api/user"
ACKIFY_OAUTH_SCOPES="read:user,user:email"
```

### Domain restriction
```bash
ACKIFY_OAUTH_ALLOWED_DOMAIN="@company.com"  # Only @company.com emails
```

### Log level setup
```bash
ACKIFY_LOG_LEVEL="info" # can be debug, info, warn(ing), error. default: info
```

### Auto-login setup
```bash
ACKIFY_OAUTH_AUTO_LOGIN="true"  # Enable silent authentication when session exists (default: false)
```

---

## 🏗️ Project Structure

Ackify follows a **monorepo architecture** with clear separation between backend and frontend:

```
ackify-ce/
├── backend/              # Go backend (API-first)
│   ├── cmd/
│   │   ├── community/    # Main application entry point
│   │   └── migrate/      # Database migration tool
│   ├── internal/
│   │   ├── domain/       # Business entities (models)
│   │   ├── application/  # Business logic (services)
│   │   ├── infrastructure/ # Technical implementations
│   │   │   ├── auth/     # OAuth2 service
│   │   │   ├── database/ # PostgreSQL repositories
│   │   │   ├── email/    # SMTP service
│   │   │   ├── config/   # Configuration management
│   │   │   └── i18n/     # Backend internationalization
│   │   └── presentation/ # HTTP layer
│   │       ├── api/      # RESTful API v1 handlers
│   │       └── handlers/ # Legacy template handlers
│   ├── pkg/              # Shared utilities
│   │   ├── crypto/       # Ed25519 signatures
│   │   ├── logger/       # Structured logging
│   │   ├── services/     # OAuth provider detection
│   │   └── web/          # HTTP server setup
│   ├── migrations/       # SQL migrations
│   ├── locales/          # Backend translations (fr, en)
│   └── templates/        # Email templates (HTML/text)
├── webapp/               # Vue.js 3 SPA (frontend)
│   ├── src/
│   │   ├── components/   # Reusable Vue components (shadcn/vue)
│   │   ├── pages/        # Page components (router views)
│   │   ├── services/     # API client services
│   │   ├── stores/       # Pinia state management
│   │   ├── router/       # Vue Router configuration
│   │   ├── locales/      # Frontend translations (fr, en, es, de, it)
│   │   └── composables/  # Vue composables
│   ├── public/           # Static assets
│   └── scripts/          # Build & i18n scripts
├── api/                  # OpenAPI specification
│   └── openapi.yaml      # Complete API documentation
├── go.mod                # Go dependencies (at root)
└── go.sum
```

## 🛡️ Security & Architecture

### Modern API-First Architecture

Ackify uses a **modern, API-first architecture** with complete separation of concerns:

**Backend (Go)**:
- **RESTful API v1**: Versioned API (`/api/v1`) with structured JSON responses
- **Clean Architecture**: Domain-driven design with clear layer separation
- **OpenAPI Specification**: Complete API documentation in `/api/openapi.yaml`
- **Secure Authentication**: OAuth2 with session-based auth + CSRF protection
- **Rate Limiting**: Protection against abuse (5 auth attempts/min, 100 general requests/min)
- **Structured Logging**: JSON logs with request IDs for distributed tracing

**Frontend (Vue.js 3 SPA)**:
- **TypeScript**: Type-safe development with full IDE support
- **Vite**: Fast HMR and optimized production builds
- **Vue Router**: Client-side routing with lazy loading
- **Pinia**: Centralized state management
- **shadcn/vue**: Accessible, customizable UI components
- **Tailwind CSS**: Utility-first styling with dark mode support
- **vue-i18n**: 5 languages (fr, en, es, de, it) with automatic detection

### Cryptographic Security
- **Ed25519**: State-of-the-art digital signatures (elliptic curve)
- **SHA-256**: Payload hashing against tampering
- **Hash Chain**: Previous signature hash for integrity verification
- **Immutable Timestamps**: PostgreSQL triggers prevent backdating
- **Encrypted Sessions**: Secure cookies with HMAC-SHA256
- **CSP Headers**: Content Security Policy for XSS protection
- **CORS**: Configurable cross-origin resource sharing

### Build & Deployment

**Multi-stage Docker Build**:
1. **Stage 1 - Frontend Build**: Node.js 22 builds Vue.js 3 SPA with Vite
2. **Stage 2 - Backend Build**: Go (latest with GOTOOLCHAIN=auto) compiles backend and embeds frontend static assets
3. **Stage 3 - Runtime**: Distroless minimal image (< 30MB)

**Key Features**:
- **Server-side injection**: `ACKIFY_BASE_URL` injected into `index.html` at runtime
- **Static embedding**: Frontend assets embedded in Go binary using `embed.FS`
- **Single binary**: Backend serves both API and frontend (no separate web server needed)
- **Graceful shutdown**: Proper HTTP server lifecycle with signal handling
- **Production-ready**: Optimized builds with dead code elimination

**Build Process**:
```dockerfile
# Frontend build (webapp/)
FROM node:22-alpine AS frontend
COPY webapp/ /build/webapp/
RUN npm ci && npm run build
# Outputs to: /build/webapp/dist/

# Backend build (backend/)
FROM golang:alpine AS backend
ENV GOTOOLCHAIN=auto
COPY backend/ /build/backend/
COPY --from=frontend /build/webapp/dist/ /build/backend/cmd/community/web/dist/
RUN go build -o community ./cmd/community
# Embeds dist/ into Go binary via embed.FS

# Runtime
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=backend /build/backend/community /app/community
CMD ["/app/community"]
```

### Technology Stack

**Backend**:
- **Go 1.24.5+**: Performance, simplicity, and strong typing
- **PostgreSQL 16+**: ACID compliance with integrity constraints
- **Chi Router**: Lightweight, idiomatic Go HTTP router
- **OAuth2**: Multi-provider authentication (Google, GitHub, GitLab, custom)
- **Ed25519**: Elliptic curve digital signatures (crypto/ed25519)
- **SMTP**: Email reminders via standard library (optional)

**Frontend**:
- **Vue 3**: Modern reactive framework with Composition API
- **TypeScript**: Full type safety across the frontend
- **Vite**: Lightning-fast HMR and optimized production builds
- **Pinia**: Intuitive state management for Vue 3
- **Vue Router**: Client-side routing with code splitting
- **Tailwind CSS**: Utility-first styling with dark mode
- **shadcn/vue**: Accessible, customizable component library
- **vue-i18n**: Internationalization (FR, EN, ES, DE, IT)

**DevOps**:
- **Docker**: Multi-stage builds with Alpine Linux
- **PostgreSQL Migrations**: Version-controlled schema evolution
- **OpenAPI**: API documentation with Swagger UI

### Internationalization (i18n)

Ackify's web interface is fully internationalized with support for **5 languages**:

- **🇫🇷 French** (default)
- **🇬🇧 English** (fallback)
- **🇪🇸 Spanish**
- **🇩🇪 German**
- **🇮🇹 Italian**

**Features:**
- Language selector with Unicode flags in header
- Automatic detection from browser or localStorage
- Dynamic page titles with i18n
- Complete translation coverage verified by CI script
- All UI elements, ARIA labels, and metadata translated

**Documentation:** See [webapp/I18N.md](webapp/I18N.md) for complete i18n guide.

**Scripts:**
```bash
cd webapp
npm run lint:i18n  # Verify translation coverage
```

---

## 📊 Database

### Schema Management

Ackify uses **versioned SQL migrations** for schema evolution:

**Migration files**: Located in `/backend/migrations/`
- `0001_init.up.sql` - Initial schema (signatures table)
- `0002_expected_signers.up.sql` - Expected signers tracking
- `0003_reminder_logs.up.sql` - Email reminder history
- `0004_add_name_to_expected_signers.up.sql` - Display names for signers
- `0005_create_documents_table.up.sql` - Document metadata
- `0006_checksum_verifications.up.sql` - Checksum verification history

**Apply migrations**:
```bash
# Using Go migrate tool
cd backend
go run ./cmd/migrate

# Or manually with psql
psql $ACKIFY_DB_DSN -f migrations/0001_init.up.sql
```

**Docker Compose**: Migrations are applied automatically on container startup.

### Database Schema

```sql
-- Main signatures table
CREATE TABLE signatures (
    id BIGSERIAL PRIMARY KEY,
    doc_id TEXT NOT NULL,                    -- Document ID
    user_sub TEXT NOT NULL,                  -- OAuth user ID
    user_email TEXT NOT NULL,               -- User email
    signed_at TIMESTAMPTZ NOT NULL,     -- Signature timestamp
    payload_hash TEXT NOT NULL,         -- Cryptographic hash
    signature TEXT NOT NULL,            -- Ed25519 signature
    nonce TEXT NOT NULL,                    -- Anti-replay
    created_at TIMESTAMPTZ DEFAULT now(),   -- Immutable
    referer TEXT,                           -- Source (optional)
    prev_hash TEXT,
    UNIQUE (doc_id, user_sub)              -- One signature per user/doc
);

-- Expected signers table (for tracking)
CREATE TABLE expected_signers (
    id BIGSERIAL PRIMARY KEY,
    doc_id TEXT NOT NULL,
    email TEXT NOT NULL,
    name TEXT NOT NULL DEFAULT '',          -- Display name (optional)
    added_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    added_by TEXT NOT NULL,                 -- Admin who added
    notes TEXT,
    UNIQUE (doc_id, email)                  -- One expectation per email/doc
);

-- Document metadata table
CREATE TABLE documents (
    doc_id TEXT PRIMARY KEY,
    title TEXT NOT NULL DEFAULT '',
    url TEXT NOT NULL DEFAULT '',           -- Document location
    checksum TEXT NOT NULL DEFAULT '',      -- SHA-256/SHA-512/MD5
    checksum_algorithm TEXT NOT NULL DEFAULT 'SHA-256',
    description TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by TEXT NOT NULL DEFAULT ''
);
```

**Guarantees**:
- ✅ **Uniqueness**: One user = one signature per document
- ✅ **Immutability**: `created_at` protected by trigger
- ✅ **Integrity**: SHA-256 hash to detect modifications
- ✅ **Non-repudiation**: Ed25519 signature cryptographically provable
- ✅ **Tracking**: Expected signers for completion monitoring
- ✅ **Metadata**: Document information with URL, checksum, and description
- ✅ **Checksum verification**: Track document integrity with verification history

### Document Integrity Verification

Ackify supports document integrity verification through checksum tracking and verification:

**Supported algorithms**: SHA-256 (default), SHA-512, MD5

**Client-side verification** (recommended):
```javascript
// Calculate checksum in browser using Web Crypto API
async function calculateChecksum(file) {
  const arrayBuffer = await file.arrayBuffer();
  const hashBuffer = await crypto.subtle.digest('SHA-256', arrayBuffer);
  const hashArray = Array.from(new Uint8Array(hashBuffer));
  return hashArray.map(b => b.toString(16).padStart(2, '0')).join('');
}
```

**Manual checksum calculation**:
```bash
# Linux/Mac
sha256sum document.pdf
sha512sum document.pdf
md5sum document.pdf

# Windows PowerShell
Get-FileHash document.pdf -Algorithm SHA256
Get-FileHash document.pdf -Algorithm SHA-512
Get-FileHash document.pdf -Algorithm MD5
```

**Note**: Checksum values are stored as metadata and can be viewed/updated through the admin document management interface. Verification is typically done client-side using the Web Crypto API or command-line tools shown above.

---

## 🚀 Production Deployment

### compose.yml
```yaml
name: ackify
services:
  ackify-migrate:
    image: btouchard/ackify-ce
    container_name: ackify-ce-migrate
    environment:
      ACKIFY_BASE_URL: https://ackify.company.com
      ACKIFY_ORGANISATION: Company
      ACKIFY_DB_DSN: postgres://user:pass@postgres:5432/ackdb?sslmode=require
      ACKIFY_OAUTH_CLIENT_ID: ${ACKIFY_OAUTH_CLIENT_ID}
      ACKIFY_OAUTH_CLIENT_SECRET: ${ACKIFY_OAUTH_CLIENT_SECRET}
      ACKIFY_OAUTH_COOKIE_SECRET: ${ACKIFY_OAUTH_COOKIE_SECRET}
    depends_on:
      ackify-db:
        condition: service_healthy
    networks:
      - internal
    command: ["/app/migrate", "up"]
    entrypoint: []
    restart: "no"

  ackify-ce:
    image: btouchard/ackify-ce:latest
    environment:
      ACKIFY_BASE_URL: https://ackify.company.com
      ACKIFY_ORGANISATION: Company
      ACKIFY_DB_DSN: postgres://user:pass@postgres:5432/ackdb?sslmode=require
      ACKIFY_OAUTH_CLIENT_ID: ${ACKIFY_OAUTH_CLIENT_ID}
      ACKIFY_OAUTH_CLIENT_SECRET: ${ACKIFY_OAUTH_CLIENT_SECRET}
      ACKIFY_OAUTH_COOKIE_SECRET: ${ACKIFY_OAUTH_COOKIE_SECRET}
    depends_on:
      ackify-migrate:
        condition: service_completed_successfully
      ackify-db:
        condition: service_healthy
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.ackify.rule=Host(`ackify.company.com`)"
      - "traefik.http.routers.ackify.tls.certresolver=letsencrypt"

  ackify-db:
    image: postgres:15-alpine
    environment:
      POSTGRES_DB: ackdb
      POSTGRES_USER: ackuser
      POSTGRES_PASSWORD: ${DB_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ${POSTGRES_USER} -d ${POSTGRES_DB}"]
      interval: 10s
      timeout: 5s
      retries: 5
```

### Production Environment Variables
```bash
# Enhanced security - generate strong secrets
ACKIFY_OAUTH_COOKIE_SECRET="$(openssl rand 64 | base64 -w 0)"
ACKIFY_ED25519_PRIVATE_KEY="$(openssl rand 64 | base64 -w 0)"

# HTTPS mandatory in production
ACKIFY_BASE_URL="https://ackify.company.com"

# Secure PostgreSQL with SSL
ACKIFY_DB_DSN="postgres://ackuser:strong_password@postgres:5432/ackdb?sslmode=require"

# Admin access (comma-separated emails)
ACKIFY_ADMIN_EMAILS="admin@company.com,cto@company.com"

# Email reminders (optional but recommended)
ACKIFY_MAIL_HOST="smtp.company.com"
ACKIFY_MAIL_PORT="587"
ACKIFY_MAIL_FROM="noreply@company.com"
ACKIFY_MAIL_FROM_NAME="Ackify - Company Name"
ACKIFY_MAIL_USERNAME="${SMTP_USERNAME}"
ACKIFY_MAIL_PASSWORD="${SMTP_PASSWORD}"

# OAuth2 configuration (example with Google)
ACKIFY_OAUTH_PROVIDER="google"
ACKIFY_OAUTH_CLIENT_ID="${GOOGLE_CLIENT_ID}"
ACKIFY_OAUTH_CLIENT_SECRET="${GOOGLE_CLIENT_SECRET}"
ACKIFY_OAUTH_ALLOWED_DOMAIN="@company.com"  # Restrict to company domain

# Logging
ACKIFY_LOG_LEVEL="info"  # Use "debug" for troubleshooting
```

### Production Tips

**Security Checklist**:
- ✅ Use HTTPS (required for secure cookies)
- ✅ Enable PostgreSQL SSL (`sslmode=require`)
- ✅ Generate strong secrets (64+ bytes)
- ✅ Restrict OAuth to company domain
- ✅ Set up admin emails list
- ✅ Monitor logs for suspicious activity
- ✅ Regular PostgreSQL backups

**Performance Optimization**:
- PostgreSQL connection pooling (handled by Go)
- CDN for static assets (if hosting separately)
- Database indexes on `(doc_id, user_sub)`
- Rate limiting enabled by default

**Monitoring**:
- Health endpoint: `GET /api/v1/health` (includes DB status)
- Structured JSON logs with request IDs
- Database metrics via PostgreSQL `pg_stat_statements`

---

## 📋 API Documentation

### OpenAPI Specification

Complete API documentation is available in OpenAPI 3.0 format:

**📁 File**: `/api/openapi.yaml`

**Features**:
- Full API v1 endpoint documentation
- Request/response schemas
- Authentication requirements
- Example payloads
- Error responses

**View online**: You can import the OpenAPI spec into:
- [Swagger Editor](https://editor.swagger.io/) - Paste the YAML content
- [Postman](https://www.postman.com/) - Import as OpenAPI 3.0
- [Insomnia](https://insomnia.rest/) - Import as OpenAPI spec
- Any OpenAPI-compatible tool

**Local viewing**:
```bash
# Using swagger-ui Docker image
docker run -p 8081:8080 -e SWAGGER_JSON=/api/openapi.yaml \
  -v $(pwd)/api:/api swaggerapi/swagger-ui
# Open http://localhost:8081
```

### API v1 Endpoints

All API v1 endpoints are prefixed with `/api/v1` and return JSON responses with standard HTTP status codes.

**Base URL structure**:
- Development: `http://localhost:8080/api/v1`
- Production: `https://your-domain.com/api/v1`

#### System & Health
- `GET /api/v1/health` - Health check with database status (public)

#### Authentication
- `POST /api/v1/auth/start` - Initiate OAuth flow (returns redirect URL)
- `GET /api/v1/auth/logout` - Logout and clear session
- `GET /api/v1/auth/check` - Check authentication status (only if auto-login enabled)
- `GET /api/v1/csrf` - Get CSRF token for authenticated requests

#### Users
- `GET /api/v1/users/me` - Get current user profile (authenticated)

#### Documents (Public)
- `GET /api/v1/documents` - List all documents with pagination
- `POST /api/v1/documents` - Create new document (requires CSRF token, rate-limited to 10/min)
- `GET /api/v1/documents/{docId}` - Get document details with signatures
- `GET /api/v1/documents/{docId}/signatures` - Get document signatures
- `GET /api/v1/documents/{docId}/expected-signers` - Get expected signers list
- `GET /api/v1/documents/find-or-create?ref={reference}` - Find or create document by reference (conditional auth for embed support)

#### Signatures
- `GET /api/v1/signatures` - Get current user's signatures with pagination (authenticated)
- `POST /api/v1/signatures` - Create new signature (authenticated + CSRF token)
- `GET /api/v1/documents/{docId}/signatures/status` - Get user's signature status (authenticated)

#### Admin Endpoints
All admin endpoints require authentication + admin privileges + CSRF token.

**Documents**:
- `GET /api/v1/admin/documents?limit=100&offset=0` - List all documents with stats
- `GET /api/v1/admin/documents/{docId}` - Get document details (admin view)
- `GET /api/v1/admin/documents/{docId}/signers` - Get document with signers and completion stats
- `GET /api/v1/admin/documents/{docId}/status` - Get document status with completion stats
- `PUT /api/v1/admin/documents/{docId}/metadata` - Create/update document metadata
- `DELETE /api/v1/admin/documents/{docId}` - Delete document entirely (including metadata and signatures)

**Expected Signers**:
- `POST /api/v1/admin/documents/{docId}/signers` - Add expected signer
- `DELETE /api/v1/admin/documents/{docId}/signers/{email}` - Remove expected signer

**Email Reminders**:
- `POST /api/v1/admin/documents/{docId}/reminders` - Send email reminders to pending readers
- `GET /api/v1/admin/documents/{docId}/reminders` - Get reminder history

### Legacy Endpoints (Server-side Rendering)

These endpoints serve server-rendered HTML or specialized content:

**Authentication**:
- `GET /api/v1/auth/callback` - OAuth2 callback handler

**Public Routes**:
- `GET /` - Vue.js 3 SPA (serves all frontend routes with query params: `/?doc=xxx`, `/signatures`, `/admin`, etc.)
- `GET /health` - Health check (alias for backward compatibility)
- `GET /oembed?url=<document_url>` - oEmbed endpoint for automatic embed discovery (returns JSON with iframe HTML pointing to `/embed?doc=xxx`)

### API Usage Examples

**Get CSRF token** (required for authenticated POST/PUT/DELETE):
```bash
curl -c cookies.txt http://localhost:8080/api/v1/csrf
# Returns: {"csrf_token": "..."}
```

**Initiate OAuth login**:
```bash
curl -X POST http://localhost:8080/api/v1/auth/start \
  -H "Content-Type: application/json" \
  -d '{"redirect_to": "/?doc=policy_2025"}'
# Returns: {"redirect_url": "https://accounts.google.com/..."}
```

**Get current user profile**:
```bash
curl -b cookies.txt http://localhost:8080/api/v1/users/me
# Returns: {"sub": "...", "email": "...", "name": "...", "is_admin": false}
```

**Create a signature**:
```bash
curl -X POST http://localhost:8080/api/v1/signatures \
  -b cookies.txt \
  -H "Content-Type: application/json" \
  -H "X-CSRF-Token: YOUR_CSRF_TOKEN" \
  -d '{"doc_id": "policy_2025"}'
# Returns: {"doc_id": "policy_2025", "user_email": "...", "signed_at": "..."}
```

**List documents with signatures**:
```bash
curl http://localhost:8080/api/v1/documents?limit=10&offset=0
# Returns: {"documents": [...], "total": 42}
```

**Get document signatures** (public):
```bash
curl http://localhost:8080/api/v1/documents/policy_2025/signatures
# Returns: {"doc_id": "policy_2025", "signatures": [...]}
```

**Admin: Add expected signers**:
```bash
curl -X POST http://localhost:8080/api/v1/admin/documents/policy_2025/signers \
  -b cookies.txt \
  -H "Content-Type: application/json" \
  -H "X-CSRF-Token: YOUR_CSRF_TOKEN" \
  -d '{"email": "john@company.com", "name": "John Doe"}'
```

**Admin: Send email reminders**:
```bash
curl -X POST http://localhost:8080/api/v1/admin/documents/policy_2025/reminders \
  -b cookies.txt \
  -H "Content-Type: application/json" \
  -H "X-CSRF-Token: YOUR_CSRF_TOKEN" \
  -d '{"emails": ["john@company.com", "jane@company.com"]}'
# Returns: {"sent": 2, "failed": 0, "errors": []}
```

**Admin: Update document metadata**:
```bash
curl -X PUT http://localhost:8080/api/v1/admin/documents/policy_2025/metadata \
  -b cookies.txt \
  -H "Content-Type: application/json" \
  -H "X-CSRF-Token: YOUR_CSRF_TOKEN" \
  -d '{"title": "Security Policy 2025", "url": "https://docs.company.com/policy", "checksum": "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", "checksum_algorithm": "SHA-256", "description": "Company security policy"}'
# Returns: {"docId": "policy_2025", "title": "Security Policy 2025", ...}
```

**oEmbed Discovery** (automatic embedding in modern editors):
```bash
# Get oEmbed data for a document URL
curl "http://localhost:8080/oembed?url=http://localhost:8080/?doc=policy_2025"
# Returns:
# {
#   "type": "rich",
#   "version": "1.0",
#   "title": "Document policy_2025 - Confirmations de lecture",
#   "provider_name": "Ackify",
#   "provider_url": "http://localhost:8080",
#   "html": "<iframe src=\"http://localhost:8080/embed?doc=policy_2025\" width=\"100%\" height=\"200\" frameborder=\"0\" style=\"border: 1px solid #ddd; border-radius: 6px;\" allowtransparency=\"true\"></iframe>",
#   "height": 200
# }
```

**How it works**:
1. User pastes `https://your-domain.com/?doc=policy_2025` in Notion, Outline, Confluence, etc.
2. The editor discovers the oEmbed endpoint via the `<link rel="alternate" type="application/json+oembed">` meta tag
3. The editor calls `/oembed?url=https://your-domain.com/?doc=policy_2025`
4. Ackify returns JSON with an iframe pointing to `/embed?doc=policy_2025`
5. The editor displays the embedded signature widget

**Supported platforms**: Notion, Outline, Confluence, AppFlowy, and any platform supporting oEmbed discovery.

### Access Control

Set `ACKIFY_ADMIN_EMAILS` with a comma-separated list of admin emails (exact match, case-insensitive):
```bash
ACKIFY_ADMIN_EMAILS="alice@company.com,bob@company.com"
```

**Admin features**:
- Document metadata management (title, URL, checksum, description)
- Expected signers tracking with completion stats
- Email reminders with history
- Document deletion (including all metadata and signatures)
- Full document and signature statistics

#### Document Metadata Management
Administrators can manage comprehensive metadata for each document:
- **Store document information**: Title, URL/location, checksum, description
- **Integrity verification**: Support for SHA-256, SHA-512, and MD5 checksums
- **Easy access**: One-click copy for checksums, clickable document URLs
- **Automatic timestamps**: Track creation and updates with PostgreSQL triggers
- **Email integration**: Document URL automatically included in reminder emails

#### Expected Signers Feature
Administrators can define and track expected signers for each document:
- **Add expected signers**: Paste emails separated by newlines, commas, or semicolons
- **Support names**: Use format "Name <email@example.com>" for personalized emails
- **Track completion**: Visual progress bar with completion percentage
- **Monitor status**: See who signed (✓) vs. who is pending (⏳)
- **Email reminders**: Send bulk or selective reminders in user's language
- **Detect unexpected signatures**: Identify users who signed but weren't expected
- **Share easily**: One-click copy of document signature link
- **Bulk management**: Add/remove signers individually or in batch

---

## 🔍 Development & Testing

### Test Coverage

**Current Status**: **72.6% code coverage** (unit + integration tests)

Our comprehensive test suite includes:
- ✅ **180+ unit tests** covering business logic, services, and utilities
- ✅ **33 integration tests** with PostgreSQL for repository layer
- ✅ **Ed25519 cryptography** tests (90% coverage)
- ✅ **HTTP handlers & middleware** tests (80%+ coverage)
- ✅ **Domain models** tests (100% coverage)
- ✅ **Email services** tests with mocks
- ✅ **OAuth2 security** tests with edge cases

**Coverage by Package**:
| Package | Coverage | Status |
|---------|----------|--------|
| `domain/models` | 100% | ✅ Complete |
| `presentation/api/health` | 100% | ✅ Complete |
| `presentation/api/users` | 100% | ✅ Complete |
| `pkg/logger` | 100% | ✅ Complete |
| `pkg/services` | 100% | ✅ Complete |
| `presentation/api/signatures` | 95.2% | ✅ Excellent |
| `presentation/api/auth` | 92.3% | ✅ Excellent |
| `application/services` | 90.6% | ✅ Excellent |
| `pkg/crypto` | 90.0% | ✅ Excellent |
| `presentation/handlers` | 85.6% | ✅ Very Good |
| `presentation/api/admin` | 84.2% | ✅ Very Good |
| `presentation/api/shared` | 80.0% | ✅ Very Good |

All tests run automatically in **GitHub Actions CI/CD** on every push and pull request. Coverage reports are uploaded to Codecov for tracking and analysis.

### Local Development Setup

**Prerequisites**:
- Go 1.24.5+
- Node.js 22+ and npm
- PostgreSQL 16+
- Docker & Docker Compose (optional but recommended)

**Backend development**:
```bash
# Navigate to backend
cd backend

# Install Go dependencies
go mod download

# Build backend
go build ./cmd/community

# Run database migrations
go run ./cmd/migrate

# Run backend (port 8080)
./community

# Linting & formatting
go fmt ./...
go vet ./...

# Run unit tests only
go test -v -short ./...

# Run unit tests with coverage
go test -coverprofile=coverage.out ./internal/... ./pkg/...

# Run integration tests (requires PostgreSQL)
docker compose -f ../compose.test.yml up -d
INTEGRATION_TESTS=1 go test -tags=integration -v ./internal/infrastructure/database/
docker compose -f ../compose.test.yml down

# Run all tests (unit + integration) with coverage
docker compose -f ../compose.test.yml up -d
INTEGRATION_TESTS=1 go test -tags=integration -coverprofile=coverage.out ./...
docker compose -f ../compose.test.yml down

# View coverage report in browser
go tool cover -html=coverage.out

# View coverage summary
go tool cover -func=coverage.out | tail -1

# Optional: static analysis
go install honnef.co/go/tools/cmd/staticcheck@latest
staticcheck ./...
```

**Frontend development**:
```bash
# Navigate to webapp
cd webapp

# Install dependencies
npm install

# Run dev server (port 5173 with HMR)
npm run dev

# Build for production
npm run build

# Preview production build
npm run preview

# Type checking
npm run type-check

# Verify i18n completeness
npm run lint:i18n
```

### Docker Development

**Option 1: Full stack with Docker Compose** (recommended):
```bash
# Development with hot reload
docker compose -f compose.local.yml up -d

# View logs
docker compose -f compose.local.yml logs -f ackify-ce

# Rebuild after changes
docker compose -f compose.local.yml up -d --force-recreate ackify-ce --build

# Stop
docker compose -f compose.local.yml down
```

**Option 2: Build and run manually**:
```bash
# Build production image
docker build -t ackify-ce:dev .

# Run with environment file
docker run -p 8080:8080 --env-file .env ackify-ce:dev

# Run with PostgreSQL
docker compose up -d
```

### Project Commands (Makefile)

```bash
# Build everything (backend + frontend)
make build

# Run tests
make test

# Clean build artifacts
make clean

# Format code
make fmt

# Run linting
make lint
```

---

## 🤝 Support

### Help & Documentation
- 🐛 **Issues**: [GitHub Issues](https://github.com/btouchard/ackify-ce/issues)
- 💬 **Discussions**: [GitHub Discussions](https://github.com/btouchard/ackify-ce/discussions)

### License (AGPLv3)
Distributed under the GNU Affero General Public License v3.0.
See [LICENSE](LICENSE) for details.

---

**Developed with ❤️ by [Benjamin TOUCHARD](https://www.kolapsis.com)**
