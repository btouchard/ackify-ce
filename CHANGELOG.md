# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [2.0.0] - 2025-10-16

### 🎉 Major Release: API-First Vue Migration

Complete architectural overhaul to a modern API-first architecture with Vue 3 SPA frontend.

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
  - Pages: Home, Sign, Signatures, Embed, Admin Dashboard, Document Details

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
  - CSRF token protection for all state-changing operations
  - Rate limiting (5 auth attempts/min, 100 general requests/min)
  - CORS configuration for development and production
  - Security headers (CSP, X-Content-Type-Options, X-Frame-Options, etc.)
  - Session-based authentication with secure cookies
  - Request ID propagation for distributed tracing

- **Public Embed Route**
  - `/embed/{docId}` route for public embedding (no authentication required)
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

For users upgrading from v1.x to v2.0:

1. **Environment Variables**: Add optional `ACKIFY_LOG_LEVEL` and `ACKIFY_OAUTH_AUTO_LOGIN` if desired
2. **Docker**: Rebuild images to include Vue SPA build
3. **API Clients**: Consider migrating to new API v1 endpoints for better structure
4. **Embed URLs**: Update to use `/embed/{docId}` instead of token-based system

### Breaking Changes

- None - v2.0 maintains backward compatibility with all v1.x features
- Template-based admin interface remains functional
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

[2.0.0]: https://github.com/btouchard/ackify-ce/compare/v1.1.3...v2.0.0
[1.1.3]: https://github.com/btouchard/ackify-ce/compare/v1.1.2...v1.1.3
[1.1.2]: https://github.com/btouchard/ackify-ce/compare/v1.1.1...v1.1.2
[1.1.1]: https://github.com/btouchard/ackify-ce/compare/v1.1.0...v1.1.1
[1.1.0]: https://github.com/btouchard/ackify-ce/releases/tag/v1.1.0
