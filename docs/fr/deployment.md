# Deployment

Guide de déploiement en production avec Docker Compose.

## Production avec Docker Compose

### Architecture Recommandée

```
[Internet] → [Reverse Proxy (Traefik/Nginx)] → [Ackify Container]
                                                        ↓
                                                 [PostgreSQL Container]
```

### compose.yml Production

Voir le fichier `/compose.yml` à la racine du projet pour la configuration complète.

**Services inclus** :
- `ackify-migrate` - Migrations PostgreSQL (run once)
- `ackify-ce` - Application principale
- `ackify-db` - PostgreSQL 16

### Configuration .env Production

```bash
# Application
APP_DNS=sign.company.com
ACKIFY_BASE_URL=https://sign.company.com
ACKIFY_ORGANISATION="ACME Corporation"
ACKIFY_LOG_LEVEL=info

# Base de données (mot de passe fort)
POSTGRES_USER=ackifyr
POSTGRES_PASSWORD=$(openssl rand -base64 32)
POSTGRES_DB=ackify

# OAuth2
ACKIFY_OAUTH_PROVIDER=google
ACKIFY_OAUTH_CLIENT_ID=your_client_id
ACKIFY_OAUTH_CLIENT_SECRET=your_client_secret
ACKIFY_OAUTH_ALLOWED_DOMAIN=@company.com

# Sécurité (générer avec openssl)
ACKIFY_OAUTH_COOKIE_SECRET=$(openssl rand -base64 64)
ACKIFY_ED25519_PRIVATE_KEY=$(openssl rand -base64 64)

# Administration
ACKIFY_ADMIN_EMAILS=admin@company.com,cto@company.com
```

## Reverse Proxy

### Traefik

Ajouter les labels dans `compose.yml` :

```yaml
services:
  ackify-ce:
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.ackify.rule=Host(`sign.company.com`)"
      - "traefik.http.routers.ackify.entrypoints=websecure"
      - "traefik.http.routers.ackify.tls.certresolver=letsencrypt"
```

### Nginx

```nginx
server {
    listen 443 ssl http2;
    server_name sign.company.com;

    ssl_certificate /etc/letsencrypt/live/sign.company.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/sign.company.com/privkey.pem;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

## Docker Healthcheck

L'image Docker Ackify inclut une commande de healthcheck intégrée pour l'orchestration de conteneurs.

### Fonctionnement

L'image inclut une directive `HEALTHCHECK` qui exécute :
```
/app/ackify health
```

Cette commande :
- Vérifie la connectivité HTTP vers le serveur API
- Vérifie la connexion à la base de données via `/api/v1/health`
- Retourne le code de sortie 0 (healthy) ou 1 (unhealthy)

### Configuration par Défaut

```yaml
healthcheck:
  test: ["CMD", "/app/ackify", "health"]
  interval: 30s      # Vérification toutes les 30 secondes
  timeout: 5s        # Timeout après 5 secondes
  start_period: 10s  # Attendre 10s avant la première vérification
  retries: 3         # Marquer unhealthy après 3 échecs
```

### Surveiller la Santé du Conteneur

```bash
# Vérifier le statut de santé du conteneur
docker compose ps

# Voir les logs de healthcheck
docker inspect --format='{{json .State.Health}}' ackify-ce | jq

# Vérification manuelle
docker compose exec ackify-ce /app/ackify health
```

### Intégration avec les Orchestrateurs

**Kubernetes** : Utilisez l'endpoint health pour les probes liveness/readiness :
```yaml
livenessProbe:
  httpGet:
    path: /api/v1/health
    port: 8080
  initialDelaySeconds: 10
  periodSeconds: 30
```

**Docker Swarm** : Le healthcheck intégré fonctionne automatiquement.

## Checklist Sécurité

- ✅ HTTPS avec certificat valide
- ✅ Secrets forts (64+ bytes)
- ✅ PostgreSQL SSL en production
- ✅ Domaine OAuth restreint
- ✅ Logs en mode info
- ✅ Backup automatique
- ✅ Monitoring actif
- ✅ Healthcheck configuré

## Backup

```bash
# Backup quotidien PostgreSQL
docker compose exec -T ackify-db pg_dump -U ackifyr ackify | gzip > backup-$(date +%Y%m%d).sql.gz

# Restauration
gunzip -c backup.sql.gz | docker compose exec -T ackify-db psql -U ackifyr ackify
```

## Mise à Jour

```bash
# Pull nouvelle image
docker compose pull ackify-ce

# Redémarrer
docker compose up -d

# Vérifier
docker compose logs -f ackify-ce
curl https://sign.company.com/api/v1/health
```

Voir [Getting Started](getting-started.md) pour plus de détails.
