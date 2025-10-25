# OAuth2 Providers

Detailed configuration of different OAuth2 providers supported by Ackify.

## Supported Providers

| Provider | Configuration | Auto-detection |
|----------|--------------|----------------|
| Google | `ACKIFY_OAUTH_PROVIDER=google` | ✅ |
| GitHub | `ACKIFY_OAUTH_PROVIDER=github` | ✅ |
| GitLab | `ACKIFY_OAUTH_PROVIDER=gitlab` | ✅ |
| Custom | Leave empty + manual URLs | ❌ |

## Google OAuth2

### Configuration steps

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a project or select an existing one
3. Enable "Google+ API" (to retrieve user profile)
4. Create OAuth 2.0 credentials:
   - **Application type**: Web application
   - **Authorized JavaScript origins**: `https://sign.your-domain.com`
   - **Authorized redirect URIs**: `https://sign.your-domain.com/api/v1/auth/callback`

### `.env` Configuration

```bash
ACKIFY_OAUTH_PROVIDER=google
ACKIFY_OAUTH_CLIENT_ID=123456789-abc.apps.googleusercontent.com
ACKIFY_OAUTH_CLIENT_SECRET=GOCSPX-xyz123abc

# Optional: restrict to @company.com emails
ACKIFY_OAUTH_ALLOWED_DOMAIN=@company.com
```

### Automatic scopes

By default, Ackify requests:
- `openid` - OAuth2 identity
- `email` - Email address
- `profile` - Full name

## GitHub OAuth2

### Configuration steps

1. Go to [GitHub Developer Settings](https://github.com/settings/developers)
2. Click "New OAuth App"
3. Fill the form:
   - **Application name**: Ackify
   - **Homepage URL**: `https://sign.your-domain.com`
   - **Authorization callback URL**: `https://sign.your-domain.com/api/v1/auth/callback`
4. Generate a client secret

### `.env` Configuration

```bash
ACKIFY_OAUTH_PROVIDER=github
ACKIFY_OAUTH_CLIENT_ID=Iv1.abc123xyz
ACKIFY_OAUTH_CLIENT_SECRET=ghp_1234567890abcdef

# Optional: restrict to verified emails from an organization
ACKIFY_OAUTH_ALLOWED_DOMAIN=@company.com
```

### Automatic scopes

By default:
- `read:user` - Profile reading
- `user:email` - Email access

## GitLab OAuth2

### GitLab.com (public)

1. Go to [GitLab Applications](https://gitlab.com/-/profile/applications)
2. Create a new application:
   - **Name**: Ackify
   - **Redirect URI**: `https://sign.your-domain.com/api/v1/auth/callback`
   - **Scopes**: `openid`, `email`, `profile`
3. Copy the Application ID and Secret

```bash
ACKIFY_OAUTH_PROVIDER=gitlab
ACKIFY_OAUTH_CLIENT_ID=abc123xyz
ACKIFY_OAUTH_CLIENT_SECRET=glpat-xyz123
```

### GitLab Self-Hosted

For a private GitLab instance:

```bash
ACKIFY_OAUTH_PROVIDER=gitlab
ACKIFY_OAUTH_GITLAB_URL=https://gitlab.company.com
ACKIFY_OAUTH_CLIENT_ID=abc123xyz
ACKIFY_OAUTH_CLIENT_SECRET=glpat-xyz123
```

**Important**: `ACKIFY_OAUTH_GITLAB_URL` must point to your GitLab instance without trailing slash.

## Custom OAuth2 Provider

To use a non-standard OAuth2 provider (Keycloak, Okta, Auth0, etc.).

### Complete configuration

```bash
# Do not define ACKIFY_OAUTH_PROVIDER (or leave empty)
ACKIFY_OAUTH_PROVIDER=

# Manual URLs
ACKIFY_OAUTH_AUTH_URL=https://auth.company.com/oauth/authorize
ACKIFY_OAUTH_TOKEN_URL=https://auth.company.com/oauth/token
ACKIFY_OAUTH_USERINFO_URL=https://auth.company.com/api/user

# Custom scopes (optional)
ACKIFY_OAUTH_SCOPES=openid,email,profile

# Custom logout URL (optional)
ACKIFY_OAUTH_LOGOUT_URL=https://auth.company.com/logout

# Credentials
ACKIFY_OAUTH_CLIENT_ID=your_client_id
ACKIFY_OAUTH_CLIENT_SECRET=your_client_secret
```

### Example with Keycloak

```bash
ACKIFY_OAUTH_PROVIDER=
ACKIFY_OAUTH_AUTH_URL=https://keycloak.company.com/realms/myrealm/protocol/openid-connect/auth
ACKIFY_OAUTH_TOKEN_URL=https://keycloak.company.com/realms/myrealm/protocol/openid-connect/token
ACKIFY_OAUTH_USERINFO_URL=https://keycloak.company.com/realms/myrealm/protocol/openid-connect/userinfo
ACKIFY_OAUTH_LOGOUT_URL=https://keycloak.company.com/realms/myrealm/protocol/openid-connect/logout
ACKIFY_OAUTH_SCOPES=openid,email,profile
ACKIFY_OAUTH_CLIENT_ID=ackify-client
ACKIFY_OAUTH_CLIENT_SECRET=secret123
```

### Example with Okta

```bash
ACKIFY_OAUTH_PROVIDER=
ACKIFY_OAUTH_AUTH_URL=https://dev-123456.okta.com/oauth2/default/v1/authorize
ACKIFY_OAUTH_TOKEN_URL=https://dev-123456.okta.com/oauth2/default/v1/token
ACKIFY_OAUTH_USERINFO_URL=https://dev-123456.okta.com/oauth2/default/v1/userinfo
ACKIFY_OAUTH_SCOPES=openid,email,profile
ACKIFY_OAUTH_CLIENT_ID=0oa123xyz
ACKIFY_OAUTH_CLIENT_SECRET=secret123
```

## Domain Restriction

For **all providers**, you can restrict access to emails from a specific domain:

```bash
# Accept only @company.com emails
ACKIFY_OAUTH_ALLOWED_DOMAIN=@company.com
```

**Behavior**:
- Users with a different email will see an error when logging in
- Verification is case-insensitive
- Works with all providers (Google, GitHub, GitLab, custom)

## Auto-Login

Enable silent auto-login for better UX:

```bash
ACKIFY_OAUTH_AUTO_LOGIN=true
```

**How it works**:
- If the user already has an active OAuth session, automatic redirect
- No click required on "Sign in"
- Useful for corporate integrations (Google Workspace, Microsoft 365)

**Warning**: Can create infinite redirects if misconfigured.

## OAuth2 Security

### PKCE (Proof Key for Code Exchange)

Ackify **automatically** implements PKCE for all providers:
- Protection against authorization code interception
- Method: S256 (SHA-256)
- Enabled by default, no configuration required

### Refresh Tokens

Refresh tokens are:
- Stored **encrypted** in PostgreSQL (AES-256-GCM)
- Used to maintain sessions for 30 days
- Automatically cleaned up after expiration (37 days)
- Protected by IP + User-Agent tracking

### Secure Sessions

```bash
# Strong secret required (minimum 32 bytes in base64)
ACKIFY_OAUTH_COOKIE_SECRET=$(openssl rand -base64 32)
```

Session cookies use:
- HMAC-SHA256 for integrity
- AES-256-GCM encryption
- `Secure` flags (HTTPS only) and `HttpOnly`
- `SameSite=Lax` for CSRF protection

## Troubleshooting

### Error "invalid redirect_uri"

Verify that:
- `ACKIFY_BASE_URL` exactly matches your domain
- The callback URL in the provider includes `/api/v1/auth/callback`
- No trailing slash in `ACKIFY_BASE_URL`

### Error "unauthorized_client"

Verify:
- The `client_id` and `client_secret` are correct
- The OAuth application is properly enabled on the provider side
- The requested scopes are authorized

### Error "access_denied"

The user refused authorization, or:
- Their email doesn't match `ACKIFY_OAUTH_ALLOWED_DOMAIN`
- The application doesn't have the required permissions

### Custom provider doesn't work

Verify:
- `ACKIFY_OAUTH_PROVIDER` is **empty** or undefined
- The 3 URLs (auth, token, userinfo) are complete and correct
- The response from `/userinfo` contains `sub`, `email`, `name`

## Testing the Configuration

```bash
# Restart after changes
docker compose restart ackify-ce

# Test OAuth connection
curl -X POST http://localhost:8080/api/v1/auth/start \
  -H "Content-Type: application/json" \
  -d '{"redirect_to": "/"}'

# Should return a redirect_url to the OAuth provider
```
