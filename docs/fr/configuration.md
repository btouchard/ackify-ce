# Configuration

Guide complet de configuration d'Ackify via variables d'environnement.

## Variables Obligatoires

Ces variables sont **requises** pour démarrer Ackify :

```bash
# URL publique de votre instance (utilisée pour OAuth callbacks)
APP_DNS=sign.your-domain.com
ACKIFY_BASE_URL=https://sign.your-domain.com

# Nom de votre organisation (affiché dans l'interface)
ACKIFY_ORGANISATION="Your Organization Name"

# Configuration PostgreSQL
POSTGRES_USER=ackifyr
POSTGRES_PASSWORD=your_secure_password
POSTGRES_DB=ackify

# OAuth2 Provider
ACKIFY_OAUTH_PROVIDER=google  # ou github, gitlab, ou vide pour custom
ACKIFY_OAUTH_CLIENT_ID=your_oauth_client_id
ACKIFY_OAUTH_CLIENT_SECRET=your_oauth_client_secret

# Secret pour chiffrer les cookies de session (générer avec: openssl rand -base64 32)
ACKIFY_OAUTH_COOKIE_SECRET=your_base64_encoded_secret_key
```

## Variables Optionnelles

### Serveur

```bash
# Adresse d'écoute HTTP (défaut: :8080)
ACKIFY_LISTEN_ADDR=:8080

# Niveau de logs: debug, info, warn, error (défaut: info)
ACKIFY_LOG_LEVEL=info
```

### Sécurité & OAuth2

```bash
# Restreindre l'accès à un domaine email spécifique
ACKIFY_OAUTH_ALLOWED_DOMAIN=@company.com

# Activer l'auto-login silencieux (défaut: false)
ACKIFY_OAUTH_AUTO_LOGIN=false

# URL de logout personnalisée (optionnel)
ACKIFY_OAUTH_LOGOUT_URL=https://your-provider.com/logout

# Scopes OAuth2 personnalisés (défaut: openid,email,profile)
ACKIFY_OAUTH_SCOPES=openid,email,profile
```

### Méthodes d'Authentification

**Important** : Au moins UNE méthode d'authentification doit être activée (OAuth ou MagicLink).

```bash
# Forcer l'activation/désactivation d'OAuth (défaut: auto-détecté depuis les credentials)
ACKIFY_AUTH_OAUTH_ENABLED=true

# Activer l'authentification MagicLink sans mot de passe (défaut: false)
# Nécessite que ACKIFY_MAIL_HOST soit configuré
ACKIFY_AUTH_MAGICLINK_ENABLED=true
```

**Auto-détection** :
- **OAuth** est automatiquement activé si `ACKIFY_OAUTH_CLIENT_ID` et `ACKIFY_OAUTH_CLIENT_SECRET` sont définis
- **MagicLink** nécessite une activation explicite avec `ACKIFY_AUTH_MAGICLINK_ENABLED=true` + configuration SMTP
- **Service SMTP/Email** est automatiquement activé quand `ACKIFY_MAIL_HOST` est configuré (indépendant de MagicLink)

**Note** : SMTP et MagicLink sont deux fonctionnalités distinctes :
- **SMTP** = Service d'envoi de rappels email aux signataires attendus (auto-détecté)
- **MagicLink** = Authentification sans mot de passe par email (nécessite activation explicite + SMTP)

### Administration

```bash
# Liste d'emails admin (séparés par virgules)
ACKIFY_ADMIN_EMAILS=admin@company.com,admin2@company.com

# Restreindre la création de documents aux admins uniquement (défaut: false)
ACKIFY_ONLY_ADMIN_CAN_CREATE=false
```

Les admins ont accès à :
- Dashboard admin (`/admin`)
- Gestion des métadonnées documents
- Tracking des signataires attendus
- Envoi de rappels email
- Suppression de documents

Quand `ACKIFY_ONLY_ADMIN_CAN_CREATE` est activé :
- ✅ Seuls les utilisateurs admin peuvent créer de nouveaux documents
- ✅ Les utilisateurs non-admin verront un message d'erreur lors d'une tentative de création
- ✅ Les deux endpoints API (`POST /documents` et `GET /documents/find-or-create`) sont protégés

### Limitation de Débit (Rate Limiting)

Configuration des limites de requêtes API pour prévenir les abus et contrôler le débit :

```bash
# Limites d'authentification Magic Link (par fenêtre de temps)
ACKIFY_AUTH_MAGICLINK_RATE_LIMIT_EMAIL=3   # Max requêtes par email (défaut: 3)
ACKIFY_AUTH_MAGICLINK_RATE_LIMIT_IP=10     # Max requêtes par IP (défaut: 10)

# Limites API générales (requêtes par minute)
ACKIFY_AUTH_RATE_LIMIT=5          # Endpoints d'authentification (défaut: 5/min)
ACKIFY_DOCUMENT_RATE_LIMIT=10     # Création de documents (défaut: 10/min)
ACKIFY_GENERAL_RATE_LIMIT=100     # Requêtes API générales (défaut: 100/min)

# Import CSV
ACKIFY_IMPORT_MAX_SIGNERS=500     # Max signataires par import CSV (défaut: 500)
```

**Quand ajuster** :
- **Tests/CI** : Augmenter les limites (ex: `1000`) pour éviter les erreurs 429 durant les tests automatisés
- **Trafic élevé** : Augmenter `GENERAL_RATE_LIMIT` pour les charges de production
- **Sécurité** : Réduire `AUTH_RATE_LIMIT` pour prévenir les attaques par force brute

### Journalisation (Logging)

```bash
# Niveau de log : debug, info, warn, error (défaut: info)
ACKIFY_LOG_LEVEL=info

# Format de log : classic ou json (défaut: classic)
ACKIFY_LOG_FORMAT=classic
```

**Formats de log** :
- `classic` : Format lisible pour le développement et déploiements simples
- `json` : JSON structuré pour aggrégateurs de logs (Datadog, ELK, Splunk)

**Exemple de sortie JSON** :
```json
{"time":"2025-11-24T10:00:00Z","level":"INFO","msg":"Server started","port":8080}
```

### Checksums Documents (Optionnel)

Configuration pour le calcul automatique de checksum lors de la création de documents depuis des URLs :

```bash
# Taille maximale de fichier à télécharger pour le calcul de checksum (défaut: 10485760 = 10MB)
ACKIFY_CHECKSUM_MAX_BYTES=10485760

# Timeout pour le téléchargement checksum en millisecondes (défaut: 5000ms = 5s)
ACKIFY_CHECKSUM_TIMEOUT_MS=5000

# Nombre maximum de redirections HTTP à suivre (défaut: 3)
ACKIFY_CHECKSUM_MAX_REDIRECTS=3

# Liste de types MIME autorisés séparés par virgules (défaut inclut PDF, images, docs Office, ODF)
ACKIFY_CHECKSUM_ALLOWED_TYPES=application/pdf,image/*,application/msword,application/vnd.openxmlformats-officedocument.wordprocessingml.document,application/vnd.ms-excel,application/vnd.openxmlformats-officedocument.spreadsheetml.sheet,application/vnd.oasis.opendocument.*
```

**Note** : Ces paramètres s'appliquent uniquement lorsque les admins créent des documents via le dashboard admin avec une URL distante. Le système tentera de télécharger et calculer le checksum SHA-256 automatiquement.

**Variables de test** (⚠️ **NE JAMAIS utiliser en production**) :
```bash
# Désactiver la protection SSRF (tests uniquement)
ACKIFY_CHECKSUM_SKIP_SSRF_CHECK=false

# Ignorer la vérification des certificats TLS (tests uniquement)
ACKIFY_CHECKSUM_INSECURE_SKIP_VERIFY=false
```

Ces variables désactivent des protections de sécurité critiques et ne doivent être utilisées **que** dans des environnements de test isolés.

## Configuration Avancée

### OAuth2 Providers

Voir [OAuth Providers](configuration/oauth-providers.md) pour la configuration détaillée de :
- Google OAuth2
- GitHub OAuth2
- GitLab OAuth2 (public + self-hosted)
- Custom OAuth2 provider

### Email (SMTP)

Voir [Email Setup](configuration/email-setup.md) pour configurer l'envoi de rappels email.

## Exemple Complet

Exemple de `.env` pour une installation en production :

```bash
# Application
APP_DNS=sign.company.com
ACKIFY_BASE_URL=https://sign.company.com
ACKIFY_ORGANISATION="ACME Corporation"
ACKIFY_LOG_LEVEL=info
ACKIFY_LISTEN_ADDR=:8080

# Base de données
POSTGRES_USER=ackifyr
POSTGRES_PASSWORD=super_secure_password_123
POSTGRES_DB=ackify

# OAuth2 (Google)
ACKIFY_OAUTH_PROVIDER=google
ACKIFY_OAUTH_CLIENT_ID=123456789-abc.apps.googleusercontent.com
ACKIFY_OAUTH_CLIENT_SECRET=GOCSPX-xyz123
ACKIFY_OAUTH_ALLOWED_DOMAIN=@company.com

# Sécurité
ACKIFY_OAUTH_COOKIE_SECRET=ZXhhbXBsZV9iYXNlNjRfc2VjcmV0X2tleQ==

# Administration
ACKIFY_ADMIN_EMAILS=admin@company.com,cto@company.com

# Email (optionnel - omettre MAIL_HOST pour désactiver)
ACKIFY_MAIL_HOST=smtp.gmail.com
ACKIFY_MAIL_PORT=587
ACKIFY_MAIL_USERNAME=noreply@company.com
ACKIFY_MAIL_PASSWORD=app_specific_password
ACKIFY_MAIL_FROM=noreply@company.com
ACKIFY_MAIL_FROM_NAME="Ackify - ACME"
ACKIFY_MAIL_TEMPLATE_DIR=templates/emails
ACKIFY_MAIL_DEFAULT_LOCALE=fr

# Checksums Documents (optionnel - pour auto-checksum depuis URLs)
ACKIFY_CHECKSUM_MAX_BYTES=10485760
ACKIFY_CHECKSUM_TIMEOUT_MS=5000
ACKIFY_CHECKSUM_MAX_REDIRECTS=3
```

## Validation de la Configuration

Après modification du `.env`, redémarrer :

```bash
docker compose restart ackify-ce
```

Vérifier les logs :

```bash
docker compose logs -f ackify-ce
```

Tester le health check :

```bash
curl http://localhost:8080/api/v1/health
```

## Variables de Production

**Checklist sécurité production** :

- ✅ Utiliser HTTPS (`ACKIFY_BASE_URL=https://...`)
- ✅ Générer des secrets forts (64+ caractères)
- ✅ Restreindre le domaine OAuth (`ACKIFY_OAUTH_ALLOWED_DOMAIN`)
- ✅ Configurer les emails admin (`ACKIFY_ADMIN_EMAILS`)
- ✅ Utiliser PostgreSQL avec SSL en production
- ✅ Logger en mode `info` (pas `debug`)
- ✅ Sauvegarder régulièrement la base de données

Voir [Deployment](deployment.md) pour plus de détails sur le déploiement en production.
