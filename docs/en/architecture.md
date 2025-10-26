# Architecture

Technical stack and design principles of Ackify.

## Overview

Ackify is a **modern monolithic application** with clear backend/frontend separation.

```
┌─────────────────────────────────────────┐
│           Client Browser                │
│  (Vue.js 3 SPA + TypeScript)            │
└──────────────┬──────────────────────────┘
               │ HTTPS / JSON
┌──────────────▼──────────────────────────┐
│         Go Backend (API-first)          │
│  ├─ RESTful API v1 (chi router)         │
│  ├─ OAuth2 Service                      │
│  ├─ Ed25519 Crypto                      │
│  └─ SMTP Email (optional)               │
└──────────────┬──────────────────────────┘
               │ PostgreSQL protocol
┌──────────────▼──────────────────────────┐
│       PostgreSQL 16 Database            │
│  (Signatures + Metadata + Sessions)     │
└─────────────────────────────────────────┘
```

## Backend (Go)

### Simplified Clean Architecture

```
backend/
├── cmd/
│   ├── community/        # Entry point + dependency injection
│   └── migrate/          # SQL migrations tool
├── internal/
│   ├── domain/
│   │   └── models/       # Business entities (User, Signature, Document)
│   ├── application/
│   │   └── services/     # Business logic (SignatureService, etc.)
│   ├── infrastructure/
│   │   ├── auth/         # OAuth2 service
│   │   ├── database/     # PostgreSQL repositories
│   │   ├── email/        # SMTP service
│   │   ├── config/       # Environment variables
│   │   └── i18n/         # Backend i18n
│   └── presentation/
│       ├── api/          # HTTP handlers API v1
│       └── handlers/     # Legacy OAuth handlers
├── pkg/
│   ├── crypto/           # Ed25519 signatures
│   ├── logger/           # Structured logging
│   ├── services/         # OAuth provider detection
│   └── web/              # HTTP server setup
├── migrations/           # SQL migrations
├── templates/            # Email templates (HTML/text)
└── locales/              # Backend translations
```

### Applied Go Principles

**Interfaces**:
- ✅ Defined in the package that uses them
- ✅ Principle "accept interfaces, return structs"
- ✅ Repositories implemented in `infrastructure/database/`

**Dependency Injection**:
- ✅ Explicit constructors in `main.go`
- ✅ No complex DI container
- ✅ Clear and visible dependencies

**Code Quality**:
- ✅ `go fmt` and `go vet` clean
- ✅ No dead code
- ✅ Simple and focused interfaces

## Frontend (Vue.js 3)

### SPA Structure

```
webapp/
├── src/
│   ├── components/       # Reusable components
│   │   ├── ui/          # shadcn/vue components
│   │   └── ...
│   ├── pages/           # Pages (router views)
│   │   ├── Home.vue
│   │   ├── Admin.vue
│   │   └── ...
│   ├── services/        # API client (axios)
│   ├── stores/          # Pinia state management
│   ├── router/          # Vue Router config
│   ├── locales/         # Translations (fr, en, es, de, it)
│   └── composables/     # Vue composables
├── public/              # Static assets
└── scripts/             # Build scripts
```

### Frontend Stack

- **Vue 3** - Composition API
- **TypeScript** - Type safety
- **Vite** - Build tool (fast HMR)
- **Pinia** - State management
- **Vue Router** - Client routing
- **Tailwind CSS** - Utility-first styling
- **shadcn/vue** - UI components
- **vue-i18n** - Internationalization

### Routing

```typescript
const routes = [
  { path: '/', component: Home },               // Public
  { path: '/signatures', component: MySignatures }, // Auth required
  { path: '/admin', component: Admin }          // Admin only
]
```

Frontend handles:
- Route `/` with query param `?doc=xxx` → Signature page
- Route `/admin` → Admin dashboard
- Route `/signatures` → My signatures

## Database

### PostgreSQL Schema

Main tables:
- `signatures` - Ed25519 signatures
- `documents` - Document metadata
- `expected_signers` - Signer tracking
- `reminder_logs` - Email history
- `checksum_verifications` - Integrity verifications
- `oauth_sessions` - OAuth2 sessions + refresh tokens

See [Database](database.md) for complete schema.

### Migrations

- Format: `XXXX_description.up.sql` / `XXXX_description.down.sql`
- Applied automatically on startup (service `ackify-migrate`)
- Tool: `/backend/cmd/migrate`

## Security

### Cryptography

**Ed25519**:
- Digital signatures (elliptic curve)
- 256-bit private key
- Guaranteed non-repudiation

**SHA-256**:
- Payload hashing before signing
- Tampering detection
- Blockchain-like chaining (`prev_hash`)

**AES-256-GCM**:
- OAuth2 refresh token encryption
- Key derived from `ACKIFY_OAUTH_COOKIE_SECRET`

### OAuth2 + PKCE

**Flow**:
1. Client generates `code_verifier` (random)
2. Calculates `code_challenge = SHA256(code_verifier)`
3. Auth request with `code_challenge`
4. Provider returns `code`
5. Token exchange with `code + code_verifier`

**Security**:
- Protection against code interception
- S256 method (SHA-256)
- Automatically enabled

### Sessions

- Secure cookies (HttpOnly, Secure, SameSite=Lax)
- HMAC-SHA256 encryption
- PostgreSQL storage with encrypted refresh tokens
- Duration: 30 days
- Automatic cleanup: 37 days

## Build & Deployment

### Multi-Stage Docker

```dockerfile
# Stage 1 - Frontend build
FROM node:22-alpine AS frontend
COPY webapp/ /build/webapp/
RUN npm ci && npm run build
# Output: webapp/dist/

# Stage 2 - Backend build + embed frontend
FROM golang:alpine AS backend
ENV GOTOOLCHAIN=auto
COPY backend/ /build/backend/
COPY --from=frontend /build/webapp/dist/ /build/backend/cmd/community/web/dist/
RUN go build -o community ./cmd/community
# Frontend embedded via embed.FS

# Stage 3 - Runtime (distroless)
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=backend /build/backend/community /app/community
CMD ["/app/community"]
```

**Result**:
- Final image < 30 MB
- Single binary (backend + frontend)
- No runtime dependencies

### Runtime Injection

`ACKIFY_BASE_URL` is injected into `index.html` at startup:

```go
// Replaces __ACKIFY_BASE_URL__ with actual value
html = strings.ReplaceAll(html, "__ACKIFY_BASE_URL__", baseURL)
```

Allows changing domain without rebuild.

## Performance

### Backend

- **Connection pooling** PostgreSQL (25 max)
- **Prepared statements** - SQL injection prevention
- **Rate limiting** - 5 auth/min, 10 doc/min, 100 req/min
- **Structured logging** - JSON with request IDs

### Frontend

- **Code splitting** - Lazy loading routes
- **Tree shaking** - Dead code elimination
- **Minification** - Optimized production builds
- **HMR** - Hot Module Replacement (dev)

### Database

- **Indexes** on (doc_id, user_sub, session_id)
- **Constraints** UNIQUE for guarantees
- **Triggers** for immutability
- **Autovacuum** enabled

## Scalability

### Current Limits

- ✅ Monolith: ~10k req/s
- ✅ PostgreSQL: Single instance
- ✅ Sessions: In-database (no Redis)

### Horizontal Scaling (future)

For > 100k req/s:
1. **Load Balancer** - Multiple backend instances
2. **PostgreSQL read replicas** - Separate read/write
3. **Redis** - Session cache + rate limiting
4. **CDN** - Static assets

## Monitoring

### Structured Logs

JSON format:
```json
{
  "level": "info",
  "timestamp": "2025-01-15T14:30:00Z",
  "request_id": "abc123",
  "method": "POST",
  "path": "/api/v1/signatures",
  "duration_ms": 42,
  "status": 201
}
```

### Health Check

```http
GET /api/v1/health
```

Response:
```json
{
  "status": "healthy",
  "database": "connected"
}
```

### Metrics (future)

- Prometheus metrics endpoint
- Grafana dashboards
- Alerting (PagerDuty, Slack)

## Tests

### Coverage

**72.6% code coverage** (unit + integration)

- Unit tests: 180+ tests
- Integration tests: 33 PostgreSQL tests
- CI/CD: GitHub Actions + Codecov

See [Development](development.md) to run tests.

## Technical Choices

### Why Go?

- ✅ Native performance (compiled)
- ✅ Simple concurrency (goroutines)
- ✅ Strong typing
- ✅ Single binary
- ✅ Simple deployment

### Why Vue 3?

- ✅ Modern Composition API
- ✅ Native TypeScript
- ✅ Reactive by default
- ✅ Rich ecosystem
- ✅ Excellent performance

### Why PostgreSQL?

- ✅ ACID compliance
- ✅ Integrity constraints
- ✅ Triggers
- ✅ JSON support
- ✅ Mature and stable

### Why Ed25519?

- ✅ Modern security (elliptic curve)
- ✅ Performance > RSA
- ✅ Short signatures (64 bytes)
- ✅ Standard crypto/ed25519 Go

## References

- [Go Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [Vue 3 Composition API](https://vuejs.org/guide/extras/composition-api-faq.html)
- [Ed25519 Spec](https://ed25519.cr.yp.to/)
- [OAuth2 + PKCE](https://oauth.net/2/pkce/)
