# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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

[1.1.2]: https://github.com/btouchard/ackify-ce/compare/v1.1.1...v1.1.2
[1.1.1]: https://github.com/btouchard/ackify-ce/compare/v1.1.0...v1.1.1
[1.1.0]: https://github.com/btouchard/ackify-ce/releases/tag/v1.1.0
