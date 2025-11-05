# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.2.1] - 2025-11-05

### 🔐 Passwordless Authentication & Enhanced Installation

Minor release adding Magic Link authentication, improved metadata extraction, and professional installation tooling.

### Added

- **Magic Link Authentication (Passwordless)**
  - Email-based passwordless authentication system
  - No password required - users receive a secure link via email
  - Multi-method support: configure OAuth and/or MagicLink independently
  - Intelligent authentication method selection page
  - Auto-redirect to login when only one method is configured
  - Secure token generation with crypto/rand (32 bytes)
  - 15-minute expiration with automatic cleanup
  - HTML and text email templates for magic links
  - New migration `0012_magic_link_authentication` with `magic_links` table
  - Indexes on token, email, and expires_at for optimal performance
  - Background worker for cleaning expired magic links

- **Enhanced Installation Experience**
  - Interactive installation script with step-by-step guidance
  - Automatic environment detection (Docker, PostgreSQL, etc.)
  - System prerequisites validation
  - Assisted configuration of environment variables
  - Support for multi-authentication method setup
  - Complete installation documentation in `install/README.md`
  - Comprehensive `.env.example` with detailed comments
  - Docker Compose templates for quick deployment

- **E2E Testing with Cypress**
  - Complete end-to-end test suite for Magic Link authentication
  - MailHog integration for email testing in development
  - GitHub Actions workflow for automated E2E tests
  - Dedicated `compose.e2e.yml` for isolated test environment
  - Test utilities for email verification and link extraction

- **Smart Document Title Extraction**
  - Enhanced automatic title detection from HTML metadata
  - Support for Open Graph tags (`og:title`)
  - Support for Twitter Card tags (`twitter:title`)
  - Intelligent fallback hierarchy: OG → Twitter → title → h1
  - Comprehensive unit tests (233 test cases)
  - Better handling of edge cases and malformed HTML

### Changed

- **Architecture Improvements**
  - Refactored OAuth code into reusable `OAuthProvider` component
  - New `SessionService` for centralized session management
  - New `MagicLinkService` for passwordless authentication logic
  - Better separation of concerns between authentication methods
  - Cleaner dependency injection in main.go

- **Configuration System**
  - Auto-detection of available authentication methods
  - New `ACKIFY_AUTH_OAUTH_ENABLED` flag (optional, auto-detected)
  - New `ACKIFY_AUTH_MAGICLINK_ENABLED` flag (optional, auto-detected)
  - MagicLink enabled automatically if `ACKIFY_MAIL_HOST` is configured
  - OAuth enabled automatically if OAuth credentials are present
  - Enhanced email configuration with detailed SMTP options
  - Better validation and error messages for configuration issues

- **Session Management**
  - 30-day persistent sessions (increased from 7 days)
  - Encrypted refresh token storage with AES-256-GCM
  - New `oauth_sessions` table for refresh token persistence
  - Automatic cleanup of expired sessions (background worker)
  - Session tracking with IP address and User-Agent

- **User Interface**
  - New authentication choice page when multiple methods available
  - Auto-redirect behavior when single authentication method
  - Window variables for dynamic config (`ACKIFY_OAUTH_ENABLED`, `ACKIFY_MAGICLINK_ENABLED`)
  - Updated localization files (en, fr, es, de, it) with Magic Link strings

### Fixed

- Improved robustness of document metadata extraction
- Better error handling in authentication flows
- More descriptive error messages for configuration issues
- Edge case handling in title extraction

### Technical Details

**New Files:**
- `backend/internal/application/services/magic_link_service.go` - MagicLink service
- `backend/internal/domain/models/magic_link.go` - MagicLink domain model
- `backend/internal/infrastructure/auth/oauth_provider.go` - OAuth provider refactored
- `backend/internal/infrastructure/auth/session_service.go` - Session management
- `backend/internal/infrastructure/auth/session_worker_test.go` - Session cleanup tests
- `backend/internal/infrastructure/database/magic_link_repository.go` - MagicLink repository
- `backend/internal/infrastructure/workers/magic_link_cleanup.go` - Cleanup worker
- `backend/internal/presentation/api/auth/magic_link_handler.go` - MagicLink endpoints
- `backend/templates/magic_link.html.tmpl` - HTML email template
- `backend/templates/magic_link.txt.tmpl` - Text email template
- `backend/migrations/0012_magic_link_authentication.{up,down}.sql`
- `webapp/src/pages/AuthChoicePage.vue` - Authentication method selection
- `webapp/cypress/` - Complete E2E test suite
- `.github/workflows/e2e-tests.yml` - E2E CI workflow
- `install/README.md` - Installation documentation

**Modified Files:**
- `backend/internal/infrastructure/config/config.go` - Enhanced configuration
- `backend/internal/infrastructure/auth/oauth.go` - Refactored to use OAuthProvider
- `backend/internal/presentation/api/router.go` - New Magic Link endpoints
- `backend/pkg/web/server.go` - Multi-auth method support
- `backend/pkg/web/static.go` - New window variables injection
- `webapp/src/router/index.ts` - Auth choice route
- `README.md`, `README_FR.md` - Updated with Magic Link documentation
- `.env.example` - Comprehensive email and auth configuration

### Migration Guide

**For users upgrading from v1.2.0 to v1.2.1:**

1. **No Breaking Changes**: v1.2.1 is 100% backward compatible
2. **Optional MagicLink**: Add email configuration to enable passwordless auth
3. **Migrations**: Applied automatically at startup
4. **Environment Variables**: Review new optional variables in `.env.example`

**To enable Magic Link authentication:**
```bash
# Add SMTP configuration
ACKIFY_MAIL_HOST="smtp.example.com"
ACKIFY_MAIL_PORT=587
ACKIFY_MAIL_USERNAME="user"
ACKIFY_MAIL_PASSWORD="pass"
ACKIFY_MAIL_FROM="noreply@example.com"

# Optional: explicitly control auth methods
ACKIFY_AUTH_OAUTH_ENABLED=true
ACKIFY_AUTH_MAGICLINK_ENABLED=true
```

## [1.2.0] - 2025-10-27

### 🎉 Major Release: API-First Vue Migration with Enhanced Security

Complete architectural overhaul to a modern API-first architecture with Vue 3 SPA frontend, signed webhooks, and improved security.

### Added

- **RESTful API v1**
  - Versioned API with `/api/v1` prefix
  - Structured JSON responses with consistent error handling
  - Public endpoints: health, documents, signatures, expected signers
  - Authentication endpoints: OAuth flow, logout, auth check
  - Authenticated endpoints: user profile, signatures, signature creation
  - Admin endpoints: document management, signer management, reminders
  - OpenAPI specification endpoint `/api/v1/openapi.json`

- **Vue 3 SPA Frontend**
  - Modern single-page application with TypeScript
  - Vite build tool with hot module replacement (HMR)
  - Pinia state management for centralized application state
  - Vue Router for client-side routing
  - Tailwind CSS for utility-first styling
  - Responsive design with mobile support
  - Version number display in footer for better traceability
  - Enhanced footer visibility with improved UX
  - Pages: Home, Sign, Signatures, Embed, Admin Dashboard, Document Details

- **Signed Webhooks Support**
  - Webhook signature verification for secure event notifications
  - HMAC-based authentication for webhook endpoints
  - Prevents unauthorized webhook injection
  - Configurable webhook endpoints for document events

- **Comprehensive Logging System**
  - Structured JSON logging with `slog` package
  - Log levels: debug, info, warn, error (configurable via `ACKIFY_LOG_LEVEL`)
  - Request ID tracking through entire request lifecycle
  - HTTP request/response logging with timing
  - Authentication flow logging
  - Signature operation logging
  - Reminder service logging
  - Database query logging
  - OAuth flow progression logging

- **Enhanced Security**
  - OAuth 2.0 Authorization Code Flow with PKCE (Proof Key for Code Exchange)
  - CSRF token protection for all state-changing operations
  - Rate limiting (5 auth attempts/min, 100 general requests/min)
  - Hard rate limiting on embed document creation endpoint
  - CORS configuration for development and production
  - Security headers (CSP, X-Content-Type-Options, X-Frame-Options, etc.)
  - Session-based authentication with secure cookies
  - Request ID propagation for distributed tracing
  - Authorization middleware for embed endpoints

- **Public Embed Route**
  - `/embed?doc={docId}` route for public embedding (no authentication required)
  - Protected document creation with rate limiting and authorization
  - oEmbed protocol support for unfurl functionality
  - CSP headers configured to allow iframe embedding on embed routes
  - Suitable for integration in documentation tools and wikis

- **Auto-Login Feature**
  - Optional `ACKIFY_OAUTH_AUTO_LOGIN` configuration
  - Silent authentication when OAuth session exists
  - `/api/v1/auth/check` endpoint for session verification
  - Seamless user experience when returning to application

- **Docker Multi-Stage Build**
  - Optimized Dockerfile with separate Node and Go build stages
  - Improved build stage efficiency
  - Smaller final image size
  - SPA assets built during Docker build process
  - Production-ready containerized deployment

### Changed

- **Architecture**
  - Migrated from template-based rendering to API-first architecture
  - Introduced clear separation between API and frontend
  - Organized API handlers into logical modules (admin, auth, documents, signatures, users)
  - Centralized middleware in `shared` package (logging, CORS, CSRF, rate limiting, security headers)

- **Routing**
  - Chi router now serves both API v1 and Vue SPA
  - SPA fallback routing for all unmatched routes
  - API endpoints prefixed with `/api/v1`
  - Static assets served from `/assets` for SPA and `/static` for legacy

- **Authentication**
  - Standardized session-based auth across API and templates
  - CSRF protection on all authenticated API endpoints
  - Rate limiting on authentication endpoints

- **Documentation**
  - Updated BUILD.md with Vue SPA build instructions
  - Updated README.md with API v1 endpoint documentation
  - Updated README_FR.md with French translations
  - Added logging configuration documentation
  - Added development environment setup instructions

### Fixed

- Consistent error handling across all API endpoints
- Proper HTTP status codes for all responses
- CORS issues in development environment
- Integration tests concurrency issues and database collisions
- Random hex generation for test database names to prevent collisions
- Migrations directory discovery in CI environment
- Missing hardcoded database struct columns removed
- Split unit and integration test coverage for better reliability
- CI/CD pipeline now pushes releases to latest tag on DockerHub

### Technical Details

**New Files:**
- `internal/presentation/api/` - Complete API v1 implementation
  - `admin/handler.go` - Admin endpoints
  - `auth/handler.go` - Authentication endpoints
  - `documents/handler.go` - Document endpoints
  - `signatures/handler.go` - Signature endpoints
  - `users/handler.go` - User endpoints
  - `health/handler.go` - Health check endpoint
  - `shared/` - Shared middleware and utilities
    - `logging.go` - Request logging middleware
    - `middleware.go` - Auth, admin, CSRF, rate limiting middleware
    - `response.go` - Standardized JSON response helpers
    - `errors.go` - Error code constants
  - `router.go` - API v1 router configuration
- `webapp/` - Complete Vue 3 SPA
  - `src/components/` - Reusable Vue components
  - `src/pages/` - Page components (Home, Sign, Signatures, Embed, Admin)
  - `src/services/` - API client services
  - `src/stores/` - Pinia state stores
  - `src/router/` - Vue Router configuration
  - `vite.config.ts` - Vite build configuration
  - `tsconfig.json` - TypeScript configuration

**Modified Files:**
- `pkg/web/server.go` - Updated to serve both API and SPA
- `internal/infrastructure/auth/oauth.go` - Added structured logging
- `internal/application/services/signature.go` - Added structured logging
- `internal/application/services/reminder.go` - Added structured logging
- `Dockerfile` - Multi-stage build for Node and Go
- `docker-compose.yml` - Updated for new architecture

**Deprecated:**
- Template-based admin routes (will be maintained for backward compatibility)
- Legacy `/status` and `/status.png` endpoints (superseded by API v1)

### Migration Guide

For users upgrading from v1.1.x to v1.2.0:

1. **Environment Variables**: Add optional `ACKIFY_LOG_LEVEL` and `ACKIFY_OAUTH_AUTO_LOGIN` if desired
2. **Docker**: Rebuild images to include Vue SPA build with multi-stage optimization
3. **API Clients**: Consider migrating to new API v1 endpoints for better structure and consistency
4. **Embed URLs**: Update to use `/embed?doc={docId}` for public document embedding
5. **Webhooks**: Configure webhook endpoints if you want to receive signed event notifications

### Breaking Changes

- None - v1.2.0 maintains backward compatibility with all v1.1.x features
- Template-based admin interface remains functional alongside new Vue SPA
- Legacy endpoints continue to work

## [1.1.3] - 2025-10-08

### Added
- **Document Metadata Management System**
  - New `documents` table for storing metadata (title, URL, checksum, description)
  - Document repository with full CRUD operations
  - Comprehensive integration tests for document operations
  - Admin UI section for viewing and editing document metadata
  - Copy-to-clipboard functionality for checksums
  - Support for SHA-256, SHA-512, and MD5 checksum algorithms
  - Automatic `updated_at` timestamp tracking with PostgreSQL trigger

- **Modern Modal Dialogs**
  - Replaced native JavaScript `alert()` and `confirm()` with styled modal dialogs
  - Consistent design across all confirmation actions
  - Better UX with warning (orange) and delete (red) visual indicators
  - Confirmation modal for email reminder sending
  - Delete confirmation modal for removing expected readers

- **SVG Favicon**
  - Added modern vector favicon with brand identity
  - Responsive and works across all modern browsers

### Changed
- **Email Reminder Improvements**
  - Email language now matches user's interface language (fr/en)
  - Document URL automatically fetched from metadata instead of manual input
  - Simplified reminder form by removing redundant URL field
  - Document URL displayed as clickable link in reminder section

- **Admin Dashboard Enhancement**
  - Document listing now includes documents from `documents` table
  - Shows documents with metadata even without signatures or expected readers

- **UI Refinements**
  - Removed "Admin connecté" status indicator from dashboard header
  - Document URL in metadata displayed as hyperlink instead of input field
  - Cleaner and more focused admin interface

### Fixed
- Template syntax error with `not` operator requiring parentheses

### Technical Details
- Added database migration `0005_create_documents_table`
- New domain model: `models.Document` and `models.DocumentInput`
- New infrastructure: `DocumentRepository` with full test coverage
- New presentation: `DocumentHandlers` with GET/POST/DELETE endpoints
- Routes: `/admin/docs/{docID}/metadata` (GET, POST, DELETE)
- Updated `ReminderService.SendReminders()` signature to include locale parameter
- Modified files:
  - `internal/domain/models/document.go` (new)
  - `internal/infrastructure/database/document_repository.go` (new)
  - `internal/infrastructure/database/document_repository_test.go` (new)
  - `internal/presentation/admin/handlers_documents.go` (new)
  - `internal/application/services/reminder.go`
  - `internal/infrastructure/database/admin_repository.go`
  - `internal/presentation/admin/handlers_expected_signers.go`
  - `internal/presentation/admin/routes_admin.go`
  - `templates/admin_dashboard.html.tpl`
  - `templates/admin_document_expected_signers.html.tpl`
  - `templates/base.html.tpl`
  - `static/favicon.svg` (new)
  - `migrations/0005_create_documents_table.{up,down}.sql` (new)

## [1.1.2] - 2025-10-03

### Added
- **SSO Provider Logout**: Complete session termination at OAuth provider level
  - Added `LogoutURL` configuration for OAuth providers
  - Automatic redirect to provider logout (Google, GitHub, GitLab, custom)
  - New environment variable `ACKIFY_OAUTH_LOGOUT_URL` for custom providers
  - Users are now properly logged out from both the application and the SSO provider

### Fixed
- **Blockchain chain isolation**: Each document now has its own independent blockchain
  - `GetLastSignature` now filters by `doc_id` to prevent cross-document chain corruption
  - Genesis signatures are correctly created per document
  - Prevents blockchain chains from mixing between different documents
  - Added comprehensive tests for multi-document blockchain integrity

### Changed
- `GetLastSignature` method signature updated to include `docID` parameter
- All repository implementations updated to support document-scoped blockchain queries

### Technical Details
- Modified files:
  - `internal/application/services/signature.go`
  - `internal/infrastructure/database/repository.go`
  - `internal/infrastructure/auth/oauth.go`
  - `internal/infrastructure/config/config.go`
  - `internal/presentation/handlers/auth.go`
  - `internal/presentation/handlers/interfaces.go`
  - `pkg/web/server.go`
- All existing tests updated and passing

## [1.1.1] - 2025-01-XX

### Changed
- Refactor template variables to separate from locale strings
- Improve database operations for UserName handling

## [1.1.0] - 2025-01-XX

### Added
- Blockchain hash determinism improvements
- ED25519 key generation documentation

### Fixed
- NULL UserName handling in database operations
- Proper string conversion for UserName field

[1.2.1]: https://github.com/btouchard/ackify-ce/compare/v1.2.0...v1.2.1
[1.2.0]: https://github.com/btouchard/ackify-ce/compare/v1.1.3...v1.2.0
[1.1.3]: https://github.com/btouchard/ackify-ce/compare/v1.1.2...v1.1.3
[1.1.2]: https://github.com/btouchard/ackify-ce/compare/v1.1.1...v1.1.2
[1.1.1]: https://github.com/btouchard/ackify-ce/compare/v1.1.0...v1.1.1
[1.1.0]: https://github.com/btouchard/ackify-ce/releases/tag/v1.1.0
