# 🔐 Ackify

> **Proof of Read. Compliance made simple.**

Service sécurisé de validation de lecture avec traçabilité cryptographique et preuves incontestables.

[![Build](https://github.com/btouchard/ackify-ce/actions/workflows/ci.yml/badge.svg)](https://github.com/btouchard/ackify-ce/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/btouchard/ackify-ce/branch/main/graph/badge.svg)](https://codecov.io/gh/btouchard/ackify-ce)
[![Security](https://img.shields.io/badge/crypto-Ed25519-blue.svg)](https://en.wikipedia.org/wiki/EdDSA)
[![Go](https://img.shields.io/badge/go-1.24.5-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/License-AGPL_v3-blue.svg)](LICENSE)

> 🌍 [English version available here](README.md)

### Visitez notre site : https://www.ackify.eu/fr

## 🎯 Pourquoi Ackify ?

**Problème** : Comment prouver qu'un collaborateur a bien lu et compris un document important ?

**Solution** : Signatures cryptographiques Ed25519 avec horodatage immutable et traçabilité complète.

### Cas d'usage concrets
- ✅ Validation de politiques de sécurité
- ✅ Attestations de formation obligatoire
- ✅ Prise de connaissance RGPD
- ✅ Accusés de réception contractuels
- ✅ Procédures qualité et compliance

### Fonctionnalités Principales

**Fonctionnalités de Base** :
- Signatures cryptographiques Ed25519 avec validation par chaîne de hachage
- Une signature par utilisateur et par document (appliqué par contraintes de base de données)
- Authentification OAuth2 (Google, GitHub, GitLab, ou fournisseur personnalisé)
- Widgets intégrables publics pour Notion, Outline, Google Docs, etc.

**Gestion des Documents** :
- Métadonnées de documents avec titre, URL et description
- Vérification de checksum (SHA-256, SHA-512, MD5) pour suivi de l'intégrité
- Historique de vérification avec piste d'audit horodatée
- Calcul côté client des checksums avec l'API Web Crypto

**Suivi & Rappels** :
- Liste des signataires attendus avec suivi de la complétion
- Rappels par email dans la langue préférée de l'utilisateur (fr, en, es, de, it)
- Barres de progression visuelles et pourcentages de complétion
- Détection automatique des signatures inattendues

**Tableau de Bord Admin** :
- Interface moderne Vue.js 3 avec mode sombre
- Gestion des documents avec opérations en masse
- Suivi des signatures et analyses
- Gestion des signataires attendus
- Système de rappels par email avec historique

**Intégration & Embedding** :
- Support oEmbed pour déploiement automatique (Slack, Teams, etc.)
- Meta tags Open Graph et Twitter Card dynamiques
- Pages embed publiques avec boutons de signature
- API RESTful v1 avec spécification OpenAPI
- Badges PNG pour fichiers README et documentation

**Sécurité & Conformité** :
- Piste d'audit immutable avec triggers PostgreSQL
- Protection CSRF et limitation de débit (5 tentatives auth/min, 10 créations documents/min, 100 requêtes générales/min)
- Sessions chiffrées avec cookies sécurisés
- En-têtes Content Security Policy (CSP)
- Application HTTPS en production

---

## 📸 Vidéos


Cliquez sur les GIFs pour ouvrir les vidéos WebM dans votre navigateur.

<table>
<tr>
  <td align="center">
    <strong>1) Création d’une signature</strong><br>
    <a href="screenshots/videos/1-initialize-sign.webm" target="_blank">
      <img src="screenshots/videos/1-initialize-sign.gif" width="380" alt="Initialisation d’une signature">
    </a>
  </td>
  <td align="center">
    <strong>2) Parcours de signature utilisateur</strong><br>
    <a href="screenshots/videos/2-user-sign-flow.webm" target="_blank">
      <img src="screenshots/videos/2-user-sign-flow.gif" width="380" alt="Parcours de signature utilisateur">
    </a>
  </td>

</tr>
</table>

## 📸 Captures d'écran

<table>
<tr>
<td align="center">
<strong>Page d'accueil</strong><br>
<a href="screenshots/1-home.png"><img src="screenshots/1-home.png" width="200" alt="Page d'accueil"></a>
</td>
<td align="center">
<strong>Demande de signature</strong><br>
<a href="screenshots/2-signing-request.png"><img src="screenshots/2-signing-request.png" width="200" alt="Demande de signature"></a>
</td>
<td align="center">
<strong>Signature confirmée</strong><br>
<a href="screenshots/3-signing-ok.png"><img src="screenshots/3-signing-ok.png" width="200" alt="Signature confirmée"></a>
</td>
</tr>
<tr>
<td align="center">
<strong>Liste des signatures</strong><br>
<a href="screenshots/4-sign-list.png"><img src="screenshots/4-sign-list.png" width="200" alt="Liste des signatures"></a>
</td>
<td align="center">
<strong>Intégration Outline</strong><br>
<a href="screenshots/5-integrated-to-outline.png"><img src="screenshots/5-integrated-to-outline.png" width="200" alt="Intégration Outline"></a>
</td>
<td align="center">
<strong>Intégration Google Docs</strong><br>
<a href="screenshots/6-integrated-to-google-doc.png"><img src="screenshots/6-integrated-to-google-doc.png" width="200" alt="Intégration Google Docs"></a>
</td>
</tr>
</table>

---

## ⚡ Démarrage Rapide

### Avec Docker (recommandé)
```bash
# Cloner le dépôt
git clone https://github.com/btouchard/ackify-ce.git
cd ackify-ce

# Configurer l'environnement
cp .env.example .env
# Éditez .env avec vos paramètres OAuth2 (voir section configuration ci-dessous)

# Démarrer les services (PostgreSQL + Ackify)
docker compose up -d

# Voir les logs
docker compose logs -f ackify-ce

# Vérifier le déploiement
curl http://localhost:8080/api/v1/health
# Attendu : {"status": "healthy", "database": "connected"}

# Accéder à l'interface web
open http://localhost:8080
# SPA Vue.js 3 moderne avec support du mode sombre
```

**Ce qui est inclus** :
- PostgreSQL 16 avec migrations automatiques
- Backend Ackify (Go) avec frontend intégré
- Endpoint de monitoring de santé
- Tableau de bord admin sur `/admin`
- Documentation API sur `/api/openapi.yaml`

### Variables d'Environnement Requises

```bash
# URL de base de l'application (requis - utilisé pour les callbacks OAuth et les URLs embed)
ACKIFY_BASE_URL="https://votre-domaine.com"

# Nom de l'organisation (requis - utilisé dans les templates email et l'affichage)
ACKIFY_ORGANISATION="Nom de Votre Organisation"

# Configuration OAuth2 (requis)
ACKIFY_OAUTH_CLIENT_ID="your-oauth-client-id"
ACKIFY_OAUTH_CLIENT_SECRET="your-oauth-client-secret"

# Connexion base de données (requis)
ACKIFY_DB_DSN="postgres://user:password@localhost/ackify?sslmode=disable"

# Sécurité des sessions (requis - générer avec : openssl rand -base64 32)
ACKIFY_OAUTH_COOKIE_SECRET="$(openssl rand -base64 32)"
```

### Variables d'Environnement Optionnelles

**Notifications Email (SMTP)** :
```bash
ACKIFY_MAIL_HOST="smtp.gmail.com"              # Serveur SMTP (si vide, email désactivé)
ACKIFY_MAIL_PORT="587"                         # Port SMTP (défaut : 587)
ACKIFY_MAIL_USERNAME="votre-email@gmail.com"   # Nom d'utilisateur SMTP
ACKIFY_MAIL_PASSWORD="votre-app-password"      # Mot de passe SMTP
ACKIFY_MAIL_TLS="true"                         # Activer TLS (défaut : true)
ACKIFY_MAIL_STARTTLS="true"                    # Activer STARTTLS (défaut : true)
ACKIFY_MAIL_TIMEOUT="10s"                      # Timeout de connexion (défaut : 10s)
ACKIFY_MAIL_FROM="noreply@entreprise.com"      # Adresse email expéditeur
ACKIFY_MAIL_FROM_NAME="Ackify"                 # Nom d'affichage expéditeur
ACKIFY_MAIL_SUBJECT_PREFIX=""                  # Préfixe optionnel pour les sujets d'email
ACKIFY_MAIL_TEMPLATE_DIR="templates/emails"    # Répertoire des templates email (défaut : templates/emails)
ACKIFY_MAIL_DEFAULT_LOCALE="en"                # Locale par défaut pour les emails (défaut : en)
```

**Configuration Serveur** :
```bash
ACKIFY_LISTEN_ADDR=":8080"                     # Adresse d'écoute HTTP (défaut : :8080)
ACKIFY_LOG_LEVEL="info"                        # Niveau de log : debug, info, warn, error (défaut : info)
```

**Accès Admin** :
```bash
ACKIFY_ADMIN_EMAILS="alice@entreprise.com,bob@entreprise.com"  # Emails admin séparés par des virgules
```

**Clés Cryptographiques** :
```bash
ACKIFY_ED25519_PRIVATE_KEY="$(openssl rand -base64 64)"  # Clé de signature Ed25519 (optionnel, auto-générée si vide)
```

**OAuth2 Avancé** :
```bash
ACKIFY_OAUTH_AUTO_LOGIN="true"                 # Activer l'authentification silencieuse (défaut : false)
ACKIFY_OAUTH_ALLOWED_DOMAIN="@entreprise.com"  # Restreindre au domaine email spécifique
ACKIFY_OAUTH_LOGOUT_URL=""                     # URL de déconnexion du provider OAuth personnalisé (optionnel)
```

**Templates & Locales** :
```bash
ACKIFY_TEMPLATES_DIR="/chemin/personnalise/templates"  # Répertoire de templates personnalisé (optionnel)
ACKIFY_LOCALES_DIR="/chemin/personnalise/locales"      # Répertoire de locales personnalisé (optionnel)
```

---

## 🚀 Utilisation Simple

### 1. Demander une signature
```
https://votre-domaine.com/?doc=procedure_securite_2025
```
→ L'utilisateur s'authentifie via OAuth2 et valide sa lecture

### 2. Intégrer dans vos pages

**Widget intégrable** (avec bouton de signature) :
```html
<!-- La SPA gère l'affichage -->
<iframe src="https://votre-domaine.com/embed?doc=procedure_securite_2025"
        width="600" height="200"
        frameborder="0"
        style="border: 1px solid #ddd; border-radius: 6px;"></iframe>
```

**Support oEmbed** (déploiement automatique dans Notion, Outline, Confluence, etc.) :
```html
<!-- Collez simplement l'URL - les plateformes avec support oEmbed l'auto-découvriront et l'intégreront -->
https://votre-domaine.com/?doc=procedure_securite_2025
```

Le endpoint oEmbed (`/oembed`) est automatiquement découvert via le meta tag `<link rel="alternate" type="application/json+oembed">`.

**Manuel oEmbed** :
```javascript
fetch('/oembed?url=https://votre-domaine.com/?doc=procedure_securite_2025')
  .then(r => r.json())
  .then(data => {
    console.log(data.html);  // <iframe src="..." width="100%" height="200"></iframe>
    console.log(data.title); // Titre du document avec nombre de signatures
  });
```

### 3. Métadonnées Dynamiques pour Unfurling

Ackify génère automatiquement des **meta tags Open Graph, Twitter Card et de découverte oEmbed dynamiques** :

```html
<!-- Meta tags auto-générés pour /?doc=doc_id -->
<meta property="og:title" content="Document: procedure_securite_2025 - 3 confirmations" />
<meta property="og:description" content="3 personnes ont confirmé avoir lu le document" />
<meta property="og:url" content="https://votre-domaine.com/?doc=doc_id" />
<meta property="og:type" content="website" />
<meta name="twitter:card" content="summary" />
<link rel="alternate" type="application/json+oembed"
      href="https://votre-domaine.com/oembed?url=https://votre-domaine.com/?doc=doc_id"
      title="Document: procedure_securite_2025 - 3 confirmations" />
```

**Résultat** : Lorsque vous collez une URL de document dans Slack, Teams, Discord, Notion, Outline, ou les réseaux sociaux :
- **Open Graph/Twitter** : Aperçu enrichi avec titre, description, nombre de signatures
- **oEmbed** (Notion, Outline, Confluence) : Widget interactif complet intégré dans la page
- **Sans authentification requise** sur la page publique, ce qui la rend parfaite pour partager la progression publiquement

---

## 🔧 Configuration OAuth2

### Providers supportés

| Provider | Configuration |
|----------|---------------|
| **Google** | `ACKIFY_OAUTH_PROVIDER=google` |
| **GitHub** | `ACKIFY_OAUTH_PROVIDER=github` |
| **GitLab** | `ACKIFY_OAUTH_PROVIDER=gitlab` + `ACKIFY_OAUTH_GITLAB_URL` |
| **Custom** | Endpoints personnalisés |

### Provider personnalisé
```bash
# Laissez ACKIFY_OAUTH_PROVIDER vide
ACKIFY_OAUTH_AUTH_URL="https://auth.company.com/oauth/authorize"
ACKIFY_OAUTH_TOKEN_URL="https://auth.company.com/oauth/token"
ACKIFY_OAUTH_USERINFO_URL="https://auth.company.com/api/user"
ACKIFY_OAUTH_SCOPES="read:user,user:email"
```

### Restriction par domaine
```bash
ACKIFY_OAUTH_ALLOWED_DOMAIN="@entreprise.com"  # Seuls les emails @entreprise.com
```

### Log level setup
```bash
ACKIFY_LOG_LEVEL="info" # peut être debug, info, warn(ing), error. défaut: info
```

### Configuration auto-login
```bash
ACKIFY_OAUTH_AUTO_LOGIN="true"  # Active l'authentification silencieuse si session existe (défaut: false)
```

---

## 🏗️ Structure du Projet

Ackify suit une **architecture monorepo** avec séparation claire entre backend et frontend :

```
ackify-ce/
├── backend/              # Backend Go (API-first)
│   ├── cmd/
│   │   ├── community/    # Point d'entrée principal de l'application
│   │   └── migrate/      # Outil de migration de base de données
│   ├── internal/
│   │   ├── domain/       # Entités métier (models)
│   │   ├── application/  # Logique métier (services)
│   │   ├── infrastructure/ # Implémentations techniques
│   │   │   ├── auth/     # Service OAuth2
│   │   │   ├── database/ # Repositories PostgreSQL
│   │   │   ├── email/    # Service SMTP
│   │   │   ├── config/   # Gestion de la configuration
│   │   │   └── i18n/     # Internationalisation backend
│   │   └── presentation/ # Couche HTTP
│   │       ├── api/      # Handlers API RESTful v1
│   │       └── handlers/ # Handlers de templates legacy
│   ├── pkg/              # Utilitaires partagés
│   │   ├── crypto/       # Signatures Ed25519
│   │   ├── logger/       # Logging structuré
│   │   ├── services/     # Détection de providers OAuth
│   │   └── web/          # Configuration serveur HTTP
│   ├── migrations/       # Migrations SQL
│   ├── locales/          # Traductions backend (fr, en)
│   └── templates/        # Templates email (HTML/texte)
├── webapp/               # SPA Vue.js 3 (frontend)
│   ├── src/
│   │   ├── components/   # Composants Vue réutilisables (shadcn/vue)
│   │   ├── pages/        # Composants de pages (vues router)
│   │   ├── services/     # Services client API
│   │   ├── stores/       # Gestion d'état Pinia
│   │   ├── router/       # Configuration Vue Router
│   │   ├── locales/      # Traductions frontend (fr, en, es, de, it)
│   │   └── composables/  # Composables Vue
│   ├── public/           # Assets statiques
│   └── scripts/          # Scripts de build & i18n
├── api/                  # Spécification OpenAPI
│   └── openapi.yaml      # Documentation API complète
├── go.mod                # Dépendances Go (à la racine)
└── go.sum
```

## 🛡️ Sécurité & Architecture

### Architecture Moderne API-First

Ackify utilise une **architecture moderne, API-first** avec séparation complète des préoccupations :

**Backend (Go)** :
- **API RESTful v1** : API versionnée (`/api/v1`) avec réponses JSON structurées
- **Clean Architecture** : Conception pilotée par le domaine avec séparation claire des couches
- **Spécification OpenAPI** : Documentation API complète dans `/api/openapi.yaml`
- **Authentification Sécurisée** : OAuth2 avec authentification basée session + protection CSRF
- **Limitation de Débit** : Protection contre les abus (5 tentatives auth/min, 100 requêtes générales/min)
- **Logging Structuré** : Logs JSON avec IDs de requête pour traçage distribué

**Frontend (SPA Vue.js 3)** :
- **TypeScript** : Développement type-safe avec support IDE complet
- **Vite** : HMR rapide et builds de production optimisés
- **Vue Router** : Routage côté client avec lazy loading
- **Pinia** : Gestion d'état centralisée
- **shadcn/vue** : Composants UI accessibles et personnalisables
- **Tailwind CSS** : Stylage utility-first avec support du mode sombre
- **vue-i18n** : 5 langues (fr, en, es, de, it) avec détection automatique

### Sécurité Cryptographique
- **Ed25519** : Signatures numériques de pointe (courbe elliptique)
- **SHA-256** : Hachage des payloads contre altération
- **Chaîne de Hachage** : Hash de signature précédente pour vérification d'intégrité
- **Horodatages Immutables** : Les triggers PostgreSQL empêchent l'antidatage
- **Sessions Chiffrées** : Cookies sécurisés avec HMAC-SHA256
- **En-têtes CSP** : Content Security Policy pour protection XSS
- **CORS** : Partage de ressources entre origines configurable

### Build & Déploiement

**Build Docker Multi-étapes** :
1. **Étape 1 - Build Frontend** : Node.js 22 construit la SPA Vue.js 3 avec Vite
2. **Étape 2 - Build Backend** : Go (latest avec GOTOOLCHAIN=auto) compile le backend et intègre les assets statiques du frontend
3. **Étape 3 - Runtime** : Image minimale Distroless (< 30MB)

**Caractéristiques Clés** :
- **Injection côté serveur** : `ACKIFY_BASE_URL` injecté dans `index.html` au runtime
- **Intégration statique** : Assets frontend intégrés dans le binaire Go via `embed.FS`
- **Binaire unique** : Le backend sert à la fois l'API et le frontend (pas besoin de serveur web séparé)
- **Arrêt gracieux** : Cycle de vie approprié du serveur HTTP avec gestion des signaux
- **Production-ready** : Builds optimisés avec élimination du code mort

**Processus de Build** :
```dockerfile
# Build frontend (webapp/)
FROM node:22-alpine AS frontend
COPY webapp/ /build/webapp/
RUN npm ci && npm run build
# Sortie vers : /build/webapp/dist/

# Build backend (backend/)
FROM golang:alpine AS backend
ENV GOTOOLCHAIN=auto
COPY backend/ /build/backend/
COPY --from=frontend /build/webapp/dist/ /build/backend/cmd/community/web/dist/
RUN go build -o community ./cmd/community
# Intègre dist/ dans le binaire Go via embed.FS

# Runtime
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=backend /build/backend/community /app/community
CMD ["/app/community"]
```

### Stack Technologique

**Backend** :
- **Go 1.24.5+** : Performance, simplicité et typage fort
- **PostgreSQL 16+** : Conformité ACID avec contraintes d'intégrité
- **Chi Router** : Routeur HTTP Go léger et idiomatique
- **OAuth2** : Authentification multi-provider (Google, GitHub, GitLab, custom)
- **Ed25519** : Signatures numériques à courbe elliptique (crypto/ed25519)
- **SMTP** : Rappels email via bibliothèque standard (optionnel)

**Frontend** :
- **Vue 3** : Framework réactif moderne avec Composition API
- **TypeScript** : Sécurité de type complète sur tout le frontend
- **Vite** : HMR ultra-rapide et builds de production optimisés
- **Pinia** : Gestion d'état intuitive pour Vue 3
- **Vue Router** : Routage côté client avec code splitting
- **Tailwind CSS** : Stylage utility-first avec mode sombre
- **shadcn/vue** : Bibliothèque de composants accessibles et personnalisables
- **vue-i18n** : Internationalisation (FR, EN, ES, DE, IT)

**DevOps** :
- **Docker** : Builds multi-étapes avec Alpine Linux
- **Migrations PostgreSQL** : Évolution de schéma versionnée
- **OpenAPI** : Documentation API avec Swagger UI

### Internationalisation (i18n)

L'interface web d'Ackify est entièrement internationalisée avec support de **5 langues** :

- **🇫🇷 Français** (par défaut)
- **🇬🇧 Anglais** (fallback)
- **🇪🇸 Espagnol**
- **🇩🇪 Allemand**
- **🇮🇹 Italien**

**Fonctionnalités** :
- Sélecteur de langue avec drapeaux Unicode dans l'en-tête
- Détection automatique depuis le navigateur ou localStorage
- Titres de page dynamiques avec i18n
- Couverture de traduction complète vérifiée par script CI
- Tous les éléments UI, labels ARIA et métadonnées traduits

**Documentation** : Voir [webapp/I18N.md](webapp/I18N.md) pour le guide i18n complet.

**Scripts** :
```bash
cd webapp
npm run lint:i18n  # Vérifier la couverture des traductions
```

---

## 📊 Base de Données

### Gestion du Schéma

Ackify utilise des **migrations SQL versionnées** pour l'évolution du schéma :

**Fichiers de migration** : Situés dans `/backend/migrations/`
- `0001_init.up.sql` - Schéma initial (table signatures)
- `0002_expected_signers.up.sql` - Suivi des signataires attendus
- `0003_reminder_logs.up.sql` - Historique des rappels email
- `0004_add_name_to_expected_signers.up.sql` - Noms d'affichage pour les signataires
- `0005_create_documents_table.up.sql` - Métadonnées de documents
- `0006_checksum_verifications.up.sql` - Historique de vérification de checksum

**Appliquer les migrations** :
```bash
# Utilisation de l'outil Go migrate
cd backend
go run ./cmd/migrate

# Ou manuellement avec psql
psql $ACKIFY_DB_DSN -f migrations/0001_init.up.sql
```

**Docker Compose** : Les migrations sont appliquées automatiquement au démarrage du conteneur.

### Schéma de Base de Données

```sql
-- Table principale des signatures
CREATE TABLE signatures (
    id BIGSERIAL PRIMARY KEY,
    doc_id TEXT NOT NULL,                    -- ID document
    user_sub TEXT NOT NULL,                  -- ID utilisateur OAuth
    user_email TEXT NOT NULL,                -- Email utilisateur
    signed_at TIMESTAMPTZ NOT NULL,          -- Horodatage signature
    payload_hash TEXT NOT NULL,              -- Hash cryptographique
    signature TEXT NOT NULL,                 -- Signature Ed25519
    nonce TEXT NOT NULL,                     -- Anti-rejeu
    created_at TIMESTAMPTZ DEFAULT now(),    -- Immutable
    referer TEXT,                            -- Source (optionnel)
    prev_hash TEXT,
    UNIQUE (doc_id, user_sub)                -- Une signature par user/doc
);

-- Table des signataires attendus (pour le suivi)
CREATE TABLE expected_signers (
    id BIGSERIAL PRIMARY KEY,
    doc_id TEXT NOT NULL,
    email TEXT NOT NULL,
    name TEXT NOT NULL DEFAULT '',           -- Nom d'affichage (optionnel)
    added_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    added_by TEXT NOT NULL,                  -- Admin qui a ajouté
    notes TEXT,
    UNIQUE (doc_id, email)                   -- Une attente par email/doc
);

-- Table des métadonnées de documents
CREATE TABLE documents (
    doc_id TEXT PRIMARY KEY,
    title TEXT NOT NULL DEFAULT '',
    url TEXT NOT NULL DEFAULT '',            -- Emplacement du document
    checksum TEXT NOT NULL DEFAULT '',       -- SHA-256/SHA-512/MD5
    checksum_algorithm TEXT NOT NULL DEFAULT 'SHA-256',
    description TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by TEXT NOT NULL DEFAULT ''
);
```

**Garanties** :
- ✅ **Unicité** : Un utilisateur = une signature par document
- ✅ **Immutabilité** : `created_at` protégé par trigger
- ✅ **Intégrité** : Hash SHA-256 pour détecter les modifications
- ✅ **Non-répudiation** : Signature Ed25519 cryptographiquement prouvable
- ✅ **Suivi** : Signataires attendus pour monitoring de complétion
- ✅ **Métadonnées** : Informations de documents avec URL, checksum et description
- ✅ **Vérification de checksum** : Suivi de l'intégrité des documents avec historique de vérification

### Vérification de l'Intégrité des Documents

Ackify prend en charge la vérification de l'intégrité des documents par suivi et vérification de checksum :

**Algorithmes supportés** : SHA-256 (par défaut), SHA-512, MD5

**Vérification côté client** (recommandé) :
```javascript
// Calculer le checksum dans le navigateur en utilisant l'API Web Crypto
async function calculateChecksum(file) {
  const arrayBuffer = await file.arrayBuffer();
  const hashBuffer = await crypto.subtle.digest('SHA-256', arrayBuffer);
  const hashArray = Array.from(new Uint8Array(hashBuffer));
  return hashArray.map(b => b.toString(16).padStart(2, '0')).join('');
}
```

**Calcul manuel du checksum** :
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

**Note** : Les valeurs de checksum sont stockées comme métadonnées et peuvent être consultées/mises à jour via l'interface de gestion des documents admin. La vérification se fait généralement côté client en utilisant l'API Web Crypto ou les outils en ligne de commande montrés ci-dessus.

---

## 🚀 Déploiement Production

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

### Variables d'Environnement Production
```bash
# Sécurité renforcée - générer des secrets forts
ACKIFY_OAUTH_COOKIE_SECRET="$(openssl rand 64 | base64 -w 0)"
ACKIFY_ED25519_PRIVATE_KEY="$(openssl rand 64 | base64 -w 0)"

# HTTPS obligatoire en production
ACKIFY_BASE_URL="https://ackify.entreprise.com"

# PostgreSQL sécurisé avec SSL
ACKIFY_DB_DSN="postgres://ackuser:strong_password@postgres:5432/ackdb?sslmode=require"

# Accès admin (emails séparés par des virgules)
ACKIFY_ADMIN_EMAILS="admin@entreprise.com,cto@entreprise.com"

# Rappels email (optionnel mais recommandé)
ACKIFY_MAIL_HOST="smtp.entreprise.com"
ACKIFY_MAIL_PORT="587"
ACKIFY_MAIL_FROM="noreply@entreprise.com"
ACKIFY_MAIL_FROM_NAME="Ackify - Nom Entreprise"
ACKIFY_MAIL_USERNAME="${SMTP_USERNAME}"
ACKIFY_MAIL_PASSWORD="${SMTP_PASSWORD}"

# Configuration OAuth2 (exemple avec Google)
ACKIFY_OAUTH_PROVIDER="google"
ACKIFY_OAUTH_CLIENT_ID="${GOOGLE_CLIENT_ID}"
ACKIFY_OAUTH_CLIENT_SECRET="${GOOGLE_CLIENT_SECRET}"
ACKIFY_OAUTH_ALLOWED_DOMAIN="@entreprise.com"  # Restreindre au domaine entreprise

# Logging
ACKIFY_LOG_LEVEL="info"  # Utiliser "debug" pour le débogage
```

### Conseils Production

**Checklist Sécurité** :
- ✅ Utiliser HTTPS (requis pour les cookies sécurisés)
- ✅ Activer SSL PostgreSQL (`sslmode=require`)
- ✅ Générer des secrets forts (64+ bytes)
- ✅ Restreindre OAuth au domaine de l'entreprise
- ✅ Configurer la liste des emails admin
- ✅ Surveiller les logs pour activité suspecte
- ✅ Sauvegardes régulières PostgreSQL

**Optimisation des Performances** :
- Pool de connexions PostgreSQL (géré par Go)
- CDN pour assets statiques (si hébergement séparé)
- Index de base de données sur `(doc_id, user_sub)`
- Limitation de débit activée par défaut

**Monitoring** :
- Endpoint de santé : `GET /api/v1/health` (inclut le statut DB)
- Logs JSON structurés avec IDs de requête
- Métriques de base de données via PostgreSQL `pg_stat_statements`

---

## 📋 Documentation API

### Spécification OpenAPI

Documentation API complète disponible au format OpenAPI 3.0 :

**📁 Fichier** : `/api/openapi.yaml`

**Fonctionnalités** :
- Documentation complète des endpoints API v1
- Schémas requête/réponse
- Exigences d'authentification
- Exemples de payloads
- Réponses d'erreur

**Visualiser en ligne** : Vous pouvez importer la spec OpenAPI dans :
- [Swagger Editor](https://editor.swagger.io/) - Coller le contenu YAML
- [Postman](https://www.postman.com/) - Importer en tant qu'OpenAPI 3.0
- [Insomnia](https://insomnia.rest/) - Importer en tant que spec OpenAPI
- Tout outil compatible OpenAPI

**Visualisation locale** :
```bash
# Utilisation de l'image Docker swagger-ui
docker run -p 8081:8080 -e SWAGGER_JSON=/api/openapi.yaml \
  -v $(pwd)/api:/api swaggerapi/swagger-ui
# Ouvrir http://localhost:8081
```

### Endpoints API v1

Tous les endpoints API v1 sont préfixés par `/api/v1` et retournent des réponses JSON avec codes de statut HTTP standards.

**Structure URL de base** :
- Développement : `http://localhost:8080/api/v1`
- Production : `https://votre-domaine.com/api/v1`

#### Système & Santé
- `GET /api/v1/health` - Health check avec statut de la base de données (public)

#### Authentification
- `POST /api/v1/auth/start` - Initier le flux OAuth (retourne l'URL de redirection)
- `GET /api/v1/auth/logout` - Déconnexion et suppression de la session
- `GET /api/v1/auth/check` - Vérifier le statut d'authentification (seulement si auto-login activé)
- `GET /api/v1/csrf` - Obtenir un token CSRF pour les requêtes authentifiées

#### Utilisateurs
- `GET /api/v1/users/me` - Obtenir le profil utilisateur actuel (authentifié)

#### Documents (Public)
- `GET /api/v1/documents` - Lister tous les documents avec pagination
- `POST /api/v1/documents` - Créer un nouveau document (nécessite token CSRF, limité à 10/min)
- `GET /api/v1/documents/{docId}` - Obtenir les détails du document avec signatures
- `GET /api/v1/documents/{docId}/signatures` - Obtenir les signatures du document
- `GET /api/v1/documents/{docId}/expected-signers` - Obtenir la liste des signataires attendus
- `GET /api/v1/documents/find-or-create?ref={reference}` - Trouver ou créer un document par référence (auth conditionnelle pour support embed)

#### Signatures
- `GET /api/v1/signatures` - Obtenir les signatures de l'utilisateur actuel avec pagination (authentifié)
- `POST /api/v1/signatures` - Créer une nouvelle signature (authentifié + token CSRF)
- `GET /api/v1/documents/{docId}/signatures/status` - Obtenir le statut de signature de l'utilisateur (authentifié)

#### Endpoints Admin
Tous les endpoints admin requièrent authentification + privilèges admin + token CSRF.

**Documents** :
- `GET /api/v1/admin/documents?limit=100&offset=0` - Lister tous les documents avec statistiques
- `GET /api/v1/admin/documents/{docId}` - Obtenir les détails du document (vue admin)
- `GET /api/v1/admin/documents/{docId}/signers` - Obtenir le document avec signataires et stats de complétion
- `GET /api/v1/admin/documents/{docId}/status` - Obtenir le statut du document avec stats de complétion
- `PUT /api/v1/admin/documents/{docId}/metadata` - Créer/mettre à jour les métadonnées du document
- `DELETE /api/v1/admin/documents/{docId}` - Supprimer le document entièrement (y compris métadonnées et signatures)

**Signataires Attendus** :
- `POST /api/v1/admin/documents/{docId}/signers` - Ajouter un signataire attendu
- `DELETE /api/v1/admin/documents/{docId}/signers/{email}` - Retirer un signataire attendu

**Rappels Email** :
- `POST /api/v1/admin/documents/{docId}/reminders` - Envoyer des rappels email aux lecteurs en attente
- `GET /api/v1/admin/documents/{docId}/reminders` - Obtenir l'historique des rappels

### Endpoints Legacy (Rendu côté serveur)

Ces endpoints servent du HTML rendu côté serveur ou du contenu spécialisé :

**Authentification** :
- `GET /api/v1/auth/callback` - Gestionnaire de callback OAuth2

**Routes Publiques** :
- `GET /` - SPA Vue.js 3 (sert toutes les routes frontend avec query params : `/?doc=xxx`, `/signatures`, `/admin`, etc.)
- `GET /health` - Health check (alias pour rétrocompatibilité)
- `GET /oembed?url=<document_url>` - Endpoint oEmbed pour découverte automatique d'embed (retourne JSON avec HTML iframe pointant vers `/embed?doc=xxx`)

### Exemples d'Utilisation de l'API

**Obtenir un token CSRF** (requis pour POST/PUT/DELETE authentifiés) :
```bash
curl -c cookies.txt http://localhost:8080/api/v1/csrf
# Retourne : {"csrf_token": "..."}
```

**Initier une connexion OAuth** :
```bash
curl -X POST http://localhost:8080/api/v1/auth/start \
  -H "Content-Type: application/json" \
  -d '{"redirect_to": "/?doc=politique_2025"}'
# Retourne : {"redirect_url": "https://accounts.google.com/..."}
```

**Obtenir le profil utilisateur actuel** :
```bash
curl -b cookies.txt http://localhost:8080/api/v1/users/me
# Retourne : {"sub": "...", "email": "...", "name": "...", "is_admin": false}
```

**Créer une signature** :
```bash
curl -X POST http://localhost:8080/api/v1/signatures \
  -b cookies.txt \
  -H "Content-Type: application/json" \
  -H "X-CSRF-Token: VOTRE_TOKEN_CSRF" \
  -d '{"doc_id": "politique_2025"}'
# Retourne : {"doc_id": "politique_2025", "user_email": "...", "signed_at": "..."}
```

**Lister les documents avec signatures** :
```bash
curl http://localhost:8080/api/v1/documents?limit=10&offset=0
# Retourne : {"documents": [...], "total": 42}
```

**Obtenir les signatures d'un document** (public) :
```bash
curl http://localhost:8080/api/v1/documents/politique_2025/signatures
# Retourne : {"doc_id": "politique_2025", "signatures": [...]}
```

**Admin : Ajouter des signataires attendus** :
```bash
curl -X POST http://localhost:8080/api/v1/admin/documents/politique_2025/signers \
  -b cookies.txt \
  -H "Content-Type: application/json" \
  -H "X-CSRF-Token: VOTRE_TOKEN_CSRF" \
  -d '{"email": "jean@entreprise.com", "name": "Jean Dupont"}'
```

**Admin : Envoyer des rappels email** :
```bash
curl -X POST http://localhost:8080/api/v1/admin/documents/politique_2025/reminders \
  -b cookies.txt \
  -H "Content-Type: application/json" \
  -H "X-CSRF-Token: VOTRE_TOKEN_CSRF" \
  -d '{"emails": ["jean@entreprise.com", "marie@entreprise.com"]}'
# Retourne : {"sent": 2, "failed": 0, "errors": []}
```

**Admin : Mettre à jour les métadonnées du document** :
```bash
curl -X PUT http://localhost:8080/api/v1/admin/documents/politique_2025/metadata \
  -b cookies.txt \
  -H "Content-Type: application/json" \
  -H "X-CSRF-Token: VOTRE_TOKEN_CSRF" \
  -d '{"title": "Politique de Sécurité 2025", "url": "https://docs.entreprise.com/politique", "checksum": "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", "checksum_algorithm": "SHA-256", "description": "Politique de sécurité de l'\''entreprise"}'
# Retourne : {"docId": "politique_2025", "title": "Politique de Sécurité 2025", ...}
```

**Découverte oEmbed** (intégration automatique dans les éditeurs modernes) :
```bash
# Obtenir les données oEmbed pour une URL de document
curl "http://localhost:8080/oembed?url=http://localhost:8080/?doc=politique_2025"
# Retourne :
# {
#   "type": "rich",
#   "version": "1.0",
#   "title": "Document politique_2025 - Confirmations de lecture",
#   "provider_name": "Ackify",
#   "provider_url": "http://localhost:8080",
#   "html": "<iframe src=\"http://localhost:8080/embed?doc=politique_2025\" width=\"100%\" height=\"200\" frameborder=\"0\" style=\"border: 1px solid #ddd; border-radius: 6px;\" allowtransparency=\"true\"></iframe>",
#   "height": 200
# }
```

**Comment ça fonctionne** :
1. L'utilisateur colle `https://votre-domaine.com/?doc=politique_2025` dans Notion, Outline, Confluence, etc.
2. L'éditeur découvre le endpoint oEmbed via la balise meta `<link rel="alternate" type="application/json+oembed">`
3. L'éditeur appelle `/oembed?url=https://votre-domaine.com/?doc=politique_2025`
4. Ackify retourne du JSON avec une iframe pointant vers `/embed?doc=politique_2025`
5. L'éditeur affiche le widget de signature intégré

**Plateformes supportées** : Notion, Outline, Confluence, AppFlowy, et toute plateforme supportant la découverte oEmbed.

### Contrôle d'Accès

Définir `ACKIFY_ADMIN_EMAILS` avec une liste d'emails admin séparés par des virgules (correspondance exacte, insensible à la casse) :
```bash
ACKIFY_ADMIN_EMAILS="alice@entreprise.com,bob@entreprise.com"
```

**Fonctionnalités Admin** :
- Gestion des métadonnées de documents (titre, URL, checksum, description)
- Suivi des signataires attendus avec stats de complétion
- Rappels email avec historique
- Suppression de document (incluant toutes les métadonnées et signatures)
- Statistiques complètes de documents et signatures

#### Gestion des Métadonnées de Documents
Les administrateurs peuvent gérer des métadonnées complètes pour chaque document :
- **Stocker les informations** : Titre, URL/emplacement, checksum, description
- **Vérification d'intégrité** : Support pour les checksums SHA-256, SHA-512 et MD5
- **Accès facile** : Copie en un clic pour les checksums, URLs de documents cliquables
- **Horodatages automatiques** : Suivi de la création et des mises à jour avec triggers PostgreSQL
- **Intégration email** : URL du document automatiquement incluse dans les emails de rappel

#### Fonctionnalité Signataires Attendus
Les administrateurs peuvent définir et suivre les signataires attendus pour chaque document :
- **Ajouter des signataires** : Coller des emails séparés par des sauts de ligne, virgules ou point-virgules
- **Support des noms** : Utiliser le format "Nom <email@example.com>" pour les emails personnalisés
- **Suivre la complétion** : Barre de progression visuelle avec pourcentage
- **Monitorer le statut** : Voir qui a signé (✓) vs. qui est en attente (⏳)
- **Rappels par email** : Envoyer des rappels en masse ou sélectifs dans la langue de l'utilisateur
- **Détecter les signatures inattendues** : Identifier les utilisateurs qui ont signé sans être attendus
- **Partage facile** : Copie en un clic du lien de signature du document
- **Gestion en masse** : Ajouter/retirer des signataires individuellement ou en lot

---

## 🔍 Développement & Tests

### Couverture des Tests

**État Actuel** : **72.6% de couverture de code** (tests unitaires + intégration)

Notre suite de tests complète inclut :
- ✅ **180+ tests unitaires** couvrant la logique métier, services et utilitaires
- ✅ **33 tests d'intégration** avec PostgreSQL pour la couche repository
- ✅ Tests **cryptographie Ed25519** (90% de couverture)
- ✅ Tests **handlers HTTP & middleware** (80%+ de couverture)
- ✅ Tests **modèles domaine** (100% de couverture)
- ✅ Tests **services email** avec mocks
- ✅ Tests **sécurité OAuth2** avec cas limites

**Couverture par Package** :
| Package | Couverture | Statut |
|---------|------------|--------|
| `domain/models` | 100% | ✅ Complet |
| `presentation/api/health` | 100% | ✅ Complet |
| `presentation/api/users` | 100% | ✅ Complet |
| `pkg/logger` | 100% | ✅ Complet |
| `pkg/services` | 100% | ✅ Complet |
| `presentation/api/signatures` | 95.2% | ✅ Excellent |
| `presentation/api/auth` | 92.3% | ✅ Excellent |
| `application/services` | 90.6% | ✅ Excellent |
| `pkg/crypto` | 90.0% | ✅ Excellent |
| `presentation/handlers` | 85.6% | ✅ Très Bon |
| `presentation/api/admin` | 84.2% | ✅ Très Bon |
| `presentation/api/shared` | 80.0% | ✅ Très Bon |

Tous les tests s'exécutent automatiquement dans **GitHub Actions CI/CD** à chaque push et pull request. Les rapports de couverture sont envoyés vers Codecov pour le suivi et l'analyse.

### Configuration de Développement Local

**Prérequis** :
- Go 1.24.5+
- Node.js 22+ et npm
- PostgreSQL 16+
- Docker & Docker Compose (optionnel mais recommandé)

**Développement backend** :
```bash
# Naviguer vers le backend
cd backend

# Installer les dépendances Go
go mod download

# Build backend
go build ./cmd/community

# Exécuter les migrations de base de données
go run ./cmd/migrate

# Lancer le backend (port 8080)
./community

# Linting & formatage
go fmt ./...
go vet ./...

# Exécuter les tests unitaires uniquement
go test -v -short ./...

# Exécuter les tests unitaires avec couverture
go test -coverprofile=coverage.out ./internal/... ./pkg/...

# Exécuter les tests d'intégration (nécessite PostgreSQL)
docker compose -f ../compose.test.yml up -d
INTEGRATION_TESTS=1 go test -tags=integration -v ./internal/infrastructure/database/
docker compose -f ../compose.test.yml down

# Exécuter tous les tests (unitaires + intégration) avec couverture
docker compose -f ../compose.test.yml up -d
INTEGRATION_TESTS=1 go test -tags=integration -coverprofile=coverage.out ./...
docker compose -f ../compose.test.yml down

# Voir le rapport de couverture dans le navigateur
go tool cover -html=coverage.out

# Voir le résumé de couverture
go tool cover -func=coverage.out | tail -1

# Optionnel : analyse statique
go install honnef.co/go/tools/cmd/staticcheck@latest
staticcheck ./...
```

**Développement frontend** :
```bash
# Naviguer vers webapp
cd webapp

# Installer les dépendances
npm install

# Lancer le serveur de dev (port 5173 avec HMR)
npm run dev

# Build pour production
npm run build

# Prévisualiser le build de production
npm run preview

# Vérification des types
npm run type-check

# Vérifier la complétude des traductions i18n
npm run lint:i18n
```

### Développement Docker

**Option 1 : Stack complet avec Docker Compose** (recommandé) :
```bash
# Développement avec rechargement à chaud
docker compose -f compose.local.yml up -d

# Voir les logs
docker compose -f compose.local.yml logs -f ackify-ce

# Rebuild après modifications
docker compose -f compose.local.yml up -d --force-recreate ackify-ce --build

# Arrêter
docker compose -f compose.local.yml down
```

**Option 2 : Build et exécution manuels** :
```bash
# Build image de production
docker build -t ackify-ce:dev .

# Exécuter avec fichier d'environnement
docker run -p 8080:8080 --env-file .env ackify-ce:dev

# Exécuter avec PostgreSQL
docker compose up -d
```

### Commandes de Projet (Makefile)

```bash
# Build complet (backend + frontend)
make build

# Exécuter les tests
make test

# Nettoyer les artefacts de build
make clean

# Formater le code
make fmt

# Exécuter le linting
make lint
```

---

## 🤝 Support

### Aide & Documentation
- 🐛 **Issues** : [GitHub Issues](https://github.com/btouchard/ackify-ce/issues)
- 💬 **Discussions** : [GitHub Discussions](https://github.com/btouchard/ackify-ce/discussions)

### Licence AGPLv3
Distribué sous la licence GNU Affero General Public License v3.0.
Voir [LICENSE](LICENSE) pour plus de détails.

---

**Développé avec ❤️ par [Benjamin TOUCHARD](https://www.kolapsis.com)**
