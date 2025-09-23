package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/btouchard/ackify-ce/internal/domain/models"
	"github.com/btouchard/ackify-ce/internal/infrastructure/config"
)

// DocumentAgg represents a document with signature count
type DocumentAgg struct {
	DocID string `json:"doc_id"`
	Count int    `json:"count"`
}

// AdminRepository provides read-only access for admin operations
type AdminRepository struct {
	db *sql.DB
}

// NewAdminRepository creates a new admin repository with its own database connection
func NewAdminRepository(ctx context.Context) (*AdminRepository, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	db, err := InitDB(ctx, Config{DSN: cfg.Database.DSN})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize admin database: %w", err)
	}

	return &AdminRepository{db: db}, nil
}

// ListDocumentsWithCounts returns all documents with their signature counts
func (r *AdminRepository) ListDocumentsWithCounts(ctx context.Context) ([]DocumentAgg, error) {
	query := `
		SELECT doc_id, COUNT(*) as count
		FROM signatures
		GROUP BY doc_id
		ORDER BY doc_id
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query documents with counts: %w", err)
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	var documents []DocumentAgg
	for rows.Next() {
		var doc DocumentAgg
		err := rows.Scan(&doc.DocID, &doc.Count)
		if err != nil {
			continue
		}
		documents = append(documents, doc)
	}

	return documents, nil
}

// ListSignaturesByDoc returns all signatures for a specific document
func (r *AdminRepository) ListSignaturesByDoc(ctx context.Context, docID string) ([]*models.Signature, error) {
	query := `
		SELECT id, doc_id, user_sub, user_email, user_name, signed_at, payload_hash, signature, nonce, created_at, referer, prev_hash
		FROM signatures
		WHERE doc_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, docID)
	if err != nil {
		return nil, fmt.Errorf("failed to query signatures for document: %w", err)
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	var signatures []*models.Signature
	for rows.Next() {
		signature := &models.Signature{}
		err := rows.Scan(
			&signature.ID,
			&signature.DocID,
			&signature.UserSub,
			&signature.UserEmail,
			&signature.UserName,
			&signature.SignedAtUTC,
			&signature.PayloadHash,
			&signature.Signature,
			&signature.Nonce,
			&signature.CreatedAt,
			&signature.Referer,
			&signature.PrevHash,
		)
		if err != nil {
			continue
		}
		signatures = append(signatures, signature)
	}

	return signatures, nil
}

// Close closes the database connection
func (r *AdminRepository) Close() error {
	if r.db != nil {
		return r.db.Close()
	}
	return nil
}
