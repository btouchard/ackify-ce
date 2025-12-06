// SPDX-License-Identifier: AGPL-3.0-or-later
package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/btouchard/ackify-ce/internal/domain/models"
)

type DocumentAgg struct {
	DocID           string `json:"doc_id"`
	Count           int    `json:"count"`            // Total signatures
	ExpectedCount   int    `json:"expected_count"`   // Nombre de signataires attendus
	SignedCount     int    `json:"signed_count"`     // Signatures attendues signées
	UnexpectedCount int    `json:"unexpected_count"` // Signatures non attendues
}

// AdminRepository provides read-only access for admin operations
type AdminRepository struct {
	db *sql.DB
}

func NewAdminRepository(db *sql.DB) *AdminRepository {
	return &AdminRepository{db: db}
}

// ListDocumentsWithCounts aggregates signature metrics across all documents for admin dashboard
func (r *AdminRepository) ListDocumentsWithCounts(ctx context.Context) ([]DocumentAgg, error) {
	query := `
		SELECT
			all_docs.doc_id,
			COALESCE(sig_counts.sig_count, 0) as count,
			COALESCE(expected_counts.expected_count, 0) as expected_count,
			COALESCE(signed_expected.signed_count, 0) as signed_count,
			COALESCE(sig_counts.sig_count, 0) - COALESCE(signed_expected.signed_count, 0) as unexpected_count
		FROM (
			SELECT DISTINCT doc_id FROM signatures
			UNION
			SELECT DISTINCT doc_id FROM expected_signers
			UNION
			SELECT DISTINCT doc_id FROM documents
		) AS all_docs
		LEFT JOIN (
			SELECT doc_id, COUNT(*) as sig_count
			FROM signatures
			GROUP BY doc_id
		) AS sig_counts USING (doc_id)
		LEFT JOIN (
			SELECT doc_id, COUNT(*) as expected_count
			FROM expected_signers
			GROUP BY doc_id
		) AS expected_counts USING (doc_id)
		LEFT JOIN (
			SELECT es.doc_id, COUNT(DISTINCT s.id) as signed_count
			FROM expected_signers es
			INNER JOIN signatures s ON es.doc_id = s.doc_id AND es.email = s.user_email
			GROUP BY es.doc_id
		) AS signed_expected USING (doc_id)
		ORDER BY all_docs.doc_id
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
		err := rows.Scan(&doc.DocID, &doc.Count, &doc.ExpectedCount, &doc.SignedCount, &doc.UnexpectedCount)
		if err != nil {
			continue
		}
		documents = append(documents, doc)
	}

	return documents, nil
}

// ListSignaturesByDoc retrieves all signatures for a document in reverse chronological order
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

// VerifyDocumentChainIntegrity validates cryptographic hash chain continuity for all signatures in a document
func (r *AdminRepository) VerifyDocumentChainIntegrity(ctx context.Context, docID string) (*ChainIntegrityResult, error) {
	signatures, err := r.ListSignaturesByDoc(ctx, docID)
	if err != nil {
		return nil, fmt.Errorf("failed to get signatures for document %s: %w", docID, err)
	}

	return r.verifyChainIntegrity(signatures), nil
}

// ChainIntegrityResult contient le résultat de la vérification d'intégrité
type ChainIntegrityResult struct {
	IsValid     bool     `json:"is_valid"`
	TotalSigs   int      `json:"total_signatures"`
	ValidSigs   int      `json:"valid_signatures"`
	InvalidSigs int      `json:"invalid_signatures"`
	Errors      []string `json:"errors,omitempty"`
	DocID       string   `json:"doc_id"`
}

// verifyChainIntegrity vérifie l'intégrité de la chaîne de signatures
func (r *AdminRepository) verifyChainIntegrity(signatures []*models.Signature) *ChainIntegrityResult {
	result := &ChainIntegrityResult{
		IsValid:     true,
		TotalSigs:   len(signatures),
		ValidSigs:   0,
		InvalidSigs: 0,
		Errors:      []string{},
	}

	if len(signatures) == 0 {
		return result
	}

	// Trier par ID pour vérifier dans l'ordre chronologique
	sortedSigs := make([]*models.Signature, len(signatures))
	copy(sortedSigs, signatures)

	// Tri manuel par ID (ordre croissant)
	for i := 0; i < len(sortedSigs)-1; i++ {
		for j := i + 1; j < len(sortedSigs); j++ {
			if sortedSigs[i].ID > sortedSigs[j].ID {
				sortedSigs[i], sortedSigs[j] = sortedSigs[j], sortedSigs[i]
			}
		}
	}

	result.DocID = sortedSigs[0].DocID

	// Vérification de la première signature (genesis)
	firstSig := sortedSigs[0]
	if firstSig.PrevHash != nil && *firstSig.PrevHash != "" {
		result.IsValid = false
		result.InvalidSigs++
		result.Errors = append(result.Errors, fmt.Sprintf("Genesis signature ID:%d has prev_hash (should be null)", firstSig.ID))
	} else {
		result.ValidSigs++
	}

	// Vérification des signatures suivantes
	for i := 1; i < len(sortedSigs); i++ {
		currentSig := sortedSigs[i]
		prevSig := sortedSigs[i-1]

		expectedPrevHash := prevSig.ComputeRecordHash()

		if currentSig.PrevHash == nil {
			result.IsValid = false
			result.InvalidSigs++
			result.Errors = append(result.Errors, fmt.Sprintf("Signature ID:%d missing prev_hash", currentSig.ID))
		} else if *currentSig.PrevHash != expectedPrevHash {
			result.IsValid = false
			result.InvalidSigs++
			result.Errors = append(result.Errors, fmt.Sprintf("Signature ID:%d has invalid prev_hash: expected %s, got %s",
				currentSig.ID, expectedPrevHash[:12]+"...", (*currentSig.PrevHash)[:12]+"..."))
		} else {
			result.ValidSigs++
		}
	}

	return result
}

// Close gracefully terminates the database connection pool to prevent resource leaks
func (r *AdminRepository) Close() error {
	if r.db != nil {
		return r.db.Close()
	}
	return nil
}
