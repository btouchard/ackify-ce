// SPDX-License-Identifier: AGPL-3.0-or-later
package models

import (
	"time"

	"github.com/google/uuid"
)

// ChecksumVerification represents a verification attempt of a document's checksum
type ChecksumVerification struct {
	ID                 int64     `json:"id" db:"id"`
	TenantID           uuid.UUID `json:"tenant_id" db:"tenant_id"`
	DocID              string    `json:"doc_id" db:"doc_id"`
	VerifiedBy         string    `json:"verified_by" db:"verified_by"`
	VerifiedAt         time.Time `json:"verified_at" db:"verified_at"`
	StoredChecksum     string    `json:"stored_checksum" db:"stored_checksum"`
	CalculatedChecksum string    `json:"calculated_checksum" db:"calculated_checksum"`
	Algorithm          string    `json:"algorithm" db:"algorithm"`
	IsValid            bool      `json:"is_valid" db:"is_valid"`
	ErrorMessage       *string   `json:"error_message,omitempty" db:"error_message"`
}

// ChecksumVerificationResult represents the result of a checksum verification operation
type ChecksumVerificationResult struct {
	Valid              bool   `json:"valid"`
	StoredChecksum     string `json:"stored_checksum"`
	CalculatedChecksum string `json:"calculated_checksum"`
	Algorithm          string `json:"algorithm"`
	Message            string `json:"message"`
	HasReferenceHash   bool   `json:"has_reference_hash"`
}
