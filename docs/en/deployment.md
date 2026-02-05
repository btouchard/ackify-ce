# Deployment

Production deployment guide with Docker Compose.

## Production with Docker Compose

### Recommended Architecture

```
[Internet] → [Reverse Proxy (Traefik/Nginx)] → [Ackify Container]
                                                        ↓
                                                 [PostgreSQL Container]
```

### Production compose.yml

See the `/compose.yml` file at the project root for complete configuration.

**Included services**:
- `ackify-migrate` - PostgreSQL migrations (run once)
- `ackify-ce` - Main application
- `ackify-db` - PostgreSQL 16

### Production .env Configuration

```bash
# Application
APP_DNS=sign.company.com
ACKIFY_BASE_URL=https://sign.company.com
ACKIFY_ORGANISATION="ACME Corporation"
ACKIFY_LOG_LEVEL=info

# Database (strong password)
POSTGRES_USER=ackifyr
POSTGRES_PASSWORD=$(openssl rand -base64 32)
POSTGRES_DB=ackify

# OAuth2
ACKIFY_OAUTH_PROVIDER=google
ACKIFY_OAUTH_CLIENT_ID=your_client_id
ACKIFY_OAUTH_CLIENT_SECRET=your_client_secret
ACKIFY_OAUTH_ALLOWED_DOMAIN=@company.com

# Security (generate with openssl)
ACKIFY_OAUTH_COOKIE_SECRET=$(openssl rand -base64 64)
ACKIFY_ED25519_PRIVATE_KEY=$(openssl rand -base64 64)

# Administration
ACKIFY_ADMIN_EMAILS=admin@company.com,cto@company.com
```

## Reverse Proxy

### Traefik

Add labels in `compose.yml`:

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

The Ackify Docker image includes a built-in healthcheck command for container orchestration.

### How It Works

The image includes a `HEALTHCHECK` directive that runs:
```
/app/ackify health
```

This command:
- Checks HTTP connectivity to the API server
- Verifies database connection via `/api/v1/health`
- Returns exit code 0 (healthy) or 1 (unhealthy)

### Default Configuration

```yaml
healthcheck:
  test: ["CMD", "/app/ackify", "health"]
  interval: 30s      # Check every 30 seconds
  timeout: 5s        # Timeout after 5 seconds
  start_period: 10s  # Wait 10s before first check
  retries: 3         # Mark unhealthy after 3 failures
```

### Monitoring Container Health

```bash
# Check container health status
docker compose ps

# View health check logs
docker inspect --format='{{json .State.Health}}' ackify-ce | jq

# Manual health check
docker compose exec ackify-ce /app/ackify health
```

### Integration with Orchestrators

**Kubernetes**: Use the health endpoint for liveness/readiness probes:
```yaml
livenessProbe:
  httpGet:
    path: /api/v1/health
    port: 8080
  initialDelaySeconds: 10
  periodSeconds: 30
```

**Docker Swarm**: The built-in healthcheck works automatically.

## Security Checklist

- ✅ HTTPS with valid certificate
- ✅ Strong secrets (64+ bytes)
- ✅ PostgreSQL SSL in production
- ✅ Restricted OAuth domain
- ✅ Logs in info mode
- ✅ Automatic backup
- ✅ Active monitoring
- ✅ Healthcheck configured

## Backup

```bash
# Daily PostgreSQL backup
docker compose exec -T ackify-db pg_dump -U ackifyr ackify | gzip > backup-$(date +%Y%m%d).sql.gz

# Restore
gunzip -c backup.sql.gz | docker compose exec -T ackify-db psql -U ackifyr ackify
```

## Update

```bash
# Pull new image
docker compose pull ackify-ce

# Restart
docker compose up -d

# Verify
docker compose logs -f ackify-ce
curl https://sign.company.com/api/v1/health
```

See [Getting Started](getting-started.md) for more details.
