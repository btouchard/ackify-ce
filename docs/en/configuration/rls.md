# Row Level Security (RLS)

PostgreSQL Row Level Security provides automatic tenant data isolation at the database level.

## Overview

RLS ensures that each tenant can only access their own data, regardless of how the application queries the database. This is a critical security feature for multi-tenant deployments.

```
┌─────────────────────────────────────────────────────────────────┐
│                    Request Flow                                  │
├─────────────────────────────────────────────────────────────────┤
│ 1. HTTP Request arrives                                          │
│ 2. RLS Middleware starts a transaction                          │
│ 3. Middleware sets: SET app.tenant_id = '<tenant-uuid>'         │
│ 4. All queries automatically filtered by tenant_id              │
│ 5. Transaction committed on success, rolled back on error       │
└─────────────────────────────────────────────────────────────────┘
```

## Configuration

### Required Variable

```bash
# Password for the ackify_app database role
ACKIFY_APP_PASSWORD=your_secure_password
```

### How It Works

1. **During migration** (`migrate up`):
   - The migrate tool reads `ACKIFY_APP_PASSWORD`
   - Creates the `ackify_app` role if it doesn't exist
   - Updates the password if the role already exists
   - Runs migrations that enable RLS policies

2. **At runtime**:
   - Application connects as `ackify_app` (not `postgres`)
   - RLS policies filter all queries by `tenant_id`
   - No data leakage possible

### compose.yml Configuration

```yaml
services:
  ackify-migrate:
    environment:
      # Superuser connection for migrations
      ACKIFY_DB_DSN: "postgres://postgres:${POSTGRES_PASSWORD}@db:5432/ackify?sslmode=disable"
      # Password for ackify_app role creation
      ACKIFY_APP_PASSWORD: "${ACKIFY_APP_PASSWORD}"

  ackify-ce:
    environment:
      # Application connects with ackify_app role (RLS enforced)
      ACKIFY_DB_DSN: "postgres://ackify_app:${ACKIFY_APP_PASSWORD}@db:5432/ackify?sslmode=disable"
```

## Security Benefits

### Automatic Filtering

Without RLS, application code must always include tenant filtering:

```sql
-- Without RLS: Easy to forget tenant_id filter
SELECT * FROM documents WHERE doc_id = '123';  -- BUG: Returns any tenant's data!
```

With RLS, filtering is automatic:

```sql
-- With RLS: Database enforces tenant isolation
SELECT * FROM documents WHERE doc_id = '123';  -- Only returns current tenant's data
```

### Defense in Depth

Even if application code has a bug that forgets tenant filtering, RLS prevents data leakage at the database level.

## Tables with RLS

RLS policies are applied to all tenant-aware tables:

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

## Troubleshooting

### Empty Results When Querying Directly

If you connect to the database with `psql` and get empty results:

```sql
-- This returns 0 rows because app.tenant_id is not set
SELECT COUNT(*) FROM documents;
```

**Solution**: Set the tenant context first:

```sql
-- Option 1: Session-level (persists until disconnect)
SELECT set_config('app.tenant_id', 'your-tenant-uuid', false);

-- Option 2: Transaction-level
BEGIN;
SELECT set_config('app.tenant_id', 'your-tenant-uuid', true);
SELECT * FROM documents;
COMMIT;
```

### Superuser Bypasses RLS

If you connect as `postgres` (superuser), RLS is bypassed:

```sql
-- As postgres: Returns ALL data (no RLS filtering)
SELECT COUNT(*) FROM documents;
```

This is by design. Use `ackify_app` for application connections.

### Migration Fails with "role does not exist"

If migrations fail because `ackify_app` doesn't exist:

1. Ensure `ACKIFY_APP_PASSWORD` is set
2. Check migrate tool logs for warnings
3. Verify the migrate tool runs before migrations

## Manual Role Management

In rare cases, you may need to manage the role manually:

```sql
-- Create role (if not using migrate tool)
CREATE ROLE ackify_app WITH
    LOGIN
    PASSWORD 'your_password'
    NOCREATEDB
    NOCREATEROLE
    NOINHERIT;

-- Grant permissions
GRANT CONNECT ON DATABASE ackify TO ackify_app;
GRANT USAGE ON SCHEMA public TO ackify_app;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO ackify_app;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO ackify_app;

-- Change password
ALTER ROLE ackify_app WITH PASSWORD 'new_password';
```

## Testing RLS

To verify RLS is working correctly:

```bash
# Connect as ackify_app
psql -U ackify_app -d ackify

# Without tenant context - should return 0 rows
SELECT COUNT(*) FROM documents;

# With tenant context - should return tenant's rows
SELECT set_config('app.tenant_id', '<tenant-uuid>', false);
SELECT COUNT(*) FROM documents;
```

## Best Practices

1. **Always use strong passwords** for `ACKIFY_APP_PASSWORD`
2. **Never connect as superuser** from the application
3. **Use SSL** for database connections in production
4. **Rotate passwords** periodically
5. **Monitor** for failed authentication attempts
