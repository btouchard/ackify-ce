# Row Level Security (RLS)

PostgreSQL Row Level Security fournit une isolation automatique des données par tenant au niveau de la base de données.

## Vue d'ensemble

RLS garantit que chaque tenant ne peut accéder qu'à ses propres données, peu importe comment l'application interroge la base de données. C'est une fonctionnalité de sécurité critique pour les déploiements multi-tenant.

```
┌─────────────────────────────────────────────────────────────────┐
│                    Flux de requête                              │
├─────────────────────────────────────────────────────────────────┤
│ 1. Requête HTTP arrive                                          │
│ 2. Le middleware RLS démarre une transaction                    │
│ 3. Le middleware définit : SET app.tenant_id = '<tenant-uuid>'  │
│ 4. Toutes les requêtes filtrées automatiquement par tenant_id   │
│ 5. Transaction validée si succès, annulée si erreur             │
└─────────────────────────────────────────────────────────────────┘
```

## Configuration

### Variable requise

```bash
# Mot de passe pour le rôle base de données ackify_app
ACKIFY_APP_PASSWORD=your_secure_password
```

### Fonctionnement

1. **Pendant la migration** (`migrate up`) :
   - L'outil migrate lit `ACKIFY_APP_PASSWORD`
   - Crée le rôle `ackify_app` s'il n'existe pas
   - Met à jour le mot de passe si le rôle existe déjà
   - Exécute les migrations qui activent les policies RLS

2. **À l'exécution** :
   - L'application se connecte en tant que `ackify_app` (pas `postgres`)
   - Les policies RLS filtrent toutes les requêtes par `tenant_id`
   - Aucune fuite de données possible

### Configuration compose.yml

```yaml
services:
  ackify-migrate:
    environment:
      # Connexion superuser pour les migrations
      ACKIFY_DB_DSN: "postgres://postgres:${POSTGRES_PASSWORD}@db:5432/ackify?sslmode=disable"
      # Mot de passe pour la création du rôle ackify_app
      ACKIFY_APP_PASSWORD: "${ACKIFY_APP_PASSWORD}"

  ackify-ce:
    environment:
      # L'application se connecte avec le rôle ackify_app (RLS appliqué)
      ACKIFY_DB_DSN: "postgres://ackify_app:${ACKIFY_APP_PASSWORD}@db:5432/ackify?sslmode=disable"
```

## Avantages sécurité

### Filtrage automatique

Sans RLS, le code applicatif doit toujours inclure le filtrage tenant :

```sql
-- Sans RLS : Facile d'oublier le filtre tenant_id
SELECT * FROM documents WHERE doc_id = '123';  -- BUG : Retourne les données de n'importe quel tenant !
```

Avec RLS, le filtrage est automatique :

```sql
-- Avec RLS : La base de données impose l'isolation tenant
SELECT * FROM documents WHERE doc_id = '123';  -- Retourne uniquement les données du tenant courant
```

### Défense en profondeur

Même si le code applicatif contient un bug qui oublie le filtrage tenant, RLS empêche les fuites de données au niveau base de données.

## Tables avec RLS

Les policies RLS sont appliquées à toutes les tables tenant-aware :

| Table | Policy |
|-------|--------|
| `documents` | `tenant_id = current_tenant_id()` |
| `signatures` | `tenant_id = current_tenant_id()` |
| `expected_signers` | `tenant_id = current_tenant_id()` |
| `webhooks` | `tenant_id = current_tenant_id()` |
| `reminder_logs` | `tenant_id = current_tenant_id()` |
| `email_queue` | `tenant_id = current_tenant_id()` |
| `checksum_verifications` | `tenant_id = current_tenant_id()` |
| `webhook_deliveries` | `tenant_id = current_tenant_id()` |
| `oauth_sessions` | `tenant_id = current_tenant_id()` |
| `magic_link_tokens` | `tenant_id IS NULL OR tenant_id = current_tenant_id()` |
| `magic_link_auth_attempts` | `tenant_id IS NULL OR tenant_id = current_tenant_id()` |

## Dépannage

### Résultats vides lors de requêtes directes

Si vous vous connectez à la base avec `psql` et obtenez des résultats vides :

```sql
-- Retourne 0 lignes car app.tenant_id n'est pas défini
SELECT COUNT(*) FROM documents;
```

**Solution** : Définir le contexte tenant d'abord :

```sql
-- Option 1 : Niveau session (persiste jusqu'à déconnexion)
SELECT set_config('app.tenant_id', 'votre-tenant-uuid', false);

-- Option 2 : Niveau transaction
BEGIN;
SELECT set_config('app.tenant_id', 'votre-tenant-uuid', true);
SELECT * FROM documents;
COMMIT;
```

### Le superuser contourne RLS

Si vous vous connectez en tant que `postgres` (superuser), RLS est contourné :

```sql
-- En tant que postgres : Retourne TOUTES les données (pas de filtrage RLS)
SELECT COUNT(*) FROM documents;
```

C'est voulu. Utilisez `ackify_app` pour les connexions applicatives.

### La migration échoue avec "role does not exist"

Si les migrations échouent parce que `ackify_app` n'existe pas :

1. Vérifiez que `ACKIFY_APP_PASSWORD` est défini
2. Consultez les logs du migrate tool pour les warnings
3. Vérifiez que le migrate tool s'exécute avant les migrations

## Gestion manuelle du rôle

Dans de rares cas, vous pourriez avoir besoin de gérer le rôle manuellement :

```sql
-- Créer le rôle (si vous n'utilisez pas le migrate tool)
CREATE ROLE ackify_app WITH
    LOGIN
    PASSWORD 'votre_mot_de_passe'
    NOCREATEDB
    NOCREATEROLE
    NOINHERIT;

-- Accorder les permissions
GRANT CONNECT ON DATABASE ackify TO ackify_app;
GRANT USAGE ON SCHEMA public TO ackify_app;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO ackify_app;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO ackify_app;

-- Changer le mot de passe
ALTER ROLE ackify_app WITH PASSWORD 'nouveau_mot_de_passe';
```

## Tester RLS

Pour vérifier que RLS fonctionne correctement :

```bash
# Se connecter en tant que ackify_app
psql -U ackify_app -d ackify

# Sans contexte tenant - devrait retourner 0 lignes
SELECT COUNT(*) FROM documents;

# Avec contexte tenant - devrait retourner les lignes du tenant
SELECT set_config('app.tenant_id', '<tenant-uuid>', false);
SELECT COUNT(*) FROM documents;
```

## Bonnes pratiques

1. **Toujours utiliser des mots de passe forts** pour `ACKIFY_APP_PASSWORD`
2. **Ne jamais se connecter en superuser** depuis l'application
3. **Utiliser SSL** pour les connexions base de données en production
4. **Faire tourner les mots de passe** périodiquement
5. **Surveiller** les tentatives d'authentification échouées
