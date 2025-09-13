package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/btouchard/ackify-ce/internal/domain/models"
)

type SignatureRepository struct {
	db *sql.DB
}

// NewSignatureRepository creates a new PostgresSQL signature repository
func NewSignatureRepository(db *sql.DB) *SignatureRepository {
	return &SignatureRepository{db: db}
}

func (r *SignatureRepository) Create(ctx context.Context, signature *models.Signature) error {
	query := `
		INSERT INTO signatures (doc_id, user_sub, user_email, user_name, signed_at, payload_hash, signature, nonce, referer, prev_hash)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at
	`

	err := r.db.QueryRowContext(
		ctx, query,
		signature.DocID,
		signature.UserSub,
		signature.UserEmail,
		signature.UserName,
		signature.SignedAtUTC,
		signature.PayloadHash,
		signature.Signature,
		signature.Nonce,
		signature.Referer,
		signature.PrevHash,
	).Scan(&signature.ID, &signature.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create signature: %w", err)
	}

	return nil
}

func (r *SignatureRepository) GetByDocAndUser(ctx context.Context, docID, userSub string) (*models.Signature, error) {
	query := `
		SELECT id, doc_id, user_sub, user_email, user_name, signed_at, payload_hash, signature, nonce, created_at, referer, prev_hash
		FROM signatures 
		WHERE doc_id = $1 AND user_sub = $2
	`

	signature := &models.Signature{}
	err := r.db.QueryRowContext(ctx, query, docID, userSub).Scan(
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
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrSignatureNotFound
		}
		return nil, fmt.Errorf("failed to get signature: %w", err)
	}

	return signature, nil
}

func (r *SignatureRepository) GetByDoc(ctx context.Context, docID string) ([]*models.Signature, error) {
	query := `
		SELECT id, doc_id, user_sub, user_email, user_name, signed_at, payload_hash, signature, nonce, created_at, referer, prev_hash
		FROM signatures 
		WHERE doc_id = $1 
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, docID)
	if err != nil {
		return nil, fmt.Errorf("failed to query signatures: %w", err)
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

func (r *SignatureRepository) GetByUser(ctx context.Context, userSub string) ([]*models.Signature, error) {
	query := `
		SELECT id, doc_id, user_sub, user_email, user_name, signed_at, payload_hash, signature, nonce, created_at, referer, prev_hash
		FROM signatures 
		WHERE user_sub = $1 
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userSub)
	if err != nil {
		return nil, fmt.Errorf("failed to query user signatures: %w", err)
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

func (r *SignatureRepository) ExistsByDocAndUser(ctx context.Context, docID, userSub string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM signatures WHERE doc_id = $1 AND user_sub = $2)`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, docID, userSub).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check signature existence: %w", err)
	}

	return exists, nil
}

func (r *SignatureRepository) CheckUserSignatureStatus(ctx context.Context, docID, userIdentifier string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM signatures 
			WHERE doc_id = $1 AND (user_sub = $2 OR LOWER(user_email) = LOWER($2))
		)
	`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, docID, userIdentifier).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check user signature status: %w", err)
	}

	return exists, nil
}

func (r *SignatureRepository) GetLastSignature(ctx context.Context) (*models.Signature, error) {
	query := `
		SELECT id, doc_id, user_sub, user_email, user_name, signed_at, payload_hash, signature, nonce, created_at, referer, prev_hash
		FROM signatures 
		ORDER BY id DESC 
		LIMIT 1
	`

	signature := &models.Signature{}
	err := r.db.QueryRowContext(ctx, query).Scan(
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
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get last signature: %w", err)
	}

	return signature, nil
}

func (r *SignatureRepository) GetAllSignaturesOrdered(ctx context.Context) ([]*models.Signature, error) {
	query := `
		SELECT id, doc_id, user_sub, user_email, user_name, signed_at, payload_hash, signature, nonce, created_at, referer, prev_hash
		FROM signatures 
		ORDER BY id ASC`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query all signatures: %w", err)
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
