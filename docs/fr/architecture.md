# Architecture

Stack technique et principes de conception d'Ackify.

## Vue d'Ensemble

Ackify est une **application monolithique moderne** avec séparation claire backend/frontend.

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
│  └─ SMTP Email (optionnel)              │
└──────────────┬──────────────────────────┘
               │ PostgreSQL protocol
┌──────────────▼──────────────────────────┐
│       PostgreSQL 16 Database            │
│  (Signatures + Metadata + Sessions)     │
└─────────────────────────────────────────┘
```

## Backend (Go)

### Clean Architecture Simplifiée

```
backend/
├── cmd/
│   ├── community/        # Point d'entrée + injection dépendances
│   └── migrate/          # Outil migrations SQL
├── internal/
│   ├── domain/
│   │   └── models/       # Entités métier (User, Signature, Document)
│   ├── application/
│   │   └── services/     # Logique métier (SignatureService, etc.)
│   ├── infrastructure/
│   │   ├── auth/         # OAuth2 service
│   │   ├── database/     # Repositories PostgreSQL
│   │   ├── email/        # SMTP service
│   │   ├── config/       # Variables d'environnement
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

### Principes Go Appliqués

**Interfaces** :
- ✅ Définies dans le package qui les utilise
- ✅ Principe "accept interfaces, return structs"
- ✅ Repositories implémentés dans `infrastructure/database/`

**Injection de Dépendances** :
- ✅ Constructeurs explicites dans `main.go`
- ✅ Pas de container DI complexe
- ✅ Dépendances claires et visibles

**Code Quality** :
- ✅ `go fmt` et `go vet` clean
- ✅ Pas de code mort
- ✅ Interfaces simples et focalisées

## Frontend (Vue.js 3)

### Structure SPA

```
webapp/
├── src/
│   ├── components/       # Composants réutilisables
│   │   ├── ui/          # shadcn/vue components
│   │   └── ...
│   ├── pages/           # Pages (router views)
│   │   ├── Home.vue
│   │   ├── Admin.vue
│   │   └── ...
│   ├── services/        # API client (axios)
│   ├── stores/          # Pinia state management
│   ├── router/          # Vue Router config
│   ├── locales/         # Traductions (fr, en, es, de, it)
│   └── composables/     # Vue composables
├── public/              # Assets statiques
└── scripts/             # Build scripts
```

### Stack Frontend

- **Vue 3** - Composition API
- **TypeScript** - Type safety
- **Vite** - Build tool (HMR rapide)
- **Pinia** - State management
- **Vue Router** - Client routing
- **Tailwind CSS** - Utility-first styling
- **shadcn/vue** - UI components
- **vue-i18n** - Internationalisation

### Routing

```typescript
const routes = [
  { path: '/', component: Home },               // Public
  { path: '/signatures', component: MySignatures }, // Auth required
  { path: '/admin', component: Admin }          // Admin only
]
```

Le frontend gère :
- Route `/` avec query param `?doc=xxx` → Page de signature
- Route `/admin` → Dashboard admin
- Route `/signatures` → Mes signatures

## Base de Données

### Schéma PostgreSQL

Tables principales :
- `signatures` - Signatures Ed25519
- `documents` - Métadonnées documents
- `expected_signers` - Tracking signataires
- `reminder_logs` - Historique emails
- `checksum_verifications` - Vérifications intégrité
- `oauth_sessions` - Sessions OAuth2 + refresh tokens

Voir [Database](database.md) pour le schéma complet.

### Migrations

- Format : `XXXX_description.up.sql` / `XXXX_description.down.sql`
- Appliquées automatiquement au démarrage (service `ackify-migrate`)
- Outil : `/backend/cmd/migrate`

## Sécurité

### Cryptographie

**Ed25519** :
- Signatures digitales (courbe elliptique)
- Clé privée 256 bits
- Non-répudiation garantie

**SHA-256** :
- Hachage payload avant signature
- Détection de tampering
- Chaînage blockchain-like (`prev_hash`)

**AES-256-GCM** :
- Chiffrement refresh tokens OAuth2
- Clé dérivée de `ACKIFY_OAUTH_COOKIE_SECRET`

### OAuth2 + PKCE

**Flow** :
1. Client génère `code_verifier` (random)
2. Calcule `code_challenge = SHA256(code_verifier)`
3. Auth request avec `code_challenge`
4. Provider retourne `code`
5. Token exchange avec `code + code_verifier`

**Sécurité** :
- Protection contre interception du code
- Méthode S256 (SHA-256)
- Activé automatiquement

### Sessions

- Cookies sécurisés (HttpOnly, Secure, SameSite=Lax)
- Chiffrement HMAC-SHA256
- Stockage PostgreSQL avec refresh tokens chiffrés
- Durée : 30 jours
- Cleanup automatique : 37 jours

## Build & Déploiement

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

**Résultat** :
- Image finale < 30 MB
- Binaire unique (backend + frontend)
- Aucune dépendance runtime

### Injection Runtime

Le `ACKIFY_BASE_URL` est injecté dans `index.html` au démarrage :

```go
// Remplace __ACKIFY_BASE_URL__ par la valeur réelle
html = strings.ReplaceAll(html, "__ACKIFY_BASE_URL__", baseURL)
```

Permet de changer le domaine sans rebuild.

## Performance

### Backend

- **Connection pooling** PostgreSQL (25 max)
- **Prepared statements** - Anti-injection SQL
- **Rate limiting** - 5 auth/min, 10 doc/min, 100 req/min
- **Structured logging** - JSON avec request IDs

### Frontend

- **Code splitting** - Lazy loading routes
- **Tree shaking** - Dead code elimination
- **Minification** - Production builds optimisés
- **HMR** - Hot Module Replacement (dev)

### Database

- **Index** sur (doc_id, user_sub, session_id)
- **Constraints** UNIQUE pour garanties
- **Triggers** pour immutabilité
- **Autovacuum** activé

## Scalabilité

### Limites Actuelles

- ✅ Monolithe : ~10k req/s
- ✅ PostgreSQL : Single instance
- ✅ Sessions : In-database (pas de Redis)

### Scaling Horizontal (futur)

Pour > 100k req/s :
1. **Load Balancer** - Multiple instances backend
2. **PostgreSQL read replicas** - Séparation read/write
3. **Redis** - Cache sessions + rate limiting
4. **CDN** - Assets statiques

## Monitoring

### Logs Structurés

Format JSON :
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

Response :
```json
{
  "status": "healthy",
  "database": "connected"
}
```

### Métriques (futur)

- Prometheus metrics endpoint
- Grafana dashboards
- Alerting (PagerDuty, Slack)

## Tests

### Coverage

**72.6% code coverage** (unit + integration)

- Unit tests : 180+ tests
- Integration tests : 33 tests PostgreSQL
- CI/CD : GitHub Actions + Codecov

Voir [Development](development.md) pour lancer les tests.

## Choix Techniques

### Pourquoi Go ?

- ✅ Performance native (compiled)
- ✅ Concurrency simple (goroutines)
- ✅ Typage fort
- ✅ Binaire unique
- ✅ Déploiement simple

### Pourquoi Vue 3 ?

- ✅ Composition API moderne
- ✅ TypeScript natif
- ✅ Reactive par défaut
- ✅ Ecosystème riche
- ✅ Performances excellentes

### Pourquoi PostgreSQL ?

- ✅ ACID compliance
- ✅ Contraintes d'intégrité
- ✅ Triggers
- ✅ JSON support
- ✅ Mature et stable

### Pourquoi Ed25519 ?

- ✅ Sécurité moderne (courbe elliptique)
- ✅ Performance > RSA
- ✅ Signatures courtes (64 bytes)
- ✅ Standard crypto/ed25519 Go

## Références

- [Go Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [Vue 3 Composition API](https://vuejs.org/guide/extras/composition-api-faq.html)
- [Ed25519 Spec](https://ed25519.cr.yp.to/)
- [OAuth2 + PKCE](https://oauth.net/2/pkce/)
