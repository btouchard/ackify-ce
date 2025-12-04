// SPDX-License-Identifier: AGPL-3.0-or-later
package coreapp

import (
	"context"

	"github.com/btouchard/ackify-ce/backend/internal/application/services"
	"github.com/btouchard/ackify-ce/backend/internal/domain/models"
)

type DocumentService interface {
	// Creation
	CreateDocument(ctx context.Context, req services.CreateDocumentRequest) (*models.Document, error)
	FindOrCreateDocument(ctx context.Context, ref string) (*models.Document, bool, error)
	FindByReference(ctx context.Context, ref string, refType string) (*models.Document, error)
	// Read
	List(ctx context.Context, limit, offset int) ([]*models.Document, error)
	Search(ctx context.Context, query string, limit, offset int) ([]*models.Document, error)
	Count(ctx context.Context, searchQuery string) (int, error)
	GetByDocID(ctx context.Context, docID string) (*models.Document, error)
	// Write
	CreateOrUpdate(ctx context.Context, docID string, input models.DocumentInput, createdBy string) (*models.Document, error)
	Delete(ctx context.Context, docID string) error
}

type SignatureService interface {
	CreateSignature(ctx context.Context, request *models.SignatureRequest) error
	GetSignatureStatus(ctx context.Context, docID string, user *models.User) (*models.SignatureStatus, error)
	GetSignatureByDocAndUser(ctx context.Context, docID string, user *models.User) (*models.Signature, error)
	GetDocumentSignatures(ctx context.Context, docID string) ([]*models.Signature, error)
	GetUserSignatures(ctx context.Context, user *models.User) ([]*models.Signature, error)
}

type ExpectedSignerService interface {
	ListByDocID(ctx context.Context, docID string) ([]*models.ExpectedSigner, error)
	ListWithStatusByDocID(ctx context.Context, docID string) ([]*models.ExpectedSignerWithStatus, error)
	AddExpected(ctx context.Context, docID string, contacts []models.ContactInfo, addedBy string) error
	Remove(ctx context.Context, docID, email string) error
	GetStats(ctx context.Context, docID string) (*models.DocCompletionStats, error)
}

type ReminderService interface {
	SendReminders(ctx context.Context, docID, sentBy string, specificEmails []string, docURL string, locale string) (*models.ReminderSendResult, error)
	GetReminderHistory(ctx context.Context, docID string) ([]*models.ReminderLog, error)
	GetReminderStats(ctx context.Context, docID string) (*models.ReminderStats, error)
}

type WebhookPublisher interface {
	Publish(ctx context.Context, eventType string, payload map[string]interface{}) error
}

type DocumentAuthorizer interface {
	CanCreateDocument(ctx context.Context, user *models.User) bool
}

type CoreDeps struct {
	Documents          DocumentService
	DocumentAuthorizer DocumentAuthorizer
	Signatures         SignatureService
	ExpectedSigners    ExpectedSignerService
	Reminders          ReminderService
	WebhookPublisher   WebhookPublisher
	BaseURL            string
	ImportMaxSigners   int
}
