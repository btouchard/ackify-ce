# OAuth2 Providers

Configuration détaillée des différents providers OAuth2 supportés par Ackify.

## Providers Supportés

| Provider | Configuration | Auto-détection |
|----------|--------------|----------------|
| Google | `ACKIFY_OAUTH_PROVIDER=google` | ✅ |
| GitHub | `ACKIFY_OAUTH_PROVIDER=github` | ✅ |
| GitLab | `ACKIFY_OAUTH_PROVIDER=gitlab` | ✅ |
| Custom | Laisser vide + URLs manuelles | ❌ |

## Google OAuth2

### Étapes de configuration

1. Aller sur [Google Cloud Console](https://console.cloud.google.com/)
2. Créer un projet ou sélectionner un existant
3. Activer "Google+ API" (pour récupérer le profil utilisateur)
4. Créer des credentials OAuth 2.0 :
   - **Application type** : Web application
   - **Authorized JavaScript origins** : `https://sign.your-domain.com`
   - **Authorized redirect URIs** : `https://sign.your-domain.com/api/v1/auth/callback`

### Configuration `.env`

```bash
ACKIFY_OAUTH_PROVIDER=google
ACKIFY_OAUTH_CLIENT_ID=123456789-abc.apps.googleusercontent.com
ACKIFY_OAUTH_CLIENT_SECRET=GOCSPX-xyz123abc

# Optionnel : restreindre aux emails @company.com
ACKIFY_OAUTH_ALLOWED_DOMAIN=@company.com
```

### Scopes automatiques

Par défaut, Ackify demande :
- `openid` - Identité OAuth2
- `email` - Adresse email
- `profile` - Nom complet

## GitHub OAuth2

### Étapes de configuration

1. Aller sur [GitHub Developer Settings](https://github.com/settings/developers)
2. Cliquer sur "New OAuth App"
3. Remplir le formulaire :
   - **Application name** : Ackify
   - **Homepage URL** : `https://sign.your-domain.com`
   - **Authorization callback URL** : `https://sign.your-domain.com/api/v1/auth/callback`
4. Générer un client secret

### Configuration `.env`

```bash
ACKIFY_OAUTH_PROVIDER=github
ACKIFY_OAUTH_CLIENT_ID=Iv1.abc123xyz
ACKIFY_OAUTH_CLIENT_SECRET=ghp_1234567890abcdef

# Optionnel : restreindre aux emails vérifiés d'une organisation
ACKIFY_OAUTH_ALLOWED_DOMAIN=@company.com
```

### Scopes automatiques

Par défaut :
- `read:user` - Lecture du profil
- `user:email` - Accès aux emails

## GitLab OAuth2

### GitLab.com (public)

1. Aller sur [GitLab Applications](https://gitlab.com/-/profile/applications)
2. Créer une nouvelle application :
   - **Name** : Ackify
   - **Redirect URI** : `https://sign.your-domain.com/api/v1/auth/callback`
   - **Scopes** : `openid`, `email`, `profile`
3. Copier l'Application ID et le Secret

```bash
ACKIFY_OAUTH_PROVIDER=gitlab
ACKIFY_OAUTH_CLIENT_ID=abc123xyz
ACKIFY_OAUTH_CLIENT_SECRET=glpat-xyz123
```

### GitLab Self-Hosted

Pour une instance GitLab privée :

```bash
ACKIFY_OAUTH_PROVIDER=gitlab
ACKIFY_OAUTH_GITLAB_URL=https://gitlab.company.com
ACKIFY_OAUTH_CLIENT_ID=abc123xyz
ACKIFY_OAUTH_CLIENT_SECRET=glpat-xyz123
```

**Important** : `ACKIFY_OAUTH_GITLAB_URL` doit pointer vers votre instance GitLab sans trailing slash.

## Custom OAuth2 Provider

Pour utiliser un provider OAuth2 non standard (Keycloak, Okta, Auth0, etc.).

### Configuration complète

```bash
# Ne pas définir ACKIFY_OAUTH_PROVIDER (ou laisser vide)
ACKIFY_OAUTH_PROVIDER=

# URLs manuelles
ACKIFY_OAUTH_AUTH_URL=https://auth.company.com/oauth/authorize
ACKIFY_OAUTH_TOKEN_URL=https://auth.company.com/oauth/token
ACKIFY_OAUTH_USERINFO_URL=https://auth.company.com/api/user

# Scopes personnalisés (optionnel)
ACKIFY_OAUTH_SCOPES=openid,email,profile

# URL de logout personnalisée (optionnel)
ACKIFY_OAUTH_LOGOUT_URL=https://auth.company.com/logout

# Credentials
ACKIFY_OAUTH_CLIENT_ID=your_client_id
ACKIFY_OAUTH_CLIENT_SECRET=your_client_secret
```

### Exemple avec Keycloak

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

### Exemple avec Okta

```bash
ACKIFY_OAUTH_PROVIDER=
ACKIFY_OAUTH_AUTH_URL=https://dev-123456.okta.com/oauth2/default/v1/authorize
ACKIFY_OAUTH_TOKEN_URL=https://dev-123456.okta.com/oauth2/default/v1/token
ACKIFY_OAUTH_USERINFO_URL=https://dev-123456.okta.com/oauth2/default/v1/userinfo
ACKIFY_OAUTH_SCOPES=openid,email,profile
ACKIFY_OAUTH_CLIENT_ID=0oa123xyz
ACKIFY_OAUTH_CLIENT_SECRET=secret123
```

## Restriction de Domaine

Pour **tous les providers**, vous pouvez restreindre l'accès aux emails d'un domaine spécifique :

```bash
# Accepter uniquement les emails @company.com
ACKIFY_OAUTH_ALLOWED_DOMAIN=@company.com
```

**Comportement** :
- Les utilisateurs avec un email différent verront une erreur lors de la connexion
- La vérification est case-insensitive
- Fonctionne avec tous les providers (Google, GitHub, GitLab, custom)

## Auto-Login

Activer l'auto-login silencieux pour une meilleure UX :

```bash
ACKIFY_OAUTH_AUTO_LOGIN=true
```

**Fonctionnement** :
- Si l'utilisateur a déjà une session OAuth active, redirection automatique
- Pas de clic requis sur "Sign in"
- Utile pour les intégrations corporate (Google Workspace, Microsoft 365)

**Attention** : Peut créer des redirections infinies si mal configuré.

## Sécurité OAuth2

### PKCE (Proof Key for Code Exchange)

Ackify implémente **automatiquement** PKCE pour tous les providers :
- Protection contre l'interception du code d'autorisation
- Méthode : S256 (SHA-256)
- Activé par défaut, aucune configuration requise

### Refresh Tokens

Les refresh tokens sont :
- Stockés **chiffrés** dans PostgreSQL (AES-256-GCM)
- Utilisés pour maintenir les sessions 30 jours
- Automatiquement nettoyés après expiration (37 jours)
- Protégés par IP + User-Agent tracking

### Sessions Sécurisées

```bash
# Secret fort requis (minimum 32 bytes en base64)
ACKIFY_OAUTH_COOKIE_SECRET=$(openssl rand -base64 32)
```

Les cookies de session utilisent :
- HMAC-SHA256 pour l'intégrité
- Chiffrement AES-256-GCM
- Flags `Secure` (HTTPS uniquement) et `HttpOnly`
- `SameSite=Lax` pour protection CSRF

## Troubleshooting

### Erreur "invalid redirect_uri"

Vérifier que :
- `ACKIFY_BASE_URL` correspond exactement à votre domaine
- La callback URL dans le provider inclut `/api/v1/auth/callback`
- Pas de trailing slash dans `ACKIFY_BASE_URL`

### Erreur "unauthorized_client"

Vérifier :
- Le `client_id` et `client_secret` sont corrects
- L'application OAuth est bien activée côté provider
- Les scopes demandés sont autorisés

### Erreur "access_denied"

L'utilisateur a refusé l'autorisation, ou :
- Son email ne correspond pas à `ACKIFY_OAUTH_ALLOWED_DOMAIN`
- L'application n'a pas les permissions requises

### Custom provider ne fonctionne pas

Vérifier :
- `ACKIFY_OAUTH_PROVIDER` est **vide** ou non défini
- Les 3 URLs (auth, token, userinfo) sont complètes et correctes
- La réponse de `/userinfo` contient bien `sub`, `email`, `name`

## Tester la Configuration

```bash
# Redémarrer après modification
docker compose restart ackify-ce

# Tester la connexion OAuth
curl -X POST http://localhost:8080/api/v1/auth/start \
  -H "Content-Type: application/json" \
  -d '{"redirect_to": "/"}'

# Devrait retourner une redirect_url vers le provider OAuth
```
