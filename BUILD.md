# Ackify Community Edition - Build & Deployment Guide

## Overview

Ackify Community Edition (CE) is the open-source version of Ackify, a document signature validation platform. This guide covers building and deploying the Community Edition.

## Prerequisites

- Go 1.24.5 or later
- Docker and Docker Compose (for containerized deployment)
- PostgreSQL 16+ (for database)

## Building from Source

### 1. Clone the Repository

```bash
git clone https://github.com/btouchard/ackify-ce.git
cd ackify-ce
```

### 2. Build the Application

```bash
# Build Community Edition
go build ./cmd/community

# Or build with specific output name
go build -o ackify-ce ./cmd/community
```

### 3. Run Tests

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./tests/
```

## Configuration

### Environment Variables

Copy the example environment file and modify it:

```bash
cp .env.example .env
```

Required environment variables:

- `APP_BASE_URL`: Public URL of your application
- `OAUTH_CLIENT_ID`: OAuth2 client ID
- `OAUTH_CLIENT_SECRET`: OAuth2 client secret
- `DB_DSN`: PostgreSQL connection string
- `OAUTH_COOKIE_SECRET`: Base64-encoded secret for session cookies

### OAuth2 Providers

Supported providers:
- `google` (default)
- `github`
- `gitlab`
- Custom (specify `OAUTH_AUTH_URL`, `OAUTH_TOKEN_URL`, `OAUTH_USERINFO_URL`)

## Deployment Options

### Option 1: Direct Binary

1. Build the application
2. Set environment variables
3. Run the binary:

```bash
./community
```

### Option 2: Docker Compose (Recommended)

1. Configure environment variables in `.env` file
2. Start services:

```bash
docker compose up -d
```

3. Check logs:

```bash
docker compose logs ackify-ce
```

4. Stop services:

```bash
docker compose down
```

### Option 3: Docker Build

```bash
# Build Docker image
docker build -t ackify-ce:latest .

# Run with environment file
docker run --env-file .env -p 8080:8080 ackify-ce:latest
```

## Database Setup

The application requires PostgreSQL. When using Docker Compose, the database is automatically created and configured.

For manual setup:

1. Create a PostgreSQL database
2. The application will automatically create required tables on first run
3. Set the `DB_DSN` environment variable to your database connection string

## Health Checks

The application provides a health endpoint:

```bash
curl http://localhost:8080/health
```

## Production Considerations

1. **HTTPS**: Always use HTTPS in production (set `APP_BASE_URL` with https://)
2. **Secrets**: Use strong, randomly generated secrets for `OAUTH_COOKIE_SECRET`
3. **Database**: Use a dedicated PostgreSQL instance with proper backups
4. **Monitoring**: Monitor the `/health` endpoint for application status
5. **Logs**: Configure proper log aggregation and monitoring

## API Endpoints

- `GET /` - Homepage
- `GET /health` - Health check
- `GET /sign?doc=<id>` - Document signing interface
- `POST /sign` - Create signature
- `GET /status?doc=<id>` - Get document signature status (JSON)
- `GET /status.png?doc=<id>&user=<email>` - Signature status badge

## Troubleshooting

### Common Issues

1. **Port already in use**: Change `LISTEN_ADDR` in environment variables
2. **Database connection failed**: Check `DB_DSN` and ensure PostgreSQL is running
3. **OAuth2 errors**: Verify `OAUTH_CLIENT_ID` and `OAUTH_CLIENT_SECRET`

### Logs

Enable debug logging by setting `LOG_LEVEL=debug` in your environment.

## Contributing

This is the Community Edition. Contributions are welcome! Please see the main repository for contribution guidelines.

## License

Community Edition is released under the Server Side Public License (SSPL).