# Database

Schéma PostgreSQL, migrations, et garanties d'intégrité.

## Vue d'Ensemble

Ackify utilise **PostgreSQL 16+** avec :
- Migrations versionnées SQL
- Contraintes d'intégrité strictes
- Triggers pour immutabilité
- Index pour performances

## Schéma Principal

### Table `signatures`

Stocke les signatures cryptographiques Ed25519.

```sql
CREATE TABLE signatures (
    id BIGSERIAL PRIMARY KEY,
    doc_id TEXT NOT NULL,
    user_sub TEXT NOT NULL,                 -- OAuth user ID (sub claim)
    user_email TEXT NOT NULL,
    user_name TEXT,                         -- Nom utilisateur (optionnel)
    signed_at TIMESTAMPTZ NOT NULL,
    payload_hash TEXT NOT NULL,             -- SHA-256 du payload
    signature TEXT NOT NULL,                -- Signature Ed25519 (base64)
    nonce TEXT NOT NULL,                    -- Anti-replay attack
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    referer TEXT,                           -- Source (optionnel)
    prev_hash TEXT,                         -- Hash de la signature précédente (chaînage)
    UNIQUE (doc_id, user_sub)              -- UNE signature par user/document
);

CREATE INDEX idx_signatures_doc_id ON signatures(doc_id);
CREATE INDEX idx_signatures_user_sub ON signatures(user_sub);
```

**Garanties** :
- ✅ Une signature par utilisateur/document (contrainte UNIQUE)
- ✅ Horodatage immutable via trigger PostgreSQL
- ✅ Chaînage hash (blockchain-like) via `prev_hash`
- ✅ Non-répudiation cryptographique (Ed25519)

### Table `documents`

Métadonnées des documents.

```sql
CREATE TABLE documents (
    doc_id TEXT PRIMARY KEY,
    title TEXT NOT NULL DEFAULT '',
    url TEXT NOT NULL DEFAULT '',           -- URL du document source
    checksum TEXT NOT NULL DEFAULT '',      -- SHA-256, SHA-512, ou MD5
    checksum_algorithm TEXT NOT NULL DEFAULT 'SHA-256',
    description TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by TEXT NOT NULL DEFAULT ''     -- user_sub de l'admin créateur
);
```

**Utilisation** :
- Titre, description affichés dans l'interface
- URL incluse dans les emails de rappel
- Checksum pour vérification d'intégrité (optionnel)

### Table `expected_signers`

Signataires attendus pour tracking.

```sql
CREATE TABLE expected_signers (
    id BIGSERIAL PRIMARY KEY,
    doc_id TEXT NOT NULL,
    email TEXT NOT NULL,
    name TEXT NOT NULL DEFAULT '',          -- Nom pour personnalisation
    added_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    added_by TEXT NOT NULL,                 -- Admin qui a ajouté
    notes TEXT,
    UNIQUE (doc_id, email)
);

CREATE INDEX idx_expected_signers_doc_id ON expected_signers(doc_id);
```

**Fonctionnalités** :
- Tracking de complétion (% signé)
- Envoi de rappels email
- Détection de signatures inattendues

### Table `reminder_logs`

Historique des rappels email.

```sql
CREATE TABLE reminder_logs (
    id BIGSERIAL PRIMARY KEY,
    doc_id TEXT NOT NULL,
    recipient_email TEXT NOT NULL,
    sent_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    sent_by TEXT NOT NULL,                  -- Admin qui a envoyé
    template_used TEXT NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('sent', 'failed', 'bounced')),
    error_message TEXT,
    FOREIGN KEY (doc_id, recipient_email)
        REFERENCES expected_signers(doc_id, email)
);

CREATE INDEX idx_reminder_logs_doc_id ON reminder_logs(doc_id);
```

### Table `checksum_verifications`

Historique des vérifications d'intégrité.

```sql
CREATE TABLE checksum_verifications (
    id BIGSERIAL PRIMARY KEY,
    doc_id TEXT NOT NULL,
    verified_by TEXT NOT NULL,
    verified_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    stored_checksum TEXT NOT NULL,
    calculated_checksum TEXT NOT NULL,
    algorithm TEXT NOT NULL,
    is_valid BOOLEAN NOT NULL,
    error_message TEXT,
    FOREIGN KEY (doc_id) REFERENCES documents(doc_id)
);

CREATE INDEX idx_checksum_verifications_doc_id ON checksum_verifications(doc_id);
```

### Table `oauth_sessions`

Sessions OAuth2 avec refresh tokens chiffrés.

```sql
CREATE TABLE oauth_sessions (
    id BIGSERIAL PRIMARY KEY,
    session_id TEXT NOT NULL UNIQUE,           -- ID session Gorilla
    user_sub TEXT NOT NULL,                    -- OAuth user ID
    refresh_token_encrypted BYTEA NOT NULL,    -- Chiffré AES-256-GCM
    access_token_expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_refreshed_at TIMESTAMPTZ,
    user_agent TEXT,
    ip_address INET
);

CREATE INDEX idx_oauth_sessions_session_id ON oauth_sessions(session_id);
CREATE INDEX idx_oauth_sessions_user_sub ON oauth_sessions(user_sub);
CREATE INDEX idx_oauth_sessions_updated_at ON oauth_sessions(updated_at);
```

**Sécurité** :
- Refresh tokens chiffrés (AES-256-GCM)
- Cleanup automatique après 37 jours
- Tracking IP + User-Agent pour détecter vols

### Table `email_queue`

File d'attente d'emails asynchrone avec mécanisme de retry.

```sql
CREATE TABLE email_queue (
    id BIGSERIAL PRIMARY KEY,

    -- Métadonnées email
    to_addresses TEXT[] NOT NULL,              -- Adresses email destinataires
    cc_addresses TEXT[],                       -- Adresses CC (optionnel)
    bcc_addresses TEXT[],                      -- Adresses BCC (optionnel)
    subject TEXT NOT NULL,                     -- Sujet de l'email
    template TEXT NOT NULL,                    -- Nom du template (ex: 'reminder')
    locale TEXT NOT NULL DEFAULT 'fr',         -- Langue email (en, fr, es, de, it)
    data JSONB NOT NULL DEFAULT '{}',          -- Variables du template
    headers JSONB,                             -- Headers email personnalisés (optionnel)

    -- Gestion de la file
    status TEXT NOT NULL DEFAULT 'pending'     -- pending, processing, sent, failed, cancelled
        CHECK (status IN ('pending', 'processing', 'sent', 'failed', 'cancelled')),
    priority INT NOT NULL DEFAULT 0,           -- Plus élevé = traité en premier (0=normal, 10=high, 100=urgent)
    retry_count INT NOT NULL DEFAULT 0,        -- Nombre de tentatives de retry
    max_retries INT NOT NULL DEFAULT 3,        -- Limite maximale de retry

    -- Suivi
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    scheduled_for TIMESTAMPTZ NOT NULL DEFAULT now(),  -- Heure de traitement la plus tôt
    processed_at TIMESTAMPTZ,                  -- Quand l'email a été envoyé
    next_retry_at TIMESTAMPTZ,                 -- Heure de retry calculée (exponential backoff)

    -- Suivi des erreurs
    last_error TEXT,                           -- Dernier message d'erreur
    error_details JSONB,                       -- Informations d'erreur détaillées

    -- Suivi des références (optionnel)
    reference_type TEXT,                       -- ex: 'reminder', 'notification'
    reference_id TEXT,                         -- ex: doc_id
    created_by TEXT                            -- Utilisateur qui a mis en file l'email
);

-- Index pour traitement efficace de la file
CREATE INDEX idx_email_queue_status_scheduled
    ON email_queue(status, scheduled_for)
    WHERE status IN ('pending', 'processing');

CREATE INDEX idx_email_queue_priority_scheduled
    ON email_queue(priority DESC, scheduled_for ASC)
    WHERE status = 'pending';

CREATE INDEX idx_email_queue_retry
    ON email_queue(next_retry_at)
    WHERE status = 'processing' AND retry_count < max_retries;

CREATE INDEX idx_email_queue_reference
    ON email_queue(reference_type, reference_id);

CREATE INDEX idx_email_queue_created_at
    ON email_queue(created_at DESC);
```

**Fonctionnalités** :
- **Traitement asynchrone** : Emails traités par worker en arrière-plan
- **Mécanisme de retry** : Exponential backoff (1min, 2min, 4min, 8min, 16min, 32min...)
- **Support de priorité** : Emails haute priorité traités en premier
- **Envoi programmé** : Retarder la livraison d'email avec `scheduled_for`
- **Suivi des erreurs** : Logging détaillé des erreurs et historique des retries
- **Suivi des références** : Lier les emails aux documents ou autres entités

**Calcul automatique du retry** :
```sql
-- Fonction pour calculer le temps de retry suivant avec exponential backoff
CREATE OR REPLACE FUNCTION calculate_next_retry_time(retry_count INT)
RETURNS TIMESTAMPTZ AS $$
BEGIN
    -- Exponential backoff: 1min, 2min, 4min, 8min, 16min, 32min...
    RETURN now() + (interval '1 minute' * power(2, retry_count));
END;
$$ LANGUAGE plpgsql;
```

**Configuration du worker** :
- Taille de lot : 10 emails par lot
- Intervalle de polling : 5 secondes
- Envois concurrents : 5 emails simultanés
- Cleanup des anciens emails : Rétention de 7 jours pour emails envoyés/échoués

## Migrations

### Gestion des Migrations

Les migrations sont dans `/backend/migrations/` avec le format :

```
XXXX_description.up.sql     # Migration "up"
XXXX_description.down.sql   # Rollback "down"
```

**Fichiers actuels** :
- `0001_init.up.sql` - Table signatures
- `0002_expected_signers.up.sql` - Expected signers
- `0003_reminder_logs.up.sql` - Reminder logs
- `0004_add_name_to_expected_signers.up.sql` - Noms signataires
- `0005_create_documents_table.up.sql` - Documents metadata
- `0006_create_new_tables.up.sql` - Checksum verifications et email queue
- `0007_oauth_sessions.up.sql` - OAuth sessions avec refresh tokens

### Appliquer les Migrations

**Via Docker Compose** (automatique) :

```bash
docker compose up -d
# Le service ackify-migrate applique les migrations au démarrage
```

**Manuellement** :

```bash
cd backend
go run ./cmd/migrate up
```

**Rollback dernière migration** :

```bash
go run ./cmd/migrate down
```

### Migrations Personnalisées

Pour créer une nouvelle migration :

1. Créer `XXXX_my_feature.up.sql` :
```sql
-- Migration up
ALTER TABLE signatures ADD COLUMN new_field TEXT;
```

2. Créer `XXXX_my_feature.down.sql` :
```sql
-- Rollback
ALTER TABLE signatures DROP COLUMN new_field;
```

3. Appliquer :
```bash
go run ./cmd/migrate up
```

## Triggers PostgreSQL

### Immutabilité de `created_at`

Trigger qui empêche la modification de `created_at` :

```sql
CREATE OR REPLACE FUNCTION prevent_created_at_update()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.created_at <> OLD.created_at THEN
        RAISE EXCEPTION 'created_at cannot be modified';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER prevent_signatures_created_at_update
    BEFORE UPDATE ON signatures
    FOR EACH ROW
    EXECUTE FUNCTION prevent_created_at_update();
```

**Garantie** : Aucune signature ne peut être backdatée.

### Auto-update de `updated_at`

Pour les tables avec `updated_at` :

```sql
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_documents_updated_at
    BEFORE UPDATE ON documents
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
```

## Requêtes Utiles

### Voir les signatures d'un document

```sql
SELECT
    user_email,
    user_name,
    signed_at,
    payload_hash,
    signature
FROM signatures
WHERE doc_id = 'my_document'
ORDER BY signed_at DESC;
```

### Statut de complétion

```sql
WITH expected AS (
    SELECT COUNT(*) as total
    FROM expected_signers
    WHERE doc_id = 'my_document'
),
signed AS (
    SELECT COUNT(*) as count
    FROM signatures s
    INNER JOIN expected_signers e ON s.user_email = e.email AND s.doc_id = e.doc_id
    WHERE s.doc_id = 'my_document'
)
SELECT
    e.total as expected,
    s.count as signed,
    ROUND(100.0 * s.count / NULLIF(e.total, 0), 2) as completion_pct
FROM expected e, signed s;
```

### Signataires manquants

```sql
SELECT
    e.email,
    e.name,
    e.added_at
FROM expected_signers e
LEFT JOIN signatures s ON e.email = s.user_email AND e.doc_id = s.doc_id
WHERE e.doc_id = 'my_document' AND s.id IS NULL
ORDER BY e.added_at;
```

### Signatures inattendues

```sql
SELECT
    s.user_email,
    s.signed_at
FROM signatures s
LEFT JOIN expected_signers e ON s.user_email = e.email AND s.doc_id = e.doc_id
WHERE s.doc_id = 'my_document' AND e.id IS NULL
ORDER BY s.signed_at DESC;
```

### Statut de la file d'emails

```sql
-- Voir les emails en attente
SELECT
    id,
    to_addresses,
    subject,
    status,
    priority,
    retry_count,
    scheduled_for,
    created_at
FROM email_queue
WHERE status IN ('pending', 'processing')
ORDER BY priority DESC, scheduled_for ASC
LIMIT 20;

-- Emails échoués nécessitant attention
SELECT
    id,
    to_addresses,
    subject,
    retry_count,
    max_retries,
    last_error,
    next_retry_at
FROM email_queue
WHERE status = 'failed'
ORDER BY created_at DESC;

-- Statistiques des emails par statut
SELECT
    status,
    COUNT(*) as count,
    MIN(created_at) as oldest,
    MAX(created_at) as newest
FROM email_queue
GROUP BY status
ORDER BY status;
```

## Sauvegarde & Restauration

### Backup PostgreSQL

```bash
# Backup complet
docker compose exec ackify-db pg_dump -U ackifyr ackify > backup.sql

# Backup avec compression
docker compose exec ackify-db pg_dump -U ackifyr ackify | gzip > backup.sql.gz
```

### Restore

```bash
# Restore depuis backup
cat backup.sql | docker compose exec -T ackify-db psql -U ackifyr ackify

# Restore depuis backup compressé
gunzip -c backup.sql.gz | docker compose exec -T ackify-db psql -U ackifyr ackify
```

### Backup Automatisé

Exemple de cron pour backup quotidien :

```bash
0 2 * * * docker compose -f /path/to/compose.yml exec -T ackify-db pg_dump -U ackifyr ackify | gzip > /backups/ackify-$(date +\%Y\%m\%d).sql.gz
```

## Performance

### Index

Les index sont automatiquement créés pour :
- `signatures(doc_id)` - Requêtes par document
- `signatures(user_sub)` - Requêtes par utilisateur
- `expected_signers(doc_id)` - Tracking complétion
- `oauth_sessions(session_id)` - Lookup sessions

### Connection Pooling

Le backend Go gère automatiquement le pooling de connexions :
- Max open connections : 25
- Max idle connections : 5
- Connection max lifetime : 5 minutes

### Vacuum & Analyze

PostgreSQL gère automatiquement via `autovacuum`. Pour forcer :

```sql
VACUUM ANALYZE signatures;
VACUUM ANALYZE documents;
```

## Monitoring

### Taille des tables

```sql
SELECT
    schemaname,
    tablename,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS size
FROM pg_tables
WHERE schemaname = 'public'
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;
```

### Statistiques

```sql
SELECT * FROM pg_stat_user_tables WHERE schemaname = 'public';
```

### Connexions actives

```sql
SELECT
    datname,
    usename,
    application_name,
    client_addr,
    state,
    query
FROM pg_stat_activity
WHERE datname = 'ackify';
```

## Sécurité

### En Production

- ✅ Utiliser SSL : `?sslmode=require` dans le DSN
- ✅ Mot de passe fort pour PostgreSQL
- ✅ Restreindre les connexions réseau
- ✅ Sauvegardes chiffrées
- ✅ Rotation régulière des secrets

### Configuration SSL

```bash
# Dans .env
ACKIFY_DB_DSN=postgres://user:pass@host:5432/ackify?sslmode=require
```

### Audit Trail

Toutes les opérations importantes sont tracées :
- `signatures.created_at` - Horodatage signature
- `expected_signers.added_by` - Qui a ajouté
- `reminder_logs.sent_by` - Qui a envoyé le rappel
- `checksum_verifications.verified_by` - Qui a vérifié

## Troubleshooting

### Migrations bloquées

```bash
# Vérifier le statut
docker compose logs ackify-migrate

# Forcer le rollback
docker compose exec ackify-ce /app/migrate down
docker compose exec ackify-ce /app/migrate up
```

### Contrainte UNIQUE violée

Erreur : `duplicate key value violates unique constraint`

**Cause** : L'utilisateur a déjà signé ce document.

**Solution** : C'est un comportement normal (une signature par user/doc).

### Connection refused

Vérifier que PostgreSQL est démarré :

```bash
docker compose ps ackify-db
docker compose logs ackify-db
```
