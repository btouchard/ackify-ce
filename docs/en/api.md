# API Reference

Complete REST API documentation for Ackify.

## Base URL

```
https://your-domain.com/api/v1
```

## Authentication

Most endpoints require authentication via session cookie (OAuth2 or MagicLink).

**Headers**:
- `X-CSRF-Token` - Required for POST/PUT/DELETE requests

Get CSRF token:
```http
GET /api/v1/csrf
```

## Endpoints

### Health

#### Health Check

```http
GET /api/v1/health
```

**Response** (200 OK):
```json
{
  "status": "healthy",
  "database": "connected"
}
```

---

### Authentication

#### Start OAuth2 Flow

```http
POST /api/v1/auth/start
```

**Body**:
```json
{
  "redirect": "/?doc=policy_2025"
}
```

#### Request MagicLink

```http
POST /api/v1/auth/magic-link/request
```

**Body**:
```json
{
  "email": "user@example.com",
  "redirect": "/?doc=policy_2025"
}
```

#### Verify MagicLink

```http
GET /api/v1/auth/magic-link/verify?token=xxx
```

#### Logout

```http
GET /api/v1/auth/logout
```

---

### Users

#### Get Current User

```http
GET /api/v1/users/me
```

**Response** (200 OK):
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

#### Find or Create Document

```http
GET /api/v1/documents/find-or-create?doc=policy_2025
```

**Response** (200 OK):
```json
{
  "data": {
    "docId": "policy_2025",
    "title": "Security Policy 2025",
    "url": "https://example.com/policy.pdf",
    "checksum": "sha256:abc123...",
    "checksumAlgorithm": "SHA-256",
    "signatureCount": 42,
    "isNew": false
  }
}
```

**Fields**:
- `signatureCount` - Total number of signatures (visible to all users)
- `isNew` - Whether the document was just created

#### Get Document Details

```http
GET /api/v1/documents/{docId}
```

#### List Document Signatures

```http
GET /api/v1/documents/{docId}/signatures
```

**Access Control**:
| User Type | Result |
|-----------|--------|
| Document owner or Admin | All signatures with emails |
| Authenticated user (not owner) | Only their own signature (if signed) |
| Non-authenticated | Empty list |

> **Note**: The signature **count** is always available via `signatureCount` in the document response. This endpoint returns the **detailed list** with email addresses.

**Response** (200 OK):
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

#### List Expected Signers

```http
GET /api/v1/documents/{docId}/expected-signers
```

**Access Control**: Same as `/signatures` endpoint (owner/admin only).

**Response** (200 OK):
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

#### Create Signature

```http
POST /api/v1/signatures
X-CSRF-Token: xxx
```

**Body**:
```json
{
  "docId": "policy_2025"
}
```

**Response** (201 Created):
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

**Errors**:
- `409 Conflict` - User has already signed this document

#### Get My Signatures

```http
GET /api/v1/signatures
```

Returns all signatures for the current authenticated user.

#### Get Signature Status

```http
GET /api/v1/documents/{docId}/signatures/status
```

Returns whether the current user has signed the document.

---

### Admin Endpoints

All admin endpoints require the user to be in `ACKIFY_ADMIN_EMAILS`.

#### List All Documents

```http
GET /api/v1/admin/documents
```

#### Get Document with Signers

```http
GET /api/v1/admin/documents/{docId}/signers
```

#### Add Expected Signer

```http
POST /api/v1/admin/documents/{docId}/signers
X-CSRF-Token: xxx
```

**Body**:
```json
{
  "email": "newuser@example.com",
  "notes": "Optional note"
}
```

#### Remove Expected Signer

```http
DELETE /api/v1/admin/documents/{docId}/signers/{email}
X-CSRF-Token: xxx
```

#### Send Email Reminders

```http
POST /api/v1/admin/documents/{docId}/reminders
X-CSRF-Token: xxx
```

#### Delete Document

```http
DELETE /api/v1/admin/documents/{docId}
X-CSRF-Token: xxx
```

---

## Error Responses

All errors follow this format:

```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "Human readable message",
    "details": {}
  }
}
```

**Common Error Codes**:
- `UNAUTHORIZED` (401) - Authentication required
- `FORBIDDEN` (403) - Insufficient permissions
- `NOT_FOUND` (404) - Resource not found
- `CONFLICT` (409) - Resource already exists (e.g., duplicate signature)
- `RATE_LIMITED` (429) - Too many requests
- `VALIDATION_ERROR` (400) - Invalid request body

---

## Rate Limiting

| Endpoint Category | Limit |
|-------------------|-------|
| Authentication | 5 requests/minute |
| Signatures | 100 requests/minute |
| General API | 100 requests/minute |

---

## OpenAPI Specification

The complete OpenAPI 3.0 specification is available at:

```
GET /api/v1/openapi.json
```
