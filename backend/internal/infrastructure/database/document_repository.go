// SPDX-License-Identifier: AGPL-3.0-or-later
package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/btouchard/ackify-ce/backend/internal/domain/models"
	"github.com/btouchard/ackify-ce/backend/internal/infrastructure/dbctx"
	"github.com/btouchard/ackify-ce/backend/internal/infrastructure/tenant"
	"github.com/btouchard/ackify-ce/backend/pkg/logger"
)

// DocumentRepository handles document metadata persistence
type DocumentRepository struct {
	db      *sql.DB
	tenants tenant.Provider
}

// NewDocumentRepository creates a new DocumentRepository
func NewDocumentRepository(db *sql.DB, tenants tenant.Provider) *DocumentRepository {
	return &DocumentRepository{db: db, tenants: tenants}
}

// Create persists a new document with metadata including optional checksum validation data
func (r *DocumentRepository) Create(ctx context.Context, docID string, input models.DocumentInput, createdBy string) (*models.Document, error) {
	tenantID, err := r.tenants.CurrentTenant(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}

	query := `
		INSERT INTO documents (tenant_id, doc_id, title, url, checksum, checksum_algorithm, description, read_mode, allow_download, require_full_read, verify_checksum, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING doc_id, tenant_id, title, url, checksum, checksum_algorithm, description, read_mode, allow_download, require_full_read, verify_checksum, created_at, updated_at, created_by, deleted_at
	`

	// Use NULL for empty checksum fields to avoid constraint violation
	var checksum, checksumAlgorithm interface{}
	if input.Checksum != "" {
		checksum = input.Checksum
		checksumAlgorithm = input.ChecksumAlgorithm
	} else {
		checksum = ""
		checksumAlgorithm = "SHA-256"
	}

	// Handle read_mode with default
	readMode := input.ReadMode
	if readMode == "" {
		readMode = "integrated"
	}

	// Handle boolean defaults
	allowDownload := true
	if input.AllowDownload != nil {
		allowDownload = *input.AllowDownload
	}
	requireFullRead := false
	if input.RequireFullRead != nil {
		requireFullRead = *input.RequireFullRead
	}
	verifyChecksum := true
	if input.VerifyChecksum != nil {
		verifyChecksum = *input.VerifyChecksum
	}

	doc := &models.Document{}
	err = dbctx.GetQuerier(ctx, r.db).QueryRowContext(
		ctx,
		query,
		tenantID,
		docID,
		input.Title,
		input.URL,
		checksum,
		checksumAlgorithm,
		input.Description,
		readMode,
		allowDownload,
		requireFullRead,
		verifyChecksum,
		createdBy,
	).Scan(
		&doc.DocID,
		&doc.TenantID,
		&doc.Title,
		&doc.URL,
		&doc.Checksum,
		&doc.ChecksumAlgorithm,
		&doc.Description,
		&doc.ReadMode,
		&doc.AllowDownload,
		&doc.RequireFullRead,
		&doc.VerifyChecksum,
		&doc.CreatedAt,
		&doc.UpdatedAt,
		&doc.CreatedBy,
		&doc.DeletedAt,
	)

	if err != nil {
		logger.Logger.Error("Failed to create document", "error", err.Error(), "doc_id", docID)
		return nil, fmt.Errorf("failed to create document: %w", err)
	}

	return doc, nil
}

// GetByDocID retrieves document metadata by document ID (excluding soft-deleted documents)
// RLS policy automatically filters by tenant_id
func (r *DocumentRepository) GetByDocID(ctx context.Context, docID string) (*models.Document, error) {
	query := `
		SELECT doc_id, tenant_id, title, url, checksum, checksum_algorithm, description, read_mode, allow_download, require_full_read, verify_checksum, created_at, updated_at, created_by, deleted_at
		FROM documents
		WHERE doc_id = $1 AND deleted_at IS NULL
	`

	doc := &models.Document{}
	err := dbctx.GetQuerier(ctx, r.db).QueryRowContext(ctx, query, docID).Scan(
		&doc.DocID,
		&doc.TenantID,
		&doc.Title,
		&doc.URL,
		&doc.Checksum,
		&doc.ChecksumAlgorithm,
		&doc.Description,
		&doc.ReadMode,
		&doc.AllowDownload,
		&doc.RequireFullRead,
		&doc.VerifyChecksum,
		&doc.CreatedAt,
		&doc.UpdatedAt,
		&doc.CreatedBy,
		&doc.DeletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		logger.Logger.Error("Failed to get document", "error", err.Error(), "doc_id", docID)
		return nil, fmt.Errorf("failed to get document: %w", err)
	}

	return doc, nil
}

// FindByReference searches for a document by reference (URL, path, or doc_id)
// RLS policy automatically filters by tenant_id
func (r *DocumentRepository) FindByReference(ctx context.Context, ref string, refType string) (*models.Document, error) {
	var query string
	var args []interface{}

	switch refType {
	case "url":
		// Search by URL field (excluding soft-deleted)
		query = `
			SELECT doc_id, tenant_id, title, url, checksum, checksum_algorithm, description, read_mode, allow_download, require_full_read, verify_checksum, created_at, updated_at, created_by, deleted_at
			FROM documents
			WHERE url = $1 AND deleted_at IS NULL
			LIMIT 1
		`
		args = []interface{}{ref}

	case "path":
		// Search by URL field (paths are also stored in url field, excluding soft-deleted)
		query = `
			SELECT doc_id, tenant_id, title, url, checksum, checksum_algorithm, description, read_mode, allow_download, require_full_read, verify_checksum, created_at, updated_at, created_by, deleted_at
			FROM documents
			WHERE url = $1 AND deleted_at IS NULL
			LIMIT 1
		`
		args = []interface{}{ref}

	case "reference":
		// Search by doc_id (excluding soft-deleted)
		query = `
			SELECT doc_id, tenant_id, title, url, checksum, checksum_algorithm, description, read_mode, allow_download, require_full_read, verify_checksum, created_at, updated_at, created_by, deleted_at
			FROM documents
			WHERE doc_id = $1 AND deleted_at IS NULL
			LIMIT 1
		`
		args = []interface{}{ref}

	default:
		return nil, fmt.Errorf("unknown reference type: %s", refType)
	}

	doc := &models.Document{}
	err := dbctx.GetQuerier(ctx, r.db).QueryRowContext(ctx, query, args...).Scan(
		&doc.DocID,
		&doc.TenantID,
		&doc.Title,
		&doc.URL,
		&doc.Checksum,
		&doc.ChecksumAlgorithm,
		&doc.Description,
		&doc.ReadMode,
		&doc.AllowDownload,
		&doc.RequireFullRead,
		&doc.VerifyChecksum,
		&doc.CreatedAt,
		&doc.UpdatedAt,
		&doc.CreatedBy,
		&doc.DeletedAt,
	)

	if err == sql.ErrNoRows {
		logger.Logger.Debug("Document not found by reference",
			"reference", ref,
			"type", refType)
		return nil, nil
	}

	if err != nil {
		logger.Logger.Error("Failed to find document by reference",
			"error", err.Error(),
			"reference", ref,
			"type", refType)
		return nil, fmt.Errorf("failed to find document: %w", err)
	}

	logger.Logger.Debug("Document found by reference",
		"doc_id", doc.DocID,
		"reference", ref,
		"type", refType)

	return doc, nil
}

// Update modifies existing document metadata while preserving creation timestamp and creator
// RLS policy automatically filters by tenant_id
func (r *DocumentRepository) Update(ctx context.Context, docID string, input models.DocumentInput) (*models.Document, error) {
	query := `
		UPDATE documents
		SET title = $2, url = $3, checksum = $4, checksum_algorithm = $5, description = $6, read_mode = $7, allow_download = $8, require_full_read = $9, verify_checksum = $10
		WHERE doc_id = $1 AND deleted_at IS NULL
		RETURNING doc_id, tenant_id, title, url, checksum, checksum_algorithm, description, read_mode, allow_download, require_full_read, verify_checksum, created_at, updated_at, created_by, deleted_at
	`

	// Use empty string for empty checksum fields (table has NOT NULL DEFAULT '')
	checksum := input.Checksum
	checksumAlgorithm := input.ChecksumAlgorithm
	if checksumAlgorithm == "" {
		checksumAlgorithm = "SHA-256" // Default algorithm
	}

	// Handle read_mode with default
	readMode := input.ReadMode
	if readMode == "" {
		readMode = "integrated"
	}

	// Handle boolean defaults
	allowDownload := true
	if input.AllowDownload != nil {
		allowDownload = *input.AllowDownload
	}
	requireFullRead := false
	if input.RequireFullRead != nil {
		requireFullRead = *input.RequireFullRead
	}
	verifyChecksum := true
	if input.VerifyChecksum != nil {
		verifyChecksum = *input.VerifyChecksum
	}

	doc := &models.Document{}
	err := dbctx.GetQuerier(ctx, r.db).QueryRowContext(
		ctx,
		query,
		docID,
		input.Title,
		input.URL,
		checksum,
		checksumAlgorithm,
		input.Description,
		readMode,
		allowDownload,
		requireFullRead,
		verifyChecksum,
	).Scan(
		&doc.DocID,
		&doc.TenantID,
		&doc.Title,
		&doc.URL,
		&doc.Checksum,
		&doc.ChecksumAlgorithm,
		&doc.Description,
		&doc.ReadMode,
		&doc.AllowDownload,
		&doc.RequireFullRead,
		&doc.VerifyChecksum,
		&doc.CreatedAt,
		&doc.UpdatedAt,
		&doc.CreatedBy,
		&doc.DeletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("document not found")
	}

	if err != nil {
		logger.Logger.Error("Failed to update document", "error", err.Error(), "doc_id", docID)
		return nil, fmt.Errorf("failed to update document: %w", err)
	}

	return doc, nil
}

// CreateOrUpdate performs upsert operation, creating new document or updating existing one atomically
func (r *DocumentRepository) CreateOrUpdate(ctx context.Context, docID string, input models.DocumentInput, createdBy string) (*models.Document, error) {
	tenantID, err := r.tenants.CurrentTenant(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}

	query := `
		INSERT INTO documents (tenant_id, doc_id, title, url, checksum, checksum_algorithm, description, read_mode, allow_download, require_full_read, verify_checksum, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		ON CONFLICT (doc_id) DO UPDATE SET
			title = EXCLUDED.title,
			url = EXCLUDED.url,
			checksum = EXCLUDED.checksum,
			checksum_algorithm = EXCLUDED.checksum_algorithm,
			description = EXCLUDED.description,
			read_mode = EXCLUDED.read_mode,
			allow_download = EXCLUDED.allow_download,
			require_full_read = EXCLUDED.require_full_read,
			verify_checksum = EXCLUDED.verify_checksum,
			deleted_at = NULL
		RETURNING doc_id, tenant_id, title, url, checksum, checksum_algorithm, description, read_mode, allow_download, require_full_read, verify_checksum, created_at, updated_at, created_by, deleted_at
	`

	// Use empty string for empty checksum fields (table has NOT NULL DEFAULT '')
	checksum := input.Checksum
	checksumAlgorithm := input.ChecksumAlgorithm
	if checksumAlgorithm == "" {
		checksumAlgorithm = "SHA-256" // Default algorithm
	}

	// Handle read_mode with default
	readMode := input.ReadMode
	if readMode == "" {
		readMode = "integrated"
	}

	// Handle boolean defaults
	allowDownload := true
	if input.AllowDownload != nil {
		allowDownload = *input.AllowDownload
	}
	requireFullRead := false
	if input.RequireFullRead != nil {
		requireFullRead = *input.RequireFullRead
	}
	verifyChecksum := true
	if input.VerifyChecksum != nil {
		verifyChecksum = *input.VerifyChecksum
	}

	doc := &models.Document{}
	err = dbctx.GetQuerier(ctx, r.db).QueryRowContext(
		ctx,
		query,
		tenantID,
		docID,
		input.Title,
		input.URL,
		checksum,
		checksumAlgorithm,
		input.Description,
		readMode,
		allowDownload,
		requireFullRead,
		verifyChecksum,
		createdBy,
	).Scan(
		&doc.DocID,
		&doc.TenantID,
		&doc.Title,
		&doc.URL,
		&doc.Checksum,
		&doc.ChecksumAlgorithm,
		&doc.Description,
		&doc.ReadMode,
		&doc.AllowDownload,
		&doc.RequireFullRead,
		&doc.VerifyChecksum,
		&doc.CreatedAt,
		&doc.UpdatedAt,
		&doc.CreatedBy,
		&doc.DeletedAt,
	)

	if err != nil {
		logger.Logger.Error("Failed to create or update document", "error", err.Error(), "doc_id", docID)
		return nil, fmt.Errorf("failed to create or update document: %w", err)
	}

	return doc, nil
}

// Delete soft-deletes document by setting deleted_at timestamp, preserving metadata and signature history
// RLS policy automatically filters by tenant_id
func (r *DocumentRepository) Delete(ctx context.Context, docID string) error {
	query := `UPDATE documents SET deleted_at = now() WHERE doc_id = $1 AND deleted_at IS NULL`

	result, err := dbctx.GetQuerier(ctx, r.db).ExecContext(ctx, query, docID)
	if err != nil {
		logger.Logger.Error("Failed to delete document", "error", err.Error(), "doc_id", docID)
		return fmt.Errorf("failed to delete document: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("document not found or already deleted")
	}

	return nil
}

// List retrieves paginated documents ordered by creation date, newest first (excluding soft-deleted)
// RLS policy automatically filters by tenant_id
func (r *DocumentRepository) List(ctx context.Context, limit, offset int) ([]*models.Document, error) {
	query := `
		SELECT doc_id, tenant_id, title, url, checksum, checksum_algorithm, description, read_mode, allow_download, require_full_read, verify_checksum, created_at, updated_at, created_by, deleted_at
		FROM documents
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := dbctx.GetQuerier(ctx, r.db).QueryContext(ctx, query, limit, offset)
	if err != nil {
		logger.Logger.Error("Failed to list documents", "error", err.Error())
		return nil, fmt.Errorf("failed to list documents: %w", err)
	}
	defer rows.Close()

	documents := []*models.Document{}
	for rows.Next() {
		doc := &models.Document{}
		err := rows.Scan(
			&doc.DocID,
			&doc.TenantID,
			&doc.Title,
			&doc.URL,
			&doc.Checksum,
			&doc.ChecksumAlgorithm,
			&doc.Description,
			&doc.ReadMode,
			&doc.AllowDownload,
			&doc.RequireFullRead,
			&doc.VerifyChecksum,
			&doc.CreatedAt,
			&doc.UpdatedAt,
			&doc.CreatedBy,
			&doc.DeletedAt,
		)
		if err != nil {
			logger.Logger.Error("Failed to scan document row", "error", err.Error())
			return nil, fmt.Errorf("failed to scan document: %w", err)
		}
		documents = append(documents, doc)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating documents: %w", err)
	}

	return documents, nil
}

// Search retrieves paginated documents matching the search query (excluding soft-deleted)
// Searches in doc_id, title, url, and description fields using case-insensitive pattern matching
// RLS policy automatically filters by tenant_id
func (r *DocumentRepository) Search(ctx context.Context, query string, limit, offset int) ([]*models.Document, error) {
	searchQuery := `
		SELECT doc_id, tenant_id, title, url, checksum, checksum_algorithm, description, read_mode, allow_download, require_full_read, verify_checksum, created_at, updated_at, created_by, deleted_at
		FROM documents
		WHERE deleted_at IS NULL
		AND (
			doc_id ILIKE $1
			OR title ILIKE $1
			OR url ILIKE $1
			OR description ILIKE $1
		)
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	searchPattern := "%" + query + "%"
	rows, err := dbctx.GetQuerier(ctx, r.db).QueryContext(ctx, searchQuery, searchPattern, limit, offset)
	if err != nil {
		logger.Logger.Error("Failed to search documents", "error", err.Error(), "query", query)
		return nil, fmt.Errorf("failed to search documents: %w", err)
	}
	defer rows.Close()

	documents := []*models.Document{}
	for rows.Next() {
		doc := &models.Document{}
		err := rows.Scan(
			&doc.DocID,
			&doc.TenantID,
			&doc.Title,
			&doc.URL,
			&doc.Checksum,
			&doc.ChecksumAlgorithm,
			&doc.Description,
			&doc.ReadMode,
			&doc.AllowDownload,
			&doc.RequireFullRead,
			&doc.VerifyChecksum,
			&doc.CreatedAt,
			&doc.UpdatedAt,
			&doc.CreatedBy,
			&doc.DeletedAt,
		)
		if err != nil {
			logger.Logger.Error("Failed to scan document row", "error", err.Error())
			return nil, fmt.Errorf("failed to scan document: %w", err)
		}
		documents = append(documents, doc)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating documents: %w", err)
	}

	logger.Logger.Debug("Document search completed",
		"query", query,
		"results", len(documents),
		"limit", limit,
		"offset", offset)

	return documents, nil
}

// Count returns the total number of documents matching the optional search query (excluding soft-deleted)
func (r *DocumentRepository) Count(ctx context.Context, searchQuery string) (int, error) {
	var query string
	var args []interface{}

	if searchQuery != "" {
		// Count with search filter
		query = `
			SELECT COUNT(*)
			FROM documents
			WHERE deleted_at IS NULL
			AND (
				doc_id ILIKE $1
				OR title ILIKE $1
				OR url ILIKE $1
				OR description ILIKE $1
			)
		`
		searchPattern := "%" + searchQuery + "%"
		args = []interface{}{searchPattern}
	} else {
		// Count all documents
		query = `
			SELECT COUNT(*)
			FROM documents
			WHERE deleted_at IS NULL
		`
		args = []interface{}{}
	}

	var count int
	err := dbctx.GetQuerier(ctx, r.db).QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		logger.Logger.Error("Failed to count documents", "error", err.Error(), "search", searchQuery)
		return 0, fmt.Errorf("failed to count documents: %w", err)
	}

	logger.Logger.Debug("Document count completed", "count", count, "search", searchQuery)
	return count, nil
}

// ListByCreatedBy retrieves paginated documents created by a specific user (excluding soft-deleted)
// RLS policy automatically filters by tenant_id
func (r *DocumentRepository) ListByCreatedBy(ctx context.Context, createdBy string, limit, offset int) ([]*models.Document, error) {
	query := `
		SELECT doc_id, tenant_id, title, url, checksum, checksum_algorithm, description, read_mode, allow_download, require_full_read, verify_checksum, created_at, updated_at, created_by, deleted_at
		FROM documents
		WHERE deleted_at IS NULL AND created_by = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := dbctx.GetQuerier(ctx, r.db).QueryContext(ctx, query, createdBy, limit, offset)
	if err != nil {
		logger.Logger.Error("Failed to list documents by creator", "error", err.Error(), "created_by", createdBy)
		return nil, fmt.Errorf("failed to list documents: %w", err)
	}
	defer rows.Close()

	documents := []*models.Document{}
	for rows.Next() {
		doc := &models.Document{}
		err := rows.Scan(
			&doc.DocID,
			&doc.TenantID,
			&doc.Title,
			&doc.URL,
			&doc.Checksum,
			&doc.ChecksumAlgorithm,
			&doc.Description,
			&doc.ReadMode,
			&doc.AllowDownload,
			&doc.RequireFullRead,
			&doc.VerifyChecksum,
			&doc.CreatedAt,
			&doc.UpdatedAt,
			&doc.CreatedBy,
			&doc.DeletedAt,
		)
		if err != nil {
			logger.Logger.Error("Failed to scan document row", "error", err.Error())
			return nil, fmt.Errorf("failed to scan document: %w", err)
		}
		documents = append(documents, doc)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating documents: %w", err)
	}

	return documents, nil
}

// SearchByCreatedBy retrieves paginated documents matching search query created by a specific user (excluding soft-deleted)
// RLS policy automatically filters by tenant_id
func (r *DocumentRepository) SearchByCreatedBy(ctx context.Context, createdBy, searchQuery string, limit, offset int) ([]*models.Document, error) {
	query := `
		SELECT doc_id, tenant_id, title, url, checksum, checksum_algorithm, description, read_mode, allow_download, require_full_read, verify_checksum, created_at, updated_at, created_by, deleted_at
		FROM documents
		WHERE deleted_at IS NULL AND created_by = $1
		AND (
			doc_id ILIKE $2
			OR title ILIKE $2
			OR url ILIKE $2
			OR description ILIKE $2
		)
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4
	`

	searchPattern := "%" + searchQuery + "%"
	rows, err := dbctx.GetQuerier(ctx, r.db).QueryContext(ctx, query, createdBy, searchPattern, limit, offset)
	if err != nil {
		logger.Logger.Error("Failed to search documents by creator", "error", err.Error(), "created_by", createdBy, "query", searchQuery)
		return nil, fmt.Errorf("failed to search documents: %w", err)
	}
	defer rows.Close()

	documents := []*models.Document{}
	for rows.Next() {
		doc := &models.Document{}
		err := rows.Scan(
			&doc.DocID,
			&doc.TenantID,
			&doc.Title,
			&doc.URL,
			&doc.Checksum,
			&doc.ChecksumAlgorithm,
			&doc.Description,
			&doc.ReadMode,
			&doc.AllowDownload,
			&doc.RequireFullRead,
			&doc.VerifyChecksum,
			&doc.CreatedAt,
			&doc.UpdatedAt,
			&doc.CreatedBy,
			&doc.DeletedAt,
		)
		if err != nil {
			logger.Logger.Error("Failed to scan document row", "error", err.Error())
			return nil, fmt.Errorf("failed to scan document: %w", err)
		}
		documents = append(documents, doc)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating documents: %w", err)
	}

	logger.Logger.Debug("Document search by creator completed",
		"created_by", createdBy,
		"query", searchQuery,
		"results", len(documents),
		"limit", limit,
		"offset", offset)

	return documents, nil
}

// CountByCreatedBy returns the total number of documents created by a specific user (excluding soft-deleted)
func (r *DocumentRepository) CountByCreatedBy(ctx context.Context, createdBy, searchQuery string) (int, error) {
	var query string
	var args []interface{}

	if searchQuery != "" {
		query = `
			SELECT COUNT(*)
			FROM documents
			WHERE deleted_at IS NULL AND created_by = $1
			AND (
				doc_id ILIKE $2
				OR title ILIKE $2
				OR url ILIKE $2
				OR description ILIKE $2
			)
		`
		searchPattern := "%" + searchQuery + "%"
		args = []interface{}{createdBy, searchPattern}
	} else {
		query = `
			SELECT COUNT(*)
			FROM documents
			WHERE deleted_at IS NULL AND created_by = $1
		`
		args = []interface{}{createdBy}
	}

	var count int
	err := dbctx.GetQuerier(ctx, r.db).QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		logger.Logger.Error("Failed to count documents by creator", "error", err.Error(), "created_by", createdBy, "search", searchQuery)
		return 0, fmt.Errorf("failed to count documents: %w", err)
	}

	logger.Logger.Debug("Document count by creator completed", "count", count, "created_by", createdBy, "search", searchQuery)
	return count, nil
}
