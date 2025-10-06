// SPDX-License-Identifier: AGPL-3.0-or-later
package models

import "time"

// ExpectedSigner represents an expected signer for a document
type ExpectedSigner struct {
	ID      int64     `json:"id" db:"id"`
	DocID   string    `json:"doc_id" db:"doc_id"`
	Email   string    `json:"email" db:"email"`
	AddedAt time.Time `json:"added_at" db:"added_at"`
	AddedBy string    `json:"added_by" db:"added_by"`
	Notes   *string   `json:"notes,omitempty" db:"notes"`
}

// ExpectedSignerWithStatus combines expected signer info with signature status
type ExpectedSignerWithStatus struct {
	ExpectedSigner
	HasSigned bool       `json:"has_signed"`
	SignedAt  *time.Time `json:"signed_at,omitempty"`
	UserName  *string    `json:"user_name,omitempty"`
}

// DocCompletionStats provides completion statistics for a document
type DocCompletionStats struct {
	DocID          string  `json:"doc_id"`
	ExpectedCount  int     `json:"expected_count"`
	SignedCount    int     `json:"signed_count"`
	PendingCount   int     `json:"pending_count"`
	CompletionRate float64 `json:"completion_rate"` // Percentage 0-100
}
