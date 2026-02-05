# Cryptographic Signatures

Complete signature flow with Ed25519 and security guarantees.

## Principle

Ackify uses **Ed25519** (elliptic curve) to create non-repudiable cryptographic signatures.

**Guarantees**:
- ✅ **Non-repudiation** - The signature proves the signer's identity
- ✅ **Integrity** - SHA-256 hash detects any modification
- ✅ **Immutable timestamp** - PostgreSQL triggers prevent backdating
- ✅ **Uniqueness** - One signature per user/document

## Signature Flow

### 1. User accesses the document

```
https://sign.company.com/?doc=policy_2025
```

The Vue.js frontend loads and displays:
- Document title (if metadata exists)
- Number of existing signatures
- "Sign this document" button

### 2. Session verification

The frontend calls:
```http
GET /api/v1/users/me
```

**If not logged in** → OAuth2 redirect
**If logged in** → Display signature button

### 3. Signature

When clicking "Sign", the frontend:

1. Gets a CSRF token:
```http
GET /api/v1/csrf
```

2. Sends the signature:
```http
POST /api/v1/signatures
Content-Type: application/json
X-CSRF-Token: abc123

{
  "doc_id": "policy_2025"
}
```

### 4. Backend Processing

The backend (Go):

1. **Verifies the session** - User authenticated
2. **Generates Ed25519 signature**:
   ```go
   payload := fmt.Sprintf("%s:%s:%s:%s", docID, userSub, userEmail, timestamp)
   hash := sha256.Sum256([]byte(payload))
   signature := ed25519.Sign(privateKey, hash[:])
   ```
3. **Calculates prev_hash** - Hash of the last signature (chaining)
4. **Inserts into database**:
   ```sql
   INSERT INTO signatures (doc_id, user_sub, user_email, signed_at, payload_hash, signature, nonce, prev_hash)
   VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
   ```
5. **Returns the signature** to the frontend

### 5. Confirmation

The frontend displays:
- ✅ Signature confirmed
- Timestamp
- Link to signatures list

## Signature Structure

```json
{
  "docId": "policy_2025",
  "userEmail": "alice@company.com",
  "userName": "Alice Smith",
  "signedAt": "2025-01-15T14:30:00Z",
  "payloadHash": "sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
  "signature": "ed25519:3045022100...",
  "nonce": "abc123xyz",
  "prevHash": "sha256:prev..."
}
```

**Fields**:
- `payloadHash` - SHA-256 of the payload (doc_id:user_sub:email:timestamp)
- `signature` - Ed25519 signature in base64
- `nonce` - Anti-replay protection
- `prevHash` - Hash of the previous signature (blockchain-like)

## Signature Verification

### Manual (via API)

```http
GET /api/v1/documents/policy_2025/signatures
```

### Access Control

The signature list endpoint has **access restrictions** to protect user privacy:

| User Type | What They See |
|-----------|---------------|
| **Document owner** (created_by) | All signatures with emails |
| **Admin** (in ACKIFY_ADMIN_EMAILS) | All signatures with emails |
| **Authenticated user** (not owner) | Only their own signature (if they signed) |
| **Non-authenticated** | Empty list |

> **Note**: The **signature count** is always visible to everyone via the `signatureCount` field in document responses. Only the **detailed list** (with email addresses) is restricted.

**Example responses**:

As document owner/admin:
```json
{
  "data": [
    {"userEmail": "alice@example.com", "signedAt": "..."},
    {"userEmail": "bob@example.com", "signedAt": "..."},
    {"userEmail": "charlie@example.com", "signedAt": "..."}
  ]
}
```

As authenticated non-owner (who has signed):
```json
{
  "data": [
    {"userEmail": "bob@example.com", "signedAt": "..."}
  ]
}
```

As non-authenticated:
```json
{
  "data": []
}
```

The same access control applies to the expected signers endpoint (`/expected-signers`).

### Programmatic (Go)

```go
import "crypto/ed25519"

func VerifySignature(publicKey ed25519.PublicKey, payload, signature []byte) bool {
    hash := sha256.Sum256(payload)
    return ed25519.Verify(publicKey, hash[:], signature)
}
```

## PostgreSQL Constraints

### One signature per user/document

```sql
UNIQUE (doc_id, user_sub)
```

**Behavior**:
- If the user tries to sign twice → 409 Conflict error
- The frontend detects this and displays "Already signed"

### Immutability of `created_at`

PostgreSQL trigger:
```sql
CREATE TRIGGER prevent_signatures_created_at_update
    BEFORE UPDATE ON signatures
    FOR EACH ROW
    EXECUTE FUNCTION prevent_created_at_update();
```

**Guarantee**: Impossible to backdate a signature.

## Chaining (Blockchain-like)

Each signature references the previous one via `prev_hash`:

```
Signature 1 → hash1
Signature 2 → hash2 (prev_hash = hash1)
Signature 3 → hash3 (prev_hash = hash2)
```

**Tampering detection**:
- If a signature is modified, the `prev_hash` of the next one no longer matches
- Allows detection of any history modification

## Security

### Ed25519 Private Key

Auto-generated on first startup or via:

```bash
ACKIFY_ED25519_PRIVATE_KEY=$(openssl rand -base64 64)
```

**Important**:
- The private key never leaves the server
- Stored in memory only (not in database)
- Backup required if you want to keep the same key after redeployment

### Anti-Replay Protection

The unique `nonce` prevents signature reuse:
```go
nonce := fmt.Sprintf("%s-%d", userSub, time.Now().UnixNano())
```

### Rate Limiting

Signatures are limited to **100 requests/minute** per IP.

## Use Cases

### Policy Read Validation

```
Document: "Security Policy 2025"
URL: https://sign.company.com/?doc=security_policy_2025
```

**Workflow**:
1. Admin sends the link to employees
2. Each employee clicks, reads, and signs
3. Admin sees completion in `/admin`

### Training Acknowledgment

```
Document: "GDPR Training 2025"
Expected signers: 50 employees
```

**Features**:
- Completion tracking (42/50 = 84%)
- Automatic email reminders
- Signature export

### Contractual Acknowledgment

```
Document: "Terms of Service v3"
Checksum: SHA-256 of the PDF
```

**Verification**:
- User calculates the PDF checksum
- Compares with stored metadata
- Signs if identical

See [Checksums](checksums.md) for more details.

## API Reference

See [API Documentation](../api.md) for all signature-related endpoints.
