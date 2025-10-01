# 🔐 Ackify

> **Proof of Read. Compliance made simple.**

Secure document reading validation service with cryptographic traceability and irrefutable proof.

[![Build](https://img.shields.io/badge/build-passing-brightgreen.svg)](https://github.com/btouchard/ackify-ce)
[![Security](https://img.shields.io/badge/crypto-Ed25519-blue.svg)](https://en.wikipedia.org/wiki/EdDSA)
[![Go](https://img.shields.io/badge/go-1.24.5-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/License-AGPL_v3-blue.svg)](LICENSE)

> 🇫🇷 [Version française disponible ici](README_FR.md)

### Visite our website here : https://www.ackify.eu

## 🎯 Why Ackify?

**Problem**: How to prove that a collaborator has actually read and understood an important document?

**Solution**: Ed25519 cryptographic signatures with immutable timestamps and complete traceability.

### Real-world use cases
- ✅ Security policy validation
- ✅ Mandatory training attestations  
- ✅ GDPR acknowledgment
- ✅ Contractual acknowledgments
- ✅ Quality and compliance procedures

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
git clone https://github.com/btouchard/ackify-ce.git
cd ackify-ce

# Minimal configuration
cp .env.example .env
# Edit .env with your OAuth2 settings

# Start
docker compose up -d

# Test
curl http://localhost:8080/health   # alias: /healthz
```

### Required variables
```bash
ACKIFY_BASE_URL="https://your-domain.com"
ACKIFY_OAUTH_CLIENT_ID="your-oauth-client-id"        # Google/GitHub/GitLab
ACKIFY_OAUTH_CLIENT_SECRET="your-oauth-client-secret"
ACKIFY_DB_DSN="postgres://user:password@localhost/ackify?sslmode=disable"
ACKIFY_OAUTH_COOKIE_SECRET="$(openssl rand -base64 32)"
```

---

## 🚀 Simple Usage

### 1. Request a signature
```
https://your-domain.com/sign?doc=security_procedure_2025
```
→ User authenticates via OAuth2 and validates their reading

### 2. Verify signatures
```bash
# JSON API - Complete list
curl "https://your-domain.com/status?doc=security_procedure_2025"

# PNG Badge - Individual status  
curl "https://your-domain.com/status.png?doc=security_procedure_2025&user=john.doe@company.com"
```

### 3. Integrate into your pages
```html
<!-- Embeddable widget -->
<iframe src="https://your-domain.com/embed?doc=security_procedure_2025" 
        width="500" height="300"></iframe>

<!-- Via oEmbed -->
<script>
fetch('/oembed?url=https://your-domain.com/embed?doc=security_procedure_2025')
  .then(r => r.json())
  .then(data => document.getElementById('signatures').innerHTML = data.html);
</script>
```

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

---

## 🛡️ Security & Architecture

### Cryptographic security
- **Ed25519**: State-of-the-art digital signatures
- **SHA-256**: Payload hashing against tampering
- **Immutable timestamps**: PostgreSQL triggers
- **Encrypted sessions**: Secure cookies
- **CSP headers**: XSS protection

### Go architecture
```
cmd/ackapp/              # Entry point
internal/
  domain/                # Business logic
    models/              # Entities
    repositories/        # Persistence interfaces
  application/           # Use cases  
    services/            # Business implementations
  infrastructure/        # Adapters
    auth/               # OAuth2
    database/           # PostgreSQL
    config/             # Configuration
  presentation/          # HTTP
    handlers/           # Controllers + interfaces
    templates/          # HTML views
pkg/                    # Shared utilities
```

### Technology stack
- **Go 1.24.5**: Performance and simplicity
- **PostgreSQL**: Integrity constraints 
- **OAuth2**: Multi-provider
- **Docker**: Simplified deployment
- **Traefik**: HTTPS reverse proxy

---

## 📊 Database

```sql
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
```

**Guarantees**:
- ✅ **Uniqueness**: One user = one signature per document
- ✅ **Immutability**: `created_at` protected by trigger
- ✅ **Integrity**: SHA-256 hash to detect modifications
- ✅ **Non-repudiation**: Ed25519 signature cryptographically provable

---

## 🚀 Production Deployment

### docker-compose.yml
```yaml
version: '3.8'
services:
  ackapp:
    image: btouchard/ackify-ce:latest
    environment:
      ACKIFY_BASE_URL: https://ackify.company.com
      ACKIFY_DB_DSN: postgres://user:pass@postgres:5432/ackdb?sslmode=require
      ACKIFY_OAUTH_CLIENT_ID: ${ACKIFY_OAUTH_CLIENT_ID}
      ACKIFY_OAUTH_CLIENT_SECRET: ${ACKIFY_OAUTH_CLIENT_SECRET}
      ACKIFY_OAUTH_COOKIE_SECRET: ${ACKIFY_OAUTH_COOKIE_SECRET}
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.ackify.rule=Host(`ackify.company.com`)"
      - "traefik.http.routers.ackify.tls.certresolver=letsencrypt"

  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_DB: ackdb
      POSTGRES_USER: ackuser
      POSTGRES_PASSWORD: ${DB_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data
```

### Production variables
```bash
# Enhanced security
ACKIFY_OAUTH_COOKIE_SECRET="$(openssl rand 64 | base64 -w 0)"
ACKIFY_ED25519_PRIVATE_KEY="$(openssl rand 64 | base64 -w 0)"

# HTTPS mandatory
ACKIFY_BASE_URL="https://ackify.company.com"

# Secure PostgreSQL
ACKIFY_DB_DSN="postgres://user:pass@postgres:5432/ackdb?sslmode=require"
```

---

## 📋 Complete API

### Authentication
- `GET /login?next=<url>` - OAuth2 login
- `GET /logout` - Logout
- `GET /oauth2/callback` - OAuth2 callback

### Signatures  
- `GET /sign?doc=<id>` - Signature interface
- `POST /sign` - Create signature
- `GET /signatures` - My signatures (auth required)

### Consultation
- `GET /status?doc=<id>` - JSON all signatures
- `GET /status.png?doc=<id>&user=<email>` - PNG badge

### Integration
- `GET /oembed?url=<embed_url>` - oEmbed metadata  
- `GET /embed?doc=<id>` - HTML widget

### Monitoring
- `GET /health` - Health check (alias: `/healthz`)

### Admin
- `GET /admin` - Dashboard (restricted)
- `GET /admin/docs/{docID}` - Signatures for a document
- `GET /admin/api/chain-integrity/{docID}` - Chain integrity JSON

Access control: set `ACKIFY_ADMIN_EMAILS` with a comma-separated list of admin emails (exact match, case-insensitive). Example:
```bash
ACKIFY_ADMIN_EMAILS="alice@company.com,bob@company.com"
```

---

## 🔍 Development & Testing

### Local build
```bash
# Dependencies
go mod tidy

# Build
go build ./cmd/community

# Linting
go fmt ./...
go vet ./...

# Tests (TODO: add tests)
go test -v ./...
```

### Docker development
```bash
# Build image
docker build -t ackify-ce:dev .

# Run with local database
docker run -p 8080:8080 --env-file .env ackify-ce:dev

# Optional: static analysis
go install honnef.co/go/tools/cmd/staticcheck@latest
staticcheck ./...
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
