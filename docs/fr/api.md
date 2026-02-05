# Référence API

Documentation complète de l'API REST Ackify.

## URL de Base

```
https://votre-domaine.com/api/v1
```

## Authentification

La plupart des endpoints requièrent une authentification via cookie de session (OAuth2 ou MagicLink).

**Headers** :
- `X-CSRF-Token` - Requis pour les requêtes POST/PUT/DELETE

Obtenir un token CSRF :
```http
GET /api/v1/csrf
```

## Endpoints

### Santé

#### Health Check

```http
GET /api/v1/health
```

**Réponse** (200 OK) :
```json
{
  "status": "healthy",
  "database": "connected"
}
```

---

### Authentification

#### Démarrer le Flow OAuth2

```http
POST /api/v1/auth/start
```

**Body** :
```json
{
  "redirect": "/?doc=policy_2025"
}
```

#### Demander un MagicLink

```http
POST /api/v1/auth/magic-link/request
```

**Body** :
```json
{
  "email": "user@example.com",
  "redirect": "/?doc=policy_2025"
}
```

#### Vérifier un MagicLink

```http
GET /api/v1/auth/magic-link/verify?token=xxx
```

#### Déconnexion

```http
GET /api/v1/auth/logout
```

---

### Utilisateurs

#### Obtenir l'Utilisateur Courant

```http
GET /api/v1/users/me
```

**Réponse** (200 OK) :
```json
{
  "data": {
    "sub": "google-oauth2|123456",
    "email": "user@example.com",
    "name": "John Doe",
    "isAdmin": false,
    "canCreateDocuments": true
  }
}
```

---

### Documents

#### Trouver ou Créer un Document

```http
GET /api/v1/documents/find-or-create?ref=policy_2025
```

**Réponse** (200 OK) :
```json
{
  "data": {
    "docId": "policy_2025",
    "title": "Politique de Sécurité 2025",
    "url": "https://example.com/policy.pdf",
    "checksum": "sha256:abc123...",
    "checksumAlgorithm": "SHA-256",
    "signatureCount": 42,
    "isNew": false
  }
}
```

**Champs** :
- `signatureCount` - Nombre total de signatures (visible par tous)
- `isNew` - Indique si le document vient d'être créé

#### Obtenir les Détails d'un Document

```http
GET /api/v1/documents/{docId}
```

#### Lister les Signatures d'un Document

```http
GET /api/v1/documents/{docId}/signatures
```

**Contrôle d'Accès** :
| Type d'utilisateur | Résultat |
|-------------------|----------|
| Propriétaire du document ou Admin | Toutes les signatures avec emails |
| Utilisateur authentifié (non propriétaire) | Uniquement sa propre signature (s'il a signé) |
| Non authentifié | Liste vide |

> **Note** : Le **compteur** de signatures est toujours disponible via `signatureCount` dans la réponse du document. Cet endpoint retourne la **liste détaillée** avec les adresses email.

**Réponse** (200 OK) :
```json
{
  "data": [
    {
      "id": 1,
      "docId": "policy_2025",
      "userEmail": "alice@example.com",
      "userName": "Alice Smith",
      "signedAt": "2025-01-15T14:30:00Z",
      "payloadHash": "sha256:e3b0c44...",
      "signature": "ed25519:3045022100..."
    }
  ]
}
```

#### Lister les Signataires Attendus

```http
GET /api/v1/documents/{docId}/expected-signers
```

**Contrôle d'Accès** : Identique à l'endpoint `/signatures` (propriétaire/admin uniquement).

**Réponse** (200 OK) :
```json
{
  "data": [
    {
      "email": "bob@example.com",
      "addedAt": "2025-01-10T10:00:00Z",
      "hasSigned": false
    }
  ]
}
```

---

### Signatures

#### Créer une Signature

```http
POST /api/v1/signatures
X-CSRF-Token: xxx
```

**Body** :
```json
{
  "docId": "policy_2025"
}
```

**Réponse** (201 Created) :
```json
{
  "data": {
    "id": 123,
    "docId": "policy_2025",
    "userEmail": "user@example.com",
    "signedAt": "2025-01-15T14:30:00Z",
    "payloadHash": "sha256:...",
    "signature": "ed25519:..."
  }
}
```

**Erreurs** :
- `409 Conflict` - L'utilisateur a déjà signé ce document

#### Obtenir Mes Signatures

```http
GET /api/v1/signatures
```

Retourne toutes les signatures de l'utilisateur authentifié courant.

#### Obtenir le Statut de Signature

```http
GET /api/v1/documents/{docId}/signatures/status
```

Retourne si l'utilisateur courant a signé le document.

---

### Endpoints Admin

Tous les endpoints admin requièrent que l'utilisateur soit dans `ACKIFY_ADMIN_EMAILS`.

#### Lister Tous les Documents

```http
GET /api/v1/admin/documents
```

#### Obtenir un Document avec Signataires

```http
GET /api/v1/admin/documents/{docId}/signers
```

#### Ajouter un Signataire Attendu

```http
POST /api/v1/admin/documents/{docId}/signers
X-CSRF-Token: xxx
```

**Body** :
```json
{
  "email": "newuser@example.com",
  "notes": "Note optionnelle"
}
```

#### Retirer un Signataire Attendu

```http
DELETE /api/v1/admin/documents/{docId}/signers/{email}
X-CSRF-Token: xxx
```

#### Envoyer des Rappels Email

```http
POST /api/v1/admin/documents/{docId}/reminders
X-CSRF-Token: xxx
```

#### Supprimer un Document

```http
DELETE /api/v1/admin/documents/{docId}
X-CSRF-Token: xxx
```

---

## Réponses d'Erreur

Toutes les erreurs suivent ce format :

```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "Message lisible",
    "details": {}
  }
}
```

**Codes d'Erreur Courants** :
- `UNAUTHORIZED` (401) - Authentification requise
- `FORBIDDEN` (403) - Permissions insuffisantes
- `NOT_FOUND` (404) - Ressource non trouvée
- `CONFLICT` (409) - Ressource existe déjà (ex: signature dupliquée)
- `RATE_LIMITED` (429) - Trop de requêtes
- `VALIDATION_ERROR` (400) - Corps de requête invalide

---

## Rate Limiting

| Catégorie d'Endpoint | Limite |
|---------------------|--------|
| Authentification | 5 requêtes/minute |
| Signatures | 100 requêtes/minute |
| API Générale | 100 requêtes/minute |

---

## Spécification OpenAPI

La spécification OpenAPI 3.0 complète est disponible à :

```
GET /api/v1/openapi.json
```
