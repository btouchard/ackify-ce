# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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

[1.1.3]: https://github.com/btouchard/ackify-ce/compare/v1.1.2...v1.1.3
[1.1.2]: https://github.com/btouchard/ackify-ce/compare/v1.1.1...v1.1.2
[1.1.1]: https://github.com/btouchard/ackify-ce/compare/v1.1.0...v1.1.1
[1.1.0]: https://github.com/btouchard/ackify-ce/releases/tag/v1.1.0
