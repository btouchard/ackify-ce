# 🔐 Ackify

> **Proof of Read. Compliance made simple.**

Service sécurisé de validation de lecture avec traçabilité cryptographique et preuves incontestables.

[![Build](https://img.shields.io/badge/build-passing-brightgreen.svg)](https://github.com/btouchard/ackify)
[![Security](https://img.shields.io/badge/crypto-Ed25519-blue.svg)](https://en.wikipedia.org/wiki/EdDSA)
[![Go](https://img.shields.io/badge/go-1.24.5-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/license-SSPL-blue.svg)](LICENSE)

> 🌍 [English version available here](README_EN.md)

## 🎯 Pourquoi Ackify ?

**Problème** : Comment prouver qu'un collaborateur a bien lu et compris un document important ?

**Solution** : Signatures cryptographiques Ed25519 avec horodatage immutable et traçabilité complète.

### Cas d'usage concrets
- ✅ Validation de politiques de sécurité
- ✅ Attestations de formation obligatoire  
- ✅ Prise de connaissance RGPD
- ✅ Accusés de réception contractuels
- ✅ Procédures qualité et compliance

---

## ⚡ Démarrage Rapide

### Avec Docker (recommandé)
```bash
# Installation automatique
curl -fsSL https://raw.githubusercontent.com/btouchard/ackify/main/install/install.sh | bash

# Ou téléchargement manuel
curl -O https://raw.githubusercontent.com/btouchard/ackify/main/install/docker-compose.yml
curl -O https://raw.githubusercontent.com/btouchard/ackify/main/install/.env.example

# Configuration
cp .env.example .env
# Éditez .env avec vos paramètres OAuth2

# Génération des secrets
export OAUTH_COOKIE_SECRET=$(openssl rand -base64 32)
export ED25519_PRIVATE_KEY_B64=$(openssl rand 64 | base64 -w 0)

# Démarrage
docker compose up -d

# Test
curl http://localhost:8080/healthz
```

### Variables obligatoires
```bash
APP_BASE_URL="https://votre-domaine.com"
OAUTH_CLIENT_ID="your-oauth-client-id"        # Google/GitHub/GitLab
OAUTH_CLIENT_SECRET="your-oauth-client-secret"
DB_DSN="postgres://user:password@localhost/ackify?sslmode=disable"
OAUTH_COOKIE_SECRET="$(openssl rand -base64 32)"
```

---

## 🚀 Utilisation Simple

### 1. Demander une signature
```
https://votre-domaine.com/sign?doc=procedure_securite_2025
```
→ L'utilisateur s'authentifie via OAuth2 et valide sa lecture

### 2. Vérifier les signatures
```bash
# API JSON - Liste complète
curl "https://votre-domaine.com/status?doc=procedure_securite_2025"

# Badge PNG - Statut individuel  
curl "https://votre-domaine.com/status.png?doc=procedure_securite_2025&user=jean.dupont@entreprise.com"
```

### 3. Intégrer dans vos pages
```html
<!-- Widget intégrable -->
<iframe src="https://votre-domaine.com/embed?doc=procedure_securite_2025" 
        width="500" height="300"></iframe>

<!-- Via oEmbed -->
<script>
fetch('/oembed?url=https://votre-domaine.com/embed?doc=procedure_securite_2025')
  .then(r => r.json())
  .then(data => document.getElementById('signatures').innerHTML = data.html);
</script>
```

---

## 🔧 Configuration OAuth2

### Providers supportés

| Provider | Configuration |
|----------|---------------|
| **Google** | `OAUTH_PROVIDER=google` |
| **GitHub** | `OAUTH_PROVIDER=github` |
| **GitLab** | `OAUTH_PROVIDER=gitlab` + `OAUTH_GITLAB_URL` |
| **Custom** | Endpoints personnalisés |

### Provider personnalisé
```bash
# Laissez OAUTH_PROVIDER vide
OAUTH_AUTH_URL="https://auth.company.com/oauth/authorize"
OAUTH_TOKEN_URL="https://auth.company.com/oauth/token"  
OAUTH_USERINFO_URL="https://auth.company.com/api/user"
OAUTH_SCOPES="read:user,user:email"
```

### Restriction par domaine
```bash
OAUTH_ALLOWED_DOMAIN="@entreprise.com"  # Seuls les emails @entreprise.com
```

---

## 🛡️ Sécurité & Architecture

### Sécurité cryptographique
- **Ed25519** : Signatures numériques de pointe
- **SHA-256** : Hachage des payloads contre le tampering
- **Horodatage immutable** : Triggers PostgreSQL
- **Sessions chiffrées** : Cookies sécurisés
- **CSP headers** : Protection XSS

### Architecture Go
```
cmd/ackapp/              # Point d'entrée
internal/
  domain/                # Logique métier
    models/              # Entités
    repositories/        # Interfaces persistance
  application/           # Use cases  
    services/            # Implémentations métier
  infrastructure/        # Adaptateurs
    auth/               # OAuth2
    database/           # PostgreSQL
    config/             # Configuration
  presentation/          # HTTP
    handlers/           # Contrôleurs + interfaces
    templates/          # Vues HTML
pkg/                    # Utilitaires partagés
```

### Stack technique
- **Go 1.24.5** : Performance et simplicité
- **PostgreSQL** : Contraintes d'intégrité 
- **OAuth2** : Multi-providers
- **Docker** : Déploiement simplifié
- **Traefik** : Reverse proxy HTTPS

---

## 📊 Base de Données

```sql
CREATE TABLE signatures (
    id BIGSERIAL PRIMARY KEY,
    doc_id TEXT NOT NULL,                    -- ID document
    user_sub TEXT NOT NULL,                  -- ID OAuth utilisateur
    user_email TEXT NOT NULL,                -- Email utilisateur
    signed_at TIMESTAMPTZ NOT NULL,          -- Timestamp signature
    payload_hash TEXT NOT NULL,              -- Hash cryptographique
    signature TEXT NOT NULL,                 -- Signature Ed25519
    nonce TEXT NOT NULL,                     -- Anti-replay
    created_at TIMESTAMPTZ DEFAULT now(),    -- Immutable
    referer TEXT,                            -- Source (optionnel)
    prev_hash TEXT,                          -- Prev Hash
    UNIQUE (doc_id, user_sub)                -- Une signature par user/doc
);
```

**Garanties** :
- ✅ **Unicité** : Un utilisateur = une signature par document
- ✅ **Immutabilité** : `created_at` protégé par trigger
- ✅ **Intégrité** : Hachage SHA-256 pour détecter modifications
- ✅ **Non-répudiation** : Signature Ed25519 cryptographiquement prouvable

---

## 🚀 Déploiement Production

### docker-compose.yml
```yaml
version: '3.8'
services:
  ackapp:
    image: btouchard/ackify:latest
    environment:
      APP_BASE_URL: https://ackify.company.com
      DB_DSN: postgres://user:pass@postgres:5432/ackdb?sslmode=require
      OAUTH_CLIENT_ID: ${OAUTH_CLIENT_ID}
      OAUTH_CLIENT_SECRET: ${OAUTH_CLIENT_SECRET}
      OAUTH_COOKIE_SECRET: ${OAUTH_COOKIE_SECRET}
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

### Variables production
```bash
# Sécurité renforcée
OAUTH_COOKIE_SECRET="$(openssl rand -base64 64)"  # AES-256
ED25519_PRIVATE_KEY_B64="$(openssl genpkey -algorithm Ed25519 | base64 -w 0)"

# HTTPS obligatoire
APP_BASE_URL="https://ackify.company.com"

# PostgreSQL sécurisé
DB_DSN="postgres://user:pass@postgres:5432/ackdb?sslmode=require"
```

---

## 📋 API Complète

### Authentification
- `GET /login?next=<url>` - Connexion OAuth2
- `GET /logout` - Déconnexion
- `GET /oauth2/callback` - Callback OAuth2

### Signatures  
- `GET /sign?doc=<id>` - Interface de signature
- `POST /sign` - Créer signature
- `GET /signatures` - Mes signatures (auth requis)

### Consultation
- `GET /status?doc=<id>` - JSON toutes signatures
- `GET /status.png?doc=<id>&user=<email>` - Badge PNG

### Intégration
- `GET /oembed?url=<embed_url>` - Métadonnées oEmbed  
- `GET /embed?doc=<id>` - Widget HTML

### Supervision
- `GET /healthz` - Health check

---

## 🔍 Développement & Tests

### Build local
```bash
# Dépendances
go mod tidy

# Build
go build ./cmd/ackify

# Linting
go fmt ./...
go vet ./...

# Tests (TODO: ajouter des tests)
go test -v ./...
```

### Docker development
```bash
# Build image
docker build -t ackify:dev .

# Run avec base locale
docker run -p 8080:8080 --env-file .env ackify:dev
```

---

## 🤝 Support

### Aide & Documentation
- 🐛 **Issues** : [GitHub Issues](https://github.com/btouchard/ackify/issues)
- 💬 **Discussions** : [GitHub Discussions](https://github.com/btouchard/ackify/discussions)

### Licence SSPL
Usage libre pour projets internes. Restriction pour services commerciaux concurrents.
Voir [LICENSE](LICENSE) pour détails complets.

---

**Développé avec ❤️ par [Benjamin TOUCHARD](mailto:benjamin@kolapsis.com)**