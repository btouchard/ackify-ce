// SPDX-License-Identifier: AGPL-3.0-or-later
package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/btouchard/ackify-ce/backend/internal/domain/models"
	"github.com/btouchard/ackify-ce/backend/pkg/logger"
)

// DocumentRepository handles document metadata persistence
type DocumentRepository struct {
	db *sql.DB
}

// NewDocumentRepository creates a new DocumentRepository
func NewDocumentRepository(db *sql.DB) *DocumentRepository {
	return &DocumentRepository{db: db}
}

// Create persists a new document with metadata including optional checksum validation data
func (r *DocumentRepository) Create(ctx context.Context, docID string, input models.DocumentInput, createdBy string) (*models.Document, error) {
	query := `
		INSERT INTO documents (doc_id, title, url, checksum, checksum_algorithm, description, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING doc_id, title, url, checksum, checksum_algorithm, description, created_at, updated_at, created_by, deleted_at
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

	doc := &models.Document{}
	err := r.db.QueryRowContext(
		ctx,
		query,
		docID,
		input.Title,
		input.URL,
		checksum,
		checksumAlgorithm,
		input.Description,
		createdBy,
	).Scan(
		&doc.DocID,
		&doc.Title,
		&doc.URL,
		&doc.Checksum,
		&doc.ChecksumAlgorithm,
		&doc.Description,
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
func (r *DocumentRepository) GetByDocID(ctx context.Context, docID string) (*models.Document, error) {
	query := `
		SELECT doc_id, title, url, checksum, checksum_algorithm, description, created_at, updated_at, created_by, deleted_at
		FROM documents
		WHERE doc_id = $1 AND deleted_at IS NULL
	`

	doc := &models.Document{}
	err := r.db.QueryRowContext(ctx, query, docID).Scan(
		&doc.DocID,
		&doc.Title,
		&doc.URL,
		&doc.Checksum,
		&doc.ChecksumAlgorithm,
		&doc.Description,
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
func (r *DocumentRepository) FindByReference(ctx context.Context, ref string, refType string) (*models.Document, error) {
	var query string
	var args []interface{}

	switch refType {
	case "url":
		// Search by URL field (excluding soft-deleted)
		query = `
			SELECT doc_id, title, url, checksum, checksum_algorithm, description, created_at, updated_at, created_by, deleted_at
			FROM documents
			WHERE url = $1 AND deleted_at IS NULL
			LIMIT 1
		`
		args = []interface{}{ref}

	case "path":
		// Search by URL field (paths are also stored in url field, excluding soft-deleted)
		query = `
			SELECT doc_id, title, url, checksum, checksum_algorithm, description, created_at, updated_at, created_by, deleted_at
			FROM documents
			WHERE url = $1 AND deleted_at IS NULL
			LIMIT 1
		`
		args = []interface{}{ref}

	case "reference":
		// Search by doc_id (excluding soft-deleted)
		query = `
			SELECT doc_id, title, url, checksum, checksum_algorithm, description, created_at, updated_at, created_by, deleted_at
			FROM documents
			WHERE doc_id = $1 AND deleted_at IS NULL
			LIMIT 1
		`
		args = []interface{}{ref}

	default:
		return nil, fmt.Errorf("unknown reference type: %s", refType)
	}

	doc := &models.Document{}
	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&doc.DocID,
		&doc.Title,
		&doc.URL,
		&doc.Checksum,
		&doc.ChecksumAlgorithm,
		&doc.Description,
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
func (r *DocumentRepository) Update(ctx context.Context, docID string, input models.DocumentInput) (*models.Document, error) {
	query := `
		UPDATE documents
		SET title = $2, url = $3, checksum = $4, checksum_algorithm = $5, description = $6
		WHERE doc_id = $1 AND deleted_at IS NULL
		RETURNING doc_id, title, url, checksum, checksum_algorithm, description, created_at, updated_at, created_by, deleted_at
	`

	// Use empty string for empty checksum fields (table has NOT NULL DEFAULT '')
	checksum := input.Checksum
	checksumAlgorithm := input.ChecksumAlgorithm
	if checksumAlgorithm == "" {
		checksumAlgorithm = "SHA-256" // Default algorithm
	}

	doc := &models.Document{}
	err := r.db.QueryRowContext(
		ctx,
		query,
		docID,
		input.Title,
		input.URL,
		checksum,
		checksumAlgorithm,
		input.Description,
	).Scan(
		&doc.DocID,
		&doc.Title,
		&doc.URL,
		&doc.Checksum,
		&doc.ChecksumAlgorithm,
		&doc.Description,
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
	query := `
		INSERT INTO documents (doc_id, title, url, checksum, checksum_algorithm, description, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (doc_id) DO UPDATE SET
			title = EXCLUDED.title,
			url = EXCLUDED.url,
			checksum = EXCLUDED.checksum,
			checksum_algorithm = EXCLUDED.checksum_algorithm,
			description = EXCLUDED.description,
			deleted_at = NULL
		RETURNING doc_id, title, url, checksum, checksum_algorithm, description, created_at, updated_at, created_by, deleted_at
	`

	// Use empty string for empty checksum fields (table has NOT NULL DEFAULT '')
	checksum := input.Checksum
	checksumAlgorithm := input.ChecksumAlgorithm
	if checksumAlgorithm == "" {
		checksumAlgorithm = "SHA-256" // Default algorithm
	}

	doc := &models.Document{}
	err := r.db.QueryRowContext(
		ctx,
		query,
		docID,
		input.Title,
		input.URL,
		checksum,
		checksumAlgorithm,
		input.Description,
		createdBy,
	).Scan(
		&doc.DocID,
		&doc.Title,
		&doc.URL,
		&doc.Checksum,
		&doc.ChecksumAlgorithm,
		&doc.Description,
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
func (r *DocumentRepository) Delete(ctx context.Context, docID string) error {
	query := `UPDATE documents SET deleted_at = now() WHERE doc_id = $1 AND deleted_at IS NULL`

	result, err := r.db.ExecContext(ctx, query, docID)
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
func (r *DocumentRepository) List(ctx context.Context, limit, offset int) ([]*models.Document, error) {
	query := `
		SELECT doc_id, title, url, checksum, checksum_algorithm, description, created_at, updated_at, created_by, deleted_at
		FROM documents
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
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
			&doc.Title,
			&doc.URL,
			&doc.Checksum,
			&doc.ChecksumAlgorithm,
			&doc.Description,
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
