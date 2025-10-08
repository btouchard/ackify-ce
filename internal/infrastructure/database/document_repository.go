// SPDX-License-Identifier: AGPL-3.0-or-later
package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/btouchard/ackify-ce/internal/domain/models"
	"github.com/btouchard/ackify-ce/pkg/logger"
)

// DocumentRepository handles document metadata persistence
type DocumentRepository struct {
	db *sql.DB
}

// NewDocumentRepository creates a new DocumentRepository
func NewDocumentRepository(db *sql.DB) *DocumentRepository {
	return &DocumentRepository{db: db}
}

// Create creates a new document metadata entry
func (r *DocumentRepository) Create(ctx context.Context, docID string, input models.DocumentInput, createdBy string) (*models.Document, error) {
	query := `
		INSERT INTO documents (doc_id, title, url, checksum, checksum_algorithm, description, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING doc_id, title, url, checksum, checksum_algorithm, description, created_at, updated_at, created_by
	`

	doc := &models.Document{}
	err := r.db.QueryRowContext(
		ctx,
		query,
		docID,
		input.Title,
		input.URL,
		input.Checksum,
		input.ChecksumAlgorithm,
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
	)

	if err != nil {
		logger.Logger.Error("Failed to create document", "error", err.Error(), "doc_id", docID)
		return nil, fmt.Errorf("failed to create document: %w", err)
	}

	return doc, nil
}

// GetByDocID retrieves document metadata by document ID
func (r *DocumentRepository) GetByDocID(ctx context.Context, docID string) (*models.Document, error) {
	query := `
		SELECT doc_id, title, url, checksum, checksum_algorithm, description, created_at, updated_at, created_by
		FROM documents
		WHERE doc_id = $1
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

// Update updates document metadata
func (r *DocumentRepository) Update(ctx context.Context, docID string, input models.DocumentInput) (*models.Document, error) {
	query := `
		UPDATE documents
		SET title = $2, url = $3, checksum = $4, checksum_algorithm = $5, description = $6
		WHERE doc_id = $1
		RETURNING doc_id, title, url, checksum, checksum_algorithm, description, created_at, updated_at, created_by
	`

	doc := &models.Document{}
	err := r.db.QueryRowContext(
		ctx,
		query,
		docID,
		input.Title,
		input.URL,
		input.Checksum,
		input.ChecksumAlgorithm,
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

// CreateOrUpdate creates or updates document metadata
func (r *DocumentRepository) CreateOrUpdate(ctx context.Context, docID string, input models.DocumentInput, createdBy string) (*models.Document, error) {
	query := `
		INSERT INTO documents (doc_id, title, url, checksum, checksum_algorithm, description, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (doc_id) DO UPDATE SET
			title = EXCLUDED.title,
			url = EXCLUDED.url,
			checksum = EXCLUDED.checksum,
			checksum_algorithm = EXCLUDED.checksum_algorithm,
			description = EXCLUDED.description
		RETURNING doc_id, title, url, checksum, checksum_algorithm, description, created_at, updated_at, created_by
	`

	doc := &models.Document{}
	err := r.db.QueryRowContext(
		ctx,
		query,
		docID,
		input.Title,
		input.URL,
		input.Checksum,
		input.ChecksumAlgorithm,
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
	)

	if err != nil {
		logger.Logger.Error("Failed to create or update document", "error", err.Error(), "doc_id", docID)
		return nil, fmt.Errorf("failed to create or update document: %w", err)
	}

	return doc, nil
}

// Delete deletes document metadata
func (r *DocumentRepository) Delete(ctx context.Context, docID string) error {
	query := `DELETE FROM documents WHERE doc_id = $1`

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
		return fmt.Errorf("document not found")
	}

	return nil
}

// List retrieves all documents with pagination
func (r *DocumentRepository) List(ctx context.Context, limit, offset int) ([]*models.Document, error) {
	query := `
		SELECT doc_id, title, url, checksum, checksum_algorithm, description, created_at, updated_at, created_by
		FROM documents
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
