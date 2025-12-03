// SPDX-License-Identifier: AGPL-3.0-or-later
package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/btouchard/ackify-ce/backend/internal/domain/models"
	"github.com/btouchard/ackify-ce/backend/internal/infrastructure/tenant"
	"github.com/btouchard/ackify-ce/backend/pkg/logger"
)

// ExpectedSignerRepository handles database operations for expected signers
type ExpectedSignerRepository struct {
	db      *sql.DB
	tenants tenant.Provider
}

// NewExpectedSignerRepository creates a new expected signer repository
func NewExpectedSignerRepository(db *sql.DB, tenants tenant.Provider) *ExpectedSignerRepository {
	return &ExpectedSignerRepository{db: db, tenants: tenants}
}

// AddExpected batch-inserts multiple expected signers with conflict-safe deduplication on (doc_id, email)
func (r *ExpectedSignerRepository) AddExpected(ctx context.Context, docID string, contacts []models.ContactInfo, addedBy string) error {
	if len(contacts) == 0 {
		return nil
	}

	tenantID, err := r.tenants.CurrentTenant(ctx)
	if err != nil {
		return fmt.Errorf("failed to get tenant: %w", err)
	}

	// Build batch INSERT with ON CONFLICT DO NOTHING
	valueStrings := make([]string, 0, len(contacts))
	valueArgs := make([]interface{}, 0, len(contacts)*5)

	for i, contact := range contacts {
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d)", i*5+1, i*5+2, i*5+3, i*5+4, i*5+5))
		valueArgs = append(valueArgs, tenantID, docID, contact.Email, contact.Name, addedBy)
	}

	query := fmt.Sprintf(`
		INSERT INTO expected_signers (tenant_id, doc_id, email, name, added_by)
		VALUES %s
		ON CONFLICT (doc_id, email) DO NOTHING
	`, strings.Join(valueStrings, ","))

	_, err = r.db.ExecContext(ctx, query, valueArgs...)
	if err != nil {
		return fmt.Errorf("failed to add expected signers: %w", err)
	}

	return nil
}

// ListByDocID retrieves all expected signers for a document, ordered chronologically by when they were added
func (r *ExpectedSignerRepository) ListByDocID(ctx context.Context, docID string) ([]*models.ExpectedSigner, error) {
	tenantID, err := r.tenants.CurrentTenant(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}

	query := `
		SELECT id, tenant_id, doc_id, email, name, added_at, added_by, notes
		FROM expected_signers
		WHERE tenant_id = $1 AND doc_id = $2
		ORDER BY added_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, tenantID, docID)
	if err != nil {
		return nil, fmt.Errorf("failed to query expected signers: %w", err)
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			logger.Logger.Error("failed to close rows", "error", err)
		}
	}(rows)

	var signers []*models.ExpectedSigner
	for rows.Next() {
		signer := &models.ExpectedSigner{}
		err := rows.Scan(
			&signer.ID,
			&signer.TenantID,
			&signer.DocID,
			&signer.Email,
			&signer.Name,
			&signer.AddedAt,
			&signer.AddedBy,
			&signer.Notes,
		)
		if err != nil {
			continue
		}
		signers = append(signers, signer)
	}

	return signers, nil
}

// ListWithStatusByDocID enriches signer data with signature completion status and reminder tracking metrics
func (r *ExpectedSignerRepository) ListWithStatusByDocID(ctx context.Context, docID string) ([]*models.ExpectedSignerWithStatus, error) {
	tenantID, err := r.tenants.CurrentTenant(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}

	query := `
		SELECT
			es.id,
			es.tenant_id,
			es.doc_id,
			es.email,
			es.name,
			es.added_at,
			es.added_by,
			es.notes,
			CASE WHEN s.id IS NOT NULL THEN true ELSE false END as has_signed,
			s.signed_at,
			s.user_name,
			MAX(rl.sent_at) as last_reminder_sent,
			COUNT(CASE WHEN rl.status = 'sent' THEN 1 END) as reminder_count,
			EXTRACT(DAY FROM (NOW() - es.added_at))::int as days_since_added,
			EXTRACT(DAY FROM (NOW() - MAX(rl.sent_at)))::int as days_since_last_reminder
		FROM expected_signers es
		LEFT JOIN signatures s ON es.tenant_id = s.tenant_id AND es.doc_id = s.doc_id AND es.email = s.user_email
		LEFT JOIN reminder_logs rl ON es.tenant_id = rl.tenant_id AND es.doc_id = rl.doc_id AND es.email = rl.recipient_email
		WHERE es.tenant_id = $1 AND es.doc_id = $2
		GROUP BY es.id, es.tenant_id, es.doc_id, es.email, es.name, es.added_at, es.added_by, es.notes, s.id, s.signed_at, s.user_name
		ORDER BY has_signed DESC, es.added_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, tenantID, docID)
	if err != nil {
		return nil, fmt.Errorf("failed to query expected signers with status: %w", err)
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			logger.Logger.Error("failed to close rows", "error", err)
		}
	}(rows)

	var signers []*models.ExpectedSignerWithStatus
	for rows.Next() {
		signer := &models.ExpectedSignerWithStatus{}
		var lastReminderSent sql.NullTime
		var daysSinceLastReminder sql.NullInt64

		err := rows.Scan(
			&signer.ID,
			&signer.TenantID,
			&signer.DocID,
			&signer.Email,
			&signer.Name,
			&signer.AddedAt,
			&signer.AddedBy,
			&signer.Notes,
			&signer.HasSigned,
			&signer.SignedAt,
			&signer.UserName,
			&lastReminderSent,
			&signer.ReminderCount,
			&signer.DaysSinceAdded,
			&daysSinceLastReminder,
		)
		if err != nil {
			continue
		}

		if lastReminderSent.Valid {
			signer.LastReminderSent = &lastReminderSent.Time
		}

		if daysSinceLastReminder.Valid {
			days := int(daysSinceLastReminder.Int64)
			signer.DaysSinceLastReminder = &days
		}

		signers = append(signers, signer)
	}

	return signers, nil
}

// Remove deletes a specific expected signer by document ID and email address
func (r *ExpectedSignerRepository) Remove(ctx context.Context, docID, email string) error {
	tenantID, err := r.tenants.CurrentTenant(ctx)
	if err != nil {
		return fmt.Errorf("failed to get tenant: %w", err)
	}

	query := `
		DELETE FROM expected_signers
		WHERE tenant_id = $1 AND doc_id = $2 AND email = $3
	`

	result, err := r.db.ExecContext(ctx, query, tenantID, docID, email)
	if err != nil {
		return fmt.Errorf("failed to remove expected signer: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("expected signer not found")
	}

	return nil
}

// RemoveAllForDoc purges all expected signers associated with a document in a single operation
func (r *ExpectedSignerRepository) RemoveAllForDoc(ctx context.Context, docID string) error {
	tenantID, err := r.tenants.CurrentTenant(ctx)
	if err != nil {
		return fmt.Errorf("failed to get tenant: %w", err)
	}

	query := `
		DELETE FROM expected_signers
		WHERE tenant_id = $1 AND doc_id = $2
	`

	_, err = r.db.ExecContext(ctx, query, tenantID, docID)
	if err != nil {
		return fmt.Errorf("failed to remove all expected signers: %w", err)
	}

	return nil
}

// IsExpected efficiently verifies if an email address is in the expected signer list for a document
func (r *ExpectedSignerRepository) IsExpected(ctx context.Context, docID, email string) (bool, error) {
	tenantID, err := r.tenants.CurrentTenant(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to get tenant: %w", err)
	}

	query := `
		SELECT EXISTS(
			SELECT 1 FROM expected_signers
			WHERE tenant_id = $1 AND doc_id = $2 AND email = $3
		)
	`

	var exists bool
	err = r.db.QueryRowContext(ctx, query, tenantID, docID, email).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check if email is expected: %w", err)
	}

	return exists, nil
}

// GetStats calculates signature completion metrics including percentage progress for a document
func (r *ExpectedSignerRepository) GetStats(ctx context.Context, docID string) (*models.DocCompletionStats, error) {
	tenantID, err := r.tenants.CurrentTenant(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}

	query := `
		SELECT
			COUNT(*) as expected_count,
			COUNT(s.id) as signed_count
		FROM expected_signers es
		LEFT JOIN signatures s ON es.tenant_id = s.tenant_id AND es.doc_id = s.doc_id AND es.email = s.user_email
		WHERE es.tenant_id = $1 AND es.doc_id = $2
	`

	stats := &models.DocCompletionStats{
		DocID: docID,
	}

	err = r.db.QueryRowContext(ctx, query, tenantID, docID).Scan(&stats.ExpectedCount, &stats.SignedCount)
	if err != nil {
		return nil, fmt.Errorf("failed to get stats: %w", err)
	}

	stats.PendingCount = stats.ExpectedCount - stats.SignedCount

	if stats.ExpectedCount > 0 {
		stats.CompletionRate = float64(stats.SignedCount) / float64(stats.ExpectedCount) * 100
	} else {
		stats.CompletionRate = 0
	}

	return stats, nil
}
