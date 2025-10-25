# Checksums

Document integrity verification with tracking.

## Overview

Ackify allows storing and verifying document checksums (fingerprints) to ensure their integrity.

**Supported algorithms**:
- SHA-256 (recommended)
- SHA-512
- MD5 (legacy)

## Calculating a Checksum

### Command Line

```bash
# Linux/Mac - SHA-256
sha256sum document.pdf
# Output: e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855  document.pdf

# SHA-512
sha512sum document.pdf

# MD5
md5sum document.pdf

# Windows PowerShell
Get-FileHash document.pdf -Algorithm SHA256
Get-FileHash document.pdf -Algorithm SHA512
Get-FileHash document.pdf -Algorithm MD5
```

### Client-Side (JavaScript)

The Vue.js frontend uses the **Web Crypto API**:

```javascript
async function calculateChecksum(file) {
  const arrayBuffer = await file.arrayBuffer()
  const hashBuffer = await crypto.subtle.digest('SHA-256', arrayBuffer)
  const hashArray = Array.from(new Uint8Array(hashBuffer))
  return hashArray.map(b => b.toString(16).padStart(2, '0')).join('')
}

// Usage
const file = document.querySelector('input[type="file"]').files[0]
const checksum = await calculateChecksum(file)
console.log('SHA-256:', checksum)
```

## Storing the Checksum

### Via Admin Dashboard

1. Go to `/admin`
2. Select a document
3. Click "Edit Metadata"
4. Fill in:
   - **Checksum**: e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855
   - **Algorithm**: SHA-256
   - **Document URL**: https://docs.company.com/policy.pdf

### Via API

```http
PUT /api/v1/admin/documents/policy_2025/metadata
Content-Type: application/json
X-CSRF-Token: abc123

{
  "title": "Security Policy 2025",
  "url": "https://docs.company.com/policy.pdf",
  "checksum": "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
  "checksumAlgorithm": "SHA-256",
  "description": "Annual security policy"
}
```

## Verification

### User Interface

The frontend displays:
```
Document: Security Policy 2025
Checksum (SHA-256): e3b0c44...52b855 [Copy]
URL: https://docs.company.com/policy.pdf [Open]

[Upload file to verify]
```

**User workflow**:
1. Downloads document from URL
2. Uploads to verification interface
3. Checksum is calculated client-side
4. Automatic comparison with stored value
5. ✅ Match or ❌ Mismatch

### Manual Verification

```bash
# 1. Download the document
wget https://docs.company.com/policy.pdf

# 2. Calculate checksum
sha256sum policy.pdf
# e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855

# 3. Compare with stored value (via API)
curl http://localhost:8080/api/v1/documents/policy_2025
# "checksum": "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"

# 4. If identical → Document is intact
```


## Use Cases

### Document Compliance

```
Document: "ISO 27001 Certification"
Checksum: SHA-256 of official PDF
```

**Workflow**:
- Store checksum of certified document
- Each reviewer verifies integrity before signing
- Audit trail of all verifications

### Legal Contract

```
Document: "Service Agreement v2.3"
Checksum: SHA-512 for maximum security
URL: https://legal.company.com/contracts/sa-v2.3.pdf
```

**Guarantees**:
- Signed document matches exactly the checksum version
- Detection of any modification
- Traceability of verifications

### Training with Materials

```
Document: "GDPR Training Materials"
Checksum: SHA-256 of ZIP file
```

**Usage**:
- Participants download ZIP
- Verify checksum before starting
- Sign after completion

## Security

### Algorithm Choice

| Algorithm | Security | Performance | Recommendation |
|-----------|----------|-------------|----------------|
| SHA-256 | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ✅ Recommended |
| SHA-512 | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ | Maximum security |
| MD5 | ⭐⭐ | ⭐⭐⭐⭐⭐ | ❌ Legacy only |

**Recommendation**: Use **SHA-256** by default.

### MD5 Limitations

MD5 is **deprecated** for security:
- Collisions possible (two different files = same hash)
- Usable only for legacy compatibility

### Web Crypto API

Client-side verification uses browser's native API:
- No external dependency
- Native performance
- Supported by all modern browsers

## Integration with Signatures

Complete workflow:

```
1. Admin uploads document → calculates checksum → stores metadata
2. User downloads document → verifies checksum client-side
3. If checksum OK → User signs document
4. Signature linked to doc_id with stored checksum
```

**Guarantee**: Signature proves user read **exactly** the checksum version.

## Best Practices

### Storage

- ✅ Always store checksum **before** sending signature link
- ✅ Include document URL in metadata
- ✅ Use SHA-256 minimum
- ✅ Document the algorithm used

### Verification

- ✅ Encourage users to verify before signing
- ✅ Display checksum visibly (with Copy button)
- ✅ Alert on mismatch

### Audit

- ✅ Monitor document integrity
- ✅ Review checksums regularly

## Limitations

- **Manual verification only** - Users must manually calculate and compare checksums
- **No server-side verification API** - Checksum verification is performed client-side or manually
- **No automated audit trail** - The `checksum_verifications` table exists in the database schema but is not currently used by the API
- No checksum signing (future feature: sign checksum with Ed25519)
- No cloud storage integration (S3, GCS) for automatic retrieval

## Current Implementation

Currently, Ackify supports:
- ✅ Storing checksums in document metadata (via admin dashboard or API)
- ✅ Displaying checksums to users for manual verification
- ✅ Client-side checksum calculation using Web Crypto API
- ✅ Automatic checksum computation for remote URLs (admin only)

Future features may include:
- API endpoints for checksum verification tracking
- Automated verification workflows
- Integration with external verification services
