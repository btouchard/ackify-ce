// SPDX-License-Identifier: AGPL-3.0-or-later
package services

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/btouchard/ackify-ce/backend/pkg/logger"
	"github.com/btouchard/ackify-ce/backend/pkg/models"
)

// ChecksumVerificationRepository defines the interface for checksum verification persistence
type ChecksumVerificationRepository interface {
	RecordVerification(ctx context.Context, verification *models.ChecksumVerification) error
	GetVerificationHistory(ctx context.Context, docID string, limit int) ([]*models.ChecksumVerification, error)
	GetLastVerification(ctx context.Context, docID string) (*models.ChecksumVerification, error)
}

// DocumentRepository defines the interface for document metadata operations
type DocumentRepository interface {
	GetByDocID(ctx context.Context, docID string) (*models.Document, error)
}

// ChecksumService orchestrates document integrity verification with audit trail persistence
type ChecksumService struct {
	verificationRepo ChecksumVerificationRepository
	documentRepo     DocumentRepository
}

// NewChecksumService initializes checksum verification service with required repository dependencies
func NewChecksumService(
	verificationRepo ChecksumVerificationRepository,
	documentRepo DocumentRepository,
) *ChecksumService {
	return &ChecksumService{
		verificationRepo: verificationRepo,
		documentRepo:     documentRepo,
	}
}

// ValidateChecksumFormat ensures checksum matches expected hexadecimal length for SHA-256/SHA-512/MD5
func (s *ChecksumService) ValidateChecksumFormat(checksum, algorithm string) error {
	// Remove common separators and whitespace
	checksum = normalizeChecksum(checksum)

	var expectedLength int
	switch algorithm {
	case "SHA-256":
		expectedLength = 64
	case "SHA-512":
		expectedLength = 128
	case "MD5":
		expectedLength = 32
	default:
		return fmt.Errorf("unsupported algorithm: %s", algorithm)
	}

	// Check length
	if len(checksum) != expectedLength {
		return fmt.Errorf("invalid checksum length for %s: expected %d hexadecimal characters, got %d", algorithm, expectedLength, len(checksum))
	}

	// Check if it's a valid hex string
	hexPattern := regexp.MustCompile("^[a-fA-F0-9]+$")
	if !hexPattern.MatchString(checksum) {
		return fmt.Errorf("invalid checksum format: must contain only hexadecimal characters (0-9, a-f, A-F)")
	}

	return nil
}

// VerifyChecksum compares calculated hash against stored reference and creates immutable audit record
func (s *ChecksumService) VerifyChecksum(ctx context.Context, docID, calculatedChecksum, verifiedBy string) (*models.ChecksumVerificationResult, error) {
	// Get document metadata
	doc, err := s.documentRepo.GetByDocID(ctx, docID)
	if err != nil {
		return nil, fmt.Errorf("failed to get document: %w", err)
	}

	if doc == nil {
		return nil, fmt.Errorf("document not found: %s", docID)
	}

	// Normalize checksums for comparison
	normalizedCalculated := normalizeChecksum(calculatedChecksum)
	normalizedStored := normalizeChecksum(doc.Checksum)

	// Determine the algorithm to use (from document or default to SHA-256)
	algorithm := doc.ChecksumAlgorithm
	if algorithm == "" {
		algorithm = "SHA-256"
	}

	// Validate the calculated checksum format
	if err := s.ValidateChecksumFormat(normalizedCalculated, algorithm); err != nil {
		// Record failed verification with error
		errorMsg := err.Error()
		verification := &models.ChecksumVerification{
			DocID:              docID,
			VerifiedBy:         verifiedBy,
			VerifiedAt:         time.Now(),
			StoredChecksum:     normalizedStored,
			CalculatedChecksum: normalizedCalculated,
			Algorithm:          algorithm,
			IsValid:            false,
			ErrorMessage:       &errorMsg,
		}
		_ = s.verificationRepo.RecordVerification(ctx, verification)

		return nil, fmt.Errorf("invalid checksum format: %w", err)
	}

	// Check if document has a reference checksum
	if !doc.HasChecksum() {
		result := &models.ChecksumVerificationResult{
			Valid:              false,
			StoredChecksum:     "",
			CalculatedChecksum: normalizedCalculated,
			Algorithm:          algorithm,
			Message:            "No reference checksum configured for this document",
			HasReferenceHash:   false,
		}
		return result, nil
	}

	// Compare checksums (case-insensitive)
	isValid := strings.EqualFold(normalizedCalculated, normalizedStored)

	// Record verification
	verification := &models.ChecksumVerification{
		DocID:              docID,
		VerifiedBy:         verifiedBy,
		VerifiedAt:         time.Now(),
		StoredChecksum:     normalizedStored,
		CalculatedChecksum: normalizedCalculated,
		Algorithm:          algorithm,
		IsValid:            isValid,
		ErrorMessage:       nil,
	}

	if err := s.verificationRepo.RecordVerification(ctx, verification); err != nil {
		logger.Logger.Error("Failed to record verification", "error", err.Error(), "doc_id", docID)
		// Continue even if recording fails - return the result
	}

	var message string
	if isValid {
		message = "Checksums match - document integrity verified"
	} else {
		message = "Checksums do not match - document may have been modified"
	}

	result := &models.ChecksumVerificationResult{
		Valid:              isValid,
		StoredChecksum:     normalizedStored,
		CalculatedChecksum: normalizedCalculated,
		Algorithm:          algorithm,
		Message:            message,
		HasReferenceHash:   true,
	}

	return result, nil
}

// GetVerificationHistory retrieves paginated audit trail of all checksum validation attempts
func (s *ChecksumService) GetVerificationHistory(ctx context.Context, docID string, limit int) ([]*models.ChecksumVerification, error) {
	if limit <= 0 {
		limit = 20
	}

	return s.verificationRepo.GetVerificationHistory(ctx, docID, limit)
}

// GetSupportedAlgorithms returns available hash algorithms for client-side documentation
func (s *ChecksumService) GetSupportedAlgorithms() []string {
	return []string{"SHA-256", "SHA-512", "MD5"}
}

// GetChecksumInfo exposes document hash metadata for public verification interfaces
func (s *ChecksumService) GetChecksumInfo(ctx context.Context, docID string) (map[string]interface{}, error) {
	doc, err := s.documentRepo.GetByDocID(ctx, docID)
	if err != nil {
		return nil, fmt.Errorf("failed to get document: %w", err)
	}

	if doc == nil {
		return nil, fmt.Errorf("document not found: %s", docID)
	}

	algorithm := doc.ChecksumAlgorithm
	if algorithm == "" {
		algorithm = "SHA-256"
	}

	info := map[string]interface{}{
		"doc_id":               docID,
		"has_checksum":         doc.HasChecksum(),
		"algorithm":            algorithm,
		"checksum_length":      doc.GetExpectedChecksumLength(),
		"supported_algorithms": s.GetSupportedAlgorithms(),
	}

	return info, nil
}

// normalizeChecksum removes common separators and converts to lowercase
func normalizeChecksum(checksum string) string {
	// Remove spaces, hyphens, underscores
	checksum = strings.ReplaceAll(checksum, " ", "")
	checksum = strings.ReplaceAll(checksum, "-", "")
	checksum = strings.ReplaceAll(checksum, "_", "")
	checksum = strings.TrimSpace(checksum)
	// Convert to lowercase for case-insensitive comparison
	return strings.ToLower(checksum)
}
