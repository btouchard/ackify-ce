# Document Storage

Upload and store documents directly in Ackify.

## Overview

Ackify supports optional document storage, allowing users to upload files directly instead of providing external URLs. Documents are stored securely and served through authenticated API endpoints.

**Storage options:**
- **Disabled** (default) - Users provide document URLs
- **Local filesystem** - Documents stored on the server
- **S3-compatible** - AWS S3, MinIO, Wasabi, DigitalOcean Spaces, etc.

## Supported File Types

| Type | Extensions | MIME Types |
|------|------------|------------|
| PDF | `.pdf` | `application/pdf` |
| Images | `.png`, `.jpg`, `.jpeg`, `.gif`, `.webp` | `image/*` |
| Office | `.doc`, `.docx` | `application/msword`, `application/vnd.openxmlformats-*` |
| Text | `.txt` | `text/plain` |
| HTML | `.html`, `.htm` | `text/html` |

## Configuration

### Local Storage

Store documents on the server filesystem using a Docker volume.

```env
ACKIFY_STORAGE_TYPE=local
ACKIFY_STORAGE_LOCAL_PATH=/data/documents
ACKIFY_STORAGE_MAX_SIZE_MB=50
```

**Docker Compose volume:**
```yaml
services:
  ackify-ce:
    volumes:
      - ackify_storage:/data/documents

volumes:
  ackify_storage:
```

### S3-Compatible Storage

Works with any S3-compatible storage provider.

```env
ACKIFY_STORAGE_TYPE=s3
ACKIFY_STORAGE_MAX_SIZE_MB=50
ACKIFY_STORAGE_S3_ENDPOINT=https://s3.amazonaws.com
ACKIFY_STORAGE_S3_BUCKET=ackify-documents
ACKIFY_STORAGE_S3_ACCESS_KEY=your_access_key
ACKIFY_STORAGE_S3_SECRET_KEY=your_secret_key
ACKIFY_STORAGE_S3_REGION=us-east-1
ACKIFY_STORAGE_S3_USE_SSL=true
```

### MinIO (Self-hosted S3)

MinIO is a popular open-source S3-compatible storage solution.

```env
ACKIFY_STORAGE_TYPE=s3
ACKIFY_STORAGE_S3_ENDPOINT=http://minio:9000
ACKIFY_STORAGE_S3_BUCKET=ackify-documents
ACKIFY_STORAGE_S3_ACCESS_KEY=minioadmin
ACKIFY_STORAGE_S3_SECRET_KEY=minioadmin
ACKIFY_STORAGE_S3_REGION=us-east-1
ACKIFY_STORAGE_S3_USE_SSL=false
```

**Docker Compose with MinIO:**
```yaml
services:
  minio:
    image: minio/minio:latest
    environment:
      MINIO_ROOT_USER: minioadmin
      MINIO_ROOT_PASSWORD: minioadmin
    volumes:
      - minio_data:/data
    ports:
      - "9000:9000"
      - "9001:9001"
    command: server /data --console-address ":9001"
    healthcheck:
      test: ["CMD", "mc", "ready", "local"]
      interval: 10s
      timeout: 5s
      retries: 5

volumes:
  minio_data:
```

## Usage

### User Interface

When storage is enabled, an upload button appears next to the document URL input:

1. Click the upload button (or drag & drop)
2. Select a file from your computer
3. The file name and size are displayed
4. Click "Upload" to submit

**Features:**
- Progress bar during upload
- Automatic title from filename
- File type validation
- Size limit enforcement

### API Endpoints

#### Upload Document

```http
POST /api/v1/documents/upload
Content-Type: multipart/form-data
X-CSRF-Token: abc123

file: (binary)
title: Optional document title
```

**Response:**
```json
{
  "success": true,
  "data": {
    "doc_id": "abc123",
    "title": "document.pdf",
    "storage_key": "abc123/document.pdf",
    "storage_provider": "local",
    "file_size": 1048576,
    "mime_type": "application/pdf",
    "checksum": "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
    "checksum_algorithm": "SHA-256",
    "created_at": "2025-01-07T12:00:00Z",
    "is_new": true
  }
}
```

#### Get Document Content

```http
GET /api/v1/storage/{docId}/content
```

Returns the document file with appropriate `Content-Type` header.

**Note:** Requires authenticated session.

## Security

### Authentication

- All storage endpoints require authentication
- Documents are only accessible to authenticated users
- CSRF protection on upload endpoint

### Checksum Verification

- SHA-256 checksum calculated automatically on upload
- Stored in document metadata
- Ensures file integrity

### File Validation

- File type checked against allowed MIME types
- File size validated against `ACKIFY_STORAGE_MAX_SIZE_MB`
- Filename sanitized to prevent path traversal

## Database Schema

Uploaded documents add fields to the `documents` table:

```sql
ALTER TABLE documents ADD COLUMN storage_key TEXT;
ALTER TABLE documents ADD COLUMN storage_provider TEXT;
ALTER TABLE documents ADD COLUMN file_size BIGINT;
ALTER TABLE documents ADD COLUMN mime_type TEXT;
```

## Best Practices

### Storage Selection

| Scenario | Recommended Storage |
|----------|---------------------|
| Single server deployment | Local |
| Multiple servers / scaling | S3 |
| Air-gapped environment | Local |
| Cloud-native deployment | S3 |
| Development / testing | Local or MinIO |

### Backup

**Local storage:**
- Include `ackify_storage` volume in backup strategy
- Use `docker-volume-backup` or similar tools

**S3 storage:**
- Configure bucket versioning
- Enable cross-region replication if needed
- Use bucket lifecycle policies for retention

### Performance

- Enable S3 SSL in production (`ACKIFY_STORAGE_S3_USE_SSL=true`)
- Use regional S3 endpoints for lower latency
- Consider CDN for frequently accessed documents

## Limitations

- Maximum file size: configurable, default 50MB
- No virus scanning (implement at infrastructure level)
- No document preview generation
- No automatic compression

## Troubleshooting

### Upload fails with "Storage not configured"

Ensure `ACKIFY_STORAGE_TYPE` is set to `local` or `s3`.

### S3 connection errors

1. Verify endpoint URL format (include `http://` or `https://`)
2. Check access key and secret key
3. Verify bucket exists and is accessible
4. Check SSL setting matches endpoint

### File too large error

Increase `ACKIFY_STORAGE_MAX_SIZE_MB` or reduce file size.

### Permission denied (local storage)

Ensure the container has write access to the storage path:
```bash
docker exec ackify-ce ls -la /data/documents
```
