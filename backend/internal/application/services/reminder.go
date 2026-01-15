// SPDX-License-Identifier: AGPL-3.0-or-later
package services

import (
	"context"
	"fmt"
	"time"

	"github.com/btouchard/ackify-ce/backend/internal/infrastructure/email"
	"github.com/btouchard/ackify-ce/backend/internal/infrastructure/i18n"
	"github.com/btouchard/ackify-ce/backend/pkg/logger"
	"github.com/btouchard/ackify-ce/backend/pkg/models"
)

// expectedSignerRepository defines minimal interface for expected signer operations
type expectedSignerRepository interface {
	ListWithStatusByDocID(ctx context.Context, docID string) ([]*models.ExpectedSignerWithStatus, error)
}

// reminderRepository defines minimal interface for reminder logging and history
type reminderRepository interface {
	LogReminder(ctx context.Context, log *models.ReminderLog) error
	GetReminderHistory(ctx context.Context, docID string) ([]*models.ReminderLog, error)
	GetReminderStats(ctx context.Context, docID string) (*models.ReminderStats, error)
}

// magicLinkService defines minimal interface for creating reminder auth tokens
type magicLinkService interface {
	CreateReminderAuthToken(ctx context.Context, email string, docID string) (string, error)
}

// ReminderService manages email notifications to pending signers with delivery tracking
type ReminderService struct {
	expectedSignerRepo expectedSignerRepository
	reminderRepo       reminderRepository
	emailSender        email.Sender
	magicLinkService   magicLinkService
	i18n               *i18n.I18n
	baseURL            string
}

// NewReminderService initializes reminder service with email sender and repository dependencies
func NewReminderService(
	expectedSignerRepo expectedSignerRepository,
	reminderRepo reminderRepository,
	emailSender email.Sender,
	magicLinkService magicLinkService,
	i18nService *i18n.I18n,
	baseURL string,
) *ReminderService {
	return &ReminderService{
		expectedSignerRepo: expectedSignerRepo,
		reminderRepo:       reminderRepo,
		emailSender:        emailSender,
		magicLinkService:   magicLinkService,
		i18n:               i18nService,
		baseURL:            baseURL,
	}
}

// SendReminders dispatches email notifications to all or selected pending signers with result aggregation
func (s *ReminderService) SendReminders(
	ctx context.Context,
	docID string,
	sentBy string,
	specificEmails []string,
	docURL string,
	locale string,
) (*models.ReminderSendResult, error) {

	logger.Logger.Info("Starting reminder sending process",
		"doc_id", docID,
		"sent_by", sentBy,
		"specific_emails_count", len(specificEmails),
		"locale", locale)

	allSigners, err := s.expectedSignerRepo.ListWithStatusByDocID(ctx, docID)
	if err != nil {
		logger.Logger.Error("Failed to get expected signers for reminders",
			"doc_id", docID,
			"error", err.Error())
		return nil, fmt.Errorf("failed to get expected signers: %w", err)
	}

	logger.Logger.Debug("Retrieved expected signers",
		"doc_id", docID,
		"total_signers", len(allSigners))

	var pendingSigners []*models.ExpectedSignerWithStatus
	for _, signer := range allSigners {
		if !signer.HasSigned {
			if len(specificEmails) > 0 {
				if containsEmail(specificEmails, signer.Email) {
					pendingSigners = append(pendingSigners, signer)
				}
			} else {
				pendingSigners = append(pendingSigners, signer)
			}
		}
	}

	logger.Logger.Info("Identified pending signers",
		"doc_id", docID,
		"pending_count", len(pendingSigners),
		"total_signers", len(allSigners))

	if len(pendingSigners) == 0 {
		logger.Logger.Info("No pending signers found, no reminders to send",
			"doc_id", docID)
		return &models.ReminderSendResult{
			TotalAttempted:   0,
			SuccessfullySent: 0,
			Failed:           0,
		}, nil
	}

	result := &models.ReminderSendResult{
		TotalAttempted: len(pendingSigners),
	}

	for _, signer := range pendingSigners {
		err := s.sendSingleReminder(ctx, docID, signer.Email, signer.Name, sentBy, docURL, locale)
		if err != nil {
			result.Failed++
			result.Errors = append(result.Errors, fmt.Sprintf("%s: %v", signer.Email, err))
		} else {
			result.SuccessfullySent++
		}
	}

	logger.Logger.Info("Reminder batch completed",
		"doc_id", docID,
		"total_attempted", result.TotalAttempted,
		"successfully_sent", result.SuccessfullySent,
		"failed", result.Failed)

	return result, nil
}

// sendSingleReminder sends a reminder to a single signer
func (s *ReminderService) sendSingleReminder(
	ctx context.Context,
	docID string,
	recipientEmail string,
	recipientName string,
	sentBy string,
	docURL string,
	locale string,
) error {

	logger.Logger.Debug("Sending reminder to signer",
		"doc_id", docID,
		"recipient_email", recipientEmail,
		"recipient_name", recipientName,
		"sent_by", sentBy)

	// Générer un token d'authentification pour ce lecteur
	token, err := s.magicLinkService.CreateReminderAuthToken(ctx, recipientEmail, docID)
	if err != nil {
		logger.Logger.Error("Failed to create reminder auth token",
			"doc_id", docID,
			"recipient_email", recipientEmail,
			"error", err.Error())
		return fmt.Errorf("failed to create auth token: %w", err)
	}

	// Construire l'URL d'authentification qui redirigera vers la page de signature
	authSignURL := fmt.Sprintf("%s/api/v1/auth/reminder-link/verify?token=%s", s.baseURL, token)

	logger.Logger.Debug("Generated auth sign URL for reminder",
		"doc_id", docID,
		"recipient_email", recipientEmail,
		"url", authSignURL)

	log := &models.ReminderLog{
		DocID:          docID,
		RecipientEmail: recipientEmail,
		SentAt:         time.Now(),
		SentBy:         sentBy,
		TemplateUsed:   "signature_reminder",
		Status:         "sent",
	}

	err = email.SendSignatureReminderEmail(ctx, s.emailSender, s.i18n, []string{recipientEmail}, locale, docID, docURL, authSignURL, recipientName)
	if err != nil {
		log.Status = "failed"
		errMsg := err.Error()
		log.ErrorMessage = &errMsg

		logger.Logger.Warn("Failed to send reminder email",
			"doc_id", docID,
			"recipient_email", recipientEmail,
			"error", err.Error())

		if logErr := s.reminderRepo.LogReminder(ctx, log); logErr != nil {
			logger.Logger.Error("Failed to log reminder error",
				"doc_id", docID,
				"recipient_email", recipientEmail,
				"log_error", logErr.Error(),
				"original_error", err.Error())
		}

		return fmt.Errorf("failed to send email: %w", err)
	}

	logger.Logger.Info("Reminder email sent successfully",
		"doc_id", docID,
		"recipient_email", recipientEmail)

	if err := s.reminderRepo.LogReminder(ctx, log); err != nil {
		logger.Logger.Error("Failed to log successful reminder",
			"doc_id", docID,
			"recipient_email", recipientEmail,
			"error", err.Error())
		return fmt.Errorf("email sent but failed to log: %w", err)
	}

	return nil
}

// GetReminderStats retrieves aggregated reminder metrics for monitoring dashboard
func (s *ReminderService) GetReminderStats(ctx context.Context, docID string) (*models.ReminderStats, error) {
	return s.reminderRepo.GetReminderStats(ctx, docID)
}

// GetReminderHistory retrieves complete email send log with success/failure tracking
func (s *ReminderService) GetReminderHistory(ctx context.Context, docID string) ([]*models.ReminderLog, error) {
	return s.reminderRepo.GetReminderHistory(ctx, docID)
}

func containsEmail(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
