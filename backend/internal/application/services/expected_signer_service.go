// SPDX-License-Identifier: AGPL-3.0-or-later
package services

import (
	"context"

	"github.com/btouchard/ackify-ce/backend/internal/domain/models"
)

type expectedSignerRepo interface {
	ListByDocID(ctx context.Context, docID string) ([]*models.ExpectedSigner, error)
	ListWithStatusByDocID(ctx context.Context, docID string) ([]*models.ExpectedSignerWithStatus, error)
	AddExpected(ctx context.Context, docID string, contacts []models.ContactInfo, addedBy string) error
	Remove(ctx context.Context, docID, email string) error
	GetStats(ctx context.Context, docID string) (*models.DocCompletionStats, error)
}

// ExpectedSignerService handles expected signer operations
type ExpectedSignerService struct {
	repo expectedSignerRepo
}

// NewExpectedSignerService creates a new expected signer service
func NewExpectedSignerService(repo expectedSignerRepo) *ExpectedSignerService {
	return &ExpectedSignerService{repo: repo}
}

// ListByDocID returns all expected signers for a document
func (s *ExpectedSignerService) ListByDocID(ctx context.Context, docID string) ([]*models.ExpectedSigner, error) {
	return s.repo.ListByDocID(ctx, docID)
}

// ListWithStatusByDocID returns all expected signers with their signature status
func (s *ExpectedSignerService) ListWithStatusByDocID(ctx context.Context, docID string) ([]*models.ExpectedSignerWithStatus, error) {
	return s.repo.ListWithStatusByDocID(ctx, docID)
}

// AddExpected adds expected signers to a document
func (s *ExpectedSignerService) AddExpected(ctx context.Context, docID string, contacts []models.ContactInfo, addedBy string) error {
	return s.repo.AddExpected(ctx, docID, contacts, addedBy)
}

// Remove removes an expected signer from a document
func (s *ExpectedSignerService) Remove(ctx context.Context, docID, email string) error {
	return s.repo.Remove(ctx, docID, email)
}

// GetStats returns completion statistics for a document
func (s *ExpectedSignerService) GetStats(ctx context.Context, docID string) (*models.DocCompletionStats, error) {
	return s.repo.GetStats(ctx, docID)
}
